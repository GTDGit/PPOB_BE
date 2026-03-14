package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/google/uuid"
)

// SandboxService provides safe dummy flows for app testing.
type SandboxService struct {
	historyRepo     repository.HistoryRepository
	balanceRepo     repository.BalanceRepository
	depositRepo     repository.DepositRepository
	notificationSvc *NotificationService
}

// SandboxCheckoutRequest represents a dummy checkout request.
type SandboxCheckoutRequest struct {
	UserID          string
	TransactionType string
	ServiceType     string
	ProductName     string
	Target          string
	Description     string
	Amount          int64
	AdminFee        int64
}

// NewSandboxService creates a new sandbox service.
func NewSandboxService(
	historyRepo repository.HistoryRepository,
	balanceRepo repository.BalanceRepository,
	depositRepo repository.DepositRepository,
	notificationSvc *NotificationService,
) *SandboxService {
	return &SandboxService{
		historyRepo:     historyRepo,
		balanceRepo:     balanceRepo,
		depositRepo:     depositRepo,
		notificationSvc: notificationSvc,
	}
}

// Checkout creates a dummy successful transaction and deducts user balance.
func (s *SandboxService) Checkout(ctx context.Context, req SandboxCheckoutRequest) (*domain.SandboxCheckoutResponse, error) {
	if req.Amount <= 0 {
		return nil, domain.ErrValidationFailed("Nominal transaksi wajib lebih dari 0")
	}

	transactionType, serviceType := normalizeSandboxTypes(req.TransactionType, req.ServiceType, req.ProductName)
	totalPayment := req.Amount + req.AdminFee

	dbtx, err := s.historyRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin sandbox transaction: %w", err)
	}
	defer dbtx.Rollback()

	balance, err := s.balanceRepo.FindByUserIDForUpdate(ctx, dbtx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load balance: %w", err)
	}
	if balance == nil {
		return nil, domain.ErrNotFound("Balance")
	}
	if balance.Amount < totalPayment {
		return nil, domain.ErrInsufficientBalance
	}

	balanceBefore := balance.Amount
	balance.Amount -= totalPayment
	if err := s.balanceRepo.UpdateWithTx(ctx, dbtx, balance); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	now := time.Now()
	transactionID := "trx_sb_" + uuid.New().String()[:8]
	referenceNumber := "SBX" + strings.ToUpper(uuid.New().String()[:10])
	transaction := &domain.Transaction{
		ID:           transactionID,
		UserID:       req.UserID,
		Type:         transactionType,
		ServiceType:  serviceType,
		Target:       req.Target,
		ProductName:  req.ProductName,
		Amount:       req.Amount,
		AdminFee:     req.AdminFee,
		Discount:     0,
		TotalPayment: totalPayment,
		Status:       domain.TransactionStatusSuccess,
		ProviderRef:  &referenceNumber,
		CompletedAt:  &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if strings.EqualFold(serviceType, domain.ServicePLNPrepaid) {
		token := generateSandboxPLNToken()
		transaction.Token = &token
	}

	if err := s.historyRepo.CreateWithTx(ctx, dbtx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create sandbox transaction: %w", err)
	}

	if err := dbtx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit sandbox transaction: %w", err)
	}

	if err := s.createNotification(ctx, req.UserID, domain.NotificationCategoryTransaction,
		"Transaksi Berhasil",
		fmt.Sprintf("%s berhasil diproses. Total pembayaran %s.", req.ProductName, formatSandboxCurrency(totalPayment)),
		map[string]string{"transactionId": transactionID},
	); err != nil {
		return nil, err
	}

	return &domain.SandboxCheckoutResponse{
		TransactionID: transactionID,
		Status:        domain.TransactionStatusSuccess,
		Title:         "Transaksi dummy berhasil",
		Message:       fmt.Sprintf("%s berhasil diproses untuk pengujian.", req.ProductName),
		Balance: &domain.SandboxBalanceDelta{
			Before:          balanceBefore,
			BeforeFormatted: formatSandboxCurrency(balanceBefore),
			After:           balance.Amount,
			AfterFormatted:  formatSandboxCurrency(balance.Amount),
		},
	}, nil
}

