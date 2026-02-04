package domain

import "time"

// Contact represents a saved favorite contact
type Contact struct {
	ID         string     `db:"id" json:"id"`
	UserID     string     `db:"user_id" json:"userId"`
	Name       string     `db:"name" json:"name"`
	Type       string     `db:"type" json:"type"`         // phone, pln, pdam, bpjs, bank, etc
	Value      string     `db:"value" json:"value"`       // phone number, meter number, account number, etc
	Metadata   *string    `db:"metadata" json:"metadata"` // JSON for extra info (bank code, operator, etc)
	LastUsedAt *time.Time `db:"last_used_at" json:"lastUsedAt"`
	UsageCount int        `db:"usage_count" json:"usageCount"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
}

// Contact types
const (
	ContactTypePhone  = "phone"
	ContactTypePLN    = "pln"
	ContactTypePDAM   = "pdam"
	ContactTypeBPJS   = "bpjs"
	ContactTypeTelkom = "telkom"
	ContactTypeBank   = "bank"
)
