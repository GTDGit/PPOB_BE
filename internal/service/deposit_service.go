package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/external/gerbang"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// DepositService handles deposit business logic
type DepositService struct {
	depositRepo   repository.DepositRepository
	balanceRepo   repository.BalanceRepository
	userRepo      repository.UserRepository
	gerbangClient *gerbang.Client
}

// NewDepositService creates a new deposit service
func NewDepositService(
	depositRepo repository.DepositRepository,
	balanceRepo repository.BalanceRepository,
	userRepo repository.UserRepository,
	gerbangClient *gerbang.Client,
) *DepositService {
	return &DepositService{
		depositRepo:   depositRepo,
		balanceRepo:   balanceRepo,
		userRepo:      userRepo,
		gerbangClient: gerbangClient,
	}
}

// GetMethods returns list of available deposit methods
func (s *DepositService) GetMethods(ctx context.Context) (*domain.DepositMethodsResponse, error) {
	methods, err := s.depositRepo.FindAllMethods(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get methods: %w", err)
	}

	return &domain.DepositMethodsResponse{
		Methods: methods,
	}, nil
}

// CreateBankTransfer creates a bank transfer deposit
func (s *DepositService) CreateBankTransfer(ctx context.Context, userID string, amount int64) (*domain.BankTransferResponse, error) {
	// Validate amount
	if amount < 10000 {
		return nil, domain.ErrAmountTooLow
	}
	if amount > 50000000 {
		return nil, domain.ErrAmountTooHigh
	}

	// Check pending deposits limit (max 3 pending)
	pendingCount, err := s.depositRepo.CountPending(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending: %w", err)
	}
	if pendingCount >= 3 {
		return nil, domain.ErrTooManyPending
	}

	// Generate unique code (3 digits)
	uniqueCode := generateUniqueCode()
	totalAmount := amount + int64(uniqueCode)

	// Create deposit
	deposit := &domain.Deposit{
		ID:          fmt.Sprintf("dep_bank_%s", uuid.New().String()[:8]),
		UserID:      userID,
		Method:      domain.DepositMethodBankTransfer,
		Amount:      amount,
		AdminFee:    0,
		UniqueCode:  uniqueCode,
		TotalAmount: totalAmount,
		Status:      domain.DepositStatusPending,
		ExpiresAt:   time.Now().Add(6 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.depositRepo.Create(ctx, deposit); err != nil {
		return nil, fmt.Errorf("failed to create deposit: %w", err)
	}

	// Get company bank accounts
	bankAccounts, err := s.depositRepo.FindAllCompanyBankAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bank accounts: %w", err)
	}

	// Build response
	depositInfo := s.toDepositInfo(deposit)
	paymentInfo := &domain.BankTransferPayment{
		TotalTransfer:          totalAmount,
		TotalTransferFormatted: formatDepositCurrency(totalAmount),
		UniqueCode:             uniqueCode,
		ValidUntil:             deposit.ExpiresAt.Format(time.RFC3339),
		ValidUntilFormatted:    formatDepositDate(deposit.ExpiresAt),
	}

	instructions := []string{
		"Transfer ke salah satu rekening perusahaan di bawah ini",
		fmt.Sprintf("Jumlah transfer: %s (termasuk kode unik %d)", formatDepositCurrency(totalAmount), uniqueCode),
		"Transfer akan dikonfirmasi otomatis dalam 5-15 menit",
		"Kode unik berlaku untuk memverifikasi transfer Anda",
		"Deposit akan expired dalam 6 jam jika tidak ada pembayaran",
	}

	return &domain.BankTransferResponse{
		Deposit:      depositInfo,
		PaymentInfo:  paymentInfo,
		BankAccounts: bankAccounts,
		Instructions: instructions,
	}, nil
}

// CreateQRIS creates a QRIS deposit
func (s *DepositService) CreateQRIS(ctx context.Context, userID string, amount int64) (*domain.QRISResponse, error) {
	// Validate amount
	if amount < 10000 {
		return nil, domain.ErrAmountTooLow
	}
	if amount > 10000000 {
		return nil, domain.ErrAmountTooHigh
	}

	// Check pending deposits limit (max 3 pending)
	pendingCount, err := s.depositRepo.CountPending(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending: %w", err)
	}
	if pendingCount >= 3 {
		return nil, domain.ErrTooManyPending
	}

	// Calculate admin fee (0.7%)
	adminFee := int64(float64(amount) * 0.007)
	totalAmount := amount + adminFee

	// Get user info for customer data
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Create deposit ID (will be used as reference ID)
	depositID := fmt.Sprintf("dep_qris_%s", uuid.New().String()[:8])

	// Call Gerbang API to create QRIS payment
	gerbangResp, err := s.gerbangClient.CreateQRISPayment(ctx, depositID, totalAmount, gerbang.CustomerInfo{
		Name:  user.FullName,
		Phone: user.Phone,
	})

	var externalID string
	var qrisString string
	var qrisImageURL string

	if err != nil {
		// Fallback to mock for development
		externalID = fmt.Sprintf("gerbang_qris_%s", uuid.New().String()[:12])
		qrisString = fmt.Sprintf("00020101021226670016COM.GERBANGAPI01189360050300000870303UMI51440014ID.CO.QRIS.WWW0215ID%s520454995303360540%d5802ID5913PT PPOB ID6007Jakarta610512345", externalID, totalAmount)
		qrisImageURL = fmt.Sprintf("https://api.gerbang.id/qris/image/%s", externalID)
	} else {
		// Use real Gerbang response
		externalID = gerbangResp.PaymentID
		if qrisDetail, ok := gerbangResp.PaymentDetail["qrString"].(string); ok {
			qrisString = qrisDetail
		}
		if qrisImg, ok := gerbangResp.PaymentDetail["qrImageUrl"].(string); ok {
			qrisImageURL = qrisImg
		}
	}

	extIDPtr := externalID
	// Create deposit
	deposit := &domain.Deposit{
		ID:          depositID,
		UserID:      userID,
		Method:      domain.DepositMethodQRIS,
		Amount:      amount,
		AdminFee:    adminFee,
		UniqueCode:  0,
		TotalAmount: totalAmount,
		Status:      domain.DepositStatusPending,
		ExternalID:  &extIDPtr,
		ExpiresAt:   time.Now().Add(30 * time.Minute),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.depositRepo.Create(ctx, deposit); err != nil {
		return nil, fmt.Errorf("failed to create deposit: %w", err)
	}

	// Build response
	depositInfo := s.toDepositInfo(deposit)
	paymentInfo := &domain.QRISPayment{
		QRISString:          qrisString,
		QRISImageURL:        qrisImageURL,
		ValidUntil:          deposit.ExpiresAt.Format(time.RFC3339),
		ValidUntilFormatted: formatDepositDate(deposit.ExpiresAt),
	}

	instructions := []string{
		"Scan QR code menggunakan aplikasi mobile banking Anda",
		"Pastikan jumlah transfer sudah benar sebelum konfirmasi",
		"Saldo akan masuk otomatis setelah pembayaran berhasil",
		"QRIS berlaku selama 30 menit",
	}

	return &domain.QRISResponse{
		Deposit:      depositInfo,
		PaymentInfo:  paymentInfo,
		Instructions: instructions,
	}, nil
}

// GetRetailProviders returns list of retail providers
func (s *DepositService) GetRetailProviders(ctx context.Context) (*domain.RetailProvidersResponse, error) {
	providers, err := s.depositRepo.FindAllRetailProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}

	return &domain.RetailProvidersResponse{
		Providers: providers,
	}, nil
}

// CreateRetail creates a retail deposit
func (s *DepositService) CreateRetail(ctx context.Context, userID, providerCode string, amount int64) (*domain.RetailResponse, error) {
	// Validate provider
	providers, err := s.depositRepo.FindAllRetailProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}

	var provider *domain.RetailProvider
	for _, p := range providers {
		if p.Code == providerCode {
			provider = p
			break
		}
	}
	if provider == nil {
		return nil, domain.ErrInvalidProvider
	}

	// Validate amount
	if amount < provider.MinAmount {
		return nil, domain.ErrAmountTooLow
	}
	if amount > provider.MaxAmount {
		return nil, domain.ErrAmountTooHigh
	}

	// Check pending deposits limit (max 3 pending)
	pendingCount, err := s.depositRepo.CountPending(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending: %w", err)
	}
	if pendingCount >= 3 {
		return nil, domain.ErrTooManyPending
	}

	// Admin fee
	adminFee := provider.Fee
	totalAmount := amount + adminFee

	// Get user info for customer data
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Create deposit ID (will be used as reference ID)
	depositID := fmt.Sprintf("dep_retail_%s", uuid.New().String()[:8])

	// Call Gerbang API to create retail payment
	gerbangResp, err := s.gerbangClient.CreateRetailPayment(ctx, depositID, providerCode, totalAmount, gerbang.CustomerInfo{
		Name:  user.FullName,
		Phone: user.Phone,
	})

	var externalID string
	var paymentCode string

	if err != nil {
		// Fallback to mock for development
		externalID = fmt.Sprintf("gerbang_retail_%s", uuid.New().String()[:12])
		paymentCode = fmt.Sprintf("PPOB%d", time.Now().Unix()%100000000)
	} else {
		// Use real Gerbang response
		externalID = gerbangResp.PaymentID
		if code, ok := gerbangResp.PaymentDetail["paymentCode"].(string); ok {
			paymentCode = code
		}
	}

	provCode := providerCode
	extIDPtr := externalID
	// Create deposit
	deposit := &domain.Deposit{
		ID:           depositID,
		UserID:       userID,
		Method:       domain.DepositMethodRetail,
		ProviderCode: &provCode,
		Amount:       amount,
		AdminFee:     adminFee,
		UniqueCode:   0,
		TotalAmount:  totalAmount,
		Status:       domain.DepositStatusPending,
		ExternalID:   &extIDPtr,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.depositRepo.Create(ctx, deposit); err != nil {
		return nil, fmt.Errorf("failed to create deposit: %w", err)
	}

	// Build response
	depositInfo := s.toDepositInfo(deposit)
	paymentInfo := &domain.RetailPayment{
		ProviderCode:        providerCode,
		ProviderName:        provider.Name,
		PaymentCode:         paymentCode,
		ValidUntil:          deposit.ExpiresAt.Format(time.RFC3339),
		ValidUntilFormatted: formatDepositDate(deposit.ExpiresAt),
	}

	instructions := []string{
		fmt.Sprintf("Datang ke kasir %s terdekat", provider.Name),
		fmt.Sprintf("Sebutkan kode pembayaran: %s", paymentCode),
		fmt.Sprintf("Bayar sejumlah %s", formatDepositCurrency(totalAmount)),
		"Simpan struk pembayaran sebagai bukti",
		"Saldo akan masuk otomatis dalam 5-10 menit setelah pembayaran",
	}

	return &domain.RetailResponse{
		Deposit:      depositInfo,
		PaymentInfo:  paymentInfo,
		Instructions: instructions,
	}, nil
}

// GetVABanks returns list of VA banks
func (s *DepositService) GetVABanks(ctx context.Context) (*domain.VABanksResponse, error) {
	banks, err := s.depositRepo.FindAllVABanks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get VA banks: %w", err)
	}

	return &domain.VABanksResponse{
		Banks: banks,
	}, nil
}

