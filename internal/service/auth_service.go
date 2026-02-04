package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/GTDGit/PPOB_BE/internal/config"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/GTDGit/PPOB_BE/pkg/hash"
	"github.com/GTDGit/PPOB_BE/pkg/jwt"
	"github.com/GTDGit/PPOB_BE/pkg/redis"
	"github.com/GTDGit/PPOB_BE/pkg/validator"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo     repository.UserRepository
	deviceRepo   repository.DeviceRepository
	sessionRepo  repository.SessionRepository
	balanceRepo  repository.BalanceRepository
	settingsRepo repository.UserSettingsRepository
	jwtGen       *jwt.Generator
	otpService   *OTPService
	emailService *EmailService
	redisClient  *redis.Client
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	deviceRepo repository.DeviceRepository,
	sessionRepo repository.SessionRepository,
	balanceRepo repository.BalanceRepository,
	settingsRepo repository.UserSettingsRepository,
	otpService *OTPService,
	emailService *EmailService,
	redisClient *redis.Client,
	jwtConfig config.JWTConfig,
) *AuthService {
	jwtGen := jwt.NewGenerator(jwtConfig.Secret, jwtConfig.AccessTTL, jwtConfig.RefreshTTL)
	return &AuthService{
		userRepo:     userRepo,
		deviceRepo:   deviceRepo,
		sessionRepo:  sessionRepo,
		balanceRepo:  balanceRepo,
		settingsRepo: settingsRepo,
		jwtGen:       jwtGen,
		otpService:   otpService,
		emailService: emailService,
		redisClient:  redisClient,
	}
}

// StartAuthRequest represents start auth request
type StartAuthRequest struct {
	Phone      string
	DeviceID   string
	DeviceName string
	OTPMethod  string
	IPAddress  string
}

// StartAuth handles the initiation of authentication
func (s *AuthService) StartAuth(ctx context.Context, req StartAuthRequest) (*domain.StartAuthResponse, error) {
	// Validate phone number
	if !validator.ValidatePhone(req.Phone) {
		return nil, domain.ErrInvalidPhone
	}

	// Normalize phone number
	phone := validator.NormalizePhone(req.Phone)

	// Check if user exists
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check user: %w", err)
	}

	// Determine flow
	flow := domain.FlowRegister
	if user != nil {
		flow = domain.FlowLogin

		// Check if device is recognized
		device, _ := s.deviceRepo.FindByUserIDAndDeviceID(ctx, user.ID, req.DeviceID)
		if device != nil && device.IsActive {
			// Device recognized - can use PIN login directly
			return &domain.StartAuthResponse{
				Step:        domain.StepInputPIN,
				Flow:        flow,
				MaskedPhone: validator.MaskPhone(phone),
				UserName:    stringPtr(user.GetDisplayName()),
			}, nil
		}
	}

	// Send OTP
	otpResp, err := s.otpService.SendOTP(ctx, SendOTPRequest{
		Phone:     phone,
		Flow:      flow,
		DeviceID:  req.DeviceID,
		IPAddress: req.IPAddress,
		OTPMethod: req.OTPMethod,
	})
	if err != nil {
		return nil, err
	}

	return &domain.StartAuthResponse{
		Step:         domain.StepVerifyOTP,
		Flow:         flow,
		OTPSessionID: otpResp.SessionID,
		ExpiresIn:    otpResp.ExpiresIn,
		OTPMethod:    otpResp.Channel,
		MaskedPhone:  validator.MaskPhone(phone),
	}, nil
}

// VerifyOTPAuthRequest represents OTP verification request
type VerifyOTPAuthRequest struct {
	Phone        string
	OTP          string
	OTPSessionID string
	DeviceID     string
	DeviceName   string
	IPAddress    string
}

