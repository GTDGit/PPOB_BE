package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/config"
	internalbrevo "github.com/GTDGit/PPOB_BE/internal/external/brevo"
	internalses "github.com/GTDGit/PPOB_BE/internal/external/ses"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/google/uuid"
)

const (
	emailProviderBrevo = "brevo"
	emailProviderSES   = "ses"
)

type EmailService struct {
	emailCfg    config.EmailConfig
	brevoCfg    config.BrevoConfig
	brevoClient *internalbrevo.Client
	sesClient   *internalses.Client
	adminRepo   *repository.AdminRepository
}

type MailReplyRequest struct {
	Category          string
	MailboxID         string
	ThreadID          string
	MessageID         string
	FromAddress       string
	FromName          string
	ToAddresses       []string
	CcAddresses       []string
	BccAddresses      []string
	ReplyToAddresses  []string
	Subject           string
	HTMLBody          string
	TextBody          string
	Headers           map[string]string
	ConfigurationSet  string
	Tags              map[string]string
}

func NewEmailService(emailCfg config.EmailConfig, brevoCfg config.BrevoConfig, adminRepo *repository.AdminRepository) (*EmailService, error) {
	service := &EmailService{
		emailCfg:    emailCfg,
		brevoCfg:    brevoCfg,
		brevoClient: internalbrevo.NewClient(brevoCfg),
		adminRepo:   adminRepo,
	}

	if strings.EqualFold(emailCfg.Provider, emailProviderSES) {
		sesClient, err := internalses.NewClient(emailCfg.SES)
		if err != nil {
			return nil, err
		}
		service.sesClient = sesClient
	}

	return service, nil
}

func (s *EmailService) SendVerificationEmail(ctx context.Context, email, name, token string) error {
	verificationLink := fmt.Sprintf("%s/email-verify?token=%s&email=%s", s.brevoCfg.FrontendURL, token, email)

	if s.useBrevo() {
		if !s.brevoClient.IsEnabled() {
			return nil
		}

		resp, err := s.brevoClient.SendVerificationEmail(ctx, email, name, verificationLink)
		s.logTransactionalDispatch(ctx, "verification_email", "", email, s.brevoCfg.SenderEmail, s.brevoCfg.SenderName, brevoMessageID(resp), statusFromProvider(s.providerName()), err, map[string]interface{}{
			"name": name,
		})
		if err != nil {
			return fmt.Errorf("failed to send verification email: %w", err)
		}
		return nil
	}

	return s.sendTransactionalEmail(ctx, transactionalSendRequest{
		Category:    "verification_email",
		ToEmail:     email,
		ToName:      name,
		Subject:     "Verifikasi Email Anda",
		ActionLabel: "Verifikasi Email",
		ActionLink:  verificationLink,
		Headline:    "Verifikasi email akun Anda",
		Intro:       fmt.Sprintf("Halo %s, klik tombol di bawah untuk memverifikasi email akun PPOB.ID Anda.", fallbackDisplayName(name, "Pelanggan PPOB.ID")),
		Secondary:   "Jika Anda tidak merasa meminta verifikasi ini, abaikan email ini.",
		TextLines: []string{
			"Halo " + fallbackDisplayName(name, "Pelanggan PPOB.ID"),
			"Klik tautan berikut untuk memverifikasi email akun Anda:",
			verificationLink,
			"Jika Anda tidak merasa meminta verifikasi ini, abaikan email ini.",
		},
	})
}