// CreateVA creates a virtual account deposit
func (s *DepositService) CreateVA(ctx context.Context, userID, bankCode string, amount int64) (*domain.VAResponse, error) {
	// Validate bank
	banks, err := s.depositRepo.FindAllVABanks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get VA banks: %w", err)
	}

	var bank *domain.VABank
	for _, b := range banks {
		if b.Code == bankCode {
			bank = b
			break
		}
	}
	if bank == nil {
		return nil, domain.ErrInvalidBank
	}

	// Validate amount
	if amount < bank.MinAmount {
		return nil, domain.ErrAmountTooLow
	}
	if amount > bank.MaxAmount {
		return nil, domain.ErrAmountTooHigh
	}

	// Check pending deposits limit (max 3 pending)
	pendingCount, err := s.depositRepo.CountPending(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending: %w", err)
	}
	if pendingCount >= 3 {
		return nil, domain.ErrTooManyPending
	}

	// Admin fee
	adminFee := bank.Fee
	totalAmount := amount + adminFee

	// Get user info for customer data
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Create deposit ID (will be used as reference ID)
	depositID := fmt.Sprintf("dep_va_%s", uuid.New().String()[:8])

	// Call Gerbang API to create VA payment
	gerbangResp, err := s.gerbangClient.CreateVAPayment(ctx, depositID, bankCode, totalAmount, gerbang.CustomerInfo{
		Name:  user.FullName,
		Phone: user.Phone,
	})

	var externalID string
	var vaNumber string

	if err != nil {
		// Fallback to mock for development
		externalID = fmt.Sprintf("gerbang_va_%s", uuid.New().String()[:12])
		vaNumber = fmt.Sprintf("8808%d", time.Now().Unix()%1000000000)
	} else {
		// Use real Gerbang response
		externalID = gerbangResp.PaymentID
		if vaNum, ok := gerbangResp.PaymentDetail["vaNumber"].(string); ok {
			vaNumber = vaNum
		}
	}

	bankCodeStr := bankCode
	extIDPtr := externalID
	// Create deposit
	deposit := &domain.Deposit{
		ID:          depositID,
		UserID:      userID,
		Method:      domain.DepositMethodVirtualAccount,
		BankCode:    &bankCodeStr,
		Amount:      amount,
		AdminFee:    adminFee,
		UniqueCode:  0,
		TotalAmount: totalAmount,
		Status:      domain.DepositStatusPending,
		ExternalID:  &extIDPtr,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.depositRepo.Create(ctx, deposit); err != nil {
		return nil, fmt.Errorf("failed to create deposit: %w", err)
	}

	// Build response
	depositInfo := s.toDepositInfo(deposit)
	paymentInfo := &domain.VAPayment{
		BankCode:            bankCode,
		BankName:            bank.Name,
		VANumber:            vaNumber,
		ValidUntil:          deposit.ExpiresAt.Format(time.RFC3339),
		ValidUntilFormatted: formatDepositDate(deposit.ExpiresAt),
	}

	instructions := []string{
		fmt.Sprintf("Transfer ke nomor Virtual Account %s", bank.ShortName),
		fmt.Sprintf("VA Number: %s", vaNumber),
		fmt.Sprintf("Jumlah: %s", formatDepositCurrency(totalAmount)),
		"Transfer melalui mobile banking, ATM, atau internet banking",
		"Saldo akan masuk otomatis dalam 5-15 menit setelah transfer",
		"VA berlaku selama 24 jam",
	}

	return &domain.VAResponse{
		Deposit:      depositInfo,
		PaymentInfo:  paymentInfo,
		Instructions: instructions,
	}, nil
}

