package domain

import "time"

// PushToken represents a registered mobile/web push token.
type PushToken struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"userId"`
	DeviceID  string    `db:"device_id" json:"deviceId"`
	Token     string    `db:"token" json:"token"`
	Platform  string    `db:"platform" json:"platform"`
	IsActive  bool      `db:"is_active" json:"isActive"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

// PushTokenResponse is returned after registering or deactivating a token.
type PushTokenResponse struct {
	Registered bool   `json:"registered"`
	DeviceID   string `json:"deviceId"`
	Platform   string `json:"platform,omitempty"`
}
