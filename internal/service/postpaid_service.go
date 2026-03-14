package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/external/gerbang"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/google/uuid"
)

// PostpaidService handles postpaid business logic
type PostpaidService struct {
	postpaidRepo  repository.PostpaidRepository
	balanceRepo   repository.BalanceRepository
	voucherRepo   repository.VoucherRepository
	userRepo      repository.UserRepository
	productRepo   repository.ProductRepository
	gerbangClient *gerbang.Client
	allowDummy    bool
}

// NewPostpaidService creates a new postpaid service
func NewPostpaidService(
	postpaidRepo repository.PostpaidRepository,
	balanceRepo repository.BalanceRepository,
	voucherRepo repository.VoucherRepository,
	userRepo repository.UserRepository,
	productRepo repository.ProductRepository,
	gerbangClient *gerbang.Client,
	allowDummy bool,
) *PostpaidService {
	return &PostpaidService{
		postpaidRepo:  postpaidRepo,
		balanceRepo:   balanceRepo,
		voucherRepo:   voucherRepo,
		userRepo:      userRepo,
		productRepo:   productRepo,
		gerbangClient: gerbangClient,
		allowDummy:    allowDummy,
	}
}

// Inquiry performs bill inquiry
func (s *PostpaidService) Inquiry(ctx context.Context, userID, serviceType, target string, providerID, period *string) (*domain.PostpaidInquiryResponse, error) {
	// Validate service type
	if !isValidPostpaidService(serviceType) {
		return nil, domain.ErrInvalidServiceType
	}

	// Validate target
	if err := validatePostpaidTarget(serviceType, target); err != nil {
		return nil, err
	}

	// Call Gerbang API for inquiry
	inquiryID := fmt.Sprintf("inq_post_%s", uuid.New().String()[:12])
	product, err := s.findPostpaidProduct(ctx, serviceType)
	if err != nil {
		return nil, err
	}

	gerbangResp, err := s.gerbangClient.CreateInquiry(ctx, inquiryID, product.SKUCode, target)

	var inquiry *domain.PostpaidInquiry
	var customerName string
	var billAmount int64
	var adminFee int64
	var hasBill bool
	var externalID string
	var inquiryPeriod string

	if err != nil {
		// Check if it's a "no bill" error
		if gerbangErr, ok := err.(*gerbang.Error); ok && gerbangErr.Code == 404 {
			hasBill = false
			customerName = ""
			inquiryPeriod = defaultPostpaidPeriod(period)
		} else if s.allowDummy {
			slog.Warn("falling back to dummy postpaid inquiry",
				slog.String("service_type", serviceType),
				slog.String("target", target),
				slog.String("error", err.Error()),
			)

			hasBill = true
			customerName = s.dummyPostpaidCustomerName(serviceType, target)
			billAmount = s.dummyPostpaidAmount(serviceType, target, product)
			adminFee = product.Admin
			externalID = buildDummyPaymentID("dummy_inquiry")
			inquiryPeriod = defaultPostpaidPeriod(period)
		} else {
			return nil, fmt.Errorf("provider inquiry failed: %w", err)
		}
	} else {
		// Use real Gerbang response
		hasBill = true
		externalID = gerbangResp.TransactionID
		customerName = gerbangResp.CustomerName
		billAmount = gerbangResp.Amount
		if billAmount == 0 {
			billAmount = gerbangResp.Price
		}
		adminFee = gerbangResp.Admin
		inquiryPeriod = gerbangResp.Period
	}

	// Create inquiry record
	inquiry = &domain.PostpaidInquiry{
		ID:           inquiryID,
		UserID:       userID,
		ServiceType:  serviceType,
		Target:       target,
		ProviderID:   providerID,
		CustomerID:   target,
		CustomerName: customerName,
		Period:       inquiryPeriod,
		BillAmount:   billAmount,
		AdminFee:     adminFee,
		Penalty:      0,
		TotalPayment: billAmount + adminFee,
		HasBill:      hasBill,
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		CreatedAt:    time.Now(),
	}

	if gerbangResp != nil && gerbangResp.TotalAmount > 0 {
		inquiry.TotalPayment = gerbangResp.TotalAmount
	}
	if inquiry.Period == "" {
		inquiry.Period = defaultPostpaidPeriod(period)
	}

	if externalID != "" {
		inquiry.ExternalID = &externalID
	}

	// Save inquiry
	if err := s.postpaidRepo.CreateInquiry(ctx, inquiry); err != nil {
		return nil, fmt.Errorf("failed to create inquiry: %w", err)
	}

	// Get user balance
	balance, err := s.balanceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	if balance == nil {
		return nil, domain.ErrValidationFailed("Balance not found")
	}

	// Build response
	response := &domain.PostpaidInquiryResponse{
		PinRequired: true,
		Notices:     []string{},
	}

	// Inquiry info
	if hasBill {
		expiresAtStr := inquiry.ExpiresAt.Format(time.RFC3339)
		response.Inquiry = &domain.PostpaidInquiryInfo{
			InquiryID:   &inquiry.ID,
			ServiceType: inquiry.ServiceType,
			Target:      inquiry.Target,
			TargetValid: true,
			ExpiresAt:   &expiresAtStr,
		}
	} else {
		response.Inquiry = &domain.PostpaidInquiryInfo{
			InquiryID:   nil,
			ServiceType: inquiry.ServiceType,
			Target:      inquiry.Target,
			TargetValid: true,
			NoBill:      true,
		}
		noBillMsg := "Tidak ada tagihan untuk periode ini"
		response.Message = &noBillMsg
	}

	// Customer info
	response.Customer = &domain.PostpaidCustomerInfo{
		CustomerID: inquiry.CustomerID,
		Name:       inquiry.CustomerName,
	}

	// PLN specific fields
	if serviceType == domain.ServicePLNPostpaid {
		segmentPower := "R1/1300VA"
		address := "JL. MERDEKA NO 123, JAKARTA"
		response.Customer.SegmentPower = &segmentPower
		response.Customer.Address = &address
	}

	// Bill info (only if has bill)
	if hasBill {
		period := inquiry.Period
		periodCode := "202501"
		dueDate := "2025-02-20"
		response.Bill = &domain.BillInfo{
			Period:                &period,
			PeriodCode:            &periodCode,
			Amount:                inquiry.BillAmount,
			AmountFormatted:       formatCurrency(inquiry.BillAmount),
			AdminFee:              inquiry.AdminFee,
			AdminFeeFormatted:     formatCurrency(inquiry.AdminFee),
			Penalty:               inquiry.Penalty,
			PenaltyFormatted:      formatCurrency(inquiry.Penalty),
			TotalPayment:          inquiry.TotalPayment,
			TotalPaymentFormatted: formatCurrency(inquiry.TotalPayment),
			DueDate:               &dueDate,
		}

		// PLN meter info
		if serviceType == domain.ServicePLNPostpaid {
			response.Bill.StandMeter = &domain.MeterInfo{
				Previous: 12500,
				Current:  12750,
				Usage:    250,
			}
		}

		// Payment info
		balanceSufficient := balance.Amount >= inquiry.TotalPayment
		response.Payment = &domain.PostpaidPaymentInfo{
			Method:                    "balance",
			BalanceAvailable:          balance.Amount,
			BalanceAvailableFormatted: formatCurrency(balance.Amount),
			BalanceSufficient:         balanceSufficient,
		}
	}

	return response, nil
}