// VerifyOTP verifies OTP and returns appropriate next step
func (s *AuthService) VerifyOTP(ctx context.Context, req VerifyOTPAuthRequest) (*domain.VerifyOTPResponse, error) {
	// Validate OTP format
	if !validator.ValidateOTP(req.OTP) {
		return nil, domain.ErrOTPInvalid
	}

	// Verify OTP
	otpResp, err := s.otpService.VerifyOTP(ctx, VerifyOTPRequest{
		SessionID: req.OTPSessionID,
		Phone:     req.Phone,
		OTP:       req.OTP,
	})
	if err != nil {
		return nil, err
	}

	if !otpResp.Valid {
		return nil, domain.ErrOTPInvalid
	}

	// Check if user exists
	phone := validator.NormalizePhone(req.Phone)
	user, _ := s.userRepo.FindByPhone(ctx, phone)

	if user == nil {
		// New user - need to complete profile
		return &domain.VerifyOTPResponse{
			Step:      domain.StepCompleteProfile,
			TempToken: otpResp.TempToken,
		}, nil
	}

	// Existing user - check profile completion
	if user.FullName == "" {
		return &domain.VerifyOTPResponse{
			Step:      domain.StepCompleteProfile,
			TempToken: otpResp.TempToken,
		}, nil
	}

	// Check if PIN is set
	if user.PINHash == "" {
		return &domain.VerifyOTPResponse{
			Step:      domain.StepSetPIN,
			TempToken: otpResp.TempToken,
		}, nil
	}

	// User has PIN, complete login
	return s.completeLogin(ctx, user, req.DeviceID, req.DeviceName, req.IPAddress)
}

