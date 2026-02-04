package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// DepositRepository defines the interface for deposit operations
type DepositRepository interface {
	// Deposit CRUD
	Create(ctx context.Context, deposit *domain.Deposit) error
	FindByID(ctx context.Context, id string) (*domain.Deposit, error)
	FindByUserAndID(ctx context.Context, userID, id string) (*domain.Deposit, error)
	FindByUserID(ctx context.Context, userID string, filter DepositFilter) ([]*domain.Deposit, int, error)
	UpdateStatus(ctx context.Context, id, status string, paidAt *time.Time) error
	UpdateStatusWithTx(ctx context.Context, tx *sqlx.Tx, id, status string, paidAt *time.Time) error
	CountPending(ctx context.Context, userID string) (int, error)
	FindByExternalID(ctx context.Context, externalID string) (*domain.Deposit, error)
	FindByExternalIDForUpdate(ctx context.Context, tx *sqlx.Tx, externalID string) (*domain.Deposit, error)

	// Transaction support
	BeginTx(ctx context.Context) (*sqlx.Tx, error)

	// Methods & Providers
	FindAllMethods(ctx context.Context) ([]*domain.DepositMethod, error)
	FindAllRetailProviders(ctx context.Context) ([]*domain.RetailProvider, error)
	FindAllVABanks(ctx context.Context) ([]*domain.VABank, error)
	FindAllCompanyBankAccounts(ctx context.Context) ([]*domain.CompanyBankAccount, error)
}

// DepositFilter represents filter options for deposit list
type DepositFilter struct {
	Status  string // pending, success, expired, failed, all
	Method  string // bank_transfer, qris, retail, virtual_account, all
	Page    int
	PerPage int
}

// depositRepository implements DepositRepository
type depositRepository struct {
	db *sqlx.DB
}

// Column constants for explicit SELECT
const (
	depositColumns = `id, user_id, method, provider_code, bank_code, amount, admin_fee, 
		unique_code, total_amount, status, payment_data, external_id, reference_number, 
		payer_name, payer_bank, payer_account, expires_at, paid_at, created_at, updated_at`

	depositMethodColumns = `code, name, description, icon, icon_url, fee_type, fee_amount, 
		min_amount, max_amount, schedule, position, status`

	retailProviderColumns = `code, name, icon, icon_url, fee, min_amount, max_amount, status`

	vaBankColumns = `code, name, short_name, icon, icon_url, fee, min_amount, max_amount, status`

	companyBankColumns = `bank_code, bank_name, bank_short_name, bank_icon, account_number, account_name, status`
)

// NewDepositRepository creates a new deposit repository
func NewDepositRepository(db *sqlx.DB) DepositRepository {
	return &depositRepository{db: db}
}

// Create creates a new deposit
func (r *depositRepository) Create(ctx context.Context, deposit *domain.Deposit) error {
	query := `
		INSERT INTO deposits (
			id, user_id, method, provider_code, bank_code, amount, admin_fee, 
			unique_code, total_amount, status, payment_data, external_id, 
			reference_number, payer_name, payer_bank, payer_account, expires_at, 
			paid_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		deposit.ID,
		deposit.UserID,
		deposit.Method,
		deposit.ProviderCode,
		deposit.BankCode,
		deposit.Amount,
		deposit.AdminFee,
		deposit.UniqueCode,
		deposit.TotalAmount,
		deposit.Status,
		deposit.PaymentData,
		deposit.ExternalID,
		deposit.ReferenceNumber,
		deposit.PayerName,
		deposit.PayerBank,
		deposit.PayerAccount,
		deposit.ExpiresAt,
		deposit.PaidAt,
		deposit.CreatedAt,
		deposit.UpdatedAt,
	)

	return err
}

// FindByID finds a deposit by ID
func (r *depositRepository) FindByID(ctx context.Context, id string) (*domain.Deposit, error) {
	query := `SELECT ` + depositColumns + ` FROM deposits WHERE id = $1`

	var deposit domain.Deposit
	err := r.db.GetContext(ctx, &deposit, query, id)
	if err != nil {
		return nil, err
	}

	return &deposit, nil
}

// FindByUserAndID finds a deposit by user ID and deposit ID (ownership validation)
func (r *depositRepository) FindByUserAndID(ctx context.Context, userID, id string) (*domain.Deposit, error) {
	query := `SELECT ` + depositColumns + ` FROM deposits WHERE id = $1 AND user_id = $2`

	var deposit domain.Deposit
	err := r.db.GetContext(ctx, &deposit, query, id, userID)
	if err != nil {
		return nil, err
	}

	return &deposit, nil
}

// FindByUserID finds deposits by user ID with filters and pagination
func (r *depositRepository) FindByUserID(ctx context.Context, userID string, filter DepositFilter) ([]*domain.Deposit, int, error) {
	// Build WHERE clause
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIdx := 2

	// Filter by status
	if filter.Status != "" && filter.Status != "all" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}

	// Filter by method
	if filter.Method != "" && filter.Method != "all" {
		whereClause += fmt.Sprintf(" AND method = $%d", argIdx)
		args = append(args, filter.Method)
		argIdx++
	}

	// Count total
	countQuery := `SELECT COUNT(*) FROM deposits ` + whereClause
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 50 {
		perPage = 50
	}

	offset := (page - 1) * perPage

	// Query with pagination
	query := `SELECT ` + depositColumns + ` FROM deposits ` + whereClause +
		` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argIdx) +
		` OFFSET $` + fmt.Sprintf("%d", argIdx+1)

	args = append(args, perPage, offset)

	var deposits []*domain.Deposit
	err = r.db.SelectContext(ctx, &deposits, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return deposits, total, nil
}

