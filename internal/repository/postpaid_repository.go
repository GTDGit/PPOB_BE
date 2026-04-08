package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/jmoiron/sqlx"
)

// PostpaidRepository handles postpaid data access
type PostpaidRepository interface {
	CreateInquiry(ctx context.Context, inquiry *domain.PostpaidInquiry) error
	FindInquiryByID(ctx context.Context, id string) (*domain.PostpaidInquiry, error)
	FindInquiryByUserAndID(ctx context.Context, userID, id string) (*domain.PostpaidInquiry, error)

	CreateTransaction(ctx context.Context, tx *domain.PostpaidTransaction) error
	CreateTransactionWithTx(ctx context.Context, dbtx *sqlx.Tx, tx *domain.PostpaidTransaction) error
	FindTransactionByID(ctx context.Context, id string) (*domain.PostpaidTransaction, error)
	FindTransactionByInquiryID(ctx context.Context, inquiryID string) (*domain.PostpaidTransaction, error)
	UpdateTransactionStatus(ctx context.Context, id, status string) error
	UpdateTransactionStatusWithTx(ctx context.Context, tx *sqlx.Tx, id, status string) error

	// Transaction methods for atomic operations
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

type postpaidRepository struct {
	db *sqlx.DB
}

const postpaidInquiryColumns = `id, user_id, service_type, target, provider_id, customer_id,
	customer_name, period, bill_amount, admin_fee, penalty, total_payment, has_bill,
	external_id, expires_at, created_at`

const postpaidTransactionColumns = `id, public_id, user_id, inquiry_id, service_type, target, provider_id,
	customer_id, customer_name, period, bill_amount, admin_fee, penalty, voucher_discount,
	total_payment, balance_before, balance_after, reference_number, serial_number, external_id,
	status, failed_reason, completed_at, created_at, updated_at`

// NewPostpaidRepository creates a new postpaid repository
func NewPostpaidRepository(db *sqlx.DB) PostpaidRepository {
	return &postpaidRepository{db: db}
}

// CreateInquiry creates a new inquiry
func (r *postpaidRepository) CreateInquiry(ctx context.Context, inquiry *domain.PostpaidInquiry) error {
	query := `
		INSERT INTO postpaid_inquiries (
			id, user_id, service_type, target, provider_id, customer_id, customer_name,
			period, bill_amount, admin_fee, penalty, total_payment, has_bill,
			external_id, created_at, expires_at
		) VALUES (
			:id, :user_id, :service_type, :target, :provider_id, :customer_id, :customer_name,
			:period, :bill_amount, :admin_fee, :penalty, :total_payment, :has_bill,
			:external_id, :created_at, :expires_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, inquiry)
	return err
}

// FindInquiryByID finds inquiry by ID
func (r *postpaidRepository) FindInquiryByID(ctx context.Context, id string) (*domain.PostpaidInquiry, error) {
	query := `SELECT ` + postpaidInquiryColumns + ` FROM postpaid_inquiries WHERE id = $1 LIMIT 1`

	var inquiry domain.PostpaidInquiry
	if err := r.db.GetContext(ctx, &inquiry, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &inquiry, nil
}

// FindInquiryByUserAndID finds inquiry by user and ID (ownership check)
func (r *postpaidRepository) FindInquiryByUserAndID(ctx context.Context, userID, id string) (*domain.PostpaidInquiry, error) {
	query := `SELECT ` + postpaidInquiryColumns + ` FROM postpaid_inquiries WHERE id = $1 AND user_id = $2 LIMIT 1`

	var inquiry domain.PostpaidInquiry
	if err := r.db.GetContext(ctx, &inquiry, query, id, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &inquiry, nil
}

// CreateTransaction creates a new transaction
func (r *postpaidRepository) CreateTransaction(ctx context.Context, tx *domain.PostpaidTransaction) error {
	return r.createTransaction(ctx, r.db, tx)
}

func (r *postpaidRepository) CreateTransactionWithTx(ctx context.Context, dbtx *sqlx.Tx, tx *domain.PostpaidTransaction) error {
	return r.createTransaction(ctx, dbtx, tx)
}

func (r *postpaidRepository) createTransaction(ctx context.Context, exec sqlx.ExtContext, tx *domain.PostpaidTransaction) error {
	if tx.ID == "" {
		tx.ID = NewUUID()
	}
	if tx.PublicID == nil || *tx.PublicID == "" {
		var (
			publicID string
			err      error
		)
		switch typedExec := exec.(type) {
		case *sqlx.DB:
			publicID, err = GeneratePublicTransactionID(ctx, typedExec)
		case *sqlx.Tx:
			publicID, err = GeneratePublicTransactionID(ctx, typedExec)
		default:
			err = fmt.Errorf("unsupported postpaid transaction executor")
		}
		if err != nil {
			return err
		}
		tx.PublicID = &publicID
	}

	query := `
		INSERT INTO postpaid_transactions (
			id, public_id, user_id, inquiry_id, service_type, target, provider_id, customer_id,
			customer_name, period, bill_amount, admin_fee, penalty, voucher_discount,
			total_payment, balance_before, balance_after, reference_number, serial_number,
			external_id, status, failed_reason, completed_at, created_at, updated_at
		) VALUES (
			:id, :public_id, :user_id, :inquiry_id, :service_type, :target, :provider_id, :customer_id,
			:customer_name, :period, :bill_amount, :admin_fee, :penalty, :voucher_discount,
			:total_payment, :balance_before, :balance_after, :reference_number, :serial_number,
			:external_id, :status, :failed_reason, :completed_at, :created_at, :updated_at
		)
	`
	_, err := sqlx.NamedExecContext(ctx, exec, query, tx)
	return err
}

// FindTransactionByID finds transaction by ID
func (r *postpaidRepository) FindTransactionByID(ctx context.Context, id string) (*domain.PostpaidTransaction, error) {
	query := `SELECT ` + postpaidTransactionColumns + ` FROM postpaid_transactions WHERE id = $1 OR public_id = $1 ORDER BY CASE WHEN id = $1 THEN 0 ELSE 1 END LIMIT 1`

	var tx domain.PostpaidTransaction
	if err := r.db.GetContext(ctx, &tx, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &tx, nil
}

// FindTransactionByInquiryID finds transaction by inquiry ID (idempotency check)
func (r *postpaidRepository) FindTransactionByInquiryID(ctx context.Context, inquiryID string) (*domain.PostpaidTransaction, error) {
	query := `SELECT ` + postpaidTransactionColumns + ` FROM postpaid_transactions WHERE inquiry_id = $1 LIMIT 1`

	var tx domain.PostpaidTransaction
	if err := r.db.GetContext(ctx, &tx, query, inquiryID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &tx, nil
}

// UpdateTransactionStatus updates transaction status
func (r *postpaidRepository) UpdateTransactionStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE postpaid_transactions SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

func (r *postpaidRepository) UpdateTransactionStatusWithTx(ctx context.Context, tx *sqlx.Tx, id, status string) error {
	_, err := tx.ExecContext(ctx, `UPDATE postpaid_transactions SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

// BeginTx begins a database transaction
func (r *postpaidRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}
