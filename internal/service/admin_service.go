package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"mime/multipart"
	"path/filepath"

	"github.com/GTDGit/PPOB_BE/internal/config"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	internals3 "github.com/GTDGit/PPOB_BE/internal/external/s3"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/GTDGit/PPOB_BE/pkg/hash"
	"github.com/GTDGit/PPOB_BE/pkg/jwt"
	"github.com/GTDGit/PPOB_BE/pkg/validator"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type AdminService struct {
	repo         *repository.AdminRepository
	emailService *EmailService
	s3Client     *internals3.Client
	cfg          config.AdminConfig
	jwtGen       *jwt.Generator
}

type CreateAdminInviteRequest struct {
	Email    string
	Phone    string
	FullName string
	RoleID   string
}

func NewAdminService(repo *repository.AdminRepository, emailService *EmailService, s3Client *internals3.Client, cfg config.AdminConfig) *AdminService {
	return &AdminService{
		repo:         repo,
		emailService: emailService,
		s3Client:     s3Client,
		cfg:          cfg,
		jwtGen:       jwt.NewGenerator(cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL),
	}
}

func (s *AdminService) BootstrapFirstAdmin(ctx context.Context, secret, email, phone, fullName string) (map[string]interface{}, error) {
	if s.cfg.BootstrapSecret == "" || secret != s.cfg.BootstrapSecret {
		return nil, domain.NewError("BOOTSTRAP_UNAUTHORIZED", "Bootstrap secret tidak valid", 401)
	}

	totalAdmins, err := s.repo.CountAdmins(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count admins: %w", err)
	}
	if totalAdmins > 0 {
		return nil, domain.NewError("BOOTSTRAP_DISABLED", "Bootstrap hanya untuk setup pertama", 409)
	}

	return s.CreateInvite(ctx, "", CreateAdminInviteRequest{
		Email:    email,
		Phone:    phone,
		FullName: fullName,
		RoleID:   firstNonEmpty(s.cfg.BootstrapRoleID, "super_admin"),
	})
}

func (s *AdminService) Login(ctx context.Context, email, password, totpCode, ipAddress, userAgent string) (*domain.AdminAuthResponse, error) {
	if !validator.ValidateEmail(email) {
		return nil, domain.ErrValidationFailed("Email admin tidak valid")
	}

	admin, err := s.repo.FindAdminByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
	if err != nil {
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil {
		return nil, domain.NewError("ADMIN_NOT_FOUND", "Akun admin tidak ditemukan", 404)
	}
	if !admin.IsActive || admin.Status != domain.AdminStatusActive {
		return nil, domain.NewError("ADMIN_DISABLED", "Akun admin belum aktif atau dinonaktifkan", 403)
	}
	if !admin.PasswordHash.Valid || admin.PasswordHash.String == "" {
		return nil, domain.NewError("ADMIN_PASSWORD_NOT_SET", "Password admin belum diatur", 403)
	}
	if !hash.VerifyPIN(password, admin.PasswordHash.String) {
		return nil, domain.NewError("ADMIN_LOGIN_FAILED", "Email atau password salah", 401)
	}

	totpSecret, err := s.repo.FindTOTPSecretByAdminID(ctx, admin.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get totp secret: %w", err)
	}
	if totpSecret == nil || !totpSecret.ConfirmedAt.Valid {
		return nil, domain.NewError("ADMIN_TOTP_NOT_READY", "Authenticator admin belum aktif", 403)
	}
	if !validateTOTPCode(totpCode, totpSecret.Secret) {
		return nil, domain.NewError("ADMIN_TOTP_INVALID", "Kode authenticator salah", 401)
	}

	resp, err := s.createAuthSession(ctx, admin, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	_ = s.repo.UpdateLastLogin(ctx, admin.ID)
	_ = s.logAudit(ctx, admin.ID, "admin.login", "admin_user", admin.ID, nil, map[string]interface{}{
		"ipAddress": ipAddress,
	}, ipAddress, userAgent, "success", nil)

	return resp, nil
}

func (s *AdminService) Refresh(ctx context.Context, refreshToken, ipAddress, userAgent string) (*domain.AdminAuthResponse, error) {
	claims, err := s.jwtGen.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, domain.NewError("ADMIN_REFRESH_INVALID", "Refresh token admin tidak valid", 401)
	}

	session, err := s.repo.FindSessionByID(ctx, claims.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin session: %w", err)
	}
	if session == nil || session.AdminUserID != claims.UserID || session.RefreshTokenHash != hash.HashToken(refreshToken) {
		return nil, domain.NewError("ADMIN_SESSION_INVALID", "Sesi admin tidak valid", 401)
	}

	admin, err := s.repo.FindAdminByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil || !admin.IsActive || admin.Status != domain.AdminStatusActive {
		return nil, domain.NewError("ADMIN_DISABLED", "Akun admin tidak aktif", 403)
	}

	tokens, err := s.jwtGen.GenerateTokenPair(admin.ID, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	if err := s.repo.UpdateSessionRefreshToken(ctx, session.ID, hash.HashToken(tokens.RefreshToken), time.Now().Add(s.jwtGen.GetRefreshTTL())); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &domain.AdminAuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User:         admin.ToSummary(),
		Permissions:  admin.Permissions,
	}, nil
}

func (s *AdminService) Logout(ctx context.Context, adminID, sessionID string) error {
	if err := s.repo.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete admin session: %w", err)
	}
	_ = s.logAudit(ctx, adminID, "admin.logout", "admin_session", sessionID, nil, nil, "", "", "success", nil)
	return nil
}

