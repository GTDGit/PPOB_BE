package service

import (
	"context"
	"fmt"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// UserService handles user profile and settings business logic
type UserService struct {
	userRepo     repository.UserRepository
	balanceRepo  repository.BalanceRepository
	historyRepo  repository.HistoryRepository
	settingsRepo repository.UserSettingsRepository
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repository.UserRepository,
	balanceRepo repository.BalanceRepository,
	historyRepo repository.HistoryRepository,
	settingsRepo repository.UserSettingsRepository,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		balanceRepo:  balanceRepo,
		historyRepo:  historyRepo,
		settingsRepo: settingsRepo,
	}
}

// GetProfile returns user profile with stats and tier info
func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.ProfileResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Build profile user
	email := ""
	if user.Email.Valid {
		email = user.Email.String
	}
	var emailPtr *string
	if email != "" {
		emailPtr = &email
	}

	profileUser := &domain.ProfileUser{
		ID:            user.ID,
		MIC:           user.MIC,
		Phone:         user.Phone,
		FullName:      user.FullName,
		Email:         emailPtr,
		EmailVerified: user.EmailVerifiedAt.Valid,
		Gender:        user.Gender,
		Tier:          user.Tier,
		AvatarURL:     user.AvatarURL,
		KYCStatus:     user.KYCStatus,
		BusinessType:  user.BusinessType,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
	}

	balance, err := s.balanceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	if balance == nil {
		balance = &domain.Balance{
			UserID: userID,
		}
	}

	summary, err := s.historyRepo.GetUserSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile summary: %w", err)
	}

	memberSince := user.CreatedAt.Format("January 2006")
	memberDays := int(time.Since(user.CreatedAt).Hours() / 24)
	totalTransactions := 0
	totalTransactionAmount := int64(0)
	if summary != nil {
		if summary.TotalTransactions > 0 {
			totalTransactions = summary.TotalTransactions
		}
		totalTransactionAmount = summary.SuccessfulAmount
	}

	stats := &domain.ProfileStats{
		TotalTransactions:               totalTransactions,
		TotalTransactionAmount:          totalTransactionAmount,
		TotalTransactionAmountFormatted: formatHomeCurrency(totalTransactionAmount),
		MemberSince:                     memberSince,
		MemberDays:                      memberDays,
	}

	// Build tier info
	tierInfo := s.buildTierInfo(user.Tier)

	return &domain.ProfileResponse{
		User:     profileUser,
		Balance:  s.buildProfileBalance(balance),
		Stats:    stats,
		TierInfo: tierInfo,
	}, nil
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(ctx context.Context, userID string, fullName, email, gender, businessType *string) (*domain.UpdateProfileResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Update fields if provided
	if fullName != nil && *fullName != "" {
		if len(*fullName) < 3 {
			return nil, domain.ErrValidationFailed("Full name must be at least 3 characters")
		}
		user.FullName = *fullName
	}

	if email != nil && *email != "" {
		// TODO: Add email validation
		user.Email.String = *email
		user.Email.Valid = true
	}

	if gender != nil {
		user.Gender = gender
	}

	if businessType != nil {
		user.BusinessType = businessType
	}

	user.UpdatedAt = time.Now()

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Build response
	emailStr := ""
	if user.Email.Valid {
		emailStr = user.Email.String
	}
	var emailPtr *string
	if emailStr != "" {
		emailPtr = &emailStr
	}

	profileUser := &domain.ProfileUser{
		ID:            user.ID,
		MIC:           user.MIC,
		Phone:         user.Phone,
		FullName:      user.FullName,
		Email:         emailPtr,
		EmailVerified: user.EmailVerifiedAt.Valid,
		Gender:        user.Gender,
		Tier:          user.Tier,
		AvatarURL:     user.AvatarURL,
		KYCStatus:     user.KYCStatus,
		BusinessType:  user.BusinessType,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
	}

	return &domain.UpdateProfileResponse{
		Updated: true,
		User:    profileUser,
	}, nil
}

// UploadAvatar uploads user avatar (mock implementation)
func (s *UserService) UploadAvatar(ctx context.Context, userID string) (*domain.AvatarUploadResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Mock avatar URL (in production, upload to CDN/S3)
	avatarURL := fmt.Sprintf("https://cdn.ppob.id/avatars/%s.jpg", userID)
	user.AvatarURL = &avatarURL
	user.UpdatedAt = time.Now()

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update avatar: %w", err)
	}

	return &domain.AvatarUploadResponse{
		Uploaded:  true,
		AvatarURL: avatarURL,
	}, nil
}

