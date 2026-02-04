package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// TransferRepository defines the interface for transfer data operations
type TransferRepository interface {
	// Inquiry operations
	CreateInquiry(ctx context.Context, inquiry *domain.TransferInquiry) error
	FindInquiryByID(ctx context.Context, id string) (*domain.TransferInquiry, error)
	FindInquiryByUserAndID(ctx context.Context, userID, inquiryID string) (*domain.TransferInquiry, error)

	// Transaction operations
	CreateTransaction(ctx context.Context, tx *domain.TransferTransaction) error
	CreateTransactionWithTx(ctx context.Context, dbtx *sqlx.Tx, tx *domain.TransferTransaction) error
	FindTransactionByID(ctx context.Context, id string) (*domain.TransferTransaction, error)
	FindTransactionByInquiryID(ctx context.Context, inquiryID string) (*domain.TransferTransaction, error)
	UpdateTransactionStatus(ctx context.Context, id, status string) error
	UpdateTransactionStatusWithTx(ctx context.Context, tx *sqlx.Tx, id, status string) error

	// Transaction management
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

// transferRepository implements TransferRepository
type transferRepository struct {
	db *sqlx.DB
}

// NewTransferRepository creates a new transfer repository
func NewTransferRepository(db *sqlx.DB) TransferRepository {
	return &transferRepository{db: db}
}

// Column constants for explicit SELECT
const transferInquiryColumns = `id, user_id, bank_code, bank_name, account_number, account_name, 
                                amount, admin_fee, total_payment, validated, 
                                created_at, expires_at`

const transferTransactionColumns = `id, user_id, inquiry_id, bank_code, bank_name, account_number, account_name, 
                                    amount, admin_fee, voucher_discount, total_payment, 
                                    balance_before, balance_after, reference_number, external_id, note, 
                                    status, failed_reason, completed_at, created_at, updated_at`

// BeginTx begins a new database transaction
func (r *transferRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

// CreateInquiry creates a new inquiry record
func (r *transferRepository) CreateInquiry(ctx context.Context, inquiry *domain.TransferInquiry) error {
	query := `
		INSERT INTO transfer_inquiries (
			id, user_id, bank_code, bank_name, account_number, account_name,
			amount, admin_fee, total_payment, expires_at, created_at
		) VALUES (
			:id, :user_id, :bank_code, :bank_name, :account_number, :account_name,
			:amount, :admin_fee, :total_payment, :expires_at, :created_at
		)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			bank_code = EXCLUDED.bank_code,
			bank_name = EXCLUDED.bank_name,
			account_number = EXCLUDED.account_number,
			account_name = EXCLUDED.account_name,
			amount = EXCLUDED.amount,
			admin_fee = EXCLUDED.admin_fee,
			total_payment = EXCLUDED.total_payment,
			expires_at = EXCLUDED.expires_at
	`
	_, err := r.db.NamedExecContext(ctx, query, inquiry)
	return err
}

// FindInquiryByID finds an inquiry by ID
func (r *transferRepository) FindInquiryByID(ctx context.Context, id string) (*domain.TransferInquiry, error) {
	var inquiry domain.TransferInquiry
	query := fmt.Sprintf(`SELECT %s FROM transfer_inquiries WHERE id = $1`, transferInquiryColumns)
	err := r.db.GetContext(ctx, &inquiry, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &inquiry, err
}

// FindInquiryByUserAndID finds an inquiry by user ID and inquiry ID (with ownership validation)
func (r *transferRepository) FindInquiryByUserAndID(ctx context.Context, userID, inquiryID string) (*domain.TransferInquiry, error) {
	var inquiry domain.TransferInquiry
	query := fmt.Sprintf(`SELECT %s FROM transfer_inquiries WHERE id = $1 AND user_id = $2`, transferInquiryColumns)
	err := r.db.GetContext(ctx, &inquiry, query, inquiryID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &inquiry, err
}

// CreateTransaction creates a new transaction record
func (r *transferRepository) CreateTransaction(ctx context.Context, tx *domain.TransferTransaction) error {
	query := `
		INSERT INTO transfer_transactions (
			id, user_id, inquiry_id, status, bank_code, bank_name,
			account_number, account_name, amount, admin_fee, total_payment, note,
			balance_before, balance_after, reference_number, completed_at, created_at, updated_at
		) VALUES (
			:id, :user_id, :inquiry_id, :status, :bank_code, :bank_name,
			:account_number, :account_name, :amount, :admin_fee, :total_payment, :note,
			:balance_before, :balance_after, :reference_number, :completed_at, :created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, tx)
	return err
}

// CreateTransactionWithTx creates a new transaction record within a database transaction
func (r *transferRepository) CreateTransactionWithTx(ctx context.Context, dbtx *sqlx.Tx, tx *domain.TransferTransaction) error {
	query := `
		INSERT INTO transfer_transactions (
			id, user_id, inquiry_id, status, bank_code, bank_name,
			account_number, account_name, amount, admin_fee, total_payment, note,
			balance_before, balance_after, reference_number, completed_at, created_at, updated_at
		) VALUES (
			:id, :user_id, :inquiry_id, :status, :bank_code, :bank_name,
			:account_number, :account_name, :amount, :admin_fee, :total_payment, :note,
			:balance_before, :balance_after, :reference_number, :completed_at, :created_at, :updated_at
		)
	`
	_, err := dbtx.NamedExecContext(ctx, query, tx)
	return err
}

// FindTransactionByID finds a transaction by ID
func (r *transferRepository) FindTransactionByID(ctx context.Context, id string) (*domain.TransferTransaction, error) {
	var tx domain.TransferTransaction
	query := fmt.Sprintf(`SELECT %s FROM transfer_transactions WHERE id = $1`, transferTransactionColumns)
	err := r.db.GetContext(ctx, &tx, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tx, err
}

// FindTransactionByInquiryID finds a transaction by inquiry ID (for idempotency check)
func (r *transferRepository) FindTransactionByInquiryID(ctx context.Context, inquiryID string) (*domain.TransferTransaction, error) {
	var tx domain.TransferTransaction
	query := fmt.Sprintf(`SELECT %s FROM transfer_transactions WHERE inquiry_id = $1 LIMIT 1`, transferTransactionColumns)
	err := r.db.GetContext(ctx, &tx, query, inquiryID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tx, err
}

// UpdateTransactionStatus updates transaction status
func (r *transferRepository) UpdateTransactionStatus(ctx context.Context, id, status string) error {
	query := `UPDATE transfer_transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

// UpdateTransactionStatusWithTx updates transaction status within a transaction
func (r *transferRepository) UpdateTransactionStatusWithTx(ctx context.Context, tx *sqlx.Tx, id, status string) error {
	query := `UPDATE transfer_transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, status, id)
	return err
}