func (s *AdminService) GetMe(ctx context.Context, adminID string) (*domain.AdminUserSummary, []string, error) {
	admin, err := s.repo.FindAdminByID(ctx, adminID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil {
		return nil, nil, domain.NewError("ADMIN_NOT_FOUND", "Akun admin tidak ditemukan", 404)
	}
	return admin.ToSummary(), admin.Permissions, nil
}

func (s *AdminService) GetInvitePreview(ctx context.Context, rawToken string) (*domain.AdminInvitePreviewResponse, error) {
	invite, role, err := s.requireActiveInvite(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	return &domain.AdminInvitePreviewResponse{
		Email:     invite.Email,
		Phone:     invite.Phone,
		RoleID:    invite.RoleID,
		RoleName:  role.Name,
		FullName:  invite.FullName.String,
		ExpiresAt: invite.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func (s *AdminService) AcceptInvite(ctx context.Context, rawToken, fullName, password string) (*domain.AdminInviteAcceptResponse, error) {
	invite, _, err := s.requireActiveInvite(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	if !validator.ValidatePhone(invite.Phone) {
		return nil, domain.ErrValidationFailed("Nomor HP admin tidak valid")
	}
	if !validator.ValidateName(fullName) {
		return nil, domain.ErrValidationFailed("Nama admin minimal 3 karakter")
	}
	if len(strings.TrimSpace(password)) < 8 {
		return nil, domain.ErrValidationFailed("Password admin minimal 8 karakter")
	}

	passwordHash, err := hash.HashPIN(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	adminID := invite.AdminUserID.String
	if adminID == "" {
		adminID = "adm_" + uuid.New().String()[:8]
		admin := &domain.AdminUser{
			ID:           adminID,
			Email:        strings.ToLower(strings.TrimSpace(invite.Email)),
			Phone:        validator.NormalizePhone(invite.Phone),
			FullName:     sqlNullString(validator.SanitizeName(fullName)),
			PasswordHash: sqlNullString(passwordHash),
			Status:       domain.AdminStatusPendingTOTP,
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if invite.InvitedBy.Valid {
			admin.InvitedBy = invite.InvitedBy
			admin.CreatedBy = invite.InvitedBy
		}

		if err := s.repo.CreateAdminUser(ctx, admin, invite.RoleID); err != nil {
			return nil, fmt.Errorf("failed to create admin user: %w", err)
		}
		if err := s.repo.LinkInviteToAdmin(ctx, invite.ID, adminID); err != nil {
			return nil, fmt.Errorf("failed to link invite: %w", err)
		}
	} else {
		if err := s.repo.UpdateAdminPasswordAndStatus(ctx, adminID, passwordHash, domain.AdminStatusPendingTOTP); err != nil {
			return nil, fmt.Errorf("failed to update admin password: %w", err)
		}
	}

	mailboxDomain := "ppob.id"
	if s.emailService != nil && strings.TrimSpace(s.emailService.emailCfg.MailboxDomain) != "" {
		mailboxDomain = strings.TrimSpace(s.emailService.emailCfg.MailboxDomain)
	}
	if _, err := s.repo.EnsurePersonalMailboxForAdmin(ctx, adminID, validator.SanitizeName(fullName), invite.Email, mailboxDomain); err != nil {
		return nil, fmt.Errorf("failed to ensure personal mailbox: %w", err)
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.cfg.TOTPIssuer,
		AccountName: strings.ToLower(strings.TrimSpace(invite.Email)),
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
		SecretSize:  20,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate totp secret: %w", err)
	}

	if err := s.repo.UpsertTOTPSecret(ctx, adminID, key.Secret()); err != nil {
		return nil, fmt.Errorf("failed to save totp secret: %w", err)
	}

	recoveryCodes, hashedCodes, err := generateRecoveryCodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate recovery codes: %w", err)
	}
	if err := s.repo.ReplaceRecoveryCodes(ctx, adminID, hashedCodes); err != nil {
		return nil, fmt.Errorf("failed to save recovery codes: %w", err)
	}

	return &domain.AdminInviteAcceptResponse{
		AdminID:       adminID,
		Secret:        key.Secret(),
		OTPAuthURL:    key.URL(),
		RecoveryCodes: recoveryCodes,
	}, nil
}

func (s *AdminService) ConfirmInviteTOTP(ctx context.Context, rawToken, code, ipAddress, userAgent string) (*domain.AdminAuthResponse, error) {
	invite, _, err := s.requireActiveInvite(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	if !invite.AdminUserID.Valid || invite.AdminUserID.String == "" {
		return nil, domain.ErrValidationFailed("Admin belum menyelesaikan setup password")
	}

	totpSecret, err := s.repo.FindTOTPSecretByAdminID(ctx, invite.AdminUserID.String)
	if err != nil {
		return nil, fmt.Errorf("failed to get totp secret: %w", err)
	}
	if totpSecret == nil {
		return nil, domain.ErrValidationFailed("Authenticator admin belum dibuat")
	}
	if !validateTOTPCode(code, totpSecret.Secret) {
		return nil, domain.NewError("ADMIN_TOTP_INVALID", "Kode authenticator salah", 401)
	}

	if err := s.repo.ConfirmTOTP(ctx, invite.AdminUserID.String); err != nil {
		return nil, fmt.Errorf("failed to confirm totp: %w", err)
	}
	if err := s.repo.MarkInviteAccepted(ctx, invite.ID); err != nil {
		return nil, fmt.Errorf("failed to mark invite accepted: %w", err)
	}

	admin, err := s.repo.FindAdminByID(ctx, invite.AdminUserID.String)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil {
		return nil, domain.NewError("ADMIN_NOT_FOUND", "Akun admin tidak ditemukan", 404)
	}

	resp, err := s.createAuthSession(ctx, admin, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	_ = s.repo.UpdateLastLogin(ctx, admin.ID)
	_ = s.logAudit(ctx, admin.ID, "admin.activate", "admin_user", admin.ID, nil, map[string]interface{}{
		"roleId": invite.RoleID,
	}, ipAddress, userAgent, "success", nil)

	return resp, nil
}

func (s *AdminService) CreateInvite(ctx context.Context, actorID string, req CreateAdminInviteRequest) (map[string]interface{}, error) {
	if !validator.ValidateEmail(req.Email) {
		return nil, domain.ErrValidationFailed("Email admin tidak valid")
	}
	if !validator.ValidatePhone(req.Phone) {
		return nil, domain.ErrValidationFailed("Nomor HP admin tidak valid")
	}

	role, err := s.repo.FindRoleByID(ctx, req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil || !role.IsActive {
		return nil, domain.ErrValidationFailed("Role admin tidak ditemukan")
	}

	existingAdmin, err := s.repo.FindAdminByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check admin email: %w", err)
	}
	if existingAdmin != nil {
		return nil, domain.NewError("ADMIN_ALREADY_EXISTS", "Email admin sudah terdaftar", 409)
	}

	rawToken, tokenHash, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite token: %w", err)
	}

	invite := &domain.AdminInvite{
		ID:        "ainv_" + uuid.New().String()[:8],
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:     validator.NormalizePhone(req.Phone),
		FullName:  sqlNullString(validator.SanitizeName(req.FullName)),
		RoleID:    req.RoleID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.cfg.InviteTTL),
		CreatedAt: time.Now(),
	}
	if actorID != "" {
		invite.InvitedBy = sqlNullString(actorID)
	}

	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create invite: %w", err)
	}

	inviteLink := fmt.Sprintf("%s?token=%s", strings.TrimRight(s.cfg.InviteBaseURL, "/"), rawToken)
	emailSent := false
	if err := s.emailService.SendAdminInvite(ctx, invite.Email, req.FullName, inviteLink, role.Name); err == nil {
		emailSent = true
	}

	_ = s.logAudit(ctx, actorID, "admin.invite.create", "admin_invite", invite.ID, nil, map[string]interface{}{
		"email": invite.Email,
		"role":  role.ID,
	}, "", "", "success", nil)

	return map[string]interface{}{
		"inviteId":   invite.ID,
		"inviteLink": inviteLink,
		"emailSent":  emailSent,
		"expiresAt":  invite.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func (s *AdminService) RequestPasswordReset(ctx context.Context, email, ipAddress, userAgent string) error {
	if !validator.ValidateEmail(email) {
		return domain.ErrValidationFailed("Email admin tidak valid")
	}

	admin, err := s.repo.FindAdminByEmail(ctx, strings.TrimSpace(strings.ToLower(email)))
	if err != nil {
		return fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil || !admin.IsActive || admin.Status != domain.AdminStatusActive || !admin.PasswordHash.Valid {
		return nil
	}

	rawToken, tokenHash, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	reset := &domain.AdminPasswordReset{
		ID:          "apw_" + uuid.New().String()[:8],
		AdminUserID: admin.ID,
		Email:       admin.Email,
		TokenHash:   tokenHash,
		ExpiresAt:   time.Now().Add(30 * time.Minute),
		RequestedBy: sqlNullString(ipAddress),
		CreatedAt:   time.Now(),
	}
	if err := s.repo.CreatePasswordReset(ctx, reset); err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(s.cfg.FrontendURL, "/"), rawToken)
	_ = s.emailService.SendAdminPasswordReset(ctx, admin.Email, admin.DisplayName(), resetLink)
	_ = s.logAudit(ctx, admin.ID, "admin.password_reset.request", "admin_user", admin.ID, nil, map[string]interface{}{
		"email": admin.Email,
	}, ipAddress, userAgent, "success", nil)

	return nil
}

func (s *AdminService) GetPasswordResetPreview(ctx context.Context, rawToken string) (*domain.AdminPasswordResetPreviewResponse, error) {
	reset, admin, err := s.requireActivePasswordReset(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	return &domain.AdminPasswordResetPreviewResponse{
		Email:     admin.Email,
		ExpiresAt: reset.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func (s *AdminService) ConfirmPasswordReset(ctx context.Context, rawToken, newPassword, totpCode, recoveryCode, ipAddress, userAgent string) error {
	reset, admin, err := s.requireActivePasswordReset(ctx, rawToken)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(newPassword)) < 8 {
		return domain.ErrValidationFailed("Password admin minimal 8 karakter")
	}
	if err := s.validateResetSecondFactor(ctx, admin.ID, totpCode, recoveryCode); err != nil {
		return err
	}

	passwordHash, err := hash.HashPIN(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.UpdateAdminPassword(ctx, admin.ID, passwordHash); err != nil {
		return fmt.Errorf("failed to update admin password: %w", err)
	}
	if err := s.repo.MarkPasswordResetUsed(ctx, reset.ID); err != nil {
		return fmt.Errorf("failed to mark password reset used: %w", err)
	}
	if err := s.repo.DeleteSessionsByAdminID(ctx, admin.ID); err != nil {
		return fmt.Errorf("failed to delete admin sessions: %w", err)
	}

	_ = s.logAudit(ctx, admin.ID, "admin.password_reset.confirm", "admin_user", admin.ID, nil, map[string]interface{}{
		"method": map[bool]string{true: "recovery_code", false: "totp"}[strings.TrimSpace(recoveryCode) != ""],
	}, ipAddress, userAgent, "success", nil)
	return nil
}

func (s *AdminService) ListRoles(ctx context.Context) ([]domain.AdminRole, error) {
	return s.repo.ListRoles(ctx)
}

func (s *AdminService) ListPermissions(ctx context.Context) ([]domain.AdminPermission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *AdminService) createAuthSession(ctx context.Context, admin *domain.AdminUser, ipAddress, userAgent string) (*domain.AdminAuthResponse, error) {
	sessionID := "ases_" + uuid.New().String()[:8]
	tokens, err := s.jwtGen.GenerateTokenPair(admin.ID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate admin tokens: %w", err)
	}

	session := &domain.AdminSession{
		ID:               sessionID,
		AdminUserID:      admin.ID,
		RefreshTokenHash: hash.HashToken(tokens.RefreshToken),
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		ExpiresAt:        time.Now().Add(s.jwtGen.GetRefreshTTL()),
		CreatedAt:        time.Now(),
		LastUsedAt:       time.Now(),
	}
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create admin session: %w", err)
	}

	return &domain.AdminAuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
		User:         admin.ToSummary(),
		Permissions:  admin.Permissions,
	}, nil
}

func (s *AdminService) requireActiveInvite(ctx context.Context, rawToken string) (*domain.AdminInvite, *domain.AdminRole, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, nil, domain.ErrValidationFailed("Token undangan tidak valid")
	}
	invite, err := s.repo.FindInviteByTokenHash(ctx, hash.HashToken(rawToken))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get invite: %w", err)
	}
	if invite == nil {
		return nil, nil, domain.NewError("ADMIN_INVITE_NOT_FOUND", "Undangan admin tidak ditemukan", 404)
	}
	if invite.AcceptedAt.Valid {
		return nil, nil, domain.NewError("ADMIN_INVITE_USED", "Undangan admin sudah digunakan", 409)
	}
	if time.Now().After(invite.ExpiresAt) {
		return nil, nil, domain.NewError("ADMIN_INVITE_EXPIRED", "Undangan admin sudah kadaluarsa", 410)
	}
	role, err := s.repo.FindRoleByID(ctx, invite.RoleID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get invite role: %w", err)
	}
	if role == nil {
		return nil, nil, domain.ErrValidationFailed("Role undangan admin tidak ditemukan")
	}
	return invite, role, nil
}

func (s *AdminService) requireActivePasswordReset(ctx context.Context, rawToken string) (*domain.AdminPasswordReset, *domain.AdminUser, error) {
	if strings.TrimSpace(rawToken) == "" {
		return nil, nil, domain.ErrValidationFailed("Token reset password tidak valid")
	}

	reset, err := s.repo.FindPasswordResetByTokenHash(ctx, hash.HashToken(rawToken))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get password reset: %w", err)
	}
	if reset == nil {
		return nil, nil, domain.NewError("ADMIN_PASSWORD_RESET_NOT_FOUND", "Link reset password tidak ditemukan", 404)
	}
	if reset.UsedAt.Valid {
		return nil, nil, domain.NewError("ADMIN_PASSWORD_RESET_USED", "Link reset password sudah digunakan", 409)
	}
	if time.Now().After(reset.ExpiresAt) {
		return nil, nil, domain.NewError("ADMIN_PASSWORD_RESET_EXPIRED", "Link reset password sudah kadaluarsa", 410)
	}

	admin, err := s.repo.FindAdminByID(ctx, reset.AdminUserID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get admin for password reset: %w", err)
	}
	if admin == nil || !admin.IsActive || admin.Status != domain.AdminStatusActive {
		return nil, nil, domain.NewError("ADMIN_DISABLED", "Akun admin tidak aktif", 403)
	}

	return reset, admin, nil
}

func (s *AdminService) validateResetSecondFactor(ctx context.Context, adminID, totpCode, recoveryCode string) error {
	if strings.TrimSpace(totpCode) == "" && strings.TrimSpace(recoveryCode) == "" {
		return domain.ErrValidationFailed("Isi kode authenticator atau recovery code")
	}

	if strings.TrimSpace(totpCode) != "" {
		totpSecret, err := s.repo.FindTOTPSecretByAdminID(ctx, adminID)
		if err != nil {
			return fmt.Errorf("failed to get totp secret: %w", err)
		}
		if totpSecret == nil || !totpSecret.ConfirmedAt.Valid {
			return domain.NewError("ADMIN_TOTP_NOT_READY", "Authenticator admin belum aktif", 403)
		}
		if !validateTOTPCode(totpCode, totpSecret.Secret) {
			return domain.NewError("ADMIN_TOTP_INVALID", "Kode authenticator salah", 401)
		}
		return nil
	}

	codes, err := s.repo.ListRecoveryCodes(ctx, adminID)
	if err != nil {
		return fmt.Errorf("failed to get recovery codes: %w", err)
	}

	targetHash := hash.HashToken(strings.ToUpper(strings.TrimSpace(recoveryCode)))
	for _, code := range codes {
		if code.UsedAt.Valid {
			continue
		}
		if code.CodeHash == targetHash {
			if err := s.repo.MarkRecoveryCodeUsed(ctx, code.ID); err != nil {
				return fmt.Errorf("failed to mark recovery code used: %w", err)
			}
			return nil
		}
	}

	return domain.NewError("ADMIN_RECOVERY_CODE_INVALID", "Recovery code tidak valid", 401)
}

func (s *AdminService) logAudit(ctx context.Context, adminID, action, resourceType, resourceID string, oldValue, newValue interface{}, ipAddress, userAgent, status string, errMessage *string) error {
	var adminUserID sql.NullString
	if adminID != "" {
		adminUserID = sqlNullString(adminID)
	}
	var resourceTypeValue sql.NullString
	if resourceType != "" {
		resourceTypeValue = sqlNullString(resourceType)
	}
	var resourceIDValue sql.NullString
	if resourceID != "" {
		resourceIDValue = sqlNullString(resourceID)
	}
	var ipValue sql.NullString
	if ipAddress != "" {
		ipValue = sqlNullString(ipAddress)
	}
	var uaValue sql.NullString
	if userAgent != "" {
		uaValue = sqlNullString(userAgent)
	}
	var errorValue sql.NullString
	if errMessage != nil {
		errorValue = sqlNullString(*errMessage)
	}

	return s.repo.CreateAuditLog(ctx, &domain.AdminAuditLog{
		ID:           "aal_" + uuid.New().String()[:8],
		AdminUserID:  adminUserID,
		Action:       action,
		ResourceType: resourceTypeValue,
		ResourceID:   resourceIDValue,
		OldValue:     oldValue,
		NewValue:     newValue,
		IPAddress:    ipValue,
		UserAgent:    uaValue,
		Status:       firstNonEmpty(status, "success"),
		ErrorMessage: errorValue,
		CreatedAt:    time.Now(),
	})
}

func generateSecureToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	raw := hex.EncodeToString(bytes)
	return raw, hash.HashToken(raw), nil
}

func generateRecoveryCodes() ([]string, []string, error) {
	rawCodes := make([]string, 0, 8)
	hashedCodes := make([]string, 0, 8)
	for i := 0; i < 8; i++ {
		bytes := make([]byte, 4)
		if _, err := rand.Read(bytes); err != nil {
			return nil, nil, err
		}
		code := strings.ToUpper(hex.EncodeToString(bytes))
		rawCodes = append(rawCodes, code)
		hashedCodes = append(hashedCodes, hash.HashToken(code))
	}
	return rawCodes, hashedCodes, nil
}

func validateTOTPCode(code, secret string) bool {
	valid, err := totp.ValidateCustom(strings.TrimSpace(code), secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	return err == nil && valid
}

func sqlNullString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func (s *AdminService) UpdateAdminProfile(ctx context.Context, actorID, fullName string) (map[string]interface{}, error) {
	fullName = strings.TrimSpace(fullName)
	if len(fullName) < 3 {
		return nil, domain.ErrValidationFailed("Nama lengkap minimal 3 karakter")
	}
	if len(fullName) > 150 {
		return nil, domain.ErrValidationFailed("Nama lengkap maksimal 150 karakter")
	}

	if err := s.repo.UpdateAdminFullName(ctx, actorID, fullName); err != nil {
		return nil, fmt.Errorf("failed to update admin profile: %w", err)
	}

	// Sync personal mailbox display name with admin full name
	_ = s.repo.SyncPersonalMailboxDisplayName(ctx, actorID, fullName)

	admin, err := s.repo.FindAdminByID(ctx, actorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil {
		return nil, domain.NewError("ADMIN_NOT_FOUND", "Akun admin tidak ditemukan", 404)
	}

	return map[string]interface{}{
		"message": "Profil berhasil diperbarui",
		"user":    admin.ToSummary(),
	}, nil
}

func (s *AdminService) GetS3File(ctx context.Context, key string) ([]byte, string, error) {
	if s.s3Client == nil {
		return nil, "", fmt.Errorf("s3 client not available")
	}
	return s.s3Client.GetObjectBytes(ctx, key)
}

func (s *AdminService) UpdateAdminAvatar(ctx context.Context, actorID string, file *multipart.FileHeader) (string, error) {
	if s.s3Client == nil {
		return "", domain.NewError("AVATAR_UPLOAD_DISABLED", "Upload foto profil tidak tersedia", 500)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowedExts[ext] {
		return "", domain.ErrValidationFailed("Format foto harus JPG, PNG, atau WebP")
	}
	if file.Size > 2*1024*1024 {
		return "", domain.ErrValidationFailed("Ukuran foto maksimal 2MB")
	}

	avatarKey, err := s.s3Client.UploadFileKey(ctx, file, fmt.Sprintf("admin-avatars/%s", actorID))
	if err != nil {
		return "", fmt.Errorf("failed to upload avatar: %w", err)
	}

	if err := s.repo.UpdateAdminAvatar(ctx, actorID, avatarKey); err != nil {
		return "", fmt.Errorf("failed to update avatar url: %w", err)
	}

	return avatarKey, nil
}

func (s *AdminService) RemoveAdminAvatar(ctx context.Context, actorID string) error {
	if err := s.repo.ClearAdminAvatar(ctx, actorID); err != nil {
		return fmt.Errorf("failed to clear avatar: %w", err)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