// DeleteAvatar deletes user avatar
func (s *UserService) DeleteAvatar(ctx context.Context, userID string) (*domain.AvatarDeleteResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Remove avatar
	user.AvatarURL = nil
	user.UpdatedAt = time.Now()

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to delete avatar: %w", err)
	}

	return &domain.AvatarDeleteResponse{
		Deleted: true,
	}, nil
}

// GetSettings returns user settings
func (s *UserService) GetSettings(ctx context.Context, userID string) (*domain.SettingsResponse, error) {
	// Get settings
	settings, err := s.settingsRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	if settings == nil {
		return nil, domain.ErrNotFound("Settings")
	}

	return &domain.SettingsResponse{
		Settings: buildUserSettingsDetail(settings),
	}, nil
}

// UpdateSettings updates user settings
func (s *UserService) UpdateSettings(ctx context.Context, userID string, updates map[string]interface{}) (*domain.UpdateSettingsResponse, error) {
	// Get settings
	settings, err := s.settingsRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	if settings == nil {
		return nil, domain.ErrNotFound("Settings")
	}

	applySecuritySettings(settings, nestedSettingsMap(updates, "security"))
	applyTransactionSettings(settings, nestedSettingsMap(updates, "transaction"))
	applyDisplaySettings(settings, nestedSettingsMap(updates, "display"))
	applyPrivacySettings(settings, nestedSettingsMap(updates, "privacy"))

	// Backward compatibility for older flat payloads.
	applySecuritySettings(settings, updates)
	applyTransactionSettings(settings, updates)
	applyDisplaySettings(settings, updates)
	applyPrivacySettings(settings, updates)

	settings.UpdatedAt = time.Now()

	// Update settings
	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, fmt.Errorf("failed to update settings: %w", err)
	}

	return &domain.UpdateSettingsResponse{
		Updated:  true,
		Settings: buildUserSettingsDetail(settings),
	}, nil
}

func buildUserSettingsDetail(settings *domain.UserSettings) *domain.UserSettingsDetail {
	return &domain.UserSettingsDetail{
		Security: &domain.SecuritySettingsDetail{
			PINRequiredForTransaction:     settings.PINRequiredForTransaction,
			PINRequiredMinAmount:          settings.PINRequiredMinAmount,
			PINRequiredMinAmountFormatted: formatHomeCurrency(settings.PINRequiredMinAmount),
			BiometricEnabled:              settings.BiometricEnabled,
		},
		Transaction: &domain.TransactionSettingsDetail{
			DefaultSellingPriceMarkup:          settings.DefaultSellingPriceMarkup,
			DefaultSellingPriceMarkupFormatted: formatHomeCurrency(int64(settings.DefaultSellingPriceMarkup)),
			AutoSaveContact:                    settings.AutoSaveContact,
			ShowProfitOnReceipt:                settings.ShowProfitOnReceipt,
		},
		Display: &domain.DisplaySettingsDetail{
			Language: settings.Language,
			Currency: settings.Currency,
			Theme:    settings.Theme,
		},
		Privacy: &domain.PrivacySettingsDetail{
			ShowPhoneOnQris: settings.ShowPhoneOnQRIS,
			ShowNameOnQris:  settings.ShowNameOnQRIS,
		},
	}
}

func nestedSettingsMap(updates map[string]interface{}, key string) map[string]interface{} {
	raw, ok := updates[key]
	if !ok {
		return nil
	}
	nested, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	return nested
}

func applySecuritySettings(settings *domain.UserSettings, updates map[string]interface{}) {
	if updates == nil {
		return
	}
	if val, ok := updates["pinRequiredForTransaction"].(bool); ok {
		settings.PINRequiredForTransaction = val
	}
	if val, ok := updates["pinRequiredMinAmount"].(float64); ok {
		settings.PINRequiredMinAmount = int64(val)
	}
	if val, ok := updates["biometricEnabled"].(bool); ok {
		settings.BiometricEnabled = val
	}
}

func applyTransactionSettings(settings *domain.UserSettings, updates map[string]interface{}) {
	if updates == nil {
		return
	}
	if val, ok := updates["defaultSellingPriceMarkup"].(float64); ok {
		settings.DefaultSellingPriceMarkup = int(val)
	}
	if val, ok := updates["autoSaveContact"].(bool); ok {
		settings.AutoSaveContact = val
	}
	if val, ok := updates["showProfitOnReceipt"].(bool); ok {
		settings.ShowProfitOnReceipt = val
	}
}

func applyDisplaySettings(settings *domain.UserSettings, updates map[string]interface{}) {
	if updates == nil {
		return
	}
	if val, ok := updates["language"].(string); ok && val != "" {
		settings.Language = val
	}
	if val, ok := updates["currency"].(string); ok && val != "" {
		settings.Currency = val
	}
	if val, ok := updates["theme"].(string); ok && val != "" {
		settings.Theme = val
	}
}

