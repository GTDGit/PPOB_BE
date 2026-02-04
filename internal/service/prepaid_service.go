package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/external/gerbang"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/GTDGit/PPOB_BE/pkg/hash"
)

// PrepaidService handles prepaid transaction business logic
type PrepaidService struct {
	prepaidRepo   repository.PrepaidRepository
	balanceRepo   repository.BalanceRepository
	userRepo      repository.UserRepository
	gerbangClient *gerbang.Client
	// TODO: Add when implemented
	// productRepo  repository.ProductRepository
	// voucherRepo  repository.VoucherRepository
}

// NewPrepaidService creates a new prepaid service
func NewPrepaidService(
	prepaidRepo repository.PrepaidRepository,
	balanceRepo repository.BalanceRepository,
	userRepo repository.UserRepository,
	gerbangClient *gerbang.Client,
) *PrepaidService {
	return &PrepaidService{
		prepaidRepo:   prepaidRepo,
		balanceRepo:   balanceRepo,
		userRepo:      userRepo,
		gerbangClient: gerbangClient,
	}
}

// InquiryRequest represents inquiry request
type InquiryRequest struct {
	UserID      string
	ServiceType string
	Target      string
	ProviderID  *string
}

// Inquiry handles prepaid inquiry
func (s *PrepaidService) Inquiry(ctx context.Context, req InquiryRequest) (*domain.PrepaidInquiryResponse, error) {
	// Validate service type
	if !isValidServiceType(req.ServiceType) {
		return nil, domain.ErrValidationFailed("Invalid service type")
	}

	// Validate target format based on service type
	if err := validateTarget(req.ServiceType, req.Target); err != nil {
		return nil, err
	}

	// TODO: Detect operator for pulsa/data
	var operatorID *string
	if req.ServiceType == domain.ServicePulsa || req.ServiceType == domain.ServiceData {
		operator := detectOperator(req.Target)
		operatorID = &operator
	}

	// TODO: Call provider to validate target
	targetValid := true // Mock for now

	// Create inquiry record
	inquiryID := "inq_" + uuid.New().String()[:8]
	inquiry := &domain.PrepaidInquiry{
		ID:          inquiryID,
		UserID:      req.UserID,
		ServiceType: req.ServiceType,
		Target:      req.Target,
		TargetValid: targetValid,
		OperatorID:  operatorID,
		ExpiresAt:   time.Now().Add(30 * time.Minute),
		CreatedAt:   time.Now(),
	}

	if err := s.prepaidRepo.CreateInquiry(ctx, inquiry); err != nil {
		return nil, fmt.Errorf("failed to create inquiry: %w", err)
	}

	// Build response
	response := &domain.PrepaidInquiryResponse{
		Inquiry: &domain.InquiryInfo{
			InquiryID:   &inquiryID,
			ServiceType: req.ServiceType,
			Target:      req.Target,
			TargetValid: targetValid,
		},
		Products: getMockProducts(req.ServiceType), // TODO: Replace with real product fetch
		Notices:  []*domain.NoticeInfo{},
	}

	// Add operator info if available
	if operatorID != nil {
		response.Inquiry.Operator = &domain.OperatorInfo{
			ID:   *operatorID,
			Name: getOperatorName(*operatorID),
			Icon: *operatorID,
		}
	}

	expiresAt := inquiry.ExpiresAt.Format(time.RFC3339)
	response.Inquiry.ExpiresAt = &expiresAt

	return response, nil
}

// CreateOrderRequest represents create order request
type CreateOrderRequest struct {
	UserID       string
	InquiryID    string
	ProductID    string
	VoucherCodes []string
}