// Pay processes bill payment
func (s *PostpaidService) Pay(ctx context.Context, userID, inquiryID string, voucherCodes []string, pin *string) (*domain.PostpaidPayResponse, error) {
	// Get inquiry with ownership check
	inquiry, err := s.postpaidRepo.FindInquiryByUserAndID(ctx, userID, inquiryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inquiry: %w", err)
	}
	if inquiry == nil {
		return nil, domain.ErrInquiryNotFound
	}

	// Check if inquiry expired
	if time.Now().After(inquiry.ExpiresAt) {
		return nil, domain.ErrInquiryExpired
	}

	// Check if has bill
	if !inquiry.HasBill {
		return nil, domain.ErrNoBill
	}

	// Check for duplicate transaction (idempotency)
	existingTx, err := s.postpaidRepo.FindTransactionByInquiryID(ctx, inquiryID)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate: %w", err)
	}
	if existingTx != nil {
		// Return existing transaction
		return s.buildPayResponse(existingTx), nil
	}

	// TODO: Verify PIN if required
	// if pin != nil {
	//     if err := s.verifyPIN(ctx, userID, *pin); err != nil {
	//         return nil, err
	//     }
	// }

	// Calculate final payment amount with vouchers
	totalPayment := inquiry.TotalPayment
	voucherDiscount := int64(0)

	// TODO: Apply vouchers
	// if len(voucherCodes) > 0 {
	//     discount, err := s.applyVouchers(ctx, userID, voucherCodes, totalPayment)
	//     if err != nil {
	//         return nil, err
	//     }
	//     voucherDiscount = discount
	//     totalPayment -= discount
	// }

	// Begin database transaction for atomic operation
	tx, err := s.postpaidRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock user balance (SELECT FOR UPDATE)
	balance, err := s.balanceRepo.FindByUserIDForUpdate(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to lock balance: %w", err)
	}

	// Check sufficient balance
	if balance.Amount < totalPayment {
		return nil, domain.ErrInsufficientBalance
	}

	// Deduct balance
	balanceBefore := balance.Amount
	balance.Amount -= totalPayment
	balanceAfter := balance.Amount
	balance.UpdatedAt = time.Now()

	// Update balance in transaction
	if err := s.balanceRepo.UpdateWithTx(ctx, tx, balance); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	// Call Gerbang API to pay bill
	transactionID := fmt.Sprintf("trx_post_%s", uuid.New().String()[:12])
	product, err := s.findPostpaidProduct(ctx, inquiry.ServiceType)
	if err != nil {
		return nil, err
	}

	// Get inquiry external ID for payment (from Gerbang inquiry response)
	var inquiryExternalID string
	if inquiry.ExternalID != nil {
		inquiryExternalID = *inquiry.ExternalID
	} else if !s.allowDummy {
		return nil, domain.ErrServiceUnavailable
	}

	gerbangResp, err := s.gerbangClient.CreatePostpaidPayment(ctx, transactionID, inquiryExternalID, product.SKUCode, inquiry.Target)

	var externalID string
	var serialNumber *string
	var status string

	if err != nil {
		if !s.allowDummy {
			return nil, fmt.Errorf("provider call failed: %w", err)
		}

		slog.Warn("falling back to dummy postpaid payment",
			slog.String("transaction_id", transactionID),
			slog.String("inquiry_id", inquiryID),
			slog.String("service_type", inquiry.ServiceType),
			slog.String("error", err.Error()),
		)

		gerbangResp = &gerbang.TransactionResponse{
			TransactionID: buildDummyReference("POST"),
			ReferenceID:   transactionID,
			SKUCode:       product.SKUCode,
			CustomerNo:    inquiry.Target,
			CustomerName:  inquiry.CustomerName,
			Type:          "payment",
			Status:        gerbang.StatusSuccess,
			CreatedAt:     time.Now().Format(time.RFC3339),
		}
	} else {
		// Use real Gerbang response
		externalID = gerbangResp.TransactionID
		if gerbangResp.SerialNumber != nil {
			serialNumber = gerbangResp.SerialNumber
		}
		switch gerbangResp.Status {
		case gerbang.StatusSuccess:
			status = domain.PostpaidStatusSuccess
		case gerbang.StatusPending, gerbang.StatusProcessing:
			status = domain.PostpaidStatusProcessing
		default:
			if !s.allowDummy {
				return nil, fmt.Errorf("provider returned unsupported postpaid status: %s", gerbangResp.Status)
			}
			status = domain.PostpaidStatusSuccess
		}
	}

	if status == "" {
		status = domain.PostpaidStatusSuccess
	}
	if externalID == "" {
		externalID = gerbangResp.TransactionID
	}
	referenceNumber := gerbangResp.TransactionID

	// Create transaction record
	var completedAt *time.Time
	now := time.Now()
	if status == domain.PostpaidStatusSuccess {
		completedAt = &now
	}
	transaction := &domain.PostpaidTransaction{
		ID:              transactionID,
		UserID:          userID,
		InquiryID:       inquiryID,
		ServiceType:     inquiry.ServiceType,
		Target:          inquiry.Target,
		ProviderID:      inquiry.ProviderID,
		CustomerID:      inquiry.CustomerID,
		CustomerName:    inquiry.CustomerName,
		Period:          inquiry.Period,
		BillAmount:      inquiry.BillAmount,
		AdminFee:        inquiry.AdminFee,
		Penalty:         inquiry.Penalty,
		VoucherDiscount: voucherDiscount,
		TotalPayment:    totalPayment,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
		ReferenceNumber: referenceNumber,
		SerialNumber:    serialNumber,
		Status:          status,
		CompletedAt:     completedAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if externalID != "" {
		transaction.ExternalID = &externalID
	}

	// Save transaction
	if err := s.postpaidRepo.CreateTransaction(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Build response
	return s.buildPayResponse(transaction), nil
}

// buildPayResponse builds payment response
func (s *PostpaidService) buildPayResponse(tx *domain.PostpaidTransaction) *domain.PostpaidPayResponse {
	completedAt := ""
	if tx.CompletedAt != nil {
		completedAt = tx.CompletedAt.Format(time.RFC3339)
	}

	return &domain.PostpaidPayResponse{
		Transaction: &domain.PostpaidTransactionInfo{
			TransactionID: tx.ID,
			InquiryID:     tx.InquiryID,
			Status:        tx.Status,
			ServiceType:   tx.ServiceType,
			CompletedAt:   completedAt,
		},
		Customer: &domain.PostpaidCustomerInfo{
			CustomerID: tx.CustomerID,
			Name:       tx.CustomerName,
		},
		Bill: &domain.BillInfoSimple{
			Period:          tx.Period,
			Amount:          tx.BillAmount,
			AmountFormatted: formatCurrency(tx.BillAmount),
		},
		Payment: &domain.PaymentDetail{
			TotalPayment:          tx.TotalPayment,
			TotalPaymentFormatted: formatCurrency(tx.TotalPayment),
			VoucherDiscount:       tx.VoucherDiscount,
			BalanceBefore:         tx.BalanceBefore,
			BalanceAfter:          tx.BalanceAfter,
			BalanceAfterFormatted: formatCurrency(tx.BalanceAfter),
		},
		Receipt: &domain.PostpaidReceiptInfo{
			ReferenceNumber: tx.ReferenceNumber,
			SerialNumber:    tx.SerialNumber,
		},
		Message: &domain.PostpaidMessageInfo{
			Title:    getPostpaidStatusTitle(tx.ServiceType, tx.Status),
			Subtitle: getPostpaidStatusSubtitle(tx.ServiceType, tx.Status, tx.Period),
		},
	}
}

// Helper functions

func isValidPostpaidService(serviceType string) bool {
	validTypes := []string{
		domain.ServicePLNPostpaid,
		domain.ServicePhonePostpaid,
		domain.ServicePDAM,
		domain.ServiceBPJS,
		domain.ServiceTelkom,
		domain.ServicePGN,
		domain.ServicePBB,
		domain.ServiceTVCable,
	}
	for _, t := range validTypes {
		if serviceType == t {
			return true
		}
	}
	return false
}

func validatePostpaidTarget(serviceType, target string) error {
	if target == "" {
		return domain.ErrInvalidTarget
	}

	switch serviceType {
	case domain.ServicePLNPostpaid:
		// PLN meter: 11-12 digits
		if len(target) < 11 || len(target) > 12 {
			return domain.ErrInvalidTarget
		}
	case domain.ServicePhonePostpaid:
		// Phone: starts with 08, 10-13 digits
		if len(target) < 10 || len(target) > 13 || target[:2] != "08" {
			return domain.ErrInvalidTarget
		}
	case domain.ServiceBPJS:
		// BPJS: 13 digits
		if len(target) != 13 {
			return domain.ErrInvalidTarget
		}
	case domain.ServicePDAM, domain.ServiceTelkom, domain.ServicePGN, domain.ServicePBB, domain.ServiceTVCable:
		// Other services: minimum 6 digits
		if len(target) < 6 {
			return domain.ErrInvalidTarget
		}
	}

	return nil
}

func getServiceDisplayName(serviceType string) string {
	switch serviceType {
	case domain.ServicePLNPostpaid:
		return "Tagihan Listrik"
	case domain.ServicePhonePostpaid:
		return "Pulsa Pascabayar"
	case domain.ServicePDAM:
		return "PDAM"
	case domain.ServiceBPJS:
		return "BPJS Kesehatan"
	case domain.ServiceTelkom:
		return "Telkom/IndiHome"
	case domain.ServicePGN:
		return "Tagihan Gas PGN"
	case domain.ServicePBB:
		return "PBB"
	case domain.ServiceTVCable:
		return "TV Kabel"
	default:
		return "Tagihan"
	}
}

func (s *PostpaidService) dummyPostpaidCustomerName(serviceType, target string) string {
	label := strings.ToUpper(getServiceDisplayName(serviceType))
	suffix := target
	if len(suffix) > 4 {
		suffix = suffix[len(suffix)-4:]
	}
	return fmt.Sprintf("%s %s", label, suffix)
}

func (s *PostpaidService) dummyPostpaidAmount(serviceType, target string, product *domain.Product) int64 {
	base := int64(45000)
	if product != nil && product.Price > 0 {
		base = product.Price
	}

	var sum int64
	for _, char := range target {
		if char >= '0' && char <= '9' {
			sum += int64(char - '0')
		}
	}

	switch serviceType {
	case domain.ServicePLNPostpaid:
		base += 25000
	case domain.ServicePDAM:
		base += 15000
	case domain.ServiceBPJS:
		base += 10000
	}

	return base + ((sum % 9) * 5000)
}

func (s *PostpaidService) findPostpaidProduct(ctx context.Context, serviceType string) (*domain.Product, error) {
	isActive := true
	filter := repository.ProductFilter{
		Type:     domain.ProductTypePostpaid,
		IsActive: &isActive,
		Page:     1,
		PerPage:  50,
	}

	switch serviceType {
	case domain.ServicePLNPostpaid:
		filter.Category = domain.CategoryPLN
	case domain.ServicePDAM:
		filter.Category = domain.CategoryPDAM
	case domain.ServiceBPJS:
		filter.Category = domain.CategoryBPJS
	case domain.ServiceTelkom:
		filter.Category = domain.CategoryTelkom
	case domain.ServiceTVCable:
		filter.Category = domain.CategoryTV
	case domain.ServicePhonePostpaid:
		filter.Search = "pascabayar"
	case domain.ServicePGN:
		filter.Search = "PGN"
	case domain.ServicePBB:
		filter.Search = "PBB"
	}

	products, err := s.productRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to load postpaid products: %w", err)
	}
	for _, product := range products {
		if postpaidProductMatchesService(product, serviceType) {
			return product, nil
		}
	}

	return nil, domain.ErrInvalidProduct
}