func applyPrivacySettings(settings *domain.UserSettings, updates map[string]interface{}) {
	if updates == nil {
		return
	}
	if val, ok := updates["showPhoneOnQris"].(bool); ok {
		settings.ShowPhoneOnQRIS = val
	}
	if val, ok := updates["showNameOnQris"].(bool); ok {
		settings.ShowNameOnQRIS = val
	}
}

// GetReferralInfo returns referral information (mock)
func (s *UserService) GetReferralInfo(ctx context.Context, userID string) (*domain.ReferralInfoResponse, error) {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Mock referral code
	referralCode := "USER123"
	if user.ReferralCode != nil {
		referralCode = *user.ReferralCode
	}

	// Mock data
	return &domain.ReferralInfoResponse{
		ReferralCode:           referralCode,
		ReferralCount:          0, // TODO: Implement from referral tracking
		TotalEarnings:          0,
		TotalEarningsFormatted: "Rp0",
		ShareLink:              fmt.Sprintf("https://ppob.id/ref/%s", referralCode),
		Stats: &domain.ReferralStats{
			ThisMonth:              0,
			LastMonth:              0,
			EarningsToday:          0,
			EarningsTodayFormatted: "Rp0",
		},
	}, nil
}

// GetReferralHistory returns referral history (mock)
func (s *UserService) GetReferralHistory(ctx context.Context, userID string) (*domain.ReferralHistoryResponse, error) {
	// Mock empty history
	// TODO: Implement from referral tracking database
	return &domain.ReferralHistoryResponse{
		Referrals: []*domain.ReferralDetail{},
		Total:     0,
	}, nil
}

// Helper functions

func (s *UserService) buildTierInfo(tier string) *domain.ProfileTierInfo {
	tierInfo := &domain.ProfileTierInfo{
		CurrentTier: tier,
		Progress:    0,
		Benefits:    []string{},
	}

	switch tier {
	case "BRONZE":
		nextTier := "SILVER"
		points := 5000
		tierInfo.NextTier = &nextTier
		tierInfo.PointsToNextTier = &points
		tierInfo.Progress = 0
		tierInfo.Benefits = []string{
			"Cashback 0.5% setiap transaksi",
			"Akses customer service",
		}
	case "SILVER":
		nextTier := "GOLD"
		points := 7500
		tierInfo.NextTier = &nextTier
		tierInfo.PointsToNextTier = &points
		tierInfo.Progress = 25
		tierInfo.Benefits = []string{
			"Cashback 0.75% setiap transaksi",
			"Prioritas customer service",
			"Akses promo spesial",
		}
	case "GOLD":
		nextTier := "PLATINUM"
		points := 10000
		tierInfo.NextTier = &nextTier
		tierInfo.PointsToNextTier = &points
		tierInfo.Progress = 75
		tierInfo.Benefits = []string{
			"Cashback 1% setiap transaksi",
			"Prioritas customer service",
			"Akses promo eksklusif",
			"Gratis admin bank transfer",
		}
	case "PLATINUM":
		tierInfo.NextTier = nil
		tierInfo.PointsToNextTier = nil
		tierInfo.Progress = 100
		tierInfo.Benefits = []string{
			"Cashback 1.5% setiap transaksi",
			"VIP customer service 24/7",
			"Akses promo eksklusif",
			"Gratis semua admin",
			"Dedicated account manager",
		}
	}

	return tierInfo
}

func (s *UserService) buildProfileBalance(balance *domain.Balance) *domain.HomeBalanceInfo {
	lastUpdated := balance.UpdatedAt
	if lastUpdated.IsZero() {
		lastUpdated = time.Now()
	}

	balanceInfo := &domain.HomeBalanceInfo{
		Amount:      balance.Amount,
		Formatted:   formatHomeCurrency(balance.Amount),
		LastUpdated: lastUpdated.Format(time.RFC3339),
	}

	if balance.PendingAmount > 0 {
		balanceInfo.PendingBalance = &domain.HomePendingBalance{
			Amount:    balance.PendingAmount,
			Formatted: formatHomeCurrency(balance.PendingAmount),
		}
	}

	if balance.Points > 0 {
		var expiresAt *string
		if balance.PointsExpiresAt != nil {
			formatted := balance.PointsExpiresAt.Format(time.RFC3339)
			expiresAt = &formatted
		}
		balanceInfo.Points = &domain.HomePointsInfo{
			Amount:    int64(balance.Points),
			Formatted: fmt.Sprintf("%s Poin", formatNumberHome(int64(balance.Points))),
			ExpiresAt: expiresAt,
		}
	}

	return balanceInfo
}
