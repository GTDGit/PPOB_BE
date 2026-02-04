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

// TransferService handles transfer transaction business logic
type TransferService struct {
	transferRepo  repository.TransferRepository
	balanceRepo   repository.BalanceRepository
	userRepo      repository.UserRepository
	productRepo   repository.ProductRepository
	gerbangClient *gerbang.Client
}

// NewTransferService creates a new transfer service
func NewTransferService(
	transferRepo repository.TransferRepository,
	balanceRepo repository.BalanceRepository,
	userRepo repository.UserRepository,
	productRepo repository.ProductRepository,
	gerbangClient *gerbang.Client,
) *TransferService {
	return &TransferService{
		transferRepo:  transferRepo,
		balanceRepo:   balanceRepo,
		userRepo:      userRepo,
		productRepo:   productRepo,
		gerbangClient: gerbangClient,
	}
}

// TransferInquiryRequest represents inquiry request
type TransferInquiryRequest struct {
	UserID        string
	BankCode      string
	AccountNumber string
	Amount        int64
}

// Inquiry handles transfer inquiry
func (s *TransferService) Inquiry(ctx context.Context, req TransferInquiryRequest) (*domain.TransferInquiryResponse, error) {
	// CRITICAL: Check KYC status
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUnauthorizedError
	}

	// Check if user is KYC verified
	if user.KYCStatus != "verified" {
		return nil, domain.ErrKYCRequired
	}

	// Validate amount
	if req.Amount < 10000 {
		return nil, domain.ErrValidationFailed("Minimum transfer amount is Rp10.000")
	}

	// Call Gerbang API to validate account and get account name
	inquiryResp, err := s.gerbangClient.TransferInquiry(ctx, req.BankCode, req.AccountNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to inquiry account: %w", err)
	}

	// Get estimated fee from banks table
	bank, err := s.productRepo.FindBankByCode(ctx, req.BankCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank info: %w", err)
	}

	// Use transfer fee from bank info as estimated admin fee
	// Actual fee will be known during execute from Gerbang response
	adminFee := int64(0)
	if bank != nil {
		adminFee = bank.TransferFee
	} else {
		// Fallback to default if bank not found
		adminFee = calculateAdminFee(req.BankCode, req.Amount)
	}

	totalPayment := req.Amount + adminFee

	// Get user balance and check sufficiency BEFORE calling Gerbang API
	balance, err := s.balanceRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	if balance == nil {
		return nil, domain.ErrNotFound("Balance")
	}
	if balance.Amount < totalPayment {
		return nil, domain.ErrInsufficientBalance
	}

	// Create inquiry record
	inquiryID := "inq_tf_" + uuid.New().String()[:8]
	inquiry := &domain.TransferInquiry{
		ID:               inquiryID,
		UserID:           req.UserID,
		BankCode:         req.BankCode,
		BankName:         inquiryResp.BankName,
		AccountNumber:    req.AccountNumber,
		AccountName:      inquiryResp.AccountName,
		Amount:           req.Amount,
		AdminFee:         adminFee,
		TotalPayment:     totalPayment,
		GerbangInquiryID: &inquiryResp.InquiryID,
		Fee:              adminFee, // Store estimated fee, actual fee from execute
		ExpiresAt:        time.Now().Add(30 * time.Minute),
		CreatedAt:        time.Now(),
	}

	if err := s.transferRepo.CreateInquiry(ctx, inquiry); err != nil {
		return nil, fmt.Errorf("failed to create inquiry: %w", err)
	}

	// Build response
	expiresAt := inquiry.ExpiresAt.Format(time.RFC3339)
	response := &domain.TransferInquiryResponse{
		Inquiry: &domain.TransferInquiryInfo{
			InquiryID: inquiryID,
			ExpiresAt: expiresAt,
		},
		Destination: &domain.TransferDestinationInfo{
			BankCode:      req.BankCode,
			BankName:      inquiryResp.BankName,
			AccountNumber: req.AccountNumber,
			AccountName:   inquiryResp.AccountName,
		},
		Transfer: &domain.TransferAmountInfo{
			Amount:                req.Amount,
			AmountFormatted:       formatCurrency(req.Amount),
			AdminFee:              adminFee,
			AdminFeeFormatted:     formatCurrency(adminFee),
			TotalPayment:          totalPayment,
			TotalPaymentFormatted: formatCurrency(totalPayment),
		},
		PINRequired: false, // TODO: Get from user settings
		Notices: []*domain.NoticeInfo{
			{
				Type:    "info",
				Message: "Transfer akan diproses dalam 1-10 menit",
			},
		},
	}

	// Add bank short name and icon
	bankShortName := inquiryResp.BankShortName
	bankIcon := ""
	if bank != nil {
		bankIcon = bank.IconURL
	}
	response.Destination.BankShortName = &bankShortName
	if bankIcon != "" {
		response.Destination.BankIcon = &bankIcon
	}

	// Add payment info - balance already checked above
	balanceSufficient := balance.Amount >= totalPayment
	response.Payment = &domain.PaymentInfo{
		Method:                    "balance",
		BalanceAvailable:          balance.Amount,
		BalanceAvailableFormatted: formatCurrency(balance.Amount),
		BalanceSufficient:         balanceSufficient,
	}

	if !balanceSufficient {
		shortfall := totalPayment - balance.Amount
		shortfallFormatted := formatCurrency(shortfall)
		response.Payment.Shortfall = &shortfall
		response.Payment.ShortfallFormatted = &shortfallFormatted
	}

	return response, nil
}

