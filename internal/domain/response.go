package domain

import (
	"time"

	"github.com/google/uuid"
)

// APIResponse is the standard success response format
type APIResponse struct {
	Data interface{} `json:"data"`
	Meta *Meta       `json:"meta"`
}

// APIErrorResponse is the standard error response format
type APIErrorResponse struct {
	Error *AppError `json:"error"`
	Meta  *Meta     `json:"meta"`
}

// Meta contains request metadata
type Meta struct {
	RequestID string `json:"requestId"`
	Timestamp string `json:"timestamp"`
}

// NewMeta creates new response metadata
func NewMeta(requestID string) *Meta {
	if requestID == "" {
		requestID = "req_" + uuid.New().String()[:8]
	}
	return &Meta{
		RequestID: requestID,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// SuccessResponse creates a success response
func SuccessResponse(data interface{}, requestID string) *APIResponse {
	return &APIResponse{
		Data: data,
		Meta: NewMeta(requestID),
	}
}

// ErrorResponse creates an error response
func ErrorResponse(err *AppError, requestID string) *APIErrorResponse {
	return &APIErrorResponse{
		Error: err,
		Meta:  NewMeta(requestID),
	}
}

// StartAuthResponse is the response for /auth/start
type StartAuthResponse struct {
	Step         string  `json:"step"`
	Flow         string  `json:"flow"`
	OTPSessionID string  `json:"otpSessionId,omitempty"`
	ExpiresIn    int     `json:"expiresIn,omitempty"`
	OTPMethod    string  `json:"otpMethod,omitempty"`
	MaskedPhone  string  `json:"maskedPhone"`
	UserName     *string `json:"userName,omitempty"`
}

// VerifyOTPResponse is the response for /auth/verify-otp
type VerifyOTPResponse struct {
	Step         string        `json:"step"`
	TempToken    string        `json:"tempToken,omitempty"`
	AccessToken  string        `json:"accessToken,omitempty"`
	RefreshToken string        `json:"refreshToken,omitempty"`
	ExpiresIn    int           `json:"expiresIn,omitempty"`
	User         *UserResponse `json:"user,omitempty"`
}

// CompleteProfileResponse is the response for /auth/complete-profile
type CompleteProfileResponse struct {
	Step      string `json:"step"`
	TempToken string `json:"tempToken"`
}

// SetPINResponse is the response for /auth/set-pin
type SetPINResponse struct {
	Step         string        `json:"step"`
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	ExpiresIn    int           `json:"expiresIn"`
	User         *UserResponse `json:"user"`
}

// PINLoginResponse is the response for /auth/pin-login
type PINLoginResponse struct {
	Step         string        `json:"step"`
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	ExpiresIn    int           `json:"expiresIn"`
	User         *UserResponse `json:"user"`
}

// VerifyPINOnlyResponse is the response for /auth/verify-pin-only
type VerifyPINOnlyResponse struct {
	Step       string `json:"step"`
	Message    string `json:"message"`
	VerifiedAt string `json:"verifiedAt"`
}

// ResendOTPResponse is the response for /auth/resend-otp
type ResendOTPResponse struct {
	Step         string `json:"step"`
	OTPSessionID string `json:"otpSessionId"`
	ExpiresIn    int    `json:"expiresIn"`
	OTPMethod    string `json:"otpMethod"`
	ResendCount  int    `json:"resendCount"`
	MaxResend    int    `json:"maxResend"`
	NextResendAt string `json:"nextResendAt"`
}

// RefreshTokenResponse is the response for /auth/refresh-token
type RefreshTokenResponse struct {
	Step         string `json:"step"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}

// LogoutResponse is the response for /auth/logout
type LogoutResponse struct {
	Step             string `json:"step"`
	Message          string `json:"message"`
	DevicesLoggedOut int    `json:"devicesLoggedOut"`
}

// ListDevicesResponse is the response for /auth/devices
type ListDevicesResponse struct {
	Step         string            `json:"step"`
	Devices      []*DeviceResponse `json:"devices"`
	TotalDevices int               `json:"totalDevices"`
}

// RemoveDeviceResponse is the response for /auth/devices/:deviceId
type RemoveDeviceResponse struct {
	Step     string `json:"step"`
	Message  string `json:"message"`
	DeviceID string `json:"deviceId"`
}

// ChangePINVerifyResponse is the response for /auth/change-pin/verify-current
type ChangePINVerifyResponse struct {
	Step      string `json:"step"`
	TempToken string `json:"tempToken"`
	ExpiresIn int    `json:"expiresIn"`
}

// ChangePINConfirmResponse is the response for /auth/change-pin/confirm
type ChangePINConfirmResponse struct {
	Step    string `json:"step"`
	Message string `json:"message"`
}

// ChangePhoneRequestOTPResponse is the response for phone change OTP requests
type ChangePhoneRequestOTPResponse struct {
	Step         string `json:"step"`
	OTPSessionID string `json:"otpSessionId"`
	ExpiresIn    int    `json:"expiresIn"`
	OTPMethod    string `json:"otpMethod"`
	MaskedPhone  string `json:"maskedPhone"`
}

// ChangePhoneVerifyOldResponse is the response for /auth/change-phone/verify-old
type ChangePhoneVerifyOldResponse struct {
	Step      string `json:"step"`
	TempToken string `json:"tempToken"`
	ExpiresIn int    `json:"expiresIn"`
}

// ChangePhoneCompleteResponse is the response for completing phone change
type ChangePhoneCompleteResponse struct {
	Step        string `json:"step"`
	Message     string `json:"message"`
	Phone       string `json:"phone"`
	MaskedPhone string `json:"maskedPhone"`
}

// EmailVerificationSentResponse is the response for email verification request
type EmailVerificationSentResponse struct {
	Step        string `json:"step"`
	Message     string `json:"message"`
	MaskedEmail string `json:"maskedEmail"`
	ExpiresIn   int    `json:"expiresIn"`
}

// EmailVerifiedResponse is the response for email verification
type EmailVerifiedResponse struct {
	Step       string `json:"step"`
	Message    string `json:"message"`
	Email      string `json:"email"`
	VerifiedAt string `json:"verifiedAt"`
}
