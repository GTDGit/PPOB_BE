package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/GTDGit/PPOB_BE/internal/config"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/external/fazpass"
	"github.com/GTDGit/PPOB_BE/internal/external/whatsapp"
	"github.com/GTDGit/PPOB_BE/pkg/redis"
	"github.com/GTDGit/PPOB_BE/pkg/validator"
)

// OTPService handles OTP generation and verification
type OTPService struct {
	cfg            config.OTPConfig
	redisClient    *redis.Client
	whatsappClient *whatsapp.Client
	fazpassClient  *fazpass.Client
}

// NewOTPService creates a new OTP service
func NewOTPService(
	redisClient *redis.Client,
	cfg config.OTPConfig,
	whatsappCfg config.WhatsAppConfig,
	fazpassCfg config.FazpassConfig,
) *OTPService {
	return &OTPService{
		cfg:            cfg,
		redisClient:    redisClient,
		whatsappClient: whatsapp.NewClient(whatsappCfg),
		fazpassClient:  fazpass.NewClient(fazpassCfg),
	}
}

// SendOTPRequest represents OTP send request
type SendOTPRequest struct {
	Phone     string
	Flow      string // "REGISTER", "LOGIN", "RESET_PIN", "CHANGE_PHONE"
	DeviceID  string
	IPAddress string
	OTPMethod string // "wa" or "sms"
}

// SendOTPResponse represents OTP send response
type SendOTPResponse struct {
	SessionID    string
	Phone        string
	ExpiresIn    int // seconds
	ResendIn     int // seconds
	AttemptsLeft int
	Channel      string // "wa" or "sms"
	ResendCount  int
}

// VerifyOTPRequest represents OTP verification request
type VerifyOTPRequest struct {
	SessionID string
	Phone     string
	OTP       string
}

// VerifyOTPResponse represents OTP verification response
type VerifyOTPResponse struct {
	Valid     bool
	Flow      string
	TempToken string // Temporary token for next step
}

// SendOTP generates and sends OTP
func (s *OTPService) SendOTP(ctx context.Context, req SendOTPRequest) (*SendOTPResponse, error) {
	// Normalize phone number
	phone := validator.NormalizePhone(req.Phone)

	// Check rate limit
	if err := s.checkRateLimit(ctx, phone); err != nil {
		return nil, err
	}

	// Generate OTP (4 digits as per spec)
	otp, err := s.generateOTP()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Generate session ID
	sessionID := "otp_" + uuid.New().String()[:12]

	// Create OTP session
	session := &domain.OTPSession{
		SessionID:   sessionID,
		Phone:       phone,
		OTP:         otp,
		OTPMethod:   req.OTPMethod,
		Flow:        req.Flow,
		Attempts:    0,
		MaxAttempts: s.cfg.MaxAttempts,
		DeviceID:    req.DeviceID,
		IPAddress:   req.IPAddress,
		ExpiresAt:   time.Now().Add(s.cfg.TTL),
		CreatedAt:   time.Now(),
	}

	// Store session in Redis
	key := redis.OTPSessionKey(phone, sessionID)
	if err := s.redisClient.SetJSON(ctx, key, session, s.cfg.TTL); err != nil {
		return nil, fmt.Errorf("failed to store OTP session: %w", err)
	}

	// Get resend count
	resendKey := redis.OTPResendCountKey(phone)
	resendCount, _ := s.redisClient.Get(ctx, resendKey).Int()
	s.redisClient.Incr(ctx, resendKey)
	s.redisClient.Expire(ctx, resendKey, time.Hour)

	// Send OTP via WhatsApp (primary) or SMS (fallback)
	channel := "wa"
	if req.OTPMethod == "sms" {
		channel = "sms"
	}

	if channel == "wa" && s.whatsappClient.IsEnabled() {
		resp, err := s.whatsappClient.SendOTP(ctx, whatsapp.SendOTPRequest{
			Phone: phone,
			OTP:   otp,
		})
		if err != nil || !resp.Success {
			// Fallback to SMS
			channel = "sms"
			if s.fazpassClient.IsEnabled() {
				_, err = s.fazpassClient.SendOTP(ctx, fazpass.SendOTPRequest{
					Phone: phone,
					OTP:   otp,
				})
				if err != nil {
					return nil, domain.ErrOTPSendFailed
				}
			} else {
				return nil, domain.ErrOTPSendFailed
			}
		}
	} else if s.fazpassClient.IsEnabled() {
		channel = "sms"
		_, err := s.fazpassClient.SendOTP(ctx, fazpass.SendOTPRequest{
			Phone: phone,
			OTP:   otp,
		})
		if err != nil {
			return nil, domain.ErrOTPSendFailed
		}
	} else {
		// Development mode - don't send, just log (masked)
		slog.Debug("OTP generated for development",
			slog.String("phone_masked", maskPhone(phone)),
			// SECURITY: Never log actual OTP value
		)
	}

	// Update rate limit
	s.updateRateLimit(ctx, phone)

	return &SendOTPResponse{
		SessionID:    sessionID,
		Phone:        validator.MaskPhone(phone),
		ExpiresIn:    int(s.cfg.TTL.Seconds()),
		ResendIn:     int(s.cfg.ResendCooldown.Seconds()),
		AttemptsLeft: s.cfg.MaxAttempts,
		Channel:      channel,
		ResendCount:  resendCount + 1,
	}, nil
}

