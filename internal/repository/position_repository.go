package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/jmoiron/sqlx"
)

type PositionRepository struct {
	db *sqlx.DB
}

func NewPositionRepository(db *sqlx.DB) *PositionRepository {
	return &PositionRepository{db: db}
}

func (r *PositionRepository) ListPositions(ctx context.Context) ([]domain.AdminPosition, error) {
	var positions []domain.AdminPosition
	err := r.db.SelectContext(ctx, &positions, `
		SELECT id, name, created_at, updated_at
		FROM admin_positions
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	if positions == nil {
		positions = []domain.AdminPosition{}
	}
	return positions, nil
}

func (r *PositionRepository) GetPositionByID(ctx context.Context, id string) (*domain.AdminPosition, error) {
	var pos domain.AdminPosition
	err := r.db.GetContext(ctx, &pos, `SELECT id, name, created_at, updated_at FROM admin_positions WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

func (r *PositionRepository) CreatePosition(ctx context.Context, name string) (*domain.AdminPosition, error) {
	var pos domain.AdminPosition
	err := r.db.GetContext(ctx, &pos, `
		INSERT INTO admin_positions (name) VALUES ($1)
		RETURNING id, name, created_at, updated_at
	`, name)
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

func (r *PositionRepository) UpdatePosition(ctx context.Context, id, name string) (*domain.AdminPosition, error) {
	var pos domain.AdminPosition
	err := r.db.GetContext(ctx, &pos, `
		UPDATE admin_positions SET name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, created_at, updated_at
	`, id, name)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

func (r *PositionRepository) DeletePosition(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM admin_positions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PositionRepository) CountAdminsByPosition(ctx context.Context, positionID string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM admin_users WHERE position_id = $1`, positionID)
	return count, err
}

func (r *PositionRepository) ListPositionsWithCount(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryxContext(ctx, `
		SELECT
			ap.id,
			ap.name,
			ap.created_at,
			ap.updated_at,
			COUNT(au.id) AS admin_count
		FROM admin_positions ap
		LEFT JOIN admin_users au ON au.position_id = ap.id
		GROUP BY ap.id
		ORDER BY ap.name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		item := make(map[string]interface{})
		if err := rows.MapScan(item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []map[string]interface{}{}
	}
	return items, nil
}

func (r *PositionRepository) ListAdminsByPosition(ctx context.Context, positionID string) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryxContext(ctx, `
		SELECT
			au.id,
			au.email,
			COALESCE(au.full_name, '') AS full_name,
			COALESCE(au.avatar_url, '') AS avatar_url,
			au.status,
			COALESCE(STRING_AGG(DISTINCT ar.name, ', '), '') AS roles
		FROM admin_users au
		LEFT JOIN admin_user_roles aur ON aur.admin_user_id = au.id
		LEFT JOIN admin_roles ar ON ar.id = aur.role_id
		WHERE au.position_id = $1
		GROUP BY au.id
		ORDER BY au.full_name ASC
	`, positionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		item := make(map[string]interface{})
		if err := rows.MapScan(item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []map[string]interface{}{}
	}
	return items, nil
}

func (r *PositionRepository) AssignPosition(ctx context.Context, adminID, positionID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE admin_users SET position_id = $2, updated_at = NOW() WHERE id = $1
	`, adminID, positionID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("admin not found")
	}
	return nil
}

func (r *PositionRepository) RemoveAdminPosition(ctx context.Context, adminID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_users SET position_id = NULL, updated_at = NOW() WHERE id = $1
	`, adminID)
	return err
}
