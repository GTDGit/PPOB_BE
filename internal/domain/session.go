package domain

import (
	"time"
)

// Session represents a JWT session for token management
type Session struct {
	ID               string     `db:"id" json:"id"`
	UserID           string     `db:"user_id" json:"userId"`
	DeviceID         string     `db:"device_id" json:"deviceId"`
	RefreshTokenHash string     `db:"refresh_token_hash" json:"-"`
	ExpiresAt        time.Time  `db:"expires_at" json:"expiresAt"`
	IsRevoked        bool       `db:"is_revoked" json:"isRevoked"`
	RevokedAt        *time.Time `db:"revoked_at" json:"revokedAt"`
	CreatedAt        time.Time  `db:"created_at" json:"createdAt"`
}

// OTPSession represents an OTP verification session (stored in Redis)
type OTPSession struct {
	SessionID    string    `json:"sessionId"`
	Phone        string    `json:"phone"`
	OTP          string    `json:"otp"`
	OTPMethod    string    `json:"otpMethod"`
	Flow         string    `json:"flow"`
	Attempts     int       `json:"attempts"`
	MaxAttempts  int       `json:"maxAttempts"`
	ResendCount  int       `json:"resendCount"`
	LastResendAt *int64    `json:"lastResendAt"`
	IsVerified   bool      `json:"isVerified"`
	VerifiedAt   *int64    `json:"verifiedAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
	CreatedAt    time.Time `json:"createdAt"`

	// Additional data for registration flow
	DeviceID   string `json:"deviceId,omitempty"`
	DeviceName string `json:"deviceName,omitempty"`
	IPAddress  string `json:"ipAddress,omitempty"`
}

// TempToken represents temporary token data (stored in Redis)
type TempToken struct {
	Token      string `json:"token"`
	UserID     string `json:"userId"`
	Phone      string `json:"phone"`
	Flow       string `json:"flow"`
	Step       string `json:"step"`
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	ExpiresAt  int64  `json:"expiresAt"`
	CreatedAt  int64  `json:"createdAt"`

	// For profile completion
	IsProfileComplete bool `json:"isProfileComplete"`
	IsPINSet          bool `json:"isPINSet"`
}

// OTP Method enum
const (
	OTPMethodWA  = "wa"
	OTPMethodSMS = "sms"
)

// Auth Flow enum
const (
	FlowRegister    = "REGISTER"
	FlowLogin       = "LOGIN"
	FlowResetPIN    = "RESET_PIN"
	FlowChangePhone = "CHANGE_PHONE"
	FlowChangePIN   = "CHANGE_PIN"
)

// Auth Step enum
const (
	StepRequestOTP        = "REQUEST_OTP"
	StepVerifyOTP         = "VERIFY_OTP"
	StepCompleteProfile   = "COMPLETE_PROFILE"
	StepSetPIN            = "SET_PIN"
	StepInputPIN          = "INPUT_PIN"
	StepAuthenticated     = "AUTHENTICATED"
	StepChangePINConfirm  = "CHANGE_PIN_CONFIRM"
	StepVerifyNewPhone    = "VERIFY_NEW_PHONE"
	StepVerifyOldPhone    = "VERIFY_OLD_PHONE"
	StepVerifyNewPhoneOTP = "VERIFY_NEW_PHONE_OTP"
	StepPINVerified       = "PIN_VERIFIED"
	StepPINChanged        = "PIN_CHANGED"
	StepPhoneChanged      = "PHONE_CHANGED"
	StepTokenRefreshed    = "TOKEN_REFRESHED"
	StepLoggedOut         = "LOGGED_OUT"
	StepDevicesListed     = "DEVICES_LISTED"
	StepDeviceRemoved     = "DEVICE_REMOVED"
)