func (s *EmailService) SendNewLoginAlert(ctx context.Context, email, name, deviceName, ipAddress string) error {
	loginTime := time.Now().Format("02 Jan 2006 15:04 WIB")

	if s.useBrevo() {
		if !s.brevoClient.IsEnabled() {
			return nil
		}

		resp, err := s.brevoClient.SendNewLoginAlert(ctx, email, name, deviceName, ipAddress, loginTime)
		s.logTransactionalDispatch(ctx, "new_login_alert", "", email, s.brevoCfg.SenderEmail, s.brevoCfg.SenderName, brevoMessageID(resp), statusFromProvider(s.providerName()), err, map[string]interface{}{
			"deviceName": deviceName,
			"ipAddress":  ipAddress,
		})
		if err != nil {
			return fmt.Errorf("failed to send new login alert: %w", err)
		}
		return nil
	}

	return s.sendTransactionalEmail(ctx, transactionalSendRequest{
		Category:  "new_login_alert",
		ToEmail:   email,
		ToName:    name,
		Subject:   "Peringatan Keamanan: Login Baru Terdeteksi",
		Headline:  "Login baru terdeteksi",
		Intro:     fmt.Sprintf("Halo %s, kami mendeteksi login baru ke akun PPOB.ID Anda.", fallbackDisplayName(name, "Pelanggan PPOB.ID")),
		Secondary: fmt.Sprintf("Perangkat: %s | IP: %s | Waktu: %s", fallbackDisplayName(deviceName, "-"), fallbackDisplayName(ipAddress, "-"), loginTime),
		TextLines: []string{
			"Halo " + fallbackDisplayName(name, "Pelanggan PPOB.ID"),
			"Kami mendeteksi login baru ke akun PPOB.ID Anda.",
			"Perangkat: " + fallbackDisplayName(deviceName, "-"),
			"IP: " + fallbackDisplayName(ipAddress, "-"),
			"Waktu: " + loginTime,
		},
	})
}

func (s *EmailService) SendPINChangedAlert(ctx context.Context, email, name string) error {
	changeTime := time.Now().Format("02 Jan 2006 15:04 WIB")

	if s.useBrevo() {
		if !s.brevoClient.IsEnabled() {
			return nil
		}

		resp, err := s.brevoClient.SendPINChangedAlert(ctx, email, name, changeTime)
		s.logTransactionalDispatch(ctx, "pin_changed_alert", "", email, s.brevoCfg.SenderEmail, s.brevoCfg.SenderName, brevoMessageID(resp), statusFromProvider(s.providerName()), err, nil)
		if err != nil {
			return fmt.Errorf("failed to send PIN changed alert: %w", err)
		}
		return nil
	}

	return s.sendTransactionalEmail(ctx, transactionalSendRequest{
		Category:  "pin_changed_alert",
		ToEmail:   email,
		ToName:    name,
		Subject:   "PIN PPOB.ID Berhasil Diubah",
		Headline:  "PIN akun berhasil diubah",
		Intro:     fmt.Sprintf("Halo %s, PIN akun PPOB.ID Anda berhasil diubah.", fallbackDisplayName(name, "Pelanggan PPOB.ID")),
		Secondary: "Waktu perubahan: " + changeTime,
		TextLines: []string{
			"Halo " + fallbackDisplayName(name, "Pelanggan PPOB.ID"),
			"PIN akun PPOB.ID Anda berhasil diubah.",
			"Waktu perubahan: " + changeTime,
		},
	})
}

func (s *EmailService) SendPhoneChangedAlert(ctx context.Context, email, name, oldPhone, newPhone string) error {
	changeTime := time.Now().Format("02 Jan 2006 15:04 WIB")

	if s.useBrevo() {
		if !s.brevoClient.IsEnabled() {
			return nil
		}

		resp, err := s.brevoClient.SendPhoneChangedAlert(ctx, email, name, oldPhone, newPhone, changeTime)
		s.logTransactionalDispatch(ctx, "phone_changed_alert", "", email, s.brevoCfg.SenderEmail, s.brevoCfg.SenderName, brevoMessageID(resp), statusFromProvider(s.providerName()), err, nil)
		if err != nil {
			return fmt.Errorf("failed to send phone changed alert: %w", err)
		}
		return nil
	}

	return s.sendTransactionalEmail(ctx, transactionalSendRequest{
		Category:  "phone_changed_alert",
		ToEmail:   email,
		ToName:    name,
		Subject:   "Nomor Telepon PPOB.ID Berhasil Diubah",
		Headline:  "Nomor telepon berhasil diubah",
		Intro:     fmt.Sprintf("Halo %s, nomor telepon akun PPOB.ID Anda berhasil diubah.", fallbackDisplayName(name, "Pelanggan PPOB.ID")),
		Secondary: fmt.Sprintf("Nomor lama: %s | Nomor baru: %s | Waktu: %s", oldPhone, newPhone, changeTime),
		TextLines: []string{
			"Halo " + fallbackDisplayName(name, "Pelanggan PPOB.ID"),
			"Nomor telepon akun PPOB.ID Anda berhasil diubah.",
			"Nomor lama: " + oldPhone,
			"Nomor baru: " + newPhone,
			"Waktu perubahan: " + changeTime,
		},
	})
}