// CreateOrder handles order creation
func (s *PrepaidService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*domain.PrepaidOrderResponse, error) {
	// Get inquiry
	inquiry, err := s.prepaidRepo.FindInquiryByID(ctx, req.InquiryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inquiry: %w", err)
	}
	if inquiry == nil {
		return nil, domain.ErrValidationFailed("Inquiry not found")
	}

	// Check inquiry expiration
	if time.Now().After(inquiry.ExpiresAt) {
		return nil, domain.ErrInquiryExpired
	}

	// Check if inquiry belongs to user
	if inquiry.UserID != req.UserID {
		return nil, domain.ErrValidationFailed("Inquiry does not belong to user")
	}

	// TODO: Get product from repository
	product := getMockProduct(req.ProductID)
	if product == nil {
		return nil, domain.ErrInvalidProduct
	}

	// Calculate pricing
	productPrice := product.Price
	adminFee := product.AdminFee
	subtotal := productPrice + adminFee
	totalDiscount := int64(0) // TODO: Apply vouchers
	totalPayment := subtotal - totalDiscount

	// Get user balance
	balance, err := s.balanceRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	if balance == nil {
		return nil, domain.ErrValidationFailed("Balance not found")
	}

	balanceSufficient := balance.Amount >= totalPayment

	// Get user settings for PIN requirement
	// TODO: Fetch from user settings repository
	pinRequired := false

	// Create order
	orderID := "ord_" + uuid.New().String()[:8]
	order := &domain.PrepaidOrder{
		ID:            orderID,
		UserID:        req.UserID,
		InquiryID:     req.InquiryID,
		ProductID:     req.ProductID,
		Status:        domain.OrderPendingPayment,
		ServiceType:   inquiry.ServiceType,
		Target:        inquiry.Target,
		ProductPrice:  productPrice,
		AdminFee:      adminFee,
		Subtotal:      subtotal,
		TotalDiscount: totalDiscount,
		TotalPayment:  totalPayment,
		ExpiresAt:     time.Now().Add(15 * time.Minute),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.prepaidRepo.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Build response
	response := &domain.PrepaidOrderResponse{
		Order: &domain.OrderInfo{
			OrderID:     orderID,
			Status:      domain.OrderPendingPayment,
			ServiceType: inquiry.ServiceType,
			CreatedAt:   order.CreatedAt.Format(time.RFC3339),
			ExpiresAt:   order.ExpiresAt.Format(time.RFC3339),
		},
		Product: &domain.OrderProductInfo{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Nominal:     product.Nominal,
		},
		Target: &domain.OrderTargetInfo{
			Number: inquiry.Target,
		},
		Pricing: &domain.PricingInfo{
			ProductPrice:           productPrice,
			ProductPriceFormatted:  formatCurrency(productPrice),
			AdminFee:               adminFee,
			AdminFeeFormatted:      formatCurrency(adminFee),
			Subtotal:               subtotal,
			SubtotalFormatted:      formatCurrency(subtotal),
			Vouchers:               []*domain.VoucherInfo{},
			TotalDiscount:          totalDiscount,
			TotalDiscountFormatted: formatCurrency(totalDiscount),
			TotalPayment:           totalPayment,
			TotalPaymentFormatted:  formatCurrency(totalPayment),
		},
		Payment: &domain.PaymentInfo{
			Method:                    "balance",
			BalanceAvailable:          balance.Amount,
			BalanceAvailableFormatted: formatCurrency(balance.Amount),
			BalanceSufficient:         balanceSufficient,
		},
		PINRequired: pinRequired,
	}

	// Add operator info
	if inquiry.OperatorID != nil {
		response.Target.Operator = &domain.OperatorInfo{
			ID:   *inquiry.OperatorID,
			Name: getOperatorName(*inquiry.OperatorID),
		}
	}

	// Add shortfall if balance insufficient
	if !balanceSufficient {
		shortfall := totalPayment - balance.Amount
		shortfallFormatted := formatCurrency(shortfall)
		response.Payment.Shortfall = &shortfall
		response.Payment.ShortfallFormatted = &shortfallFormatted
	}

	return response, nil
}

// PayRequest represents payment request
type PayRequest struct {
	UserID  string
	OrderID string
	PIN     *string
}

// Pay handles payment execution with database transaction
func (s *PrepaidService) Pay(ctx context.Context, req PayRequest) (*domain.PrepaidPayResponse, error) {
	// Get order with ownership validation
	order, err := s.prepaidRepo.FindOrderByUserAndID(ctx, req.UserID, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return nil, domain.ErrValidationFailed("Order not found")
	}

	// Check order expiration
	if time.Now().After(order.ExpiresAt) {
		return nil, domain.ErrOrderExpired
	}

	// Check order status
	if order.Status != domain.OrderPendingPayment {
		return nil, domain.ErrValidationFailed("Order is not pending payment")
	}

	// Check for existing transaction (idempotency)
	existingTx, err := s.prepaidRepo.FindTransactionByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing transaction: %w", err)
	}
	if existingTx != nil {
		// Transaction already exists, return duplicate error
		return nil, domain.ErrDuplicateTransaction
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// TODO: Get user settings to check PIN requirement
	pinRequired := false

	// Verify PIN if required
	if pinRequired {
		if req.PIN == nil {
			return nil, domain.ErrPINRequiredError
		}
		if !hash.VerifyPIN(*req.PIN, user.PINHash) {
			return nil, domain.ErrInvalidPINError
		}
	}

	// Execute payment in database transaction
	var balanceBefore, balanceAfter int64
	var serialNumber, referenceNumber string
	var token, kwh *string
	transactionID := "trx_" + uuid.New().String()[:8]

	// Begin transaction
	tx, err := s.prepaidRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be no-op if transaction is committed

	// Lock balance and deduct atomically
	balance, err := s.balanceRepo.FindByUserIDForUpdate(ctx, tx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to lock balance: %w", err)
	}
	if balance == nil {
		return nil, domain.ErrValidationFailed("Balance not found")
	}

	// Check balance sufficiency
	if balance.Amount < order.TotalPayment {
		return nil, domain.ErrInsufficientBalance
	}

	// Store balance before deduction
	balanceBefore = balance.Amount
	balance.Amount -= order.TotalPayment

	// Update balance within transaction
	if err := s.balanceRepo.UpdateWithTx(ctx, tx, balance); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}
	balanceAfter = balance.Amount

	// Call Gerbang API to execute prepaid transaction
	// TODO: Get product SKU code from order (for now using mock)
	productSKU := "MOCK_SKU_" + order.ProductID
	referenceNumber = "PPOB" + transactionID

	gerbangResp, err := s.gerbangClient.CreatePrepaidTransaction(ctx, referenceNumber, productSKU, order.Target)
	if err != nil {
		// Log error but continue with mock response for now (development mode)
		// In production, this should fail and rollback
		serialNumber = "SN" + uuid.New().String()[:10]
		referenceNumber = "REF" + time.Now().Format("20060102150405")
	} else {
		// Use actual response from Gerbang
		if gerbangResp.SerialNumber != nil {
			serialNumber = *gerbangResp.SerialNumber
		}
		referenceNumber = gerbangResp.TransactionID
	}

	// Generate PLN-specific fields if applicable
	if order.ServiceType == domain.ServicePLNPrepaid {
		// Try to extract from Gerbang response description
		if gerbangResp != nil && gerbangResp.Description != nil {
			if tokenVal, ok := gerbangResp.Description["token"]; ok {
				if tokenStr, ok := tokenVal.(string); ok {
					token = &tokenStr
				}
			}
			if kwhVal, ok := gerbangResp.Description["kwh"]; ok {
				if kwhStr, ok := kwhVal.(string); ok {
					kwh = &kwhStr
				}
			}
		}

		// Fallback to mock if not in response
		if token == nil {
			tokenStr := generatePLNToken()
			token = &tokenStr
		}
		if kwh == nil {
			kwhStr := calculateKWH(order.TotalPayment)
			kwh = &kwhStr
		}
	}

	// Create transaction record within DB transaction
	completedAt := time.Now()
	transaction := &domain.PrepaidTransaction{
		ID:              transactionID,
		UserID:          req.UserID,
		OrderID:         req.OrderID,
		Status:          domain.TransactionSuccess,
		ServiceType:     order.ServiceType,
		Target:          order.Target,
		ProductID:       order.ProductID,
		TotalPayment:    order.TotalPayment,
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
		SerialNumber:    &serialNumber,
		ReferenceNumber: &referenceNumber,
		Token:           token,
		KWH:             kwh,
		CompletedAt:     &completedAt,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.prepaidRepo.CreateTransactionWithTx(ctx, tx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Update order status within transaction
	if err := s.prepaidRepo.UpdateOrderStatusWithTx(ctx, tx, req.OrderID, domain.OrderSuccess); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// TODO: Get product details
	product := getMockProduct(order.ProductID)

	// Build response
	completedAtStr := completedAt.Format(time.RFC3339)
	response := &domain.PrepaidPayResponse{
		Transaction: &domain.TransactionInfo{
			TransactionID: transactionID,
			OrderID:       req.OrderID,
			Status:        domain.TransactionSuccess,
			ServiceType:   order.ServiceType,
			CompletedAt:   &completedAtStr,
		},
		Product: &domain.TransactionProductInfo{
			ID:      product.ID,
			Name:    product.Name,
			Nominal: product.Nominal,
		},
		Target: &domain.TransactionTargetInfo{
			Number: order.Target,
		},
		Payment: &domain.TransactionPaymentInfo{
			TotalPayment:          order.TotalPayment,
			TotalPaymentFormatted: formatCurrency(order.TotalPayment),
			BalanceBefore:         balanceBefore,
			BalanceAfter:          balanceAfter,
			BalanceAfterFormatted: formatCurrency(balanceAfter),
		},
		Receipt: &domain.ReceiptInfo{
			SerialNumber:    &serialNumber,
			ReferenceNumber: &referenceNumber,
			Token:           token,
			KWH:             kwh,
		},
		Message: &domain.MessageInfo{
			Title:    getSuccessTitle(order.ServiceType),
			Subtitle: getSuccessSubtitle(order.ServiceType, order.TotalPayment),
		},
	}

	// Customize message for PLN Prepaid with token
	if order.ServiceType == domain.ServicePLNPrepaid && token != nil {
		response.Message.Subtitle = fmt.Sprintf("Token listrik Anda: %s", *token)
	}

	return response, nil
}

// Helper functions

func isValidServiceType(serviceType string) bool {
	validTypes := []string{
		domain.ServicePulsa,
		domain.ServiceData,
		domain.ServicePLNPrepaid,
		domain.ServiceEwallet,
		domain.ServiceGame,
	}
	for _, t := range validTypes {
		if t == serviceType {
			return true
		}
	}
	return false
}

func validateTarget(serviceType, target string) error {
	if target == "" {
		return domain.ErrInvalidTarget
	}

	switch serviceType {
	case domain.ServicePulsa, domain.ServiceData:
		// Validate phone number format (10-13 digits, starts with 08)
		if len(target) < 10 || len(target) > 13 {
			return domain.ErrInvalidTarget
		}
		if target[:2] != "08" {
			return domain.ErrInvalidTarget
		}
		// Check if all digits
		for _, c := range target {
			if c < '0' || c > '9' {
				return domain.ErrInvalidTarget
			}
		}
	case domain.ServicePLNPrepaid:
		// Validate PLN meter number (11-12 digits)
		if len(target) < 11 || len(target) > 12 {
			return domain.ErrInvalidTarget
		}
		// Check if all digits
		for _, c := range target {
			if c < '0' || c > '9' {
				return domain.ErrInvalidTarget
			}
		}
	case domain.ServiceEwallet:
		// Validate phone number format for e-wallet
		if len(target) < 10 || len(target) > 13 {
			return domain.ErrInvalidTarget
		}
		if target[:2] != "08" {
			return domain.ErrInvalidTarget
		}
		// Check if all digits
		for _, c := range target {
			if c < '0' || c > '9' {
				return domain.ErrInvalidTarget
			}
		}
	case domain.ServiceGame:
		// Game user ID - minimal validation
		if len(target) < 3 {
			return domain.ErrInvalidTarget
		}
	}

	return nil
}

func detectOperator(target string) string {
	if len(target) < 4 {
		return "unknown"
	}
	prefix := target[:4]
	operatorMap := map[string]string{
		// Telkomsel
		"0811": "telkomsel", "0812": "telkomsel", "0813": "telkomsel",
		"0821": "telkomsel", "0822": "telkomsel", "0823": "telkomsel",
		"0852": "telkomsel", "0853": "telkomsel",
		// Indosat
		"0814": "indosat", "0815": "indosat", "0816": "indosat",
		"0855": "indosat", "0856": "indosat", "0857": "indosat", "0858": "indosat",
		// XL
		"0817": "xl", "0818": "xl", "0819": "xl",
		"0859": "xl", "0877": "xl", "0878": "xl",
		// Axis
		"0831": "axis", "0832": "axis", "0833": "axis", "0838": "axis",
		// Three
		"0895": "three", "0896": "three", "0897": "three", "0898": "three", "0899": "three",
		// Smartfren
		"0881": "smartfren", "0882": "smartfren", "0883": "smartfren",
		"0884": "smartfren", "0885": "smartfren", "0886": "smartfren",
		"0887": "smartfren", "0888": "smartfren", "0889": "smartfren",
		// by.U
		"0851": "byu",
	}
	if operator, ok := operatorMap[prefix]; ok {
		return operator
	}
	return "unknown"
}

func getOperatorName(operatorID string) string {
	names := map[string]string{
		"telkomsel": "Telkomsel",
		"indosat":   "Indosat Ooredoo",
		"xl":        "XL Axiata",
		"axis":      "Axis",
		"three":     "Tri",
		"smartfren": "Smartfren",
		"byu":       "by.U",
	}
	if name, ok := names[operatorID]; ok {
		return name
	}
	return operatorID
}

// generatePLNToken generates a mock PLN token
func generatePLNToken() string {
	// Mock PLN token format: XXXX-XXXX-XXXX-XXXX-XXXX
	parts := make([]string, 5)
	for i := 0; i < 5; i++ {
		parts[i] = fmt.Sprintf("%04d", 1000+i*1111)
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", parts[0], parts[1], parts[2], parts[3], parts[4])
}

// calculateKWH calculates KWH from token amount (mock)
func calculateKWH(amount int64) string {
	// Mock calculation: assume Rp1500 per kWh
	kwh := float64(amount) / 1500.0
	return fmt.Sprintf("%.1f kWh", kwh)
}

func getMockProducts(serviceType string) []*domain.ProductInfo {
	// TODO: Replace with actual product fetch
	return []*domain.ProductInfo{
		{
			ID:             "prd_pulsa_5k",
			Name:           "Pulsa 5.000",
			Description:    "Masa Aktif 7 Hari",
			Category:       serviceType,
			Nominal:        5000,
			Price:          5000,
			PriceFormatted: "Rp5.000",
			AdminFee:       0,
			Discount:       nil,
			Status:         "active",
			Stock:          "available",
		},
		{
			ID:             "prd_pulsa_10k",
			Name:           "Pulsa 10.000",
			Description:    "Masa Aktif 7 Hari",
			Category:       serviceType,
			Nominal:        10000,
			Price:          10000,
			PriceFormatted: "Rp10.000",
			AdminFee:       0,
			Discount:       nil,
			Status:         "active",
			Stock:          "available",
		},
	}
}

func getMockProduct(productID string) *domain.ProductInfo {
	products := map[string]*domain.ProductInfo{
		"prd_pulsa_5k": {
			ID:             "prd_pulsa_5k",
			Name:           "Pulsa 5.000",
			Description:    "Masa Aktif 7 Hari",
			Category:       domain.ServicePulsa,
			Nominal:        5000,
			Price:          5000,
			PriceFormatted: "Rp5.000",
			AdminFee:       0,
			Status:         "active",
			Stock:          "available",
		},
		"prd_pulsa_10k": {
			ID:             "prd_pulsa_10k",
			Name:           "Pulsa 10.000",
			Description:    "Masa Aktif 7 Hari",
			Category:       domain.ServicePulsa,
			Nominal:        10000,
			Price:          10000,
			PriceFormatted: "Rp10.000",
			AdminFee:       0,
			Status:         "active",
			Stock:          "available",
		},
	}
	return products[productID]
}

func formatCurrency(amount int64) string {
	return fmt.Sprintf("Rp%s", formatNumber(amount))
}

func formatNumber(n int64) string {
	// Simple number formatting
	// TODO: Use proper locale-aware formatting
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}
	return result
}

func getSuccessTitle(serviceType string) string {
	titles := map[string]string{
		domain.ServicePulsa:      "Yeay Pembelian Pulsa Berhasil",
		domain.ServiceData:       "Yeay Pembelian Paket Data Berhasil",
		domain.ServicePLNPrepaid: "Yeay Pembelian Token Berhasil",
		domain.ServiceEwallet:    "Yeay Top Up E-Wallet Berhasil",
		domain.ServiceGame:       "Yeay Pembelian Voucher Game Berhasil",
	}
	if title, ok := titles[serviceType]; ok {
		return title
	}
	return "Transaksi Berhasil"
}

func getSuccessSubtitle(serviceType string, amount int64) string {
	return fmt.Sprintf("Total pembayaran %s, lihat rincian riwayat untuk informasi lebih lengkap", formatCurrency(amount))
}

// Refund restores balance for failed/cancelled transaction
func (s *PrepaidService) Refund(ctx context.Context, transactionID, reason string) error {
	// Begin transaction
	tx, err := s.prepaidRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get transaction
	transaction, err := s.prepaidRepo.FindTransactionByID(ctx, transactionID)
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
	if err := s.prepaidRepo.UpdateTransactionStatus(ctx, transactionID, domain.TransactionRefunded); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// TODO: Create balance history record when balance_history table is ready

	return tx.Commit()
}

// HandleWebhook handles prepaid transaction webhook notification from Gerbang API
// Called when transaction status changes to Success, Failed, or Pending
func (s *PrepaidService) HandleWebhook(ctx context.Context, webhookData *gerbang.TransactionWebhookData) error {
	// Find transaction by reference ID (which is our order ID)
	order, err := s.prepaidRepo.FindOrderByID(ctx, webhookData.ReferenceID)
	if err != nil {
		return fmt.Errorf("failed to find order: %w", err)
	}
	if order == nil {
		return domain.ErrNotFound("Pesanan prepaid")
	}

	// Check if already processed - return nil for idempotency
	// This ensures webhook sender gets 200 OK and won't retry forever
	if order.Status == domain.TransactionSuccess || order.Status == domain.TransactionFailed {
		return nil // Idempotent: already processed
	}

	// Begin database transaction for atomic operation
	tx, err := s.prepaidRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update order based on webhook event
	now := time.Now()
	var newStatus string
	var refundNeeded bool

	switch webhookData.Status {
	case "Success":
		newStatus = domain.TransactionSuccess
		refundNeeded = false
		// Serial number is stored in PrepaidTransaction, not PrepaidOrder
		// Will be handled when updating transaction record

	case "Failed":
		newStatus = domain.TransactionFailed
		refundNeeded = true

	case "Pending":
		newStatus = domain.TransactionProcessing // Use existing constant
		refundNeeded = false

	default:
		return fmt.Errorf("unknown transaction status: %s", webhookData.Status)
	}

	// Update order status
	order.Status = newStatus
	order.UpdatedAt = now

	if err := s.prepaidRepo.UpdateOrderStatusWithTx(ctx, tx, order.ID, newStatus); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// If transaction failed, refund balance to user
	if refundNeeded {
		// Lock and get user balance
		balance, err := s.balanceRepo.FindByUserIDForUpdate(ctx, tx, order.UserID)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}
		if balance == nil {
			return domain.ErrNotFound("Saldo")
		}

		// Restore balance
		balance.Amount += order.TotalPayment
		balance.UpdatedAt = now

		if err := s.balanceRepo.UpdateWithTx(ctx, tx, balance); err != nil {
			return fmt.Errorf("failed to restore balance: %w", err)
		}

		// TODO: Create balance history record (refund)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