func postpaidProductMatchesService(product *domain.Product, serviceType string) bool {
	if product == nil || product.Type != domain.ProductTypePostpaid {
		return false
	}

	switch serviceType {
	case domain.ServicePLNPostpaid:
		return strings.EqualFold(product.Category, domain.CategoryPLN)
	case domain.ServicePDAM:
		return strings.EqualFold(product.Category, domain.CategoryPDAM)
	case domain.ServiceBPJS:
		return strings.EqualFold(product.Category, domain.CategoryBPJS)
	case domain.ServiceTelkom:
		return strings.EqualFold(product.Category, domain.CategoryTelkom)
	case domain.ServiceTVCable:
		return strings.EqualFold(product.Category, domain.CategoryTV)
	case domain.ServicePhonePostpaid:
		return strings.Contains(strings.ToLower(product.Name), "pascabayar")
	case domain.ServicePGN:
		return strings.Contains(strings.ToLower(product.Name), "pgn")
	case domain.ServicePBB:
		return strings.Contains(strings.ToLower(product.Name), "pbb")
	default:
		return false
	}
}

func defaultPostpaidPeriod(period *string) string {
	if period != nil && *period != "" {
		return *period
	}
	return time.Now().Format("January 2006")
}

func getPostpaidStatusTitle(serviceType, status string) string {
	if status == domain.PostpaidStatusSuccess {
		return fmt.Sprintf("Pembayaran %s Berhasil", getServiceDisplayName(serviceType))
	}
	return "Pembayaran Sedang Diproses"
}