func (s *EmailService) SendAdminInvite(ctx context.Context, email, name, inviteLink, roleName string) error {
	displayName := fallbackDisplayName(name, "Admin")
	subject := "Undangan Admin PPOB.ID"
	html := buildActionEmailHTML("Undangan Admin PPOB.ID", displayName,
		fmt.Sprintf("Anda diundang sebagai %s untuk mengakses console admin PPOB.ID.", roleName),
		"Aktivasi Admin",
		inviteLink,
		"Setelah membuka tautan ini, Anda akan membuat password dan mengaktifkan authenticator.",
	)
	text := strings.Join([]string{
		"Halo " + displayName,
		"Anda diundang sebagai " + roleName + " untuk mengakses console admin PPOB.ID.",
		"Buka tautan berikut untuk aktivasi admin:",
		inviteLink,
	}, "\n")

	return s.sendCustomEmail(ctx, sendCustomEmailRequest{
		Category:  "admin_invite",
		ToEmail:   email,
		ToName:    displayName,
		Subject:   subject,
		HTMLBody:  html,
		TextBody:  text,
		ReplyTo:   []string{s.emailCfg.ReplyToEmail},
		ConfigSet: s.emailCfg.SES.ConfigurationSetTransactional,
	})
}

func (s *EmailService) SendAdminPasswordReset(ctx context.Context, email, name, resetLink string) error {
	displayName := fallbackDisplayName(name, "Admin")
	subject := "Reset Password Admin PPOB.ID"
	html := buildActionEmailHTML("Reset Password Console Admin", displayName,
		"Kami menerima permintaan reset password untuk akun admin PPOB.ID Anda.",
		"Reset Password",
		resetLink,
		"Setelah mengganti password, Anda tetap perlu memasukkan authenticator atau recovery code.",
	)
	text := strings.Join([]string{
		"Halo " + displayName,
		"Kami menerima permintaan reset password untuk akun admin PPOB.ID Anda.",
		"Buka tautan berikut untuk reset password:",
		resetLink,
		"Link berlaku selama 30 menit.",
	}, "\n")

	return s.sendCustomEmail(ctx, sendCustomEmailRequest{
		Category:  "admin_password_reset",
		ToEmail:   email,
		ToName:    displayName,
		Subject:   subject,
		HTMLBody:  html,
		TextBody:  text,
		ReplyTo:   []string{s.emailCfg.ReplyToEmail},
		ConfigSet: s.emailCfg.SES.ConfigurationSetTransactional,
	})
}

func (s *EmailService) SendEmailChangedAlert(ctx context.Context, oldEmail, name, newEmail string) error {
	changeTime := time.Now().Format("02 Jan 2006 15:04 WIB")

	if s.useBrevo() {
		if !s.brevoClient.IsEnabled() {
			return nil
		}

		resp, err := s.brevoClient.SendEmailChangedAlert(ctx, oldEmail, name, newEmail, changeTime)
		s.logTransactionalDispatch(ctx, "email_changed_alert", "", oldEmail, s.brevoCfg.SenderEmail, s.brevoCfg.SenderName, brevoMessageID(resp), statusFromProvider(s.providerName()), err, nil)
		if err != nil {
			return fmt.Errorf("failed to send email changed alert: %w", err)
		}
		return nil
	}

	return s.sendTransactionalEmail(ctx, transactionalSendRequest{
		Category:  "email_changed_alert",
		ToEmail:   oldEmail,
		ToName:    name,
		Subject:   "Email PPOB.ID Berhasil Diubah",
		Headline:  "Email akun berhasil diubah",
		Intro:     fmt.Sprintf("Halo %s, email akun PPOB.ID Anda berhasil diubah.", fallbackDisplayName(name, "Pelanggan PPOB.ID")),
		Secondary: fmt.Sprintf("Email baru: %s | Waktu: %s", newEmail, changeTime),
		TextLines: []string{
			"Halo " + fallbackDisplayName(name, "Pelanggan PPOB.ID"),
			"Email akun PPOB.ID Anda berhasil diubah.",
			"Email baru: " + newEmail,
			"Waktu perubahan: " + changeTime,
		},
	})
}