// TransferExecuteRequest represents execute request
type TransferExecuteRequest struct {
	UserID    string
	InquiryID string
	Purpose   *string // Purpose code (01, 02, 03, 99) - default "99" if nil
	Note      *string
	PIN       *string
}

// Execute handles transfer execution with database transaction
func (s *TransferService) Execute(ctx context.Context, req TransferExecuteRequest) (*domain.TransferExecuteResponse, error) {
	// Get inquiry with ownership validation
	inquiry, err := s.transferRepo.FindInquiryByUserAndID(ctx, req.UserID, req.InquiryID)
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

	// CRITICAL: Re-check KYC status for security
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUnauthorizedError
	}

	// Re-check KYC verification
	if user.KYCStatus != "verified" {
		return nil, domain.ErrKYCRequired
	}

	// Check for existing transaction (idempotency)
	existingTx, err := s.transferRepo.FindTransactionByInquiryID(ctx, req.InquiryID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing transaction: %w", err)
	}
	if existingTx != nil {
		return nil, domain.ErrDuplicateTransaction
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

	// Execute transfer in database transaction
	var balanceBefore, balanceAfter int64
	var referenceNumber string
	transactionID := "trx_tf_" + uuid.New().String()[:8]

	// Begin transaction
	tx, err := s.transferRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock balance and deduct atomically
	balance, err := s.balanceRepo.FindByUserIDForUpdate(ctx, tx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to lock balance: %w", err)
	}
	if balance == nil {
		return nil, domain.ErrValidationFailed("Balance not found")
	}

	// Prepare purpose code (default "99" if not provided)
	purposeCode := "99"
	if req.Purpose != nil && *req.Purpose != "" {
		purposeCode = *req.Purpose
	}

	// Prepare remark (note)
	remark := ""
	if req.Note != nil {
		remark = *req.Note
	}

	// Call Gerbang API to execute transfer
	gerbangReq := gerbang.GerbangTransferExecuteRequest{
		ReferenceID:   transactionID,
		InquiryID:     *inquiry.GerbangInquiryID,
		BankCode:      inquiry.BankCode,
		AccountNumber: inquiry.AccountNumber,
		AccountName:   inquiry.AccountName,
		Amount:        inquiry.Amount,
		Purpose:       purposeCode,
		Remark:        remark,
	}

	gerbangResp, err := s.gerbangClient.TransferExecute(ctx, gerbangReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transfer via Gerbang: %w", err)
	}

	// Use actual fee from Gerbang response
	actualFee := gerbangResp.Fee
	totalDeduction := inquiry.Amount + actualFee

	// Check balance sufficiency with actual fee
	if balance.Amount < totalDeduction {
		return nil, domain.ErrInsufficientBalance
	}

	// Store balance before deduction
	balanceBefore = balance.Amount
	balance.Amount -= totalDeduction

	// Update balance within transaction
	if err := s.balanceRepo.UpdateWithTx(ctx, tx, balance); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}
	balanceAfter = balance.Amount

	// Get reference number from Gerbang response
	referenceNumber = gerbangResp.TransferID

	// Create transaction record within DB transaction
	completedAt := time.Now()
	gerbangTransferID := gerbangResp.TransferID

	// Determine status from Gerbang response
	status := domain.TransactionProcessing
	if gerbangResp.Status == "Success" {
		status = domain.TransactionSuccess
	} else if gerbangResp.Status == "Failed" {
		status = domain.TransactionFailed
	}

	transaction := &domain.TransferTransaction{
		ID:                transactionID,
		UserID:            req.UserID,
		InquiryID:         req.InquiryID,
		Status:            status,
		BankCode:          inquiry.BankCode,
		BankName:          inquiry.BankName,
		AccountNumber:     inquiry.AccountNumber,
		AccountName:       inquiry.AccountName,
		Amount:            inquiry.Amount,
		AdminFee:          inquiry.AdminFee, // Keep estimate for reference
		TotalPayment:      totalDeduction,   // Actual total with real fee
		Note:              req.Note,
		BalanceBefore:     balanceBefore,
		BalanceAfter:      balanceAfter,
		ReferenceNumber:   &referenceNumber,
		GerbangTransferID: &gerbangTransferID,
		Purpose:           &purposeCode,
		Fee:               actualFee, // Actual fee from Gerbang
		CompletedAt:       &completedAt,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.transferRepo.CreateTransactionWithTx(ctx, tx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Build response
	completedAtStr := completedAt.Format(time.RFC3339)
	response := &domain.TransferExecuteResponse{
		Transaction: &domain.TransferTransactionInfo{
			TransactionID: transactionID,
			InquiryID:     req.InquiryID,
			Status:        status,
			CompletedAt:   &completedAtStr,
		},
		Destination: &domain.TransferDestInfo{
			BankName:      inquiry.BankName,
			AccountNumber: inquiry.AccountNumber,
			AccountName:   inquiry.AccountName,
		},
		Transfer: &domain.TransferExecInfo{
			Amount:          inquiry.Amount,
			AmountFormatted: formatCurrency(inquiry.Amount),
			Note:            req.Note,
		},
		Payment: &domain.TransferPaymentInfo{
			TotalPayment:          totalDeduction,
			TotalPaymentFormatted: formatCurrency(totalDeduction),
			BalanceBefore:         balanceBefore,
			BalanceAfter:          balanceAfter,
			BalanceAfterFormatted: formatCurrency(balanceAfter),
		},
		Receipt: &domain.ReceiptInfo{
			ReferenceNumber: &referenceNumber,
		},
		Message: &domain.MessageInfo{
			Title:    getStatusTitle(status),
			Subtitle: getStatusSubtitle(status, inquiry.Amount, inquiry.AccountName),
		},
	}

	return response, nil
}

// Helper to get status title
func getStatusTitle(status string) string {
	switch status {
	case domain.TransactionSuccess:
		return "Transfer Berhasil"
	case domain.TransactionProcessing:
		return "Transfer Diproses"
	case domain.TransactionFailed:
		return "Transfer Gagal"
	default:
		return "Transfer Diproses"
	}
}

// Helper to get status subtitle
func getStatusSubtitle(status string, amount int64, accountName string) string {
	switch status {
	case domain.TransactionSuccess:
		return fmt.Sprintf("Transfer %s ke %s berhasil", formatCurrency(amount), accountName)
	case domain.TransactionProcessing:
		return fmt.Sprintf("Transfer %s ke %s sedang diproses", formatCurrency(amount), accountName)
	case domain.TransactionFailed:
		return fmt.Sprintf("Transfer %s ke %s gagal", formatCurrency(amount), accountName)
	default:
		return fmt.Sprintf("Transfer %s ke %s sedang diproses", formatCurrency(amount), accountName)
	}
}

// Helper functions

// getBankName returns bank name by code
func getBankName(bankCode string) string {
	bankMap := map[string]string{
		"002": "Bank BRI",
		"008": "Bank Mandiri",
		"009": "Bank BNI",
		"014": "Bank Central Asia",
		"022": "Bank CIMB Niaga",
		"213": "Bank BTPN",
		"451": "Bank Syariah Indonesia",
		"200": "Bank BTN",
		"110": "Bank Jabar Banten",
		"111": "Bank DKI",
		"153": "Bank Sinarmas",
		"013": "Bank Permata",
		"011": "Bank Danamon",
		"016": "Bank Maybank",
		"019": "Bank Panin",
		"023": "Bank UOB",
		"028": "Bank OCBC NISP",
		"031": "Citibank",
		"032": "JP Morgan Chase Bank",
		"037": "Bank Artha Graha",
		"039": "Bank Capital",
		"041": "Bank HSBC",
		"042": "Bank of Tokyo Mitsubishi UFJ",
		"046": "Bank DBS Indonesia",
		"050": "Standard Chartered Bank",
		"054": "Bank China Construction Bank",
		"061": "Bank ANZ Indonesia",
		"069": "Bank of China",
		"087": "Bank Ekonomi",
		"089": "Bank Rakyat Indonesia Agroniaga",
		"093": "Bank Ina Perdana",
		"095": "Bank Mitra Niaga",
		"097": "Bank Mayapada",
		"109": "Bank Nusantara Parahyangan",
		"112": "Bank Jatim",
		"113": "Bank Aceh",
		"114": "Bank Sumut",
		"115": "Bank Nagari",
		"116": "Bank Riau Kepri",
		"117": "Bank Jambi",
		"118": "Bank Bengkulu",
		"119": "Bank Sumsel Babel",
		"120": "Bank Lampung",
		"121": "Bank Kalsel",
		"122": "Bank Kalbar",
		"123": "Bank Kaltim Kaltara",
		"124": "Bank Kalteng",
		"125": "Bank Sulselbar",
		"126": "Bank Sulutgo",
		"127": "Bank Sulteng",
		"128": "Bank Sultra",
		"129": "Bank Bali",
		"130": "Bank NTB",
		"131": "Bank NTB Syariah",
		"132": "Bank NTT",
		"133": "Bank Maluku Malut",
		"134": "Bank Papua",
		"135": "Bank Banten",
		"145": "Bank Pembangunan Daerah Banten",
		"146": "Bank Yogyakarta",
		"147": "Bank Jateng",
		"212": "Bank Woori Saudara",
		"405": "Bank Victoria Syariah",
		"426": "Bank Mega",
		"441": "Bank Bukopin",
		"459": "Bank Bisnis Internasional",
		"466": "Bank Sri Partha",
		"472": "Bank Jasa Jakarta",
		"484": "Bank KEB Hana",
		"485": "Bank MNC Internasional",
		"490": "Bank Yudha Bhakti",
		"494": "Bank Raya Indonesia",
		"498": "Bank BPD DIY",
		"501": "Bank Digital BCA",
		"503": "Bank Neo Commerce",
		"506": "Bank Seabank Indonesia",
		"513": "Bank Allo",
		"521": "Bank Bukopin Syariah",
		"523": "Bank Sahabat Sampoerna",
		"526": "Bank Mega Syariah",
		"535": "Bank Jago",
		"536": "Bank BNC",
		"542": "Bank Mitra Niaga",
		"547": "Bank Amar Indonesia",
		"553": "Bank Mayora",
		"555": "Bank Index Selindo",
		"558": "Bank Commonwealth",
		"562": "Bank China Construction Bank Indonesia",
		"564": "Bank QNB Indonesia",
		"567": "Bank Dinar Indonesia",
		"688": "Bank Seabank Indonesia",
	}
	if name, ok := bankMap[bankCode]; ok {
		return name
	}
	return ""
}

// getBankShortName returns bank short name by code
func getBankShortName(bankCode string) string {
	shortNames := map[string]string{
		"002": "BRI",
		"008": "Mandiri",
		"009": "BNI",
		"014": "BCA",
		"022": "CIMB Niaga",
		"213": "BTPN",
		"451": "BSI",
		"200": "BTN",
		"013": "Permata",
		"011": "Danamon",
		"016": "Maybank",
		"019": "Panin",
		"023": "UOB",
		"028": "OCBC NISP",
	}
	if shortName, ok := shortNames[bankCode]; ok {
		return shortName
	}
	return getBankName(bankCode)
}

// getBankIcon returns bank icon URL by code
func getBankIcon(bankCode string) string {
	iconMap := map[string]string{
		"002": "https://cdn.ppob.id/banks/bri.png",
		"008": "https://cdn.ppob.id/banks/mandiri.png",
		"009": "https://cdn.ppob.id/banks/bni.png",
		"014": "https://cdn.ppob.id/banks/bca.png",
		"022": "https://cdn.ppob.id/banks/cimb.png",
		"213": "https://cdn.ppob.id/banks/btpn.png",
		"451": "https://cdn.ppob.id/banks/bsi.png",
		"200": "https://cdn.ppob.id/banks/btn.png",
	}
	if icon, ok := iconMap[bankCode]; ok {
		return icon
	}
	return "https://cdn.ppob.id/banks/default.png"
}

// calculateAdminFee calculates admin fee for transfer
func calculateAdminFee(bankCode string, amount int64) int64 {
	// Same bank (BCA to BCA example)
	if bankCode == "014" {
		return 0 // Free for same bank
	}
	// Other banks
	return 6500
}

// getMockAccountName returns mock account name for testing
func getMockAccountName(accountNumber string) string {
	// Mock account validation
	// In production, this will call the bank API
	return "BUDI SANTOSO"
}

// Refund restores balance for failed/cancelled transaction
func (s *TransferService) Refund(ctx context.Context, transactionID, reason string) error {
	// Begin transaction
	tx, err := s.transferRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get transaction
	transaction, err := s.transferRepo.FindTransactionByID(ctx, transactionID)
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
	if err := s.transferRepo.UpdateTransactionStatus(ctx, transactionID, domain.TransactionRefunded); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// TODO: Create balance history record when balance_history table is ready

	return tx.Commit()
}

// HandleWebhook handles transfer webhook notification from Gerbang API
// Called when transfer status changes to Success or Failed
func (s *TransferService) HandleWebhook(ctx context.Context, webhookData *gerbang.TransferWebhookData) error {
	// Find transaction by reference ID (which is our transaction ID)
	transaction, err := s.transferRepo.FindTransactionByID(ctx, webhookData.ReferenceID)
	if err != nil {
		return fmt.Errorf("failed to find transaction: %w", err)
	}
	if transaction == nil {
		return domain.ErrNotFound("Transaksi transfer")
	}

	// Check if already processed - return nil for idempotency
	// This ensures webhook sender gets 200 OK and won't retry forever
	if transaction.Status == domain.TransactionSuccess || transaction.Status == domain.TransactionFailed {
		return nil // Idempotent: already processed, just return success
	}

	// Begin database transaction for atomic operation
	tx, err := s.transferRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update transaction based on webhook event
	now := time.Now()
	var newStatus string
	var refundNeeded bool

	switch webhookData.Status {
	case "Success":
		newStatus = domain.TransactionSuccess
		refundNeeded = false

	case "Failed":
		newStatus = domain.TransactionFailed
		refundNeeded = true

	default:
		return fmt.Errorf("unknown transfer status: %s", webhookData.Status)
	}

	// Update transaction status
	transaction.Status = newStatus
	transaction.UpdatedAt = now

	// Store provider reference if available (for logging/debugging)
	// Note: TransferTransaction domain doesn't have ProviderTransactionID field yet
	// Can be added later if needed for reconciliation

	// Log failure info if failed
	if webhookData.FailedReason != nil && webhookData.FailedCode != nil {
		// Log failure details for debugging
		// transaction.FailureInfo field can be added to domain.TransferTransaction later if needed
		// For now, the status itself (failed) is sufficient
	}

	if err := s.transferRepo.UpdateTransactionStatusWithTx(ctx, tx, transaction.ID, newStatus); err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// If transfer failed, refund balance to user
	if refundNeeded {
		// Lock and get user balance
		balance, err := s.balanceRepo.FindByUserIDForUpdate(ctx, tx, transaction.UserID)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}
		if balance == nil {
			return domain.ErrNotFound("Saldo")
		}

		// Restore balance
		balance.Amount += transaction.TotalPayment
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
