package domain

import "time"

// Notification represents a user notification
type Notification struct {
	ID               string     `db:"id" json:"id"`
	UserID           string     `db:"user_id" json:"userId"`
	Category         string     `db:"category" json:"category"` // security, transaction, deposit, promo, info, qris
	Title            string     `db:"title" json:"title"`
	Body             string     `db:"body" json:"body"`
	ShortBody        *string    `db:"short_body" json:"shortBody"` // Short summary for list view
	ImageURL         *string    `db:"image_url" json:"imageUrl"`
	ActionType       *string    `db:"action_type" json:"actionType"`              // deeplink, webview, external_url, none
	ActionValue      *string    `db:"action_value" json:"actionValue"`            // URL or deeplink path
	ActionButtonText *string    `db:"action_button_text" json:"actionButtonText"` // Custom action button text
	Metadata         *string    `db:"metadata" json:"metadata"`                   // JSON for extra data
	IsRead           bool       `db:"is_read" json:"isRead"`
	ReadAt           *time.Time `db:"read_at" json:"readAt"`
	CreatedAt        time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updatedAt"`
}

// Notification categories
const (
	NotificationCategorySecurity    = "security"
	NotificationCategoryTransaction = "transaction"
	NotificationCategoryDeposit     = "deposit"
	NotificationCategoryPromo       = "promo"
	NotificationCategoryInfo        = "info"
	NotificationCategoryQRIS        = "qris"
)

// Notification action types
const (
	NotificationActionDeeplink    = "deeplink"
	NotificationActionWebview     = "webview"
	NotificationActionExternalURL = "external_url"
	NotificationActionNone        = "none"
)
