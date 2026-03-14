package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/jmoiron/sqlx"
)

// HistoryRepository defines the interface for transaction history operations.
type HistoryRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Transaction, error)
	FindByUserAndID(ctx context.Context, userID, id string) (*domain.Transaction, error)
	FindByUserID(ctx context.Context, userID string, filter TransactionFilter) ([]*domain.Transaction, int, error)
	Create(ctx context.Context, tx *domain.Transaction) error
	CreateWithTx(ctx context.Context, dbtx *sqlx.Tx, tx *domain.Transaction) error
	UpdateStatus(ctx context.Context, id, status string, receiptData *string) error
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

// TransactionFilter represents filter options for transaction list.
type TransactionFilter struct {
	Type        string
	ServiceType string
	Status      string
	StartDate   *time.Time
	EndDate     *time.Time
	Search      string
	Page        int
	PerPage     int
}

type historyRepository struct {
	db *sqlx.DB
}

const historySelectColumns = `
	id,
	user_id,
	order_id,
	inquiry_id,
	type,
	service_type,
	target,
	product_id,
	COALESCE(product_name, '') AS product_name,
	price AS amount,
	admin_fee,
	voucher_discount AS discount,
	total_payment,
	status,
	NULL::VARCHAR(50) AS provider_ref,
	serial_number,
	token,
	NULL::TEXT AS receipt_data,
	status_message AS failure_reason,
	completed_at,
	created_at,
	updated_at
`

// NewHistoryRepository creates a new history repository.
func NewHistoryRepository(db *sqlx.DB) HistoryRepository {
	return &historyRepository{db: db}
}

// BeginTx begins a new transaction for write operations.
func (r *historyRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

// FindByID finds a transaction by ID.
func (r *historyRepository) FindByID(ctx context.Context, id string) (*domain.Transaction, error) {
	query := `SELECT ` + historySelectColumns + ` FROM transactions WHERE id = $1`

	var tx domain.Transaction
	if err := r.db.GetContext(ctx, &tx, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &tx, nil
}

// FindByUserAndID finds a transaction by user ID and transaction ID.
func (r *historyRepository) FindByUserAndID(ctx context.Context, userID, id string) (*domain.Transaction, error) {
	query := `SELECT ` + historySelectColumns + ` FROM transactions WHERE id = $1 AND user_id = $2`

	var tx domain.Transaction
	if err := r.db.GetContext(ctx, &tx, query, id, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &tx, nil
}

// FindByUserID finds transactions by user ID with filters and pagination.
func (r *historyRepository) FindByUserID(ctx context.Context, userID string, filter TransactionFilter) ([]*domain.Transaction, int, error) {
	whereClauses := []string{"user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	if filter.Type != "" && filter.Type != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, filter.Type)
		argIdx++
	}

	if filter.ServiceType != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("service_type = $%d", argIdx))
		args = append(args, filter.ServiceType)
		argIdx++
	}

	if filter.Status != "" && filter.Status != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.StartDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at >= $%d", argIdx))
		args = append(args, *filter.StartDate)
		argIdx++
	}

	if filter.EndDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_at <= $%d", argIdx))
		args = append(args, *filter.EndDate)
		argIdx++
	}

	if filter.Search != "" {
		search := "%" + strings.ToLower(strings.TrimSpace(filter.Search)) + "%"
		whereClauses = append(
			whereClauses,
			fmt.Sprintf("(LOWER(COALESCE(product_name, '')) LIKE $%d OR LOWER(target) LIKE $%d)", argIdx, argIdx),
		)
		args = append(args, search)
		argIdx++
	}

	whereSQL := "WHERE " + strings.Join(whereClauses, " AND ")

	countQuery := `SELECT COUNT(*) FROM transactions ` + whereSQL
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

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
	query := `SELECT ` + historySelectColumns + ` FROM transactions ` + whereSQL +
		fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)

	queryArgs := append(args, perPage, offset)
	var transactions []*domain.Transaction
	if err := r.db.SelectContext(ctx, &transactions, query, queryArgs...); err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// Create inserts a transaction record.
func (r *historyRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	return r.create(ctx, r.db, tx)
}

// CreateWithTx inserts a transaction record within an existing transaction.
func (r *historyRepository) CreateWithTx(ctx context.Context, dbtx *sqlx.Tx, tx *domain.Transaction) error {
	return r.create(ctx, dbtx, tx)
}

// UpdateStatus updates transaction status.
func (r *historyRepository) UpdateStatus(ctx context.Context, id, status string, receiptData *string) error {
	var statusMessage *string
	if receiptData != nil && *receiptData != "" {
		statusMessage = receiptData
	}

	query := `
		UPDATE transactions
		SET status = $1, status_message = COALESCE($2, status_message), updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, status, statusMessage, id)
	return err
}

func (r *historyRepository) create(ctx context.Context, exec sqlx.ExtContext, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, user_id, type, service_type, inquiry_id, order_id, target, product_id,
			product_name, price, admin_fee, voucher_discount, total_payment, status,
			status_message, reference_number, serial_number, token, completed_at, created_at, updated_at
		) VALUES (
			:id, :user_id, :type, :service_type, :inquiry_id, :order_id, :target, :product_id,
			:product_name, :amount, :admin_fee, :discount, :total_payment, :status,
			:failure_reason, :provider_ref, :serial_number, :token, :completed_at, :created_at, :updated_at
		)
	`

	_, err := sqlx.NamedExecContext(ctx, exec, query, tx)
	return err
}
