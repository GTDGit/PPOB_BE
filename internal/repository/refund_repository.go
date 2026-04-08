package repository

import (
	"context"
	"database/sql"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/jmoiron/sqlx"
)

type RefundRepository interface {
	Create(ctx context.Context, refund *domain.Refund) error
	CreateWithTx(ctx context.Context, tx *sqlx.Tx, refund *domain.Refund) error
	FindByID(ctx context.Context, identifier string) (*domain.Refund, error)
	FindBySourceTransactionID(ctx context.Context, sourceTransactionID string) (*domain.Refund, error)
}

type refundRepository struct {
	db *sqlx.DB
}

const refundColumns = `id, public_id, source_transaction_id, source_type, user_id, amount, reason, status, refunded_at, created_at, updated_at`

func NewRefundRepository(db *sqlx.DB) RefundRepository {
	return &refundRepository{db: db}
}

func (r *refundRepository) Create(ctx context.Context, refund *domain.Refund) error {
	return r.create(ctx, r.db, refund)
}

func (r *refundRepository) CreateWithTx(ctx context.Context, tx *sqlx.Tx, refund *domain.Refund) error {
	return r.create(ctx, tx, refund)
}

func (r *refundRepository) FindByID(ctx context.Context, identifier string) (*domain.Refund, error) {
	var refund domain.Refund
	query := `SELECT ` + refundColumns + ` FROM refunds WHERE id = $1 OR public_id = $1 ORDER BY CASE WHEN id = $1 THEN 0 ELSE 1 END LIMIT 1`
	if err := r.db.GetContext(ctx, &refund, query, identifier); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepository) FindBySourceTransactionID(ctx context.Context, sourceTransactionID string) (*domain.Refund, error) {
	var refund domain.Refund
	query := `SELECT ` + refundColumns + ` FROM refunds WHERE source_transaction_id = $1 LIMIT 1`
	if err := r.db.GetContext(ctx, &refund, query, sourceTransactionID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &refund, nil
}

func (r *refundRepository) create(ctx context.Context, exec sqlx.ExtContext, refund *domain.Refund) error {
	if refund.ID == "" {
		refund.ID = NewUUID()
	}
	if refund.PublicID == "" {
		var (
			publicID string
			err      error
		)
		switch typedExec := exec.(type) {
		case *sqlx.Tx:
			publicID, err = GeneratePublicRefundID(ctx, typedExec)
		case *sqlx.DB:
			publicID, err = GeneratePublicRefundID(ctx, typedExec)
		default:
			err = sql.ErrConnDone
		}
		if err != nil {
			return err
		}
		refund.PublicID = publicID
	}

	query := `
		INSERT INTO refunds (
			id, public_id, source_transaction_id, source_type, user_id, amount, reason, status, refunded_at, created_at, updated_at
		) VALUES (
			:id, :public_id, :source_transaction_id, :source_type, :user_id, :amount, :reason, :status, :refunded_at, :created_at, :updated_at
		)
	`
	_, err := sqlx.NamedExecContext(ctx, exec, query, refund)
	return err
}