func getPostpaidStatusSubtitle(serviceType, status, period string) string {
	if period == "" {
		period = "berjalan"
	}
	if status == domain.PostpaidStatusSuccess {
		return fmt.Sprintf("Tagihan %s periode %s telah dibayar", getServiceDisplayName(serviceType), period)
	}
	return fmt.Sprintf("Tagihan %s periode %s sedang diproses", getServiceDisplayName(serviceType), period)
}

// Refund restores balance for failed/cancelled transaction
func (s *PostpaidService) Refund(ctx context.Context, transactionID, reason string) error {
	// Begin transaction
	tx, err := s.postpaidRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get transaction
	transaction, err := s.postpaidRepo.FindTransactionByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to find transaction: %w", err)
	}
	if transaction == nil {
		return domain.ErrNotFound("Transaksi")
	}

	// Validate status - can only refund processing/failed transactions
	if transaction.Status != domain.TransactionProcessing &&
		transaction.Status != domain.TransactionFailed {
		return domain.ErrValidationFailed("Transaksi tidak dapat dikembalikan")
	}

	// Lock and restore balance
	balance, err := s.balanceRepo.FindByUserIDForUpdate(ctx, tx, transaction.UserID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}
	if balance == nil {
		return domain.ErrNotFound("Saldo")
	}

	// Restore balance
	balance.Amount += transaction.TotalPayment
	if err := s.balanceRepo.UpdateWithTx(ctx, tx, balance); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Update transaction status to refunded
	transaction.Status = domain.TransactionRefunded
	transaction.UpdatedAt = time.Now()
	if err := s.postpaidRepo.UpdateTransactionStatus(ctx, transactionID, domain.TransactionRefunded); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// TODO: Create balance history record when balance_history table is ready

	return tx.Commit()
}

