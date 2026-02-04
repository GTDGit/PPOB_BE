package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// StartAuthRequest represents start auth request body
type StartAuthRequest struct {
	Phone      string `json:"phone" binding:"required"`
	DeviceID   string `json:"deviceId" binding:"required"`
	DeviceName string `json:"deviceName" binding:"required"`
	OTPMethod  string `json:"otpMethod" binding:"required,oneof=wa sms"`
}

// StartAuth handles POST /v1/auth/start
func (h *AuthHandler) StartAuth(c *gin.Context) {
	var req StartAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	resp, err := h.authService.StartAuth(c.Request.Context(), service.StartAuthRequest{
		Phone:      req.Phone,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
		OTPMethod:  req.OTPMethod,
		IPAddress:  c.ClientIP(),
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// VerifyOTPRequest represents verify OTP request body
type VerifyOTPRequest struct {
	Phone        string `json:"phone" binding:"required"`
	OTP          string `json:"otp" binding:"required,len=4"`
	OTPSessionID string `json:"otpSessionId" binding:"required"`
	DeviceID     string `json:"deviceId" binding:"required"`
	DeviceName   string `json:"deviceName" binding:"required"`
}

// VerifyOTP handles POST /v1/auth/verify-otp
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	resp, err := h.authService.VerifyOTP(c.Request.Context(), service.VerifyOTPAuthRequest{
		Phone:        req.Phone,
		OTP:          req.OTP,
		OTPSessionID: req.OTPSessionID,
		DeviceID:     req.DeviceID,
		DeviceName:   req.DeviceName,
		IPAddress:    c.ClientIP(),
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// ResendOTPRequest represents resend OTP request body
type ResendOTPRequest struct {
	Phone        string `json:"phone" binding:"required"`
	OTPSessionID string `json:"otpSessionId" binding:"required"`
	OTPMethod    string `json:"otpMethod" binding:"required,oneof=wa sms"`
}

// ResendOTP handles POST /v1/auth/resend-otp
func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	resp, err := h.authService.ResendOTP(c.Request.Context(), service.ResendOTPRequest{
		Phone:        req.Phone,
		OTPSessionID: req.OTPSessionID,
		OTPMethod:    req.OTPMethod,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// CompleteProfileRequest represents complete profile request body
type CompleteProfileRequest struct {
	FullName     string  `json:"fullName" binding:"required,min=3"`
	Email        *string `json:"email" binding:"omitempty,email"`
	Gender       string  `json:"gender" binding:"required,oneof=MALE FEMALE"`
	ReferredBy   *string `json:"referredBy" binding:"omitempty,oneof=USER SALES"`
	ReferralCode *string `json:"referralCode"`
	BusinessType string  `json:"businessType" binding:"required,oneof=SHOP COUNTER MONEY_AGENT ONLINE_SELLER NONE OTHER"`
	Source       string  `json:"source" binding:"required,oneof=SALES FRIEND GOOGLE PLAYSTORE_APPSTORE ADS SOCIAL_MEDIA OUTDOOR OTHER"`
}

// CompleteProfile handles POST /v1/auth/complete-profile
func (h *AuthHandler) CompleteProfile(c *gin.Context) {
	var req CompleteProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	tempToken := middleware.GetTempToken(c)
	if tempToken == nil {
		respondWithError(c, domain.ErrInvalidTokenError)
		return
	}

	resp, err := h.authService.CompleteProfile(c.Request.Context(), service.CompleteProfileRequest{
		TempToken:    tempToken.Token,
		FullName:     req.FullName,
		Email:        req.Email,
		Gender:       req.Gender,
		ReferredBy:   req.ReferredBy,
		ReferralCode: req.ReferralCode,
		BusinessType: req.BusinessType,
		Source:       req.Source,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// SetPINRequest represents set PIN request body
type SetPINRequest struct {
	PIN        string `json:"pin" binding:"required,len=6"`
	PINConfirm string `json:"pinConfirm" binding:"required,len=6"`
	DeviceID   string `json:"deviceId" binding:"required"`
	DeviceName string `json:"deviceName" binding:"required"`
}

// SetPIN handles POST /v1/auth/set-pin
func (h *AuthHandler) SetPIN(c *gin.Context) {
	var req SetPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	tempToken := middleware.GetTempToken(c)
	if tempToken == nil {
		respondWithError(c, domain.ErrInvalidTokenError)
		return
	}

	resp, err := h.authService.SetPIN(c.Request.Context(), service.SetPINRequest{
		TempToken:  tempToken.Token,
		PIN:        req.PIN,
		PINConfirm: req.PINConfirm,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
		IPAddress:  c.ClientIP(),
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// PINLoginRequest represents PIN login request body
type PINLoginRequest struct {
	Phone      string `json:"phone" binding:"required"`
	PIN        string `json:"pin" binding:"required,len=6"`
	DeviceID   string `json:"deviceId" binding:"required"`
	DeviceName string `json:"deviceName" binding:"required"`
}

// PINLogin handles POST /v1/auth/pin-login
func (h *AuthHandler) PINLogin(c *gin.Context) {
	var req PINLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	resp, err := h.authService.PINLogin(c.Request.Context(), service.PINLoginRequest{
		Phone:      req.Phone,
		PIN:        req.PIN,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
		IPAddress:  c.ClientIP(),
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// VerifyPINOnlyRequest represents verify PIN only request body
type VerifyPINOnlyRequest struct {
	PIN      string  `json:"pin" binding:"required,len=6"`
	DeviceID *string `json:"deviceId"`
}

// VerifyPINOnly handles POST /v1/auth/verify-pin-only
func (h *AuthHandler) VerifyPINOnly(c *gin.Context) {
	var req VerifyPINOnlyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	resp, err := h.authService.VerifyPINOnly(c.Request.Context(), service.VerifyPINOnlyRequest{
		UserID:   userID,
		PIN:      req.PIN,
		DeviceID: req.DeviceID,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// RefreshTokenRequest represents refresh token request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshToken handles POST /v1/auth/refresh-token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), service.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
		IPAddress:    c.ClientIP(),
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// LogoutRequest represents logout request body
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
	DeviceID     string `json:"deviceId"`
	LogoutAll    bool   `json:"logoutAll"`
}

// Logout handles POST /v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	resp, err := h.authService.Logout(c.Request.Context(), service.LogoutRequest{
		UserID:       userID,
		RefreshToken: req.RefreshToken,
		DeviceID:     req.DeviceID,
		LogoutAll:    req.LogoutAll,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// ListDevices handles GET /v1/auth/devices
func (h *AuthHandler) ListDevices(c *gin.Context) {
	userID := middleware.GetUserID(c)
	deviceID := middleware.GetDeviceID(c)

	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	resp, err := h.authService.ListDevices(c.Request.Context(), userID, deviceID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// RemoveDevice handles DELETE /v1/auth/devices/:deviceId
func (h *AuthHandler) RemoveDevice(c *gin.Context) {
	userID := middleware.GetUserID(c)
	currentDeviceID := middleware.GetDeviceID(c)
	targetDeviceID := c.Param("deviceId")

	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	resp, err := h.authService.RemoveDevice(c.Request.Context(), service.RemoveDeviceRequest{
		UserID:          userID,
		CurrentDeviceID: currentDeviceID,
		TargetDeviceID:  targetDeviceID,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// ChangePINVerifyCurrentRequest represents change PIN verify current request body
type ChangePINVerifyCurrentRequest struct {
	CurrentPIN string  `json:"currentPin" binding:"required,len=6"`
	DeviceID   *string `json:"deviceId"`
}

// ChangePINVerifyCurrent handles POST /v1/auth/change-pin/verify-current
func (h *AuthHandler) ChangePINVerifyCurrent(c *gin.Context) {
	var req ChangePINVerifyCurrentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	resp, err := h.authService.ChangePINVerifyCurrent(c.Request.Context(), service.ChangePINVerifyCurrentRequest{
		UserID:     userID,
		CurrentPIN: req.CurrentPIN,
		DeviceID:   req.DeviceID,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// ChangePINConfirmRequest represents change PIN confirm request body
type ChangePINConfirmRequest struct {
	NewPIN        string  `json:"newPin" binding:"required,len=6"`
	NewPINConfirm string  `json:"newPinConfirm" binding:"required,len=6"`
	DeviceID      *string `json:"deviceId"`
}

// ChangePINConfirm handles POST /v1/auth/change-pin/confirm
func (h *AuthHandler) ChangePINConfirm(c *gin.Context) {
	var req ChangePINConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	tempToken := middleware.GetTempToken(c)
	if tempToken == nil {
		respondWithError(c, domain.ErrInvalidTokenError)
		return
	}

	resp, err := h.authService.ChangePINConfirm(c.Request.Context(), service.ChangePINConfirmRequest{
		TempToken:     tempToken.Token,
		NewPIN:        req.NewPIN,
		NewPINConfirm: req.NewPINConfirm,
		DeviceID:      req.DeviceID,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// ChangePhoneRequestOTPOldRequest represents change phone request OTP old request body
type ChangePhoneRequestOTPOldRequest struct {
	OTPMethod string `json:"otpMethod" binding:"omitempty,oneof=wa sms"`
}

// ChangePhoneRequestOTPOld handles POST /v1/auth/change-phone/verify-old/request-otp
func (h *AuthHandler) ChangePhoneRequestOTPOld(c *gin.Context) {
	var req ChangePhoneRequestOTPOldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	otpMethod := req.OTPMethod
	if otpMethod == "" {
		otpMethod = "wa"
	}

	resp, err := h.authService.ChangePhoneRequestOTPOld(c.Request.Context(), service.ChangePhoneRequestOTPOldRequest{
		UserID:    userID,
		OTPMethod: otpMethod,
		IPAddress: c.ClientIP(),
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// ChangePhoneVerifyOldRequest represents change phone verify old request body
type ChangePhoneVerifyOldRequest struct {
	OTP          string `json:"otp" binding:"required,len=4"`
	OTPSessionID string `json:"otpSessionId" binding:"required"`
}

// ChangePhoneVerifyOld handles POST /v1/auth/change-phone/verify-old
func (h *AuthHandler) ChangePhoneVerifyOld(c *gin.Context) {
	var req ChangePhoneVerifyOldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	resp, err := h.authService.ChangePhoneVerifyOld(c.Request.Context(), service.ChangePhoneVerifyOldRequest{
		UserID:       userID,
		OTP:          req.OTP,
		OTPSessionID: req.OTPSessionID,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// ChangePhoneRequestOTPNewRequest represents change phone request OTP new request body
type ChangePhoneRequestOTPNewRequest struct {
	NewPhone  string `json:"newPhone" binding:"required"`
	OTPMethod string `json:"otpMethod" binding:"required,oneof=wa sms"`
}

// ChangePhoneRequestOTPNew handles POST /v1/auth/change-phone/new/request-otp
func (h *AuthHandler) ChangePhoneRequestOTPNew(c *gin.Context) {
	var req ChangePhoneRequestOTPNewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	tempToken := middleware.GetTempToken(c)
	if tempToken == nil {
		respondWithError(c, domain.ErrInvalidTokenError)
		return
	}

	resp, err := h.authService.ChangePhoneRequestOTPNew(c.Request.Context(), service.ChangePhoneRequestOTPNewRequest{
		TempToken: tempToken.Token,
		NewPhone:  req.NewPhone,
		OTPMethod: req.OTPMethod,
		IPAddress: c.ClientIP(),
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// ChangePhoneVerifyNewRequest represents change phone verify new request body
type ChangePhoneVerifyNewRequest struct {
	NewPhone     string `json:"newPhone" binding:"required"`
	OTP          string `json:"otp" binding:"required,len=4"`
	OTPSessionID string `json:"otpSessionId" binding:"required"`
}

// ChangePhoneVerifyNew handles POST /v1/auth/change-phone/new/verify-otp
func (h *AuthHandler) ChangePhoneVerifyNew(c *gin.Context) {
	var req ChangePhoneVerifyNewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	tempToken := middleware.GetTempToken(c)
	if tempToken == nil {
		respondWithError(c, domain.ErrInvalidTokenError)
		return
	}

	resp, err := h.authService.ChangePhoneVerifyNew(c.Request.Context(), service.ChangePhoneVerifyNewRequest{
		TempToken:    tempToken.Token,
		NewPhone:     req.NewPhone,
		OTP:          req.OTP,
		OTPSessionID: req.OTPSessionID,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Notify user about phone number change
	notificationErr := h.authService.NotifyPhoneNumberChange(c.Request.Context(), service.NotifyPhoneNumberChangeRequest{
		UserID:     middleware.GetUserID(c),
		OldPhone:   tempToken.Phone, // Assuming tempToken contains the old phone number
		NewPhone:   req.NewPhone,
		ChangeTime: time.Now().Format(time.RFC3339),
	})

	if notificationErr != nil {
		// Log the notification error but don't fail the request
		slog.Warn("failed to send phone number change notification",
			slog.String("user_id", middleware.GetUserID(c)),
			slog.String("error", notificationErr.Error()),
		)
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// RequestEmailVerificationRequest represents request email verification request body
type RequestEmailVerificationRequest struct {
	Email *string `json:"email" binding:"omitempty,email"`
}

// RequestEmailVerification handles POST /v1/auth/email/request-verification
func (h *AuthHandler) RequestEmailVerification(c *gin.Context) {
	var req RequestEmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	resp, err := h.authService.RequestEmailVerification(c.Request.Context(), service.RequestEmailVerificationRequest{
		UserID: userID,
		Email:  req.Email,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// VerifyEmail handles GET /v1/auth/email/verify
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		respondWithError(c, domain.ErrValidationFailed("Token is required"))
		return
	}

	resp, err := h.authService.VerifyEmail(c.Request.Context(), service.VerifyEmailRequest{
		Token: token,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}
