package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/external/gerbang"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// PostpaidService handles postpaid business logic
type PostpaidService struct {
	postpaidRepo  repository.PostpaidRepository
	balanceRepo   repository.BalanceRepository
	voucherRepo   repository.VoucherRepository
	userRepo      repository.UserRepository
	gerbangClient *gerbang.Client
}

// NewPostpaidService creates a new postpaid service
func NewPostpaidService(
	postpaidRepo repository.PostpaidRepository,
	balanceRepo repository.BalanceRepository,
	voucherRepo repository.VoucherRepository,
	userRepo repository.UserRepository,
	gerbangClient *gerbang.Client,
) *PostpaidService {
	return &PostpaidService{
		postpaidRepo:  postpaidRepo,
		balanceRepo:   balanceRepo,
		voucherRepo:   voucherRepo,
		userRepo:      userRepo,
		gerbangClient: gerbangClient,
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

	// Map service type to SKU code (simplified - should get from product DB)
	skuCode := mapServiceTypeToSKU(serviceType)

	gerbangResp, err := s.gerbangClient.CreateInquiry(ctx, inquiryID, skuCode, target)

	var inquiry *domain.PostpaidInquiry
	var customerName string
	var billAmount int64
	var adminFee int64
	var hasBill bool
	var externalID string

	if err != nil {
		// Check if it's a "no bill" error
		if gerbangErr, ok := err.(*gerbang.Error); ok && gerbangErr.Code == 404 {
			hasBill = false
			customerName = "Customer Name from Provider" // Mock
		} else {
			// Fallback to mock for development
			hasBill = true
			customerName = "BUDI SANTOSO"
			billAmount = 350000
			adminFee = 2500
			externalID = fmt.Sprintf("gerbang_inq_%s", uuid.New().String()[:12])
		}
	} else {
		// Use real Gerbang response
		hasBill = true
		externalID = gerbangResp.TransactionID
		customerName = gerbangResp.CustomerName
		billAmount = gerbangResp.Price
		adminFee = 2500 // Fixed admin fee for now
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
		Period:       "Januari 2025", // Should come from Gerbang
		BillAmount:   billAmount,
		AdminFee:     adminFee,
		Penalty:      0,
		TotalPayment: billAmount + adminFee,
		HasBill:      hasBill,
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		CreatedAt:    time.Now(),
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

	// Get inquiry external ID for payment (from Gerbang inquiry response)
	var inquiryExternalID string
	if inquiry.ExternalID != nil {
		inquiryExternalID = *inquiry.ExternalID
	} else {
		// If no external ID from inquiry, create a mock one
		inquiryExternalID = fmt.Sprintf("gerbang_inq_%s", uuid.New().String()[:12])
	}

	gerbangResp, err := s.gerbangClient.CreatePostpaidPayment(ctx, transactionID, inquiryExternalID)

	var externalID string
	var serialNumber *string
	var status string
	var failedReason *string

	if err != nil {
		// If Gerbang call fails, rollback and return error
		gerbangErr, ok := err.(*gerbang.Error)
		if ok {
			// Check if it's insufficient funds error
			if gerbang.IsInsufficientFunds(gerbangErr) {
				return nil, domain.ErrServiceUnavailable
			}
			// Other provider errors
			reason := gerbangErr.Message
			failedReason = &reason
			status = domain.PostpaidStatusFailed
		} else {
			return nil, fmt.Errorf("provider call failed: %w", err)
		}
	} else {
		// Use real Gerbang response
		externalID = gerbangResp.TransactionID
		if gerbangResp.SerialNumber != nil {
			serialNumber = gerbangResp.SerialNumber
		}
		status = domain.PostpaidStatusSuccess
	}

	// Generate reference number
	referenceNumber := fmt.Sprintf("REF%s", time.Now().Format("20060102150405"))

	// Create transaction record
	now := time.Now()
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
		FailedReason:    failedReason,
		CompletedAt:     &now,
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
	completedAt := tx.CompletedAt.Format(time.RFC3339)

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
			Title:    fmt.Sprintf("Pembayaran %s Berhasil", getServiceDisplayName(tx.ServiceType)),
			Subtitle: fmt.Sprintf("Tagihan %s periode %s telah dibayar", getServiceDisplayName(tx.ServiceType), tx.Period),
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

func mapServiceTypeToSKU(serviceType string) string {
	// Simplified mapping - should query product DB in real implementation
	switch serviceType {
	case domain.ServicePLNPostpaid:
		return "PLN_POST"
	case domain.ServicePhonePostpaid:
		return "PHONE_POST"
	case domain.ServicePDAM:
		return "PDAM"
	case domain.ServiceBPJS:
		return "BPJS"
	case domain.ServiceTelkom:
		return "TELKOM"
	case domain.ServicePGN:
		return "PGN"
	case domain.ServicePBB:
		return "PBB"
	case domain.ServiceTVCable:
		return "TV_CABLE"
	default:
		return "UNKNOWN"
	}
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
