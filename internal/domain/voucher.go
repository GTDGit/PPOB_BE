package domain

import "time"

// Voucher represents a voucher/promo code
type Voucher struct {
	ID                 string     `db:"id" json:"id"`
	Code               string     `db:"code" json:"code"`
	Name               string     `db:"name" json:"name"`
	Description        string     `db:"description" json:"description"`
	DiscountType       string     `db:"discount_type" json:"discountType"`         // fixed, percentage
	DiscountValue      int64      `db:"discount_value" json:"discountValue"`       // amount or percentage
	MinTransaction     int64      `db:"min_transaction" json:"minTransaction"`     // minimum transaction amount
	MaxDiscount        int64      `db:"max_discount" json:"maxDiscount"`           // max discount for percentage
	ApplicableServices string     `db:"applicable_services" json:"-"`              // stored as JSON string
	MaxUsage           int        `db:"max_usage" json:"maxUsage"`                 // max total usage
	MaxUsagePerUser    int        `db:"max_usage_per_user" json:"maxUsagePerUser"` // max usage per user
	CurrentUsage       int        `db:"current_usage" json:"currentUsage"`         // current usage count (renamed from UsageCount)
	StartsAt           *time.Time `db:"starts_at" json:"startsAt"`                 // start date
	ExpiresAt          time.Time  `db:"expires_at" json:"expiresAt"`
	IsActive           bool       `db:"is_active" json:"isActive"` // active status (changed from Status string)
	TermsURL           string     `db:"terms_url" json:"termsUrl"`
	CreatedAt          time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updatedAt"`
}

// UserVoucher represents voucher owned by user
type UserVoucher struct {
	ID        string     `db:"id" json:"id"`
	UserID    string     `db:"user_id" json:"userId"`
	VoucherID string     `db:"voucher_id" json:"voucherId"`
	UsedAt    *time.Time `db:"used_at" json:"usedAt"`
	CreatedAt time.Time  `db:"created_at" json:"createdAt"`
}

// Discount types
const (
	DiscountFixed      = "fixed"
	DiscountPercentage = "percentage"
)

// Voucher status
const (
	VoucherStatusActive  = "active"
	VoucherStatusUsed    = "used"
	VoucherStatusExpired = "expired"
)