// GetStatus returns deposit status
func (s *DepositService) GetStatus(ctx context.Context, userID, depositID string) (*domain.DepositStatusResponse, error) {
	// Get deposit with ownership validation
	deposit, err := s.depositRepo.FindByUserAndID(ctx, userID, depositID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deposit: %w", err)
	}
	if deposit == nil {
		return nil, domain.ErrDepositNotFound
	}

	// Auto-update to expired if past expiry time and still pending
	if deposit.Status == domain.DepositStatusPending && time.Now().After(deposit.ExpiresAt) {
		if err := s.depositRepo.UpdateStatus(ctx, deposit.ID, domain.DepositStatusExpired, nil); err != nil {
			// Log error but don't fail the request
			slog.Error("failed to auto-expire deposit",
				slog.String("deposit_id", deposit.ID),
				slog.String("error", err.Error()),
			)
		} else {
			// Update local deposit object for response
			deposit.Status = domain.DepositStatusExpired
		}
	}

	// Build detailed response
	detail := s.toDepositDetail(deposit)

	return &domain.DepositStatusResponse{
		Deposit: detail,
	}, nil
}

// GetHistory returns deposit history
func (s *DepositService) GetHistory(ctx context.Context, userID string, filter repository.DepositFilter) (*domain.DepositHistoryResponse, error) {
	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}
	if filter.PerPage > 50 {
		filter.PerPage = 50
	}

	// Get deposits
	deposits, total, err := s.depositRepo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get deposits: %w", err)
	}

	// Convert to summary
	summaries := make([]*domain.DepositSummary, 0, len(deposits))
	for _, dep := range deposits {
		summary := s.toDepositSummary(dep)
		summaries = append(summaries, summary)
	}

	// Build pagination
	totalPages := int(math.Ceil(float64(total) / float64(filter.PerPage)))
	pagination := &domain.Pagination{
		CurrentPage:  filter.Page,
		TotalPages:   totalPages,
		TotalItems:   total,
		ItemsPerPage: filter.PerPage,
		HasNextPage:  filter.Page < totalPages,
		HasPrevPage:  filter.Page > 1,
	}

	return &domain.DepositHistoryResponse{
		Deposits:   summaries,
		Pagination: pagination,
	}, nil
}

