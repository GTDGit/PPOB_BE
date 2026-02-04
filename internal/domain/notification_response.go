package domain

// NotificationListResponse represents notification list response
type NotificationListResponse struct {
	Notifications []*NotificationSummary `json:"notifications"`
	EmptyState    *EmptyState            `json:"emptyState,omitempty"`
	Pagination    *Pagination            `json:"pagination"`
}

// NotificationSummary represents summary notification info for list
type NotificationSummary struct {
	ID                 string              `json:"id"`
	Category           string              `json:"category"`
	Title              string              `json:"title"`
	Body               string              `json:"body"`
	ShortBody          string              `json:"shortBody"`
	ImageURL           *string             `json:"imageUrl"`
	IsRead             bool                `json:"isRead"`
	CreatedAt          string              `json:"createdAt"`
	CreatedAtFormatted string              `json:"createdAtFormatted"`
	Action             *NotificationAction `json:"action,omitempty"`
}

// NotificationDetailResponse represents detailed notification response
type NotificationDetailResponse struct {
	Notification *NotificationDetail `json:"notification"`
}

// NotificationDetail represents detailed notification information
type NotificationDetail struct {
	ID                 string                 `json:"id"`
	Category           string                 `json:"category"`
	Title              string                 `json:"title"`
	Body               string                 `json:"body"`
	ImageURL           *string                `json:"imageUrl"`
	IsRead             bool                   `json:"isRead"`
	CreatedAt          string                 `json:"createdAt"`
	CreatedAtFormatted string                 `json:"createdAtFormatted"`
	Action             *NotificationAction    `json:"action,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationAction represents notification action
type NotificationAction struct {
	Type       string  `json:"type"`
	Value      string  `json:"value"`
	ButtonText *string `json:"buttonText,omitempty"`
}

// EmptyState represents empty state information
type EmptyState struct {
	Title    string  `json:"title"`
	Message  string  `json:"message"`
	ImageURL *string `json:"imageUrl,omitempty"`
}

// UnreadCountResponse represents unread count response
type UnreadCountResponse struct {
	UnreadCount int            `json:"unreadCount"`
	ByCategory  map[string]int `json:"byCategory"`
}

// MarkAsReadResponse represents mark as read response
type MarkAsReadResponse struct {
	NotificationID string `json:"notificationId"`
	IsRead         bool   `json:"isRead"`
	ReadAt         string `json:"readAt"`
}

// MarkAllAsReadResponse represents mark all as read response
type MarkAllAsReadResponse struct {
	MarkedCount int    `json:"markedCount"`
	Message     string `json:"message"`
}

// DeleteNotificationResponse represents delete notification response
type DeleteNotificationResponse struct {
	Deleted        bool   `json:"deleted"`
	NotificationID string `json:"notificationId"`
}