// VerifyOTP verifies the OTP
func (s *OTPService) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*VerifyOTPResponse, error) {
	// Normalize phone
	phone := validator.NormalizePhone(req.Phone)

	// Get OTP session from Redis
	key := redis.OTPSessionKey(phone, req.SessionID)
	var session domain.OTPSession
	if err := s.redisClient.GetJSON(ctx, key, &session); err != nil {
		return nil, domain.ErrOTPInvalid
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		s.redisClient.Del(ctx, key)
		return nil, domain.ErrOTPExpired
	}

	// Check attempts
	if session.Attempts >= session.MaxAttempts {
		s.redisClient.Del(ctx, key)
		return nil, domain.ErrOTPMaxAttempts
	}

	// Verify OTP
	if session.OTP != req.OTP {
		// Increment attempts
		session.Attempts++
		s.redisClient.SetJSON(ctx, key, &session, time.Until(session.ExpiresAt))

		remaining := session.MaxAttempts - session.Attempts
		return nil, domain.ErrWithRemainingAttempts(domain.ErrInvalidOTPError, remaining)
	}

	// OTP is valid - delete session
	s.redisClient.Del(ctx, key)

	// Generate temporary token for next step
	tempToken := s.createTempTokenInternal(ctx, phone, session.Flow, s.getNextStep(session.Flow), session.DeviceID, session.DeviceName, "")

	return &VerifyOTPResponse{
		Valid:     true,
		Flow:      session.Flow,
		TempToken: tempToken,
	}, nil
}

// ResendOTP resends OTP to the same session
func (s *OTPService) ResendOTP(ctx context.Context, phone, sessionID, otpMethod string) (*SendOTPResponse, error) {
	// Normalize phone
	phone = validator.NormalizePhone(phone)

	// Check rate limit
	if err := s.checkRateLimit(ctx, phone); err != nil {
		return nil, err
	}

	// Get existing session
	key := redis.OTPSessionKey(phone, sessionID)
	var session domain.OTPSession
	if err := s.redisClient.GetJSON(ctx, key, &session); err != nil {
		return nil, domain.ErrOTPInvalid
	}

	// Check resend count
	resendKey := redis.OTPResendCountKey(phone)
	count, _ := s.redisClient.Get(ctx, resendKey).Int()
	if count >= s.cfg.MaxResend {
		return nil, domain.ErrOTPResendLimit
	}

	// Generate new OTP
	otp, err := s.generateOTP()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Update session
	session.OTP = otp
	session.Attempts = 0
	session.ExpiresAt = time.Now().Add(s.cfg.TTL)
	if otpMethod != "" {
		session.OTPMethod = otpMethod
	}

	// Store updated session
	if err := s.redisClient.SetJSON(ctx, key, &session, s.cfg.TTL); err != nil {
		return nil, fmt.Errorf("failed to update OTP session: %w", err)
	}

	// Increment resend count
	if _, err := s.redisClient.Incr(ctx, resendKey).Result(); err != nil {
		// Log but don't fail the main operation
		slog.Warn("failed to increment resend counter",
			slog.String("key", resendKey),
			slog.String("error", err.Error()),
		)
	}

	// Send OTP
	channel := session.OTPMethod
	if channel == "" || channel == "wa" {
		if s.whatsappClient.IsEnabled() {
			resp, err := s.whatsappClient.SendOTP(ctx, whatsapp.SendOTPRequest{
				Phone: phone,
				OTP:   otp,
			})
			if err != nil || !resp.Success {
				channel = "sms"
				if s.fazpassClient.IsEnabled() {
					_, err = s.fazpassClient.SendOTP(ctx, fazpass.SendOTPRequest{
						Phone: phone,
						OTP:   otp,
					})
					if err != nil {
						return nil, domain.ErrOTPSendFailed
					}
				} else {
					return nil, domain.ErrOTPSendFailed
				}
			}
		}
	} else if s.fazpassClient.IsEnabled() {
		_, err := s.fazpassClient.SendOTP(ctx, fazpass.SendOTPRequest{
			Phone: phone,
			OTP:   otp,
		})
		if err != nil {
			return nil, domain.ErrOTPSendFailed
		}
	} else {
		// Development mode - don't send, just log (masked)
		slog.Debug("OTP generated for development",
			slog.String("phone_masked", maskPhone(phone)),
			// SECURITY: Never log actual OTP value
		)
	}

	// Update rate limit
	s.updateRateLimit(ctx, phone)

	return &SendOTPResponse{
		SessionID:    sessionID,
		Phone:        validator.MaskPhone(phone),
		ExpiresIn:    int(s.cfg.TTL.Seconds()),
		ResendIn:     int(s.cfg.ResendCooldown.Seconds()),
		AttemptsLeft: s.cfg.MaxAttempts,
		Channel:      channel,
		ResendCount:  count + 1,
	}, nil
}