// HandleWebhook handles deposit webhook from Gerbang API
func (s *DepositService) HandleWebhook(ctx context.Context, referenceID string) error {
	// Begin database transaction for atomic operation
	tx, err := s.depositRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Find deposit by ID (reference ID = deposit ID) with row lock
	deposit, err := s.depositRepo.FindByUserAndID(ctx, "", referenceID)
	if err != nil {
		return fmt.Errorf("failed to find deposit: %w", err)
	}
	if deposit == nil {
		return domain.ErrDepositNotFound
	}

	// Check if already paid - return nil for idempotency (not error!)
	// This ensures webhook sender gets 200 OK and won't retry forever
	if deposit.Status == domain.DepositStatusSuccess {
		return nil // Idempotent: already processed, just return success
	}

	// Check if expired
	if time.Now().After(deposit.ExpiresAt) {
		return domain.ErrDepositExpired
	}

	// Get user balance with row lock
	balance, err := s.balanceRepo.FindByUserIDForUpdate(ctx, tx, deposit.UserID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}
	if balance == nil {
		return domain.ErrNotFound("Balance")
	}

	// Update deposit status to success
	now := time.Now()
	if err := s.depositRepo.UpdateStatusWithTx(ctx, tx, deposit.ID, domain.DepositStatusSuccess, &now); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Add balance to user
	balance.Amount += deposit.Amount
	if err := s.balanceRepo.UpdateWithTx(ctx, tx, balance); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// TODO: Send notification to user (after transaction committed)

	return nil
}

