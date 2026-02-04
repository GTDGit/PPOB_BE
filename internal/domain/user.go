package domain

import (
	"database/sql"
	"time"
)

// User represents a user/merchant in the system
type User struct {
	ID               string         `db:"id" json:"id"`
	MIC              string         `db:"mic" json:"mic"`
	Phone            string         `db:"phone" json:"phone"`
	FullName         string         `db:"full_name" json:"fullName"`
	Email            sql.NullString `db:"email" json:"email"`
	Gender           *string        `db:"gender" json:"gender"`
	Tier             string         `db:"tier" json:"tier"`
	AvatarURL        *string        `db:"avatar_url" json:"avatarUrl"`
	KYCStatus        string         `db:"kyc_status" json:"kycStatus"`
	BusinessType     *string        `db:"business_type" json:"businessType"`
	Source           *string        `db:"source" json:"source"`
	ReferredBy       *string        `db:"referred_by" json:"referredBy"`
	ReferralCode     *string        `db:"referral_code" json:"referralCode"`
	UsedReferralCode *string        `db:"used_referral_code" json:"usedReferralCode"`
	PINHash          string         `db:"pin_hash" json:"-"`
	IsActive         bool           `db:"is_active" json:"isActive"`
	IsLocked         bool           `db:"is_locked" json:"isLocked"`
	LockedUntil      *time.Time     `db:"locked_until" json:"lockedUntil"`
	CreatedAt        time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time      `db:"updated_at" json:"updatedAt"`
	PhoneVerifiedAt  sql.NullTime   `db:"phone_verified_at" json:"phoneVerifiedAt"`
}

// UserResponse is the user object returned in API responses
type UserResponse struct {
	ID        string  `json:"id"`
	MIC       string  `json:"mic"`
	Phone     string  `json:"phone"`
	FullName  string  `json:"fullName"`
	Email     *string `json:"email"`
	Tier      string  `json:"tier"`
	AvatarURL *string `json:"avatarUrl"`
	KYCStatus string  `json:"kycStatus"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	var email *string
	if u.Email.Valid {
		email = &u.Email.String
	}
	return &UserResponse{
		ID:        u.ID,
		MIC:       u.MIC,
		Phone:     u.Phone,
		FullName:  u.FullName,
		Email:     email,
		Tier:      u.Tier,
		AvatarURL: u.AvatarURL,
		KYCStatus: u.KYCStatus,
	}
}

// GetDisplayName returns a display-friendly name
func (u *User) GetDisplayName() string {
	if u.FullName != "" {
		return u.FullName
	}
	return u.Phone
}

// Constants for user tiers
const (
	TierBasic = "BASIC"
)

// Constants for KYC status (matches database: unverified, pending, verified, rejected)
const (
	KYCStatusUnverified = "unverified"
	KYCStatusPending    = "pending"
	KYCStatusVerified   = "verified"
	KYCStatusRejected   = "rejected"
)

// Constants for user flows - use domain.FlowXxx from session.go