// CompleteDeposit simulates that a pending bank transfer deposit has been paid.
func (s *SandboxService) CompleteDeposit(ctx context.Context, userID, depositID string) (*domain.SandboxDepositCompleteResponse, error) {
	beforeBalance, err := s.balanceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load balance before deposit completion: %w", err)
	}
	if beforeBalance == nil {
		return nil, domain.ErrNotFound("Balance")
	}

	deposit, err := s.depositRepo.FindByUserAndID(ctx, userID, depositID)
	if err != nil {
		return nil, fmt.Errorf("failed to load deposit: %w", err)
	}
	if deposit == nil {
		return nil, domain.ErrDepositNotFound
	}

	if deposit.Status != domain.DepositStatusSuccess {
		if err := s.completeDepositWithBalance(ctx, depositID, deposit.UserID); err != nil {
			return nil, err
		}
	}

	afterBalance, err := s.balanceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load balance after deposit completion: %w", err)
	}
	if afterBalance == nil {
		return nil, domain.ErrNotFound("Balance")
	}

	if err := s.createNotification(ctx, userID, domain.NotificationCategoryDeposit,
		"Deposit Berhasil",
		fmt.Sprintf("Top up saldo %s berhasil masuk ke akun Anda.", formatSandboxCurrency(deposit.Amount)),
		map[string]string{"depositId": depositID},
	); err != nil {
		return nil, err
	}

	return &domain.SandboxDepositCompleteResponse{
		DepositID: depositID,
		Status:    domain.DepositStatusSuccess,
		Message:   "Deposit dummy berhasil diselesaikan.",
		Balance: &domain.SandboxBalanceDelta{
			Before:          beforeBalance.Amount,
			BeforeFormatted: formatSandboxCurrency(beforeBalance.Amount),
			After:           afterBalance.Amount,
			AfterFormatted:  formatSandboxCurrency(afterBalance.Amount),
		},
	}, nil
}

func (s *SandboxService) completeDepositWithBalance(ctx context.Context, depositID, userID string) error {
	tx, err := s.depositRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin deposit completion transaction: %w", err)
	}
	defer tx.Rollback()

	deposit, err := s.depositRepo.FindByUserAndID(ctx, userID, depositID)
	if err != nil {
		return fmt.Errorf("failed to find deposit: %w", err)
	}
	if deposit == nil {
		return domain.ErrDepositNotFound
	}
	if deposit.Status == domain.DepositStatusSuccess {
		return nil
	}
	if time.Now().After(deposit.ExpiresAt) {
		return domain.ErrDepositExpired
	}

	balance, err := s.balanceRepo.FindByUserIDForUpdate(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("failed to lock balance: %w", err)
	}
	if balance == nil {
		return domain.ErrNotFound("Balance")
	}

	now := time.Now()
	if err := s.depositRepo.UpdateStatusWithTx(ctx, tx, deposit.ID, domain.DepositStatusSuccess, &now); err != nil {
		return fmt.Errorf("failed to update deposit status: %w", err)
	}

	balance.Amount += deposit.Amount
	if err := s.balanceRepo.UpdateWithTx(ctx, tx, balance); err != nil {
		return fmt.Errorf("failed to credit balance: %w", err)
	}

	return tx.Commit()
}

func (s *SandboxService) createNotification(ctx context.Context, userID, category, title, body string, metadata map[string]string) error {
	if s.notificationSvc == nil {
		return nil
	}
	return s.notificationSvc.CreateSystemNotification(ctx, userID, category, title, body, metadata)
}

func normalizeSandboxTypes(transactionType, serviceType, productName string) (string, string) {
	txType := strings.TrimSpace(strings.ToLower(transactionType))
	svcType := strings.TrimSpace(strings.ToLower(serviceType))
	name := strings.ToLower(productName)

	if svcType == "" {
		switch {
		case strings.Contains(name, "pulsa pasca"), strings.Contains(name, "tagihan"), strings.Contains(name, "bpjs"), strings.Contains(name, "pdam"), strings.Contains(name, "telkom"), strings.Contains(name, "gas"), strings.Contains(name, "tv kabel"), strings.Contains(name, "pbb"):
			svcType = domain.ServicePDAM
		case strings.Contains(name, "token pln"):
			svcType = domain.ServicePLNPrepaid
		case strings.Contains(name, "paket data"):
			svcType = domain.ServiceData
		case strings.Contains(name, "pulsa"):
			svcType = domain.ServicePulsa
		case strings.Contains(name, "transfer bank"):
			svcType = "transfer_bank"
		case strings.Contains(name, "saldo "):
			svcType = domain.ServiceEwallet
		default:
			svcType = "sandbox_generic"
		}
	}

	if txType == "" {
		switch svcType {
		case "transfer_bank":
			txType = domain.TransactionTypeTransfer
		case domain.ServicePDAM, domain.ServiceBPJS, domain.ServiceTelkom, domain.ServiceTVCable, domain.ServicePBB, domain.ServicePLNPostpaid:
			txType = domain.TransactionTypePostpaid
		default:
			txType = domain.TransactionTypePrepaid
		}
	}

	return txType, svcType
}

func generateSandboxPLNToken() string {
	parts := make([]string, 5)
	for i := range parts {
		parts[i] = fmt.Sprintf("%04d", 1111+(i*1234))
	}
	return strings.Join(parts, "-")
}

func formatSandboxCurrency(amount int64) string {
	if amount == 0 {
		return "Rp0"
	}

	raw := fmt.Sprintf("%d", amount)
	var builder strings.Builder
	builder.WriteString("Rp")
	for i, r := range raw {
		if i > 0 && (len(raw)-i)%3 == 0 {
			builder.WriteRune('.')
		}
		builder.WriteRune(r)
	}
	return builder.String()
}