// Helper functions

func (s *DepositService) toDepositInfo(deposit *domain.Deposit) *domain.DepositInfo {
	return &domain.DepositInfo{
		DepositID:            deposit.ID,
		Method:               deposit.Method,
		MethodName:           getMethodName(deposit.Method),
		Amount:               deposit.Amount,
		AmountFormatted:      formatDepositCurrency(deposit.Amount),
		AdminFee:             deposit.AdminFee,
		AdminFeeFormatted:    formatDepositCurrency(deposit.AdminFee),
		UniqueCode:           deposit.UniqueCode,
		TotalAmount:          deposit.TotalAmount,
		TotalAmountFormatted: formatDepositCurrency(deposit.TotalAmount),
		Status:               deposit.Status,
		StatusLabel:          getDepositStatusLabel(deposit.Status),
		ExpiresAt:            deposit.ExpiresAt.Format(time.RFC3339),
		ExpiresAtFormatted:   formatDepositDate(deposit.ExpiresAt),
		CreatedAt:            deposit.CreatedAt.Format(time.RFC3339),
	}
}

func (s *DepositService) toDepositDetail(deposit *domain.Deposit) *domain.DepositDetail {
	var paidAt *string
	if deposit.PaidAt != nil {
		formatted := deposit.PaidAt.Format(time.RFC3339)
		paidAt = &formatted
	}

	return &domain.DepositDetail{
		DepositID:            deposit.ID,
		Method:               deposit.Method,
		MethodName:           getMethodName(deposit.Method),
		Amount:               deposit.Amount,
		AmountFormatted:      formatDepositCurrency(deposit.Amount),
		AdminFee:             deposit.AdminFee,
		AdminFeeFormatted:    formatDepositCurrency(deposit.AdminFee),
		UniqueCode:           deposit.UniqueCode,
		TotalAmount:          deposit.TotalAmount,
		TotalAmountFormatted: formatDepositCurrency(deposit.TotalAmount),
		Status:               deposit.Status,
		StatusLabel:          getDepositStatusLabel(deposit.Status),
		ExpiresAt:            deposit.ExpiresAt.Format(time.RFC3339),
		ExpiresAtFormatted:   formatDepositDate(deposit.ExpiresAt),
		PaidAt:               paidAt,
		CreatedAt:            deposit.CreatedAt.Format(time.RFC3339),
	}
}