// HandleWebhook handles postpaid transaction webhook notification from Gerbang API.
func (s *PostpaidService) HandleWebhook(ctx context.Context, webhookData *gerbang.TransactionWebhookData) error {
	transaction, err := s.postpaidRepo.FindTransactionByID(ctx, webhookData.ReferenceID)
	if err != nil {
		return fmt.Errorf("failed to find transaction: %w", err)
	}
	if transaction == nil {
		return domain.ErrNotFound("Transaksi postpaid")
	}
	if transaction.Status == domain.PostpaidStatusSuccess || transaction.Status == domain.PostpaidStatusFailed {
		return nil
	}

	tx, err := s.postpaidRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	newStatus := domain.PostpaidStatusProcessing
	refundNeeded := false

	switch webhookData.Status {
	case "Success":
		newStatus = domain.PostpaidStatusSuccess
	case "Failed":
		newStatus = domain.PostpaidStatusFailed
		refundNeeded = true
	case "Pending":
		newStatus = domain.PostpaidStatusProcessing
	default:
		return fmt.Errorf("unknown transaction status: %s", webhookData.Status)
	}

	if err := s.postpaidRepo.UpdateTransactionStatusWithTx(ctx, tx, transaction.ID, newStatus); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	if refundNeeded {
		balance, err := s.balanceRepo.FindByUserIDForUpdate(ctx, tx, transaction.UserID)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}
		if balance == nil {
			return domain.ErrNotFound("Saldo")
		}

		balance.Amount += transaction.TotalPayment
		balance.UpdatedAt = now
		if err := s.balanceRepo.UpdateWithTx(ctx, tx, balance); err != nil {
			return fmt.Errorf("failed to restore balance: %w", err)
		}
	}

	return tx.Commit()
}
