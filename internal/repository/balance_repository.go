package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

type BalanceRepository interface {
	FindByUserID(ctx context.Context, userID string) (*domain.Balance, error)
	FindByUserIDForUpdate(ctx context.Context, tx *sqlx.Tx, userID string) (*domain.Balance, error)
	Create(ctx context.Context, balance *domain.Balance) error
	Update(ctx context.Context, balance *domain.Balance) error
	UpdateWithTx(ctx context.Context, tx *sqlx.Tx, balance *domain.Balance) error
	DeductBalance(ctx context.Context, tx *sqlx.Tx, userID string, amount int64) (*domain.Balance, error)
}

type balanceRepository struct {
	db *sqlx.DB
}

func NewBalanceRepository(db *sqlx.DB) BalanceRepository {
	return &balanceRepository{db: db}
}

// Column constants for explicit SELECT
const balanceColumns = `id, user_id, amount, pending_amount, points, 
                        points_expires_at, updated_at`

func (r *balanceRepository) FindByUserID(ctx context.Context, userID string) (*domain.Balance, error) {
	var balance domain.Balance
	query := fmt.Sprintf(`SELECT %s FROM balances WHERE user_id = $1`, balanceColumns)
	err := r.db.GetContext(ctx, &balance, query, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

func (r *balanceRepository) Create(ctx context.Context, balance *domain.Balance) error {
	if balance.ID == "" {
		balance.ID = "bal_" + uuid.New().String()[:8]
	}
	balance.UpdatedAt = time.Now()

	query := `
		INSERT INTO balances (id, user_id, amount, pending_amount, points, points_expires_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		balance.ID, balance.UserID, balance.Amount, balance.PendingAmount,
		balance.Points, balance.PointsExpiresAt, balance.UpdatedAt,
	)
	return err
}

func (r *balanceRepository) Update(ctx context.Context, balance *domain.Balance) error {
	balance.UpdatedAt = time.Now()
	query := `
		UPDATE balances SET
			amount = $2, pending_amount = $3, points = $4, points_expires_at = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		balance.ID, balance.Amount, balance.PendingAmount,
		balance.Points, balance.PointsExpiresAt, balance.UpdatedAt,
	)
	return err
}

// FindByUserIDForUpdate gets balance with row lock (SELECT FOR UPDATE)
func (r *balanceRepository) FindByUserIDForUpdate(ctx context.Context, tx *sqlx.Tx, userID string) (*domain.Balance, error) {
	var balance domain.Balance
	query := fmt.Sprintf(`SELECT %s FROM balances WHERE user_id = $1 FOR UPDATE`, balanceColumns)
	err := tx.GetContext(ctx, &balance, query, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// UpdateWithTx updates balance within a transaction
func (r *balanceRepository) UpdateWithTx(ctx context.Context, tx *sqlx.Tx, balance *domain.Balance) error {
	balance.UpdatedAt = time.Now()
	query := `
		UPDATE balances SET
			amount = $2, pending_amount = $3, points = $4, points_expires_at = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := tx.ExecContext(ctx, query,
		balance.ID, balance.Amount, balance.PendingAmount,
		balance.Points, balance.PointsExpiresAt, balance.UpdatedAt,
	)
	return err
}

// DeductBalance deducts balance atomically within a transaction
// NOTE: Currently not used, service uses FindByUserIDForUpdate + UpdateWithTx directly.
// Kept for potential future use as a convenience method.
func (r *balanceRepository) DeductBalance(ctx context.Context, tx *sqlx.Tx, userID string, amount int64) (*domain.Balance, error) {
	// Lock and get balance
	balance, err := r.FindByUserIDForUpdate(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	if balance == nil {
		return nil, sql.ErrNoRows
	}

	// Check sufficiency
	if balance.Amount < amount {
		return nil, domain.ErrInsufficientBalance
	}

	// Store balance before deduction
	balanceBefore := balance.Amount

	// Deduct balance
	balance.Amount -= amount
	balance.UpdatedAt = time.Now()

	query := `
		UPDATE balances SET
			amount = $2, pending_amount = $3, points = $4, points_expires_at = $5, updated_at = $6
		WHERE id = $1
	`
	_, err = tx.ExecContext(ctx, query,
		balance.ID, balance.Amount, balance.PendingAmount,
		balance.Points, balance.PointsExpiresAt, balance.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Return balance with before amount set for transaction record
	result := *balance
	result.Amount = balanceBefore
	return &result, nil
}