func (s *EmailService) SendMailboxReply(ctx context.Context, req MailReplyRequest) (string, error) {
	if s.useBrevo() {
		return "", fmt.Errorf("mailbox reply is not supported on brevo provider")
	}

	messageID, err := s.sendViaSES(ctx, sesSendRequest{
		Category:     req.Category,
		ToEmail:      req.ToAddresses[0],
		ToName:       "",
		FromEmail:    req.FromAddress,
		FromName:     req.FromName,
		Subject:      req.Subject,
		HTMLBody:     req.HTMLBody,
		TextBody:     req.TextBody,
		ReplyTo:      req.ReplyToAddresses,
		CcAddresses:  req.CcAddresses,
		BccAddresses: req.BccAddresses,
		Headers:      req.Headers,
		ConfigSet:    firstNonEmpty(req.ConfigurationSet, s.emailCfg.SES.ConfigurationSetOperations),
		Tags:         req.Tags,
	})
	if err != nil {
		return "", err
	}

	return messageID, nil
}

type transactionalSendRequest struct {
	Category    string
	ToEmail     string
	ToName      string
	Subject     string
	Headline    string
	Intro       string
	Secondary   string
	ActionLabel string
	ActionLink  string
	TextLines   []string
}

type sendCustomEmailRequest struct {
	Category  string
	ToEmail   string
	ToName    string
	Subject   string
	HTMLBody  string
	TextBody  string
	ReplyTo   []string
	ConfigSet string
}

type sesSendRequest struct {
	Category     string
	ToEmail      string
	ToName       string
	FromEmail    string
	FromName     string
	Subject      string
	HTMLBody     string
	TextBody     string
	ReplyTo      []string
	CcAddresses  []string
	BccAddresses []string
	Headers      map[string]string
	ConfigSet    string
	Tags         map[string]string
}

func (s *EmailService) sendTransactionalEmail(ctx context.Context, req transactionalSendRequest) error {
	html := buildActionEmailHTML(req.Headline, fallbackDisplayName(req.ToName, "Pelanggan PPOB.ID"), req.Intro, req.ActionLabel, req.ActionLink, req.Secondary)
	text := strings.Join(req.TextLines, "\n")

	return s.sendCustomEmail(ctx, sendCustomEmailRequest{
		Category:  req.Category,
		ToEmail:   req.ToEmail,
		ToName:    req.ToName,
		Subject:   req.Subject,
		HTMLBody:  html,
		TextBody:  text,
		ReplyTo:   []string{s.emailCfg.ReplyToEmail},
		ConfigSet: s.emailCfg.SES.ConfigurationSetTransactional,
	})
}

func (s *EmailService) sendCustomEmail(ctx context.Context, req sendCustomEmailRequest) error {
	if s.useBrevo() {
		if !s.brevoClient.IsEnabled() {
			return nil
		}

		resp, err := s.brevoClient.SendRawEmail(ctx, req.ToEmail, fallbackDisplayName(req.ToName, req.ToEmail), req.Subject, req.HTMLBody)
		s.logTransactionalDispatch(ctx, req.Category, "", req.ToEmail, s.brevoCfg.SenderEmail, s.brevoCfg.SenderName, brevoMessageID(resp), statusFromProvider(s.providerName()), err, nil)
		if err != nil {
			return err
		}
		return nil
	}

	messageID, err := s.sendViaSES(ctx, sesSendRequest{
		Category:  req.Category,
		ToEmail:   req.ToEmail,
		ToName:    req.ToName,
		FromEmail: s.emailCfg.DefaultFromEmail,
		FromName:  s.emailCfg.DefaultFromName,
		Subject:   req.Subject,
		HTMLBody:  req.HTMLBody,
		TextBody:  req.TextBody,
		ReplyTo:   req.ReplyTo,
		ConfigSet: req.ConfigSet,
	})
	status := "queued"
	if err != nil {
		status = "failed"
	}
	s.logTransactionalDispatch(ctx, req.Category, "", req.ToEmail, s.emailCfg.DefaultFromEmail, s.emailCfg.DefaultFromName, messageID, status, err, nil)
	if err != nil {
		return err
	}
	return nil
}