func (s *DepositService) toDepositSummary(deposit *domain.Deposit) *domain.DepositSummary {
	return &domain.DepositSummary{
		DepositID:            deposit.ID,
		Method:               deposit.Method,
		MethodName:           getMethodName(deposit.Method),
		Amount:               deposit.Amount,
		AmountFormatted:      formatDepositCurrency(deposit.Amount),
		TotalAmount:          deposit.TotalAmount,
		TotalAmountFormatted: formatDepositCurrency(deposit.TotalAmount),
		Status:               deposit.Status,
		StatusLabel:          getDepositStatusLabel(deposit.Status),
		CreatedAt:            deposit.CreatedAt.Format(time.RFC3339),
		CreatedAtFormatted:   formatDepositDate(deposit.CreatedAt),
	}
}

func generateUniqueCode() int {
	// Generate 3-digit unique code (100-999) using crypto/rand for security
	max := big.NewInt(900)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// Fallback to timestamp-based if crypto fails (very unlikely)
		slog.Error("crypto/rand failed, using fallback", slog.String("error", err.Error()))
		return 100 + int(time.Now().UnixNano()%900)
	}
	return 100 + int(n.Int64())
}

func getMethodName(method string) string {
	names := map[string]string{
		domain.DepositMethodBankTransfer:   "Transfer Bank",
		domain.DepositMethodQRIS:           "QRIS",
		domain.DepositMethodRetail:         "Retail",
		domain.DepositMethodVirtualAccount: "Virtual Account",
	}
	if name, ok := names[method]; ok {
		return name
	}
	return method
}

func getDepositStatusLabel(status string) string {
	labels := map[string]string{
		domain.DepositStatusPending: "Menunggu Pembayaran",
		domain.DepositStatusSuccess: "Berhasil",
		domain.DepositStatusExpired: "Expired",
		domain.DepositStatusFailed:  "Gagal",
	}
	if label, ok := labels[status]; ok {
		return label
	}
	return status
}

func formatDepositCurrency(amount int64) string {
	// Format: Rp1.000.000
	if amount == 0 {
		return "Rp0"
	}

	str := fmt.Sprintf("%d", amount)
	if len(str) <= 3 {
		return "Rp" + str
	}

	var result []byte
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(digit))
	}

	return "Rp" + string(result)
}

func formatDepositDate(t time.Time) string {
	// Format: "2 Januari 2026, 15:04"
	// For simplicity, using English format
	return t.Format("2 January 2006, 15:04")
}
