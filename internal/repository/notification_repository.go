package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/jmoiron/sqlx"
)

// NotificationRepository defines the interface for notification operations.
type NotificationRepository interface {
	FindByUserID(ctx context.Context, userID string, filter NotificationFilter) ([]*domain.Notification, int, error)
	FindByID(ctx context.Context, id string) (*domain.Notification, error)
	FindByUserAndID(ctx context.Context, userID, id string) (*domain.Notification, error)
	CountUnread(ctx context.Context, userID string) (int, error)
	CountUnreadByCategory(ctx context.Context, userID string) (map[string]int, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string, category *string) (int, error)
	Delete(ctx context.Context, id string) error
	Create(ctx context.Context, notif *domain.Notification) error
	UpsertPushToken(ctx context.Context, token *domain.PushToken) error
	DeactivatePushToken(ctx context.Context, userID, deviceID string) error
}

// NotificationFilter represents filter options for notification list.
type NotificationFilter struct {
	Category string
	IsRead   *bool
	Page     int
	PerPage  int
}

type notificationRepository struct {
	db *sqlx.DB
}

const notificationColumns = `
	id, user_id, category, title, body, short_body, image_url, action_type,
	action_value, action_button_text, metadata, is_read, read_at, created_at, updated_at
`

// NewNotificationRepository creates a new notification repository.
func NewNotificationRepository(db *sqlx.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// FindByUserID finds notifications by user ID with filters and pagination.
func (r *notificationRepository) FindByUserID(ctx context.Context, userID string, filter NotificationFilter) ([]*domain.Notification, int, error) {
	whereClauses := []string{"user_id = $1"}
	args := []interface{}{userID}
	argIdx := 2

	if filter.Category != "" && filter.Category != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, filter.Category)
		argIdx++
	}

	if filter.IsRead != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_read = $%d", argIdx))
		args = append(args, *filter.IsRead)
		argIdx++
	}

	whereSQL := "WHERE " + strings.Join(whereClauses, " AND ")

	countQuery := `SELECT COUNT(*) FROM notifications ` + whereSQL
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
	query := `SELECT ` + notificationColumns + ` FROM notifications ` + whereSQL +
		fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)

	queryArgs := append(args, perPage, offset)
	var notifications []*domain.Notification
	if err := r.db.SelectContext(ctx, &notifications, query, queryArgs...); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// FindByID finds a notification by ID.
func (r *notificationRepository) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	query := `SELECT ` + notificationColumns + ` FROM notifications WHERE id = $1`

	var notif domain.Notification
	if err := r.db.GetContext(ctx, &notif, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &notif, nil
}

// FindByUserAndID finds a notification by user ID and notification ID.
func (r *notificationRepository) FindByUserAndID(ctx context.Context, userID, id string) (*domain.Notification, error) {
	query := `SELECT ` + notificationColumns + ` FROM notifications WHERE id = $1 AND user_id = $2`

	var notif domain.Notification
	if err := r.db.GetContext(ctx, &notif, query, id, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &notif, nil
}

// CountUnread counts unread notifications for a user.
func (r *notificationRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`

	var count int
	if err := r.db.GetContext(ctx, &count, query, userID); err != nil {
		return 0, err
	}

	return count, nil
}

// CountUnreadByCategory counts unread notifications by category for a user.
func (r *notificationRepository) CountUnreadByCategory(ctx context.Context, userID string) (map[string]int, error) {
	counts := map[string]int{
		domain.NotificationCategorySecurity:    0,
		domain.NotificationCategoryTransaction: 0,
		domain.NotificationCategoryDeposit:     0,
		domain.NotificationCategoryPromo:       0,
		domain.NotificationCategoryInfo:        0,
		domain.NotificationCategoryQRIS:        0,
	}

	query := `
		SELECT category, COUNT(*) AS total
		FROM notifications
		WHERE user_id = $1 AND is_read = false
		GROUP BY category
	`

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var total int
		if err := rows.Scan(&category, &total); err != nil {
			return nil, err
		}
		counts[category] = total
	}

	return counts, rows.Err()
}

// MarkAsRead marks a notification as read.
func (r *notificationRepository) MarkAsRead(ctx context.Context, id string) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// MarkAllAsRead marks all notifications as read for a user.
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string, category *string) (int, error) {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND is_read = false
	`
	args := []interface{}{userID}

	if category != nil && *category != "" {
		query += ` AND category = $2`
		args = append(args, *category)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(affected), nil
}

// Delete deletes a notification.
func (r *notificationRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notifications WHERE id = $1`, id)
	return err
}

// Create creates a new notification.
func (r *notificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	query := `
		INSERT INTO notifications (
			id, user_id, category, title, body, short_body, image_url, action_type,
			action_value, action_button_text, metadata, is_read, read_at, created_at, updated_at
		) VALUES (
			:id, :user_id, :category, :title, :body, :short_body, :image_url, :action_type,
			:action_value, :action_button_text, :metadata, :is_read, :read_at, :created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, notif)
	return err
}

// UpsertPushToken creates or updates a device push token.
func (r *notificationRepository) UpsertPushToken(ctx context.Context, token *domain.PushToken) error {
	query := `
		INSERT INTO push_tokens (
			id, user_id, device_id, token, platform, is_active, created_at, updated_at
		) VALUES (
			:id, :user_id, :device_id, :token, :platform, :is_active, :created_at, :updated_at
		)
		ON CONFLICT (user_id, device_id)
		DO UPDATE SET
			token = EXCLUDED.token,
			platform = EXCLUDED.platform,
			is_active = true,
			updated_at = NOW()
	`
	_, err := r.db.NamedExecContext(ctx, query, token)
	return err
}

// DeactivatePushToken deactivates a push token for a device.
func (r *notificationRepository) DeactivatePushToken(ctx context.Context, userID, deviceID string) error {
	query := `
		UPDATE push_tokens
		SET is_active = false, updated_at = NOW()
		WHERE user_id = $1 AND device_id = $2
	`
	_, err := r.db.ExecContext(ctx, query, userID, deviceID)
	return err
}
