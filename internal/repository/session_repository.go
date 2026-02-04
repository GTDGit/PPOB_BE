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

type SessionRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Session, error)
	FindByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) (*domain.Session, error)
	FindByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*domain.Session, error)
	Create(ctx context.Context, session *domain.Session) error
	Update(ctx context.Context, session *domain.Session) error
	Revoke(ctx context.Context, id string) error
	RevokeByUserID(ctx context.Context, userID string) (int64, error)
	RevokeByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) error
	DeleteExpired(ctx context.Context) error
}

type sessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) SessionRepository {
	return &sessionRepository{db: db}
}

// Column constants for explicit SELECT
const sessionColumns = `id, user_id, device_id, refresh_token_hash, expires_at, 
                        is_revoked, revoked_at, created_at`

func (r *sessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
	var session domain.Session
	query := fmt.Sprintf(`SELECT %s FROM sessions WHERE id = $1 AND is_revoked = false AND expires_at > NOW()`, sessionColumns)
	err := r.db.GetContext(ctx, &session, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) FindByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) (*domain.Session, error) {
	var session domain.Session
	query := fmt.Sprintf(`SELECT %s FROM sessions WHERE user_id = $1 AND device_id = $2 AND is_revoked = false AND expires_at > NOW()`, sessionColumns)
	err := r.db.GetContext(ctx, &session, query, userID, deviceID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) FindByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	var session domain.Session
	query := fmt.Sprintf(`SELECT %s FROM sessions WHERE refresh_token_hash = $1 AND is_revoked = false AND expires_at > NOW()`, sessionColumns)
	err := r.db.GetContext(ctx, &session, query, hash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*domain.Session, error) {
	var sessions []*domain.Session
	query := fmt.Sprintf(`SELECT %s FROM sessions WHERE user_id = $1 AND is_revoked = false AND expires_at > NOW() ORDER BY created_at DESC`, sessionColumns)
	err := r.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if session.ID == "" {
		session.ID = "ses_" + uuid.New().String()[:8]
	}
	session.CreatedAt = time.Now()
	session.IsRevoked = false

	// First revoke existing session for this user+device
	if err := r.RevokeByUserIDAndDeviceID(ctx, session.UserID, session.DeviceID); err != nil {
		return fmt.Errorf("failed to revoke existing session: %w", err)
	}

	query := `
		INSERT INTO sessions (id, user_id, device_id, refresh_token_hash, expires_at, is_revoked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.DeviceID,
		session.RefreshTokenHash, session.ExpiresAt,
		session.IsRevoked, session.CreatedAt,
	)
	return err
}

func (r *sessionRepository) Update(ctx context.Context, session *domain.Session) error {
	query := `UPDATE sessions SET refresh_token_hash = $2, expires_at = $3 WHERE id = $1 AND is_revoked = false`
	_, err := r.db.ExecContext(ctx, query, session.ID, session.RefreshTokenHash, session.ExpiresAt)
	return err
}

func (r *sessionRepository) Revoke(ctx context.Context, id string) error {
	query := `UPDATE sessions SET is_revoked = true, revoked_at = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	return err
}

func (r *sessionRepository) RevokeByUserID(ctx context.Context, userID string) (int64, error) {
	query := `UPDATE sessions SET is_revoked = true, revoked_at = $2 WHERE user_id = $1 AND is_revoked = false`
	result, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *sessionRepository) RevokeByUserIDAndDeviceID(ctx context.Context, userID, deviceID string) error {
	query := `UPDATE sessions SET is_revoked = true, revoked_at = $3 WHERE user_id = $1 AND device_id = $2 AND is_revoked = false`
	_, err := r.db.ExecContext(ctx, query, userID, deviceID, time.Now())
	return err
}

func (r *sessionRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW() OR (is_revoked = true AND revoked_at < $1)`
	cutoff := time.Now().AddDate(0, 0, -7) // Keep revoked sessions for 7 days
	_, err := r.db.ExecContext(ctx, query, cutoff)
	return err
}
