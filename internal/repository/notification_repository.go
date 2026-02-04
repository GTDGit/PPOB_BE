package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// NotificationRepository defines the interface for notification operations
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
}

// NotificationFilter represents filter options for notification list
type NotificationFilter struct {
	Category string // security, transaction, deposit, promo, info, qris, all
	IsRead   *bool  // nil = all, true = read only, false = unread only
	Page     int
	PerPage  int
}

// notificationRepository implements NotificationRepository
type notificationRepository struct {
	db *sqlx.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *sqlx.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// FindByUserID finds notifications by user ID with filters and pagination
func (r *notificationRepository) FindByUserID(ctx context.Context, userID string, filter NotificationFilter) ([]*domain.Notification, int, error) {
	// For now, return mock notifications
	allNotifications := r.getMockNotifications()

	// Filter by user
	var userNotifications []*domain.Notification
	for _, notif := range allNotifications {
		if notif.UserID == userID {
			userNotifications = append(userNotifications, notif)
		}
	}

	// Apply filters
	var filtered []*domain.Notification
	for _, notif := range userNotifications {
		// Filter by category
		if filter.Category != "" && filter.Category != "all" && notif.Category != filter.Category {
			continue
		}

		// Filter by isRead
		if filter.IsRead != nil {
			if *filter.IsRead != notif.IsRead {
				continue
			}
		}

		filtered = append(filtered, notif)
	}

	total := len(filtered)

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

	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(filtered) {
		return []*domain.Notification{}, total, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	paginated := filtered[start:end]

	return paginated, total, nil
}

// FindByID finds a notification by ID
func (r *notificationRepository) FindByID(ctx context.Context, id string) (*domain.Notification, error) {
	notifications := r.getMockNotifications()
	for _, notif := range notifications {
		if notif.ID == id {
			return notif, nil
		}
	}
	return nil, nil
}

// FindByUserAndID finds a notification by user ID and notification ID (ownership validation)
func (r *notificationRepository) FindByUserAndID(ctx context.Context, userID, id string) (*domain.Notification, error) {
	notifications := r.getMockNotifications()
	for _, notif := range notifications {
		if notif.ID == id && notif.UserID == userID {
			return notif, nil
		}
	}
	return nil, nil
}

// CountUnread counts unread notifications for a user
func (r *notificationRepository) CountUnread(ctx context.Context, userID string) (int, error) {
	notifications := r.getMockNotifications()
	count := 0
	for _, notif := range notifications {
		if notif.UserID == userID && !notif.IsRead {
			count++
		}
	}
	return count, nil
}

// CountUnreadByCategory counts unread notifications by category for a user
func (r *notificationRepository) CountUnreadByCategory(ctx context.Context, userID string) (map[string]int, error) {
	notifications := r.getMockNotifications()
	counts := map[string]int{
		domain.NotificationCategorySecurity:    0,
		domain.NotificationCategoryTransaction: 0,
		domain.NotificationCategoryDeposit:     0,
		domain.NotificationCategoryPromo:       0,
		domain.NotificationCategoryInfo:        0,
		domain.NotificationCategoryQRIS:        0,
	}

	for _, notif := range notifications {
		if notif.UserID == userID && !notif.IsRead {
			counts[notif.Category]++
		}
	}

	return counts, nil
}

// MarkAsRead marks a notification as read
func (r *notificationRepository) MarkAsRead(ctx context.Context, id string) error {
	// In production, update database
	// For now, just return success
	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID string, category *string) (int, error) {
	// In production, update database and return affected count
	// For now, simulate by counting unread
	notifications := r.getMockNotifications()
	count := 0
	for _, notif := range notifications {
		if notif.UserID == userID && !notif.IsRead {
			if category != nil && notif.Category != *category {
				continue
			}
			count++
		}
	}
	return count, nil
}

// Delete deletes a notification
func (r *notificationRepository) Delete(ctx context.Context, id string) error {
	// In production, delete from database
	// For now, just return success
	return nil
}

// Create creates a new notification
func (r *notificationRepository) Create(ctx context.Context, notif *domain.Notification) error {
	// In production, insert into database
	// For now, just return success
	return nil
}

// getMockNotifications returns mock notification data
func (r *notificationRepository) getMockNotifications() []*domain.Notification {
	now := time.Now()

	actionDeeplink := domain.NotificationActionDeeplink

	actionValue1 := "/security/sessions"
	actionValue2 := "/security/sessions"
	actionValue3 := "/transactions/trx_pulsa_abc123"
	actionValue4 := "/services/token-pln"
	actionValue5 := "/qris/income/qris_inc_abc123"

	imageURL4 := "https://cdn.ppob.id/notifications/promo-akhir-tahun.png"

	readAt2 := now.Add(-1 * time.Hour)
	readAt3 := now.Add(-2 * time.Hour)
	readAt5 := now.Add(-3 * time.Hour)

	return []*domain.Notification{
		// Security - Unread
		{
			ID:          "notif_sec_abc123",
			UserID:      "mock_user_id",
			Category:    domain.NotificationCategorySecurity,
			Title:       "Info PPOB",
			Body:        "Akun anda tercatat login di perangkat lain. Abaikan pesan ini jika itu anda sendiri.",
			ImageURL:    nil,
			ActionType:  &actionDeeplink,
			ActionValue: &actionValue1,
			Metadata:    nil,
			IsRead:      false,
			ReadAt:      nil,
			CreatedAt:   now.Add(-5 * time.Minute),
			UpdatedAt:   now.Add(-5 * time.Minute),
		},
		// Security - Read
		{
			ID:          "notif_sec_xyz789",
			UserID:      "mock_user_id",
			Category:    domain.NotificationCategorySecurity,
			Title:       "Info PPOB",
			Body:        "Akun anda tercatat login di perangkat lain. Abaikan pesan ini jika itu anda sendiri.",
			ImageURL:    nil,
			ActionType:  &actionDeeplink,
			ActionValue: &actionValue2,
			Metadata:    nil,
			IsRead:      true,
			ReadAt:      &readAt2,
			CreatedAt:   now.Add(-1 * time.Hour).Add(-5 * time.Minute),
			UpdatedAt:   now.Add(-1 * time.Hour),
		},
		// Transaction - Read
		{
			ID:          "notif_trx_def456",
			UserID:      "mock_user_id",
			Category:    domain.NotificationCategoryTransaction,
			Title:       "Transaksi Berhasil",
			Body:        "Pembelian Pulsa Telkomsel 50.000 ke 081234567890 berhasil.",
			ImageURL:    nil,
			ActionType:  &actionDeeplink,
			ActionValue: &actionValue3,
			Metadata:    nil,
			IsRead:      true,
			ReadAt:      &readAt3,
			CreatedAt:   now.Add(-2 * time.Hour),
			UpdatedAt:   now.Add(-2 * time.Hour),
		},
		// Promo - Unread
		{
			ID:          "notif_promo_ghi789",
			UserID:      "mock_user_id",
			Category:    domain.NotificationCategoryPromo,
			Title:       "Promo Akhir Tahun! 🎉",
			Body:        "Cashback 20% untuk semua transaksi Token PLN. Berlaku sampai 31 Desember 2025.",
			ImageURL:    &imageURL4,
			ActionType:  &actionDeeplink,
			ActionValue: &actionValue4,
			Metadata:    nil,
			IsRead:      false,
			ReadAt:      nil,
			CreatedAt:   now.Add(-1 * 24 * time.Hour),
			UpdatedAt:   now.Add(-1 * 24 * time.Hour),
		},
		// QRIS - Read
		{
			ID:          "notif_qris_jkl012",
			UserID:      "mock_user_id",
			Category:    domain.NotificationCategoryQRIS,
			Title:       "Pembayaran QRIS Diterima",
			Body:        "Anda menerima pembayaran Rp50.000 dari BUDI SANTOSO via BCA.",
			ImageURL:    nil,
			ActionType:  &actionDeeplink,
			ActionValue: &actionValue5,
			Metadata:    nil,
			IsRead:      true,
			ReadAt:      &readAt5,
			CreatedAt:   now.Add(-3 * 24 * time.Hour),
			UpdatedAt:   now.Add(-3 * 24 * time.Hour),
		},
	}
}