func (s *EmailService) sendViaSES(ctx context.Context, req sesSendRequest) (string, error) {
	if s.sesClient == nil || !s.sesClient.IsEnabled() {
		return "", nil
	}

	return s.sesClient.Send(ctx, internalses.SendMessageInput{
		FromAddress:          firstNonEmpty(req.FromEmail, s.emailCfg.DefaultFromEmail),
		FromName:             firstNonEmpty(req.FromName, s.emailCfg.DefaultFromName),
		ToAddresses:          []string{req.ToEmail},
		CcAddresses:          req.CcAddresses,
		BccAddresses:         req.BccAddresses,
		ReplyToAddresses:     req.ReplyTo,
		Subject:              req.Subject,
		HTMLBody:             req.HTMLBody,
		TextBody:             req.TextBody,
		Headers:              req.Headers,
		ConfigurationSetName: req.ConfigSet,
		Tags:                 req.Tags,
	})
}

func (s *EmailService) logTransactionalDispatch(ctx context.Context, category, mailboxID, recipient, senderAddress, senderName, providerMessageID, status string, err error, metadata map[string]interface{}) {
	if s.adminRepo == nil || strings.TrimSpace(recipient) == "" {
		return
	}

	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metadata["provider"] = s.providerName()

	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}

	_ = s.adminRepo.CreateEmailDispatchLog(ctx, map[string]interface{}{
		"id":                "edl_" + uuid.New().String()[:8],
		"category":          category,
		"mailboxId":         nullableString(mailboxID),
		"recipient":         recipient,
		"senderAddress":     senderAddress,
		"senderName":        senderName,
		"provider":          strings.ToUpper(s.providerName()),
		"providerMessageId": nullableString(providerMessageID),
		"status":            firstNonEmpty(status, "queued"),
		"errorMessage":      nullableString(errorMessage),
		"metadata":          metadata,
	})
}

func (s *EmailService) providerName() string {
	if strings.EqualFold(strings.TrimSpace(s.emailCfg.Provider), emailProviderSES) {
		return emailProviderSES
	}
	return emailProviderBrevo
}

func (s *EmailService) useBrevo() bool {
	return s.providerName() == emailProviderBrevo
}

func statusFromProvider(provider string) string {
	if provider == emailProviderBrevo {
		return "sent"
	}
	return "queued"
}

func brevoMessageID(resp *internalbrevo.SendEmailResponse) string {
	if resp == nil {
		return ""
	}
	return strings.TrimSpace(resp.MessageID)
}

func fallbackDisplayName(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func nullableString(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func buildActionEmailHTML(title, name, intro, actionLabel, actionLink, secondary string) string {
	buttonHTML := ""
	if strings.TrimSpace(actionLabel) != "" && strings.TrimSpace(actionLink) != "" {
		buttonHTML = fmt.Sprintf(`
			<p style="margin:0 0 24px;">
				<a href="%s" style="display:inline-block;background:#2563eb;color:#ffffff;text-decoration:none;padding:12px 20px;border-radius:10px;font-weight:600;">%s</a>
			</p>
			<p style="margin:0 0 18px;color:#64748b;font-size:13px;">Jika tombol tidak bekerja, buka tautan ini:</p>
			<p style="margin:0 0 24px;font-size:13px;word-break:break-all;color:#2563eb;">%s</p>
		`, actionLink, actionLabel, actionLink)
	}

	secondaryHTML := ""
	if strings.TrimSpace(secondary) != "" {
		secondaryHTML = fmt.Sprintf(`<p style="margin:18px 0 0;color:#64748b;font-size:13px;line-height:1.6;">%s</p>`, secondary)
	}

	return fmt.Sprintf(`
		<html>
			<body style="font-family:Arial,Helvetica,sans-serif;background:#f5f8ff;padding:24px;">
				<div style="max-width:560px;margin:0 auto;background:#ffffff;border-radius:16px;padding:32px;border:1px solid #dbeafe;">
					<p style="margin:0 0 8px;color:#2563eb;font-size:12px;font-weight:700;letter-spacing:0.12em;text-transform:uppercase;">PPOB.ID</p>
					<h2 style="margin:0 0 16px;color:#1849d6;">%s</h2>
					<p style="margin:0 0 14px;color:#334155;line-height:1.7;">Halo %s,</p>
					<p style="margin:0 0 22px;color:#475569;line-height:1.7;">%s</p>
					%s
					%s
				</div>
			</body>
		</html>
	`, title, name, intro, buttonHTML, secondaryHTML)
}