// UpdateStatus updates deposit status
func (r *depositRepository) UpdateStatus(ctx context.Context, id, status string, paidAt *time.Time) error {
	// In production, update database
	// For now, just return success
	return nil
}

// UpdateStatusWithTx updates deposit status within a transaction
func (r *depositRepository) UpdateStatusWithTx(ctx context.Context, tx *sqlx.Tx, id, status string, paidAt *time.Time) error {
	// TODO: Implement actual database update with transaction
	// For now, just mock success
	return nil
}

// CountPending counts pending deposits for a user
func (r *depositRepository) CountPending(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM deposits WHERE user_id = $1 AND status = $2`

	var count int
	err := r.db.GetContext(ctx, &count, query, userID, domain.DepositStatusPending)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// FindByExternalID finds a deposit by external ID
func (r *depositRepository) FindByExternalID(ctx context.Context, externalID string) (*domain.Deposit, error) {
	query := `SELECT ` + depositColumns + ` FROM deposits WHERE external_id = $1`

	var deposit domain.Deposit
	err := r.db.GetContext(ctx, &deposit, query, externalID)
	if err != nil {
		return nil, err
	}

	return &deposit, nil
}

// FindByExternalIDForUpdate finds a deposit by external ID with row lock for update
func (r *depositRepository) FindByExternalIDForUpdate(ctx context.Context, tx *sqlx.Tx, externalID string) (*domain.Deposit, error) {
	query := `SELECT ` + depositColumns + ` FROM deposits WHERE external_id = $1 FOR UPDATE`

	var deposit domain.Deposit
	err := tx.GetContext(ctx, &deposit, query, externalID)
	if err != nil {
		return nil, err
	}

	return &deposit, nil
}

// BeginTx begins a new database transaction
func (r *depositRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

// FindAllMethods returns all deposit methods
func (r *depositRepository) FindAllMethods(ctx context.Context) ([]*domain.DepositMethod, error) {
	query := `SELECT ` + depositMethodColumns + ` FROM deposit_methods WHERE status = 'active' ORDER BY position ASC`

	var methods []*domain.DepositMethod
	err := r.db.SelectContext(ctx, &methods, query)
	if err != nil {
		return nil, err
	}

	return methods, nil
}

// FindAllRetailProviders returns all retail providers
func (r *depositRepository) FindAllRetailProviders(ctx context.Context) ([]*domain.RetailProvider, error) {
	query := `SELECT ` + retailProviderColumns + ` FROM deposit_retail_providers WHERE status = 'active' ORDER BY position ASC`

	var providers []*domain.RetailProvider
	err := r.db.SelectContext(ctx, &providers, query)
	if err != nil {
		return nil, err
	}

	return providers, nil
}

// FindAllVABanks returns all VA banks
func (r *depositRepository) FindAllVABanks(ctx context.Context) ([]*domain.VABank, error) {
	query := `SELECT ` + vaBankColumns + ` FROM deposit_va_banks WHERE status = 'active' ORDER BY position ASC`

	var banks []*domain.VABank
	err := r.db.SelectContext(ctx, &banks, query)
	if err != nil {
		return nil, err
	}

	return banks, nil
}

// FindAllCompanyBankAccounts returns all company bank accounts
func (r *depositRepository) FindAllCompanyBankAccounts(ctx context.Context) ([]*domain.CompanyBankAccount, error) {
	query := `SELECT ` + companyBankColumns + ` FROM company_bank_accounts WHERE status = 'active' ORDER BY position ASC`

	var accounts []*domain.CompanyBankAccount
	err := r.db.SelectContext(ctx, &accounts, query)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}