// GetTempToken retrieves and validates a temporary token
func (s *OTPService) GetTempToken(ctx context.Context, token string) (*domain.TempToken, error) {
	key := redis.TempTokenKey(token)
	var tempToken domain.TempToken
	if err := s.redisClient.GetJSON(ctx, key, &tempToken); err != nil {
		return nil, domain.ErrInvalidToken
	}

	if time.Now().Unix() > tempToken.ExpiresAt {
		s.redisClient.Del(ctx, key)
		return nil, domain.ErrTempTokenExpiredError
	}

	return &tempToken, nil
}

// CreateTempToken creates a new temporary token
func (s *OTPService) CreateTempToken(ctx context.Context, phone, flow, step, deviceID, deviceName, userID string) (string, error) {
	return s.createTempTokenInternal(ctx, phone, flow, step, deviceID, deviceName, userID), nil
}

func (s *OTPService) createTempTokenInternal(ctx context.Context, phone, flow, step, deviceID, deviceName, userID string) string {
	token := "tmp_" + uuid.New().String()[:12]
	tempToken := &domain.TempToken{
		Token:      token,
		UserID:     userID,
		Phone:      phone,
		Flow:       flow,
		Step:       step,
		DeviceID:   deviceID,
		DeviceName: deviceName,
		ExpiresAt:  time.Now().Add(15 * time.Minute).Unix(),
		CreatedAt:  time.Now().Unix(),
	}

	key := redis.TempTokenKey(token)
	s.redisClient.SetJSON(ctx, key, tempToken, 15*time.Minute)

	return token
}

// InvalidateTempToken removes a temporary token
func (s *OTPService) InvalidateTempToken(ctx context.Context, token string) error {
	key := redis.TempTokenKey(token)
	return s.redisClient.Del(ctx, key).Err()
}

// generateOTP generates a random OTP based on configured length
func (s *OTPService) generateOTP() (string, error) {
	// Get length from config (s.cfg is already config.OTPConfig)
	length := s.cfg.Length
	if length < 4 {
		length = 4 // Minimum 4 digits for security
	}
	if length > 8 {
		length = 8 // Maximum 8 digits (practical limit for SMS)
	}

	// Calculate max value: 10^length
	maxVal := int64(1)
	for i := 0; i < length; i++ {
		maxVal *= 10
	}

	// Use crypto/rand for cryptographically secure random numbers
	max := big.NewInt(maxVal)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// Format with leading zeros (e.g., %04d for 4 digits, %06d for 6 digits)
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n.Int64()), nil
}

// checkRateLimit checks if the phone number is rate limited
func (s *OTPService) checkRateLimit(ctx context.Context, phone string) error {
	key := redis.OTPRateLimitKey(phone)
	lastSent, err := s.redisClient.Get(ctx, key).Result()
	if err == nil && lastSent != "" {
		lastTime, _ := time.Parse(time.RFC3339, lastSent)
		if time.Since(lastTime) < s.cfg.ResendCooldown {
			return domain.ErrOTPRateLimited
		}
	}
	return nil
}

// updateRateLimit updates the rate limit timestamp
func (s *OTPService) updateRateLimit(ctx context.Context, phone string) {
	key := redis.OTPRateLimitKey(phone)
	s.redisClient.Set(ctx, key, time.Now().Format(time.RFC3339), s.cfg.ResendCooldown)
}

// getNextStep returns the next step after OTP verification based on flow
func (s *OTPService) getNextStep(flow string) string {
	switch flow {
	case domain.FlowRegister:
		return domain.StepCompleteProfile
	case domain.FlowLogin:
		return domain.StepAuthenticated
	case domain.FlowResetPIN:
		return domain.StepSetPIN
	case domain.FlowChangePhone:
		return domain.StepVerifyNewPhone
	case domain.FlowChangePIN:
		return domain.StepSetPIN
	default:
		return ""
	}
}

// maskPhone masks phone number for safe logging: 081234567890 -> 0812****7890
func maskPhone(phone string) string {
	if len(phone) < 8 {
		return "****"
	}
	return phone[:4] + "****" + phone[len(phone)-4:]
}
