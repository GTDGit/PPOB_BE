package service

import (
	"context"
	"fmt"
	"strings"
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

	// Build verification link pointing to frontend app
	// Frontend will call backend API to verify the token
	verificationLink := fmt.Sprintf("%s/email-verify?token=%s&email=%s", s.cfg.FrontendURL, token, email)

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

// SendAdminInvite sends an admin invitation email.
func (s *EmailService) SendAdminInvite(ctx context.Context, email, name, inviteLink, roleName string) error {
	if !s.brevoClient.IsEnabled() {
		return nil
	}

	displayName := name
	if strings.TrimSpace(displayName) == "" {
		displayName = "Admin"
	}

	html := fmt.Sprintf(`
		<html>
			<body style="font-family:Arial,Helvetica,sans-serif;background:#f5f8ff;padding:24px;">
				<div style="max-width:560px;margin:0 auto;background:#ffffff;border-radius:16px;padding:32px;">
					<h2 style="margin:0 0 12px;color:#1849d6;">Undangan Admin PPOB.ID</h2>
					<p style="margin:0 0 16px;color:#334155;">Halo %s, Anda diundang sebagai <strong>%s</strong> untuk mengakses console admin PPOB.ID.</p>
					<p style="margin:0 0 24px;color:#475569;">Klik tombol di bawah untuk aktivasi akun, buat password, lalu setup authenticator.</p>
					<p style="margin:0 0 24px;">
						<a href="%s" style="display:inline-block;background:#2563eb;color:#ffffff;text-decoration:none;padding:12px 20px;border-radius:10px;font-weight:600;">Aktivasi Admin</a>
					</p>
					<p style="margin:0;color:#64748b;font-size:13px;">Jika tombol tidak bekerja, buka tautan ini:</p>
					<p style="margin:8px 0 0;font-size:13px;word-break:break-all;color:#2563eb;">%s</p>
				</div>
			</body>
		</html>
	`, displayName, roleName, inviteLink, inviteLink)

	_, err := s.brevoClient.SendRawEmail(ctx, email, displayName, "Undangan Admin PPOB.ID", html)
	if err != nil {
		return fmt.Errorf("failed to send admin invite email: %w", err)
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
