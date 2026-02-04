package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// PrepaidRepository defines the interface for prepaid data operations
type PrepaidRepository interface {
	// Inquiry operations
	CreateInquiry(ctx context.Context, inquiry *domain.PrepaidInquiry) error
	FindInquiryByID(ctx context.Context, id string) (*domain.PrepaidInquiry, error)

	// Order operations
	CreateOrder(ctx context.Context, order *domain.PrepaidOrder) error
	FindOrderByID(ctx context.Context, id string) (*domain.PrepaidOrder, error)
	FindOrderByUserAndID(ctx context.Context, userID, orderID string) (*domain.PrepaidOrder, error)
	UpdateOrderStatus(ctx context.Context, id, status string) error
	UpdateOrderStatusWithTx(ctx context.Context, tx *sqlx.Tx, id, status string) error

	// Transaction operations
	CreateTransaction(ctx context.Context, tx *domain.PrepaidTransaction) error
	CreateTransactionWithTx(ctx context.Context, dbtx *sqlx.Tx, tx *domain.PrepaidTransaction) error
	FindTransactionByID(ctx context.Context, id string) (*domain.PrepaidTransaction, error)
	FindTransactionByOrderID(ctx context.Context, orderID string) (*domain.PrepaidTransaction, error)
	UpdateTransactionStatus(ctx context.Context, id, status string) error

	// Transaction management
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

// prepaidRepository implements PrepaidRepository
type prepaidRepository struct {
	db *sqlx.DB
}

// NewPrepaidRepository creates a new prepaid repository
func NewPrepaidRepository(db *sqlx.DB) PrepaidRepository {
	return &prepaidRepository{db: db}
}

// Column constants for explicit SELECT
const prepaidInquiryColumns = `id, user_id, service_type, target, operator_id, operator_name, 
                               product_id, product_name, product_price, admin_fee, 
                               created_at, expires_at`

const prepaidOrderColumns = `id, user_id, inquiry_id, service_type, target, operator_id, operator_name, 
                             product_id, product_name, product_price, admin_fee, 
                             voucher_id, voucher_discount, total_payment, pin_verified, 
                             status, created_at, expires_at`

const prepaidTransactionColumns = `id, user_id, order_id, service_type, target, operator_id, operator_name, 
                                   product_id, product_name, product_price, admin_fee, voucher_discount, 
                                   total_payment, balance_before, balance_after, reference_number, 
                                   serial_number, external_id, status, failed_reason, 
                                   completed_at, created_at, updated_at`

// BeginTx begins a new database transaction
func (r *prepaidRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

// CreateInquiry creates a new inquiry record
func (r *prepaidRepository) CreateInquiry(ctx context.Context, inquiry *domain.PrepaidInquiry) error {
	query := `
		INSERT INTO prepaid_inquiries (id, user_id, service_type, target, target_valid, operator_id, customer_id, customer_name, expires_at, created_at)
		VALUES (:id, :user_id, :service_type, :target, :target_valid, :operator_id, :customer_id, :customer_name, :expires_at, :created_at)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			service_type = EXCLUDED.service_type,
			target = EXCLUDED.target,
			target_valid = EXCLUDED.target_valid,
			operator_id = EXCLUDED.operator_id,
			customer_id = EXCLUDED.customer_id,
			customer_name = EXCLUDED.customer_name,
			expires_at = EXCLUDED.expires_at
	`
	_, err := r.db.NamedExecContext(ctx, query, inquiry)
	return err
}

// FindInquiryByID finds an inquiry by ID
func (r *prepaidRepository) FindInquiryByID(ctx context.Context, id string) (*domain.PrepaidInquiry, error) {
	var inquiry domain.PrepaidInquiry
	query := fmt.Sprintf(`SELECT %s FROM prepaid_inquiries WHERE id = $1`, prepaidInquiryColumns)
	err := r.db.GetContext(ctx, &inquiry, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &inquiry, err
}

// CreateOrder creates a new order record
func (r *prepaidRepository) CreateOrder(ctx context.Context, order *domain.PrepaidOrder) error {
	query := `
		INSERT INTO prepaid_orders (
			id, user_id, inquiry_id, product_id, status, service_type, target,
			product_price, admin_fee, subtotal, total_discount, total_payment,
			expires_at, created_at, updated_at
		) VALUES (
			:id, :user_id, :inquiry_id, :product_id, :status, :service_type, :target,
			:product_price, :admin_fee, :subtotal, :total_discount, :total_payment,
			:expires_at, :created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, order)
	return err
}

// FindOrderByID finds an order by ID
func (r *prepaidRepository) FindOrderByID(ctx context.Context, id string) (*domain.PrepaidOrder, error) {
	var order domain.PrepaidOrder
	query := fmt.Sprintf(`SELECT %s FROM prepaid_orders WHERE id = $1`, prepaidOrderColumns)
	err := r.db.GetContext(ctx, &order, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &order, err
}

// FindOrderByUserAndID finds an order by user ID and order ID (with ownership validation)
func (r *prepaidRepository) FindOrderByUserAndID(ctx context.Context, userID, orderID string) (*domain.PrepaidOrder, error) {
	var order domain.PrepaidOrder
	query := fmt.Sprintf(`SELECT %s FROM prepaid_orders WHERE id = $1 AND user_id = $2`, prepaidOrderColumns)
	err := r.db.GetContext(ctx, &order, query, orderID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &order, err
}

// UpdateOrderStatus updates order status
func (r *prepaidRepository) UpdateOrderStatus(ctx context.Context, id, status string) error {
	query := `UPDATE prepaid_orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

// UpdateOrderStatusWithTx updates order status within a transaction
func (r *prepaidRepository) UpdateOrderStatusWithTx(ctx context.Context, tx *sqlx.Tx, id, status string) error {
	query := `UPDATE prepaid_orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, status, id)
	return err
}

// CreateTransaction creates a new transaction record
func (r *prepaidRepository) CreateTransaction(ctx context.Context, tx *domain.PrepaidTransaction) error {
	query := `
		INSERT INTO prepaid_transactions (
			id, user_id, order_id, status, service_type, target, product_id,
			total_payment, balance_before, balance_after,
			serial_number, reference_number, token, kwh,
			completed_at, created_at, updated_at
		) VALUES (
			:id, :user_id, :order_id, :status, :service_type, :target, :product_id,
			:total_payment, :balance_before, :balance_after,
			:serial_number, :reference_number, :token, :kwh,
			:completed_at, :created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, tx)
	return err
}

// CreateTransactionWithTx creates a new transaction record within a database transaction
func (r *prepaidRepository) CreateTransactionWithTx(ctx context.Context, dbtx *sqlx.Tx, tx *domain.PrepaidTransaction) error {
	query := `
		INSERT INTO prepaid_transactions (
			id, user_id, order_id, status, service_type, target, product_id,
			total_payment, balance_before, balance_after,
			serial_number, reference_number, token, kwh,
			completed_at, created_at, updated_at
		) VALUES (
			:id, :user_id, :order_id, :status, :service_type, :target, :product_id,
			:total_payment, :balance_before, :balance_after,
			:serial_number, :reference_number, :token, :kwh,
			:completed_at, :created_at, :updated_at
		)
	`
	_, err := dbtx.NamedExecContext(ctx, query, tx)
	return err
}

// FindTransactionByID finds a transaction by ID
func (r *prepaidRepository) FindTransactionByID(ctx context.Context, id string) (*domain.PrepaidTransaction, error) {
	var tx domain.PrepaidTransaction
	query := fmt.Sprintf(`SELECT %s FROM prepaid_transactions WHERE id = $1`, prepaidTransactionColumns)
	err := r.db.GetContext(ctx, &tx, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tx, err
}

// FindTransactionByOrderID finds a transaction by order ID (for idempotency check)
func (r *prepaidRepository) FindTransactionByOrderID(ctx context.Context, orderID string) (*domain.PrepaidTransaction, error) {
	var tx domain.PrepaidTransaction
	query := fmt.Sprintf(`SELECT %s FROM prepaid_transactions WHERE order_id = $1 LIMIT 1`, prepaidTransactionColumns)
	err := r.db.GetContext(ctx, &tx, query, orderID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tx, err
}

// UpdateTransactionStatus updates transaction status
func (r *prepaidRepository) UpdateTransactionStatus(ctx context.Context, id, status string) error {
	query := `UPDATE prepaid_transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
