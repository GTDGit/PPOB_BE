package service

import (
	"context"
	"fmt"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/config"
	"github.com/GTDGit/PPOB_BE/internal/external/brevo"
)

// EmailService handles email operations
type EmailService struct {
	cfg         config.BrevoConfig
	brevoClient *brevo.Client
}

// NewEmailService creates a new email service
func NewEmailService(cfg config.BrevoConfig) *EmailService {
	return &EmailService{
		cfg:         cfg,
		brevoClient: brevo.NewClient(cfg),
	}
}

// SendVerificationEmail sends email verification link
func (s *EmailService) SendVerificationEmail(ctx context.Context, email, name, token string) error {
	if !s.brevoClient.IsEnabled() {
		// Development mode - skip sending
		return nil // Skip if not configured
	}

	// Build verification link
	verificationLink := fmt.Sprintf("%s/v1/auth/email/verify?token=%s", s.cfg.BaseURL, token)

	// Send email
	_, err := s.brevoClient.SendVerificationEmail(ctx, email, name, verificationLink)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// SendNewLoginAlert sends alert for new device login
func (s *EmailService) SendNewLoginAlert(ctx context.Context, email, name, deviceName, ipAddress string) error {
	if !s.brevoClient.IsEnabled() {
		// Development mode - skip sending
		return nil
	}

	loginTime := time.Now().Format("02 Jan 2006 15:04 WIB")
	_, err := s.brevoClient.SendNewLoginAlert(ctx, email, name, deviceName, ipAddress, loginTime)
	if err != nil {
		return fmt.Errorf("failed to send new login alert: %w", err)
	}

	return nil
}

// SendPINChangedAlert sends alert for PIN change
func (s *EmailService) SendPINChangedAlert(ctx context.Context, email, name string) error {
	if !s.brevoClient.IsEnabled() {
		// Development mode - skip sending
		return nil
	}

	changeTime := time.Now().Format("02 Jan 2006 15:04 WIB")
	_, err := s.brevoClient.SendPINChangedAlert(ctx, email, name, changeTime)
	if err != nil {
		return fmt.Errorf("failed to send PIN changed alert: %w", err)
	}

	return nil
}

// SendPhoneChangedAlert sends alert for phone change
func (s *EmailService) SendPhoneChangedAlert(ctx context.Context, email, name, oldPhone, newPhone string) error {
	if !s.brevoClient.IsEnabled() {
		// Development mode - skip sending
		return nil
	}

	changeTime := time.Now().Format("02 Jan 2006 15:04 WIB")
	_, err := s.brevoClient.SendPhoneChangedAlert(ctx, email, name, oldPhone, newPhone, changeTime)
	if err != nil {
		return fmt.Errorf("failed to send phone changed alert: %w", err)
	}

	return nil
}

// SendEmailChangedAlert sends alert for email change to old email
func (s *EmailService) SendEmailChangedAlert(ctx context.Context, oldEmail, name, newEmail string) error {
	if !s.brevoClient.IsEnabled() {
		// Development mode - skip sending
		return nil
	}

	changeTime := time.Now().Format("02 Jan 2006 15:04 WIB")
	_, err := s.brevoClient.SendEmailChangedAlert(ctx, oldEmail, name, newEmail, changeTime)
	if err != nil {
		return fmt.Errorf("failed to send email changed alert: %w", err)
	}

	return nil
}