// completeLogin finalizes the login process
func (s *AuthService) completeLogin(ctx context.Context, user *domain.User, deviceID, deviceName, ipAddress string) (*domain.VerifyOTPResponse, error) {
	// Update/create device
	now := time.Now()
	device := &domain.Device{
		ID:           "dev_" + uuid.New().String()[:8],
		UserID:       user.ID,
		DeviceID:     deviceID,
		DeviceName:   deviceName,
		IPAddress:    &ipAddress,
		IsActive:     true,
		LastActiveAt: &now,
		CreatedAt:    time.Now(),
	}
	if err := s.deviceRepo.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to create/update device: %w", err)
	}

	// Generate tokens
	tokens, err := s.jwtGen.GenerateTokenPair(user.ID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := &domain.Session{
		ID:               "ses_" + uuid.New().String()[:8],
		UserID:           user.ID,
		DeviceID:         deviceID,
		RefreshTokenHash: hash.HashToken(tokens.RefreshToken),
		ExpiresAt:        time.Now().Add(s.jwtGen.GetRefreshTTL()),
		CreatedAt:        time.Now(),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	s.userRepo.UpdateLastLogin(ctx, user.ID)

	return &domain.VerifyOTPResponse{
		Step:         domain.StepAuthenticated,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User:         user.ToResponse(),
	}, nil
}

// ResendOTPRequest represents resend OTP request
type ResendOTPRequest struct {
	Phone        string
	OTPSessionID string
	OTPMethod    string
}

// ResendOTP resends OTP
func (s *AuthService) ResendOTP(ctx context.Context, req ResendOTPRequest) (*domain.ResendOTPResponse, error) {
	resp, err := s.otpService.ResendOTP(ctx, req.Phone, req.OTPSessionID, req.OTPMethod)
	if err != nil {
		return nil, err
	}

	return &domain.ResendOTPResponse{
		Step:         domain.StepVerifyOTP,
		OTPSessionID: resp.SessionID,
		ExpiresIn:    resp.ExpiresIn,
		OTPMethod:    resp.Channel,
		ResendCount:  resp.ResendCount,
		MaxResend:    3, // from config
		NextResendAt: time.Now().Add(60 * time.Second).Format(time.RFC3339),
	}, nil
}

// CompleteProfileRequest represents complete profile request
type CompleteProfileRequest struct {
	TempToken    string
	FullName     string
	Email        *string
	Gender       string
	ReferredBy   *string
	ReferralCode *string
	BusinessType string
	Source       string
}

// CompleteProfile completes user profile
func (s *AuthService) CompleteProfile(ctx context.Context, req CompleteProfileRequest) (*domain.CompleteProfileResponse, error) {
	// Get temp token
	tempToken, err := s.otpService.GetTempToken(ctx, req.TempToken)
	if err != nil {
		return nil, err
	}

	// Check/create user
	phone := tempToken.Phone
	user, _ := s.userRepo.FindByPhone(ctx, phone)

	if user == nil {
		// Create new user
		user = &domain.User{
			ID:               "usr_" + uuid.New().String()[:8],
			MIC:              "PID" + uuid.New().String()[:8],
			Phone:            phone,
			FullName:         req.FullName,
			Gender:           &req.Gender,
			Tier:             domain.TierBasic,
			KYCStatus:        domain.KYCStatusUnverified,
			BusinessType:     &req.BusinessType,
			Source:           &req.Source,
			ReferredBy:       req.ReferredBy,
			UsedReferralCode: req.ReferralCode,
			IsActive:         true,
			PhoneVerifiedAt:  sql.NullTime{Time: time.Now(), Valid: true},
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		if req.Email != nil && *req.Email != "" {
			user.Email = sql.NullString{String: *req.Email, Valid: true}
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		// Create initial balance
		balance := &domain.Balance{
			ID:     "bal_" + uuid.New().String()[:8],
			UserID: user.ID,
			Amount: 0,
		}
		if err := s.balanceRepo.Create(ctx, balance); err != nil {
			return nil, fmt.Errorf("failed to create balance: %w", err)
		}

		// Create default settings
		settings := repository.CreateDefaultSettings(user.ID)
		if err := s.settingsRepo.Create(ctx, settings); err != nil {
			return nil, fmt.Errorf("failed to create settings: %w", err)
		}
	} else {
		// Update existing user
		user.FullName = req.FullName
		user.Gender = &req.Gender
		user.BusinessType = &req.BusinessType
		user.Source = &req.Source
		if req.Email != nil && *req.Email != "" {
			user.Email = sql.NullString{String: *req.Email, Valid: true}
		}
		user.UpdatedAt = time.Now()

		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Generate new temp token for SET_PIN step
	newTempToken, err := s.otpService.CreateTempToken(ctx, tempToken.Phone, domain.FlowRegister, domain.StepSetPIN, tempToken.DeviceID, tempToken.DeviceName, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp token: %w", err)
	}

	// Invalidate old temp token
	s.otpService.InvalidateTempToken(ctx, req.TempToken)

	return &domain.CompleteProfileResponse{
		Step:      domain.StepSetPIN,
		TempToken: newTempToken,
	}, nil
}

// SetPINRequest represents set PIN request
type SetPINRequest struct {
	TempToken  string
	PIN        string
	PINConfirm string
	DeviceID   string
	DeviceName string
	IPAddress  string
}

// SetPIN sets PIN for user
func (s *AuthService) SetPIN(ctx context.Context, req SetPINRequest) (*domain.SetPINResponse, error) {
	// Validate PIN
	if !validator.ValidatePIN(req.PIN) {
		return nil, domain.ErrInvalidPINError
	}

	// Check confirmation
	if req.PIN != req.PINConfirm {
		return nil, domain.ErrPINMismatchError
	}

	// Check weak PIN
	if hash.IsWeakPIN(req.PIN) {
		return nil, domain.ErrWeakPINError
	}

	// Get temp token
	tempToken, err := s.otpService.GetTempToken(ctx, req.TempToken)
	if err != nil {
		return nil, err
	}

	// Get user
	var user *domain.User
	if tempToken.UserID != "" {
		user, err = s.userRepo.FindByID(ctx, tempToken.UserID)
	} else {
		user, err = s.userRepo.FindByPhone(ctx, tempToken.Phone)
	}
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Hash and set PIN
	hashedPIN, err := hash.HashPIN(req.PIN)
	if err != nil {
		return nil, fmt.Errorf("failed to hash PIN: %w", err)
	}

	if err := s.userRepo.UpdatePIN(ctx, user.ID, hashedPIN); err != nil {
		return nil, fmt.Errorf("failed to update PIN: %w", err)
	}

	// Invalidate temp token
	s.otpService.InvalidateTempToken(ctx, req.TempToken)

	// Complete login
	resp, err := s.completeLogin(ctx, user, req.DeviceID, req.DeviceName, req.IPAddress)
	if err != nil {
		return nil, err
	}

	return &domain.SetPINResponse{
		Step:         domain.StepAuthenticated,
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		User:         resp.User,
	}, nil
}

// PINLoginRequest represents PIN login request
type PINLoginRequest struct {
	Phone      string
	PIN        string
	DeviceID   string
	DeviceName string
	IPAddress  string
}

// PINLogin handles PIN-based login
func (s *AuthService) PINLogin(ctx context.Context, req PINLoginRequest) (*domain.PINLoginResponse, error) {
	// Normalize phone
	phone := validator.NormalizePhone(req.Phone)

	// Get user
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil || user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Check if PIN is locked
	if s.isPINLocked(ctx, user.ID) {
		lockUntil := s.getPINLockUntil(ctx, user.ID)
		return nil, domain.ErrWithLockUntil(domain.ErrPINLockedError, lockUntil)
	}

	// Verify PIN
	if !hash.VerifyPIN(req.PIN, user.PINHash) {
		attempts := s.incrementPINAttempts(ctx, user.ID)
		remaining := 5 - attempts
		if remaining <= 0 {
			s.lockPIN(ctx, user.ID)
			lockUntil := s.getPINLockUntil(ctx, user.ID)
			return nil, domain.ErrWithLockUntil(domain.ErrPINLockedError, lockUntil)
		}
		return nil, domain.ErrWithRemainingAttempts(domain.ErrInvalidPINError, remaining)
	}

	// Clear PIN attempts
	s.clearPINAttempts(ctx, user.ID)

	// Complete login
	now := time.Now()
	device := &domain.Device{
		ID:           "dev_" + uuid.New().String()[:8],
		UserID:       user.ID,
		DeviceID:     req.DeviceID,
		DeviceName:   req.DeviceName,
		IPAddress:    &req.IPAddress,
		IsActive:     true,
		LastActiveAt: &now,
		CreatedAt:    time.Now(),
	}
	if err := s.deviceRepo.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to create/update device: %w", err)
	}

	// Generate tokens
	tokens, err := s.jwtGen.GenerateTokenPair(user.ID, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := &domain.Session{
		ID:               "ses_" + uuid.New().String()[:8],
		UserID:           user.ID,
		DeviceID:         req.DeviceID,
		RefreshTokenHash: hash.HashToken(tokens.RefreshToken),
		ExpiresAt:        time.Now().Add(s.jwtGen.GetRefreshTTL()),
		CreatedAt:        time.Now(),
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	s.userRepo.UpdateLastLogin(ctx, user.ID)

	return &domain.PINLoginResponse{
		Step:         domain.StepAuthenticated,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User:         user.ToResponse(),
	}, nil
}

// VerifyPINOnlyRequest represents verify PIN only request
type VerifyPINOnlyRequest struct {
	UserID   string
	PIN      string
	DeviceID *string
}

// VerifyPINOnly verifies PIN without generating new tokens
func (s *AuthService) VerifyPINOnly(ctx context.Context, req VerifyPINOnlyRequest) (*domain.VerifyPINOnlyResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Check if PIN is locked
	if s.isPINLocked(ctx, user.ID) {
		lockUntil := s.getPINLockUntil(ctx, user.ID)
		return nil, domain.ErrWithLockUntil(domain.ErrPINLockedError, lockUntil)
	}

	// Verify PIN
	if !hash.VerifyPIN(req.PIN, user.PINHash) {
		attempts := s.incrementPINAttempts(ctx, user.ID)
		remaining := 5 - attempts
		if remaining <= 0 {
			s.lockPIN(ctx, user.ID)
			lockUntil := s.getPINLockUntil(ctx, user.ID)
			return nil, domain.ErrWithLockUntil(domain.ErrPINLockedError, lockUntil)
		}
		return nil, domain.ErrWithRemainingAttempts(domain.ErrInvalidPINError, remaining)
	}

	// Clear PIN attempts
	s.clearPINAttempts(ctx, user.ID)

	return &domain.VerifyPINOnlyResponse{
		Step:       domain.StepPINVerified,
		Message:    "PIN_VALID",
		VerifiedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string
	IPAddress    string
}

// RefreshToken refreshes access token
func (s *AuthService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*domain.RefreshTokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwtGen.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidRefreshTokenError
	}

	// Check session
	tokenHash := hash.HashToken(req.RefreshToken)
	session, err := s.sessionRepo.FindByRefreshTokenHash(ctx, tokenHash)
	if err != nil || session == nil || session.IsRevoked {
		return nil, domain.ErrInvalidRefreshTokenError
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		return nil, domain.ErrExpiredTokenError
	}

	// Generate new tokens (rotation)
	tokens, err := s.jwtGen.GenerateTokenPair(claims.UserID, claims.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update session
	session.RefreshTokenHash = hash.HashToken(tokens.RefreshToken)
	session.ExpiresAt = time.Now().Add(s.jwtGen.GetRefreshTTL())
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &domain.RefreshTokenResponse{
		Step:         domain.StepTokenRefreshed,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// LogoutRequest represents logout request
type LogoutRequest struct {
	UserID       string
	RefreshToken string
	DeviceID     string
	LogoutAll    bool
}

// Logout logs out user
func (s *AuthService) Logout(ctx context.Context, req LogoutRequest) (*domain.LogoutResponse, error) {
	devicesLoggedOut := 0

	if req.LogoutAll {
		// Revoke all sessions
		count, err := s.sessionRepo.RevokeByUserID(ctx, req.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to revoke sessions: %w", err)
		}
		devicesLoggedOut = int(count)
	} else {
		// Revoke session by user and device
		if err := s.sessionRepo.RevokeByUserIDAndDeviceID(ctx, req.UserID, req.DeviceID); err != nil {
			return nil, fmt.Errorf("failed to revoke session: %w", err)
		}
		devicesLoggedOut = 1
	}

	return &domain.LogoutResponse{
		Step:             domain.StepLoggedOut,
		Message:          "LOGOUT_SUCCESS",
		DevicesLoggedOut: devicesLoggedOut,
	}, nil
}

// ListDevices lists user devices
func (s *AuthService) ListDevices(ctx context.Context, userID, currentDeviceID string) (*domain.ListDevicesResponse, error) {
	devices, err := s.deviceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	deviceResponses := make([]*domain.DeviceResponse, len(devices))
	for i, d := range devices {
		deviceResponses[i] = d.ToResponse(currentDeviceID)
	}

	return &domain.ListDevicesResponse{
		Step:         domain.StepDevicesListed,
		Devices:      deviceResponses,
		TotalDevices: len(devices),
	}, nil
}

// RemoveDeviceRequest represents remove device request
type RemoveDeviceRequest struct {
	UserID          string
	CurrentDeviceID string
	TargetDeviceID  string
}

// RemoveDevice removes a device
func (s *AuthService) RemoveDevice(ctx context.Context, req RemoveDeviceRequest) (*domain.RemoveDeviceResponse, error) {
	if req.CurrentDeviceID == req.TargetDeviceID {
		return nil, domain.ErrCannotRemoveCurrentDeviceError
	}

	if err := s.deviceRepo.DeleteByUserIDAndDeviceID(ctx, req.UserID, req.TargetDeviceID); err != nil {
		return nil, fmt.Errorf("failed to remove device: %w", err)
	}

	// Also revoke sessions for this device
	s.sessionRepo.RevokeByUserIDAndDeviceID(ctx, req.UserID, req.TargetDeviceID)

	return &domain.RemoveDeviceResponse{
		Step:     domain.StepDeviceRemoved,
		Message:  "DEVICE_REMOVED_SUCCESS",
		DeviceID: req.TargetDeviceID,
	}, nil
}

// ChangePINVerifyCurrentRequest represents change PIN verify current request
type ChangePINVerifyCurrentRequest struct {
	UserID     string
	CurrentPIN string
	DeviceID   *string
}

// ChangePINVerifyCurrent verifies current PIN for PIN change
func (s *AuthService) ChangePINVerifyCurrent(ctx context.Context, req ChangePINVerifyCurrentRequest) (*domain.ChangePINVerifyResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Check if PIN is locked
	if s.isPINLocked(ctx, user.ID) {
		lockUntil := s.getPINLockUntil(ctx, user.ID)
		return nil, domain.ErrWithLockUntil(domain.ErrPINLockedError, lockUntil)
	}

	// Verify PIN
	if !hash.VerifyPIN(req.CurrentPIN, user.PINHash) {
		attempts := s.incrementPINAttempts(ctx, user.ID)
		remaining := 5 - attempts
		if remaining <= 0 {
			s.lockPIN(ctx, user.ID)
			lockUntil := s.getPINLockUntil(ctx, user.ID)
			return nil, domain.ErrWithLockUntil(domain.ErrPINLockedError, lockUntil)
		}
		return nil, domain.ErrWithRemainingAttempts(domain.ErrInvalidPINError, remaining)
	}

	// Clear PIN attempts
	s.clearPINAttempts(ctx, user.ID)

	// Create temp token for PIN change
	deviceID := ""
	if req.DeviceID != nil {
		deviceID = *req.DeviceID
	}
	tempToken, err := s.otpService.CreateTempToken(ctx, user.Phone, domain.FlowChangePIN, domain.StepChangePINConfirm, deviceID, "", user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp token: %w", err)
	}

	return &domain.ChangePINVerifyResponse{
		Step:      domain.StepChangePINConfirm,
		TempToken: tempToken,
		ExpiresIn: 900,
	}, nil
}

// ChangePINConfirmRequest represents change PIN confirm request
type ChangePINConfirmRequest struct {
	TempToken     string
	NewPIN        string
	NewPINConfirm string
	DeviceID      *string
}

// ChangePINConfirm confirms new PIN
func (s *AuthService) ChangePINConfirm(ctx context.Context, req ChangePINConfirmRequest) (*domain.ChangePINConfirmResponse, error) {
	// Validate PIN
	if !validator.ValidatePIN(req.NewPIN) {
		return nil, domain.ErrInvalidPINError
	}

	// Check confirmation
	if req.NewPIN != req.NewPINConfirm {
		return nil, domain.ErrPINMismatchError
	}

	// Check weak PIN
	if hash.IsWeakPIN(req.NewPIN) {
		return nil, domain.ErrWeakPINError
	}

	// Get temp token
	tempToken, err := s.otpService.GetTempToken(ctx, req.TempToken)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, tempToken.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Hash and update PIN
	hashedPIN, err := hash.HashPIN(req.NewPIN)
	if err != nil {
		return nil, fmt.Errorf("failed to hash PIN: %w", err)
	}

	if err := s.userRepo.UpdatePIN(ctx, user.ID, hashedPIN); err != nil {
		return nil, fmt.Errorf("failed to update PIN: %w", err)
	}

	// Invalidate temp token
	s.otpService.InvalidateTempToken(ctx, req.TempToken)

	// Send notification
	if user.Email.Valid {
		s.emailService.SendPINChangedAlert(ctx, user.Email.String, user.GetDisplayName())
	}

	return &domain.ChangePINConfirmResponse{
		Step:    domain.StepPINChanged,
		Message: "PIN_CHANGED_SUCCESS",
	}, nil
}

// ChangePhoneRequestOTPOldRequest represents change phone request OTP for old phone
type ChangePhoneRequestOTPOldRequest struct {
	UserID    string
	OTPMethod string
	IPAddress string
}

// ChangePhoneRequestOTPOld requests OTP for old phone verification
func (s *AuthService) ChangePhoneRequestOTPOld(ctx context.Context, req ChangePhoneRequestOTPOldRequest) (*domain.ChangePhoneRequestOTPResponse, error) {
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	otpResp, err := s.otpService.SendOTP(ctx, SendOTPRequest{
		Phone:     user.Phone,
		Flow:      domain.FlowChangePhone,
		DeviceID:  "",
		IPAddress: req.IPAddress,
		OTPMethod: req.OTPMethod,
	})
	if err != nil {
		return nil, err
	}

	return &domain.ChangePhoneRequestOTPResponse{
		Step:         domain.StepVerifyOldPhone,
		OTPSessionID: otpResp.SessionID,
		ExpiresIn:    otpResp.ExpiresIn,
		OTPMethod:    otpResp.Channel,
		MaskedPhone:  validator.MaskPhone(user.Phone),
	}, nil
}

// ChangePhoneVerifyOldRequest represents change phone verify old request
type ChangePhoneVerifyOldRequest struct {
	UserID       string
	OTP          string
	OTPSessionID string
}

// ChangePhoneVerifyOld verifies OTP for old phone
func (s *AuthService) ChangePhoneVerifyOld(ctx context.Context, req ChangePhoneVerifyOldRequest) (*domain.ChangePhoneVerifyOldResponse, error) {
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	otpResp, err := s.otpService.VerifyOTP(ctx, VerifyOTPRequest{
		SessionID: req.OTPSessionID,
		Phone:     user.Phone,
		OTP:       req.OTP,
	})
	if err != nil {
		return nil, err
	}

	if !otpResp.Valid {
		return nil, domain.ErrOTPInvalid
	}

	// Create temp token for new phone step
	tempToken, err := s.otpService.CreateTempToken(ctx, user.Phone, domain.FlowChangePhone, domain.StepVerifyNewPhone, "", "", user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp token: %w", err)
	}

	return &domain.ChangePhoneVerifyOldResponse{
		Step:      domain.StepVerifyNewPhone,
		TempToken: tempToken,
		ExpiresIn: 900,
	}, nil
}

// ChangePhoneRequestOTPNewRequest represents change phone request OTP for new phone
type ChangePhoneRequestOTPNewRequest struct {
	TempToken string
	NewPhone  string
	OTPMethod string
	IPAddress string
}

// ChangePhoneRequestOTPNew requests OTP for new phone
func (s *AuthService) ChangePhoneRequestOTPNew(ctx context.Context, req ChangePhoneRequestOTPNewRequest) (*domain.ChangePhoneRequestOTPResponse, error) {
	// Validate temp token
	_, err := s.otpService.GetTempToken(ctx, req.TempToken)
	if err != nil {
		return nil, err
	}

	// Validate new phone
	if !validator.ValidatePhone(req.NewPhone) {
		return nil, domain.ErrInvalidPhone
	}

	newPhone := validator.NormalizePhone(req.NewPhone)

	// Check if phone is already registered
	existingUser, _ := s.userRepo.FindByPhone(ctx, newPhone)
	if existingUser != nil {
		return nil, domain.ErrPhoneAlreadyRegisteredError
	}

	// Send OTP to new phone
	otpResp, err := s.otpService.SendOTP(ctx, SendOTPRequest{
		Phone:     newPhone,
		Flow:      domain.FlowChangePhone,
		DeviceID:  "",
		IPAddress: req.IPAddress,
		OTPMethod: req.OTPMethod,
	})
	if err != nil {
		return nil, err
	}

	return &domain.ChangePhoneRequestOTPResponse{
		Step:         domain.StepVerifyNewPhoneOTP,
		OTPSessionID: otpResp.SessionID,
		ExpiresIn:    otpResp.ExpiresIn,
		OTPMethod:    otpResp.Channel,
		MaskedPhone:  validator.MaskPhone(newPhone),
	}, nil
}

// ChangePhoneVerifyNewRequest represents change phone verify new request
type ChangePhoneVerifyNewRequest struct {
	TempToken    string
	NewPhone     string
	OTP          string
	OTPSessionID string
}

// ChangePhoneVerifyNew verifies OTP for new phone and completes phone change
func (s *AuthService) ChangePhoneVerifyNew(ctx context.Context, req ChangePhoneVerifyNewRequest) (*domain.ChangePhoneCompleteResponse, error) {
	// Get temp token
	tempToken, err := s.otpService.GetTempToken(ctx, req.TempToken)
	if err != nil {
		return nil, err
	}

	newPhone := validator.NormalizePhone(req.NewPhone)

	// Verify OTP
	otpResp, err := s.otpService.VerifyOTP(ctx, VerifyOTPRequest{
		SessionID: req.OTPSessionID,
		Phone:     newPhone,
		OTP:       req.OTP,
	})
	if err != nil {
		return nil, err
	}

	if !otpResp.Valid {
		return nil, domain.ErrOTPInvalid
	}

	// Update phone number
	if err := s.userRepo.UpdatePhone(ctx, tempToken.UserID, newPhone); err != nil {
		return nil, fmt.Errorf("failed to update phone: %w", err)
	}

	// Invalidate temp token
	s.otpService.InvalidateTempToken(ctx, req.TempToken)

	return &domain.ChangePhoneCompleteResponse{
		Step:        domain.StepPhoneChanged,
		Message:     "PHONE_CHANGED_SUCCESS",
		Phone:       newPhone,
		MaskedPhone: validator.MaskPhone(newPhone),
	}, nil
}

// NotifyPhoneNumberChangeRequest represents notification request
type NotifyPhoneNumberChangeRequest struct {
	UserID     string
	OldPhone   string
	NewPhone   string
	ChangeTime string
}

// NotifyPhoneNumberChange sends notification about phone number change
func (s *AuthService) NotifyPhoneNumberChange(ctx context.Context, req NotifyPhoneNumberChangeRequest) error {
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	if user.Email.Valid {
		return s.emailService.SendPhoneChangedAlert(ctx, user.Email.String, user.GetDisplayName(), req.OldPhone, req.NewPhone)
	}

	return nil
}

// RequestEmailVerificationRequest represents email verification request
type RequestEmailVerificationRequest struct {
	UserID string
	Email  *string
}

// RequestEmailVerification requests email verification
func (s *AuthService) RequestEmailVerification(ctx context.Context, req RequestEmailVerificationRequest) (*domain.EmailVerificationSentResponse, error) {
	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	email := ""
	if req.Email != nil && *req.Email != "" {
		email = *req.Email
	} else if user.Email.Valid {
		email = user.Email.String
	}

	if email == "" {
		return nil, domain.ErrValidationFailed("Email is required")
	}

	// Generate verification token
	token := uuid.New().String()

	// Store token in Redis (24 hours)
	key := redis.EmailVerificationKey(token)
	data := map[string]string{
		"userID": user.ID,
		"email":  email,
	}
	if err := s.redisClient.SetJSON(ctx, key, data, 24*time.Hour); err != nil {
		return nil, fmt.Errorf("failed to store verification token: %w", err)
	}

	// Send verification email
	if err := s.emailService.SendVerificationEmail(ctx, email, user.GetDisplayName(), token); err != nil {
		return nil, fmt.Errorf("failed to send verification email: %w", err)
	}

	return &domain.EmailVerificationSentResponse{
		Step:        "EMAIL_VERIFICATION_SENT",
		Message:     "VERIFICATION_EMAIL_SENT",
		MaskedEmail: validator.MaskEmail(email),
		ExpiresIn:   86400,
	}, nil
}

// VerifyEmailRequest represents verify email request
type VerifyEmailRequest struct {
	Token string
}

// VerifyEmail verifies email with token
func (s *AuthService) VerifyEmail(ctx context.Context, req VerifyEmailRequest) (*domain.EmailVerifiedResponse, error) {
	// Get token data
	key := redis.EmailVerificationKey(req.Token)
	var data map[string]string
	if err := s.redisClient.GetJSON(ctx, key, &data); err != nil {
		return nil, domain.NewError("INVALID_VERIFICATION_TOKEN", "Link verifikasi tidak valid atau sudah kadaluarsa", 400)
	}

	userID := data["userID"]
	email := data["email"]

	// Update email verification
	if err := s.userRepo.VerifyEmail(ctx, userID, email); err != nil {
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}

	// Delete token
	s.redisClient.Del(ctx, key)

	return &domain.EmailVerifiedResponse{
		Step:       "EMAIL_VERIFIED",
		Message:    "EMAIL_VERIFIED_SUCCESS",
		Email:      email,
		VerifiedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// PIN attempt helpers
func (s *AuthService) isPINLocked(ctx context.Context, userID string) bool {
	key := redis.PINLockKey(userID)
	exists, _ := s.redisClient.Exists(ctx, key).Result()
	return exists > 0
}

func (s *AuthService) getPINLockUntil(ctx context.Context, userID string) string {
	key := redis.PINLockKey(userID)
	ttl, _ := s.redisClient.TTL(ctx, key).Result()
	return time.Now().Add(ttl).Format(time.RFC3339)
}

func (s *AuthService) incrementPINAttempts(ctx context.Context, userID string) int {
	key := redis.PINAttemptsKey(userID)
	count, _ := s.redisClient.Incr(ctx, key).Result()
	s.redisClient.Expire(ctx, key, 30*time.Minute)
	return int(count)
}

func (s *AuthService) clearPINAttempts(ctx context.Context, userID string) {
	key := redis.PINAttemptsKey(userID)
	s.redisClient.Del(ctx, key)
}

func (s *AuthService) lockPIN(ctx context.Context, userID string) {
	key := redis.PINLockKey(userID)
	s.redisClient.Set(ctx, key, "1", 15*time.Minute)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
