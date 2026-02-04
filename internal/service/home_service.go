package service

import (
	"context"
	"fmt"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// HomeService handles home screen business logic
type HomeService struct {
	homeRepo    repository.HomeRepository
	userRepo    repository.UserRepository
	balanceRepo repository.BalanceRepository
}

// NewHomeService creates a new home service
func NewHomeService(
	homeRepo repository.HomeRepository,
	userRepo repository.UserRepository,
	balanceRepo repository.BalanceRepository,
) *HomeService {
	return &HomeService{
		homeRepo:    homeRepo,
		userRepo:    userRepo,
		balanceRepo: balanceRepo,
	}
}

// GetHome returns aggregated home screen data
func (s *HomeService) GetHome(ctx context.Context, userID string, servicesVersion, bannersVersion string) (*domain.HomeResponse, error) {
	// Get user info
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Get balance
	balance, err := s.balanceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	if balance == nil {
		return nil, domain.ErrNotFound("Balance")
	}

	// Build user info
	email := ""
	if user.Email.Valid {
		email = user.Email.String
	}

	userInfo := &domain.HomeUserInfo{
		ID:        user.ID,
		MIC:       user.MIC,
		FullName:  user.FullName,
		Phone:     user.Phone,
		Email:     email,
		Tier:      user.Tier,
		AvatarURL: user.AvatarURL,
		KYCStatus: user.KYCStatus,
	}

	// Build balance info
	balanceInfo := s.buildBalanceInfo(balance)

	// Build services data (with version check)
	var servicesData *domain.HomeServicesData
	currentServicesVersion := s.homeRepo.GetServicesVersion()
	if servicesVersion != currentServicesVersion {
		servicesData = &domain.HomeServicesData{
			Version:    currentServicesVersion,
			Featured:   s.homeRepo.GetFeaturedServices(),
			Categories: s.homeRepo.GetServiceCategories(),
		}
	}

	// Build banners data (with version check)
	var bannersData *domain.HomeBannersData
	currentBannersVersion := s.homeRepo.GetBannersVersion()
	if bannersVersion != currentBannersVersion {
		banners := s.homeRepo.GetBanners(domain.PlacementHome, user.Tier)
		bannersData = &domain.HomeBannersData{
			Version:            currentBannersVersion,
			Placement:          domain.PlacementHome,
			Items:              banners,
			AutoScrollInterval: 5000, // 5 seconds
			TotalItems:         len(banners),
		}
	}

	// Build notifications info (placeholder)
	notificationsInfo := &domain.HomeNotificationsInfo{
		UnreadCount: 0, // TODO: Implement notification count
	}

	return &domain.HomeResponse{
		User:          userInfo,
		Balance:       balanceInfo,
		Services:      servicesData,
		Banners:       bannersData,
		Notifications: notificationsInfo,
		Announcements: []interface{}{}, // Empty for now
	}, nil
}

// GetBalance returns user balance information
func (s *HomeService) GetBalance(ctx context.Context, userID string) (*domain.BalanceResponse, error) {
	balance, err := s.balanceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	if balance == nil {
		return nil, domain.ErrNotFound("Balance")
	}

	balanceInfo := s.buildBalanceInfo(balance)

	return &domain.BalanceResponse{
		Balance:        balanceInfo,
		PendingBalance: balanceInfo.PendingBalance,
		Points:         balanceInfo.Points,
	}, nil
}

// GetServices returns services list with version
func (s *HomeService) GetServices(ctx context.Context, version string) (*domain.ServicesResponse, bool, error) {
	currentVersion := s.homeRepo.GetServicesVersion()

	// Check if client version matches (304 Not Modified)
	if version != "" && version == currentVersion {
		return nil, true, nil // Return true to indicate not modified
	}

	services, err := s.homeRepo.GetAllServices()
	if err != nil {
		return nil, false, fmt.Errorf("failed to get services: %w", err)
	}

	return services, false, nil
}

// GetBanners returns banners list with version
func (s *HomeService) GetBanners(ctx context.Context, userID string, placement, version string) (*domain.BannersResponse, bool, error) {
	// Get user for tier targeting
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, false, domain.ErrUserNotFound
	}

	currentVersion := s.homeRepo.GetBannersVersion()

	// Check if client version matches (304 Not Modified)
	if version != "" && version == currentVersion {
		return nil, true, nil
	}

	// Default placement to home
	if placement == "" {
		placement = domain.PlacementHome
	}

	banners := s.homeRepo.GetBanners(placement, user.Tier)

	return &domain.BannersResponse{
		Version:            currentVersion,
		Placement:          placement,
		Items:              banners,
		AutoScrollInterval: 5000,
		TotalItems:         len(banners),
	}, false, nil
}

// Helper functions

func (s *HomeService) buildBalanceInfo(balance *domain.Balance) *domain.HomeBalanceInfo {
	balanceInfo := &domain.HomeBalanceInfo{
		Amount:      balance.Amount,
		Formatted:   formatHomeCurrency(balance.Amount),
		LastUpdated: balance.UpdatedAt.Format(time.RFC3339),
	}

	// Add pending balance if any
	if balance.PendingAmount > 0 {
		balanceInfo.PendingBalance = &domain.HomePendingBalance{
			Amount:    balance.PendingAmount,
			Formatted: formatHomeCurrency(balance.PendingAmount),
		}
	}

	// Add points if any
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

func formatHomeCurrency(amount int64) string {
	// Simple formatting (can be enhanced with proper locale support)
	if amount == 0 {
		return "Rp0"
	}

	// Format with thousand separators
	str := fmt.Sprintf("%d", amount)
	n := len(str)
	if n <= 3 {
		return "Rp" + str
	}

	// Add dots for thousands
	result := ""
	for i, digit := range str {
		if i > 0 && (n-i)%3 == 0 {
			result += "."
		}
		result += string(digit)
	}

	return "Rp" + result
}

func formatNumberHome(num int64) string {
	// Format number with thousand separators
	if num == 0 {
		return "0"
	}

	str := fmt.Sprintf("%d", num)
	n := len(str)
	if n <= 3 {
		return str
	}

	// Add dots for thousands
	result := ""
	for i, digit := range str {
		if i > 0 && (n-i)%3 == 0 {
			result += "."
		}
		result += string(digit)
	}

	return result
}
