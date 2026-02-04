package repository

import (
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// HomeRepository defines the interface for home screen data operations
type HomeRepository interface {
	// Services
	GetServicesVersion() string
	GetFeaturedServices() []*domain.ServiceMenu
	GetServiceCategories() []*domain.ServiceCategory
	GetAllServices() (*domain.ServicesResponse, error)

	// Banners
	GetBannersVersion() string
	GetBanners(placement string, userTier string) []*domain.Banner
}

// homeRepository implements HomeRepository
type homeRepository struct{}

// NewHomeRepository creates a new home repository
func NewHomeRepository() HomeRepository {
	return &homeRepository{}
}

// GetServicesVersion returns current services version
func (r *homeRepository) GetServicesVersion() string {
	// Format: YYYYMMDDHH (Year-Month-Day-Hour)
	return time.Now().Format("2006010215")
}

// GetFeaturedServices returns featured services for home screen
func (r *homeRepository) GetFeaturedServices() []*domain.ServiceMenu {
	return []*domain.ServiceMenu{
		{
			ID:       "pulsa",
			Name:     "Pulsa",
			Icon:     "pulsa",
			IconURL:  "https://cdn.ppob.id/icons/pulsa.png",
			Route:    "/services/pulsa",
			Status:   domain.ServiceStatusActive,
			Category: "prepaid",
			Position: 1,
		},
		{
			ID:       "paket_data",
			Name:     "Paket Data",
			Icon:     "paket_data",
			IconURL:  "https://cdn.ppob.id/icons/paket-data.png",
			Route:    "/services/paket-data",
			Status:   domain.ServiceStatusActive,
			Category: "prepaid",
			Position: 2,
		},
		{
			ID:       "token_pln",
			Name:     "Token PLN",
			Icon:     "token_pln",
			IconURL:  "https://cdn.ppob.id/icons/token-pln.png",
			Route:    "/services/token-pln",
			Status:   domain.ServiceStatusActive,
			Category: "prepaid",
			Position: 3,
		},
		{
			ID:       "tagihan_pln",
			Name:     "Tagihan PLN",
			Icon:     "tagihan_pln",
			IconURL:  "https://cdn.ppob.id/icons/tagihan-pln.png",
			Route:    "/services/tagihan-pln",
			Status:   domain.ServiceStatusActive,
			Category: "postpaid",
			Position: 4,
		},
		{
			ID:       "pdam",
			Name:     "PDAM",
			Icon:     "pdam",
			IconURL:  "https://cdn.ppob.id/icons/pdam.png",
			Route:    "/services/pdam",
			Status:   domain.ServiceStatusActive,
			Category: "postpaid",
			Position: 5,
		},
		{
			ID:       "ewallet",
			Name:     "E-Wallet",
			Icon:     "ewallet",
			IconURL:  "https://cdn.ppob.id/icons/ewallet.png",
			Route:    "/services/ewallet",
			Status:   domain.ServiceStatusActive,
			Category: "finance",
			Position: 6,
		},
		{
			ID:       "transfer_bank",
			Name:     "Transfer Bank",
			Icon:     "transfer_bank",
			IconURL:  "https://cdn.ppob.id/icons/transfer-bank.png",
			Route:    "/services/transfer-bank",
			Status:   domain.ServiceStatusActive,
			Category: "finance",
			Position: 7,
		},
		{
			ID:       "voucher_game",
			Name:     "Voucher Game",
			Icon:     "voucher_game",
			IconURL:  "https://cdn.ppob.id/icons/voucher-game.png",
			Route:    "/services/voucher-game",
			Status:   domain.ServiceStatusActive,
			Category: "prepaid",
			Position: 8,
		},
		{
			ID:       "bpjs",
			Name:     "BPJS",
			Icon:     "bpjs",
			IconURL:  "https://cdn.ppob.id/icons/bpjs.png",
			Route:    "/services/bpjs",
			Status:   domain.ServiceStatusActive,
			Category: "postpaid",
			Position: 9,
		},
		{
			ID:       "telkom",
			Name:     "Telkom",
			Icon:     "telkom",
			IconURL:  "https://cdn.ppob.id/icons/telkom.png",
			Route:    "/services/telkom",
			Status:   domain.ServiceStatusActive,
			Category: "postpaid",
			Position: 10,
		},
	}
}

// GetServiceCategories returns service categories with services
func (r *homeRepository) GetServiceCategories() []*domain.ServiceCategory {
	promoBadge := domain.BadgePromo

	return []*domain.ServiceCategory{
		{
			ID:    "prepaid",
			Name:  "Pra Bayar",
			Order: 1,
			Services: []*domain.ServiceMenu{
				{
					ID:       "pulsa",
					Name:     "Pulsa",
					Icon:     "pulsa",
					IconURL:  "https://cdn.ppob.id/icons/pulsa.png",
					Route:    "/services/pulsa",
					Status:   domain.ServiceStatusActive,
					Position: 1,
				},
				{
					ID:       "paket_data",
					Name:     "Paket Data",
					Icon:     "paket_data",
					IconURL:  "https://cdn.ppob.id/icons/paket-data.png",
					Route:    "/services/paket-data",
					Status:   domain.ServiceStatusActive,
					Badge:    &promoBadge,
					Position: 2,
				},
				{
					ID:       "token_pln",
					Name:     "Token PLN",
					Icon:     "token_pln",
					IconURL:  "https://cdn.ppob.id/icons/token-pln.png",
					Route:    "/services/token-pln",
					Status:   domain.ServiceStatusActive,
					Position: 3,
				},
				{
					ID:       "voucher_game",
					Name:     "Voucher Game",
					Icon:     "voucher_game",
					IconURL:  "https://cdn.ppob.id/icons/voucher-game.png",
					Route:    "/services/voucher-game",
					Status:   domain.ServiceStatusActive,
					Position: 4,
				},
			},
		},
		{
			ID:    "postpaid",
			Name:  "Pasca Bayar",
			Order: 2,
			Services: []*domain.ServiceMenu{
				{
					ID:       "pulsa_pascabayar",
					Name:     "Pulsa Pascabayar",
					Icon:     "pulsa_pascabayar",
					IconURL:  "https://cdn.ppob.id/icons/pulsa-pascabayar.png",
					Route:    "/services/pulsa-pascabayar",
					Status:   domain.ServiceStatusActive,
					Position: 1,
				},
				{
					ID:       "pdam",
					Name:     "PDAM",
					Icon:     "pdam",
					IconURL:  "https://cdn.ppob.id/icons/pdam.png",
					Route:    "/services/pdam",
					Status:   domain.ServiceStatusActive,
					Position: 2,
				},
				{
					ID:       "tagihan_pln",
					Name:     "Tagihan PLN",
					Icon:     "tagihan_pln",
					IconURL:  "https://cdn.ppob.id/icons/tagihan-pln.png",
					Route:    "/services/tagihan-pln",
					Status:   domain.ServiceStatusActive,
					Position: 3,
				},
				{
					ID:       "bpjs",
					Name:     "BPJS",
					Icon:     "bpjs",
					IconURL:  "https://cdn.ppob.id/icons/bpjs.png",
					Route:    "/services/bpjs",
					Status:   domain.ServiceStatusActive,
					Position: 4,
				},
				{
					ID:       "telkom",
					Name:     "Telkom",
					Icon:     "telkom",
					IconURL:  "https://cdn.ppob.id/icons/telkom.png",
					Route:    "/services/telkom",
					Status:   domain.ServiceStatusActive,
					Position: 5,
				},
				{
					ID:       "tagihan_gas",
					Name:     "Tagihan Gas",
					Icon:     "tagihan_gas",
					IconURL:  "https://cdn.ppob.id/icons/tagihan-gas.png",
					Route:    "/services/tagihan-gas",
					Status:   domain.ServiceStatusActive,
					Position: 6,
				},
				{
					ID:       "pbb",
					Name:     "PBB",
					Icon:     "pbb",
					IconURL:  "https://cdn.ppob.id/icons/pbb.png",
					Route:    "/services/pbb",
					Status:   domain.ServiceStatusActive,
					Position: 7,
				},
				{
					ID:       "tv_kabel",
					Name:     "TV Kabel",
					Icon:     "tv_kabel",
					IconURL:  "https://cdn.ppob.id/icons/tv-kabel.png",
					Route:    "/services/tv-kabel",
					Status:   domain.ServiceStatusActive,
					Position: 8,
				},
			},
		},
		{
			ID:    "finance",
			Name:  "Keuangan",
			Order: 3,
			Services: []*domain.ServiceMenu{
				{
					ID:       "ewallet",
					Name:     "Topup E-Wallet",
					Icon:     "ewallet",
					IconURL:  "https://cdn.ppob.id/icons/ewallet.png",
					Route:    "/services/ewallet",
					Status:   domain.ServiceStatusActive,
					Position: 1,
				},
				{
					ID:       "transfer_bank",
					Name:     "Transfer Bank",
					Icon:     "transfer_bank",
					IconURL:  "https://cdn.ppob.id/icons/transfer-bank.png",
					Route:    "/services/transfer-bank",
					Status:   domain.ServiceStatusActive,
					Position: 2,
				},
			},
		},
	}
}

// GetAllServices returns all services with version
func (r *homeRepository) GetAllServices() (*domain.ServicesResponse, error) {
	return &domain.ServicesResponse{
		Version:    r.GetServicesVersion(),
		Featured:   r.GetFeaturedServices(),
		Categories: r.GetServiceCategories(),
	}, nil
}

// GetBannersVersion returns current banners version
func (r *homeRepository) GetBannersVersion() string {
	// Format: YYYYMMDDHH
	return time.Now().Format("2006010215")
}

// GetBanners returns active banners filtered by placement and user tier
func (r *homeRepository) GetBanners(placement string, userTier string) []*domain.Banner {
	now := time.Now()

	// All mock banners
	allBanners := []*domain.Banner{
		{
			ID:           "banner_001",
			Title:        "Top Up Game & Pulsa Murah",
			Subtitle:     strPtr("Diskon hingga 10%"),
			ImageURL:     "https://cdn.ppob.id/banners/promo-game-jan.png",
			ThumbnailURL: strPtr("https://cdn.ppob.id/banners/promo-game-jan-thumb.png"),
			Action: &domain.BannerAction{
				Type:  domain.ActionTypeDeeplink,
				Value: "/services/voucher-game",
			},
			BackgroundColor: "#1E3A8A",
			TextColor:       strPtr("#FFFFFF"),
			StartDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Priority:        1,
			Placement:       domain.PlacementHome,
			TargetAudience: &domain.Audience{
				Tiers: []string{"BRONZE", "SILVER", "GOLD", "PLATINUM"},
			},
		},
		{
			ID:           "banner_002",
			Title:        "Cashback Token PLN",
			Subtitle:     strPtr("Cashback 5% max 10rb"),
			ImageURL:     "https://cdn.ppob.id/banners/promo-pln-jan.png",
			ThumbnailURL: strPtr("https://cdn.ppob.id/banners/promo-pln-jan-thumb.png"),
			Action: &domain.BannerAction{
				Type:  domain.ActionTypeDeeplink,
				Value: "/services/token-pln",
			},
			BackgroundColor: "#047857",
			TextColor:       strPtr("#FFFFFF"),
			StartDate:       time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Priority:        2,
			Placement:       domain.PlacementHome,
			TargetAudience: &domain.Audience{
				Tiers: []string{"SILVER", "GOLD", "PLATINUM"},
			},
		},
		{
			ID:           "banner_003",
			Title:        "Ajak Teman Dapat Bonus",
			Subtitle:     strPtr("Bonus Rp10.000 per referral"),
			ImageURL:     "https://cdn.ppob.id/banners/referral-program.png",
			ThumbnailURL: strPtr("https://cdn.ppob.id/banners/referral-program-thumb.png"),
			Action: &domain.BannerAction{
				Type:  domain.ActionTypeDeeplink,
				Value: "/referral",
			},
			BackgroundColor: "#7C3AED",
			TextColor:       strPtr("#FFFFFF"),
			StartDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Priority:        3,
			Placement:       domain.PlacementHome,
			TargetAudience: &domain.Audience{
				Tiers: []string{"BRONZE", "SILVER", "GOLD", "PLATINUM"},
			},
		},
		{
			ID:           "banner_004",
			Title:        "Promo Transfer Bank Gratis",
			Subtitle:     strPtr("Gratis admin untuk transaksi pertama"),
			ImageURL:     "https://cdn.ppob.id/banners/promo-transfer.png",
			ThumbnailURL: strPtr("https://cdn.ppob.id/banners/promo-transfer-thumb.png"),
			Action: &domain.BannerAction{
				Type:  domain.ActionTypeDeeplink,
				Value: "/services/transfer-bank",
			},
			BackgroundColor: "#DC2626",
			TextColor:       strPtr("#FFFFFF"),
			StartDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:         time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Priority:        1,
			Placement:       domain.PlacementServices,
			TargetAudience: &domain.Audience{
				Tiers: []string{"BRONZE", "SILVER", "GOLD", "PLATINUM"},
			},
		},
	}

	// Filter by placement and active date
	var filteredBanners []*domain.Banner
	for _, banner := range allBanners {
		// Check placement (if specified)
		if placement != "" && banner.Placement != placement {
			continue
		}

		// Check if banner is active (current time is between start and end date)
		if now.Before(banner.StartDate) || now.After(banner.EndDate) {
			continue
		}

		// Check user tier targeting
		if banner.TargetAudience != nil && len(banner.TargetAudience.Tiers) > 0 {
			tierMatch := false
			for _, tier := range banner.TargetAudience.Tiers {
				if tier == userTier {
					tierMatch = true
					break
				}
			}
			if !tierMatch {
				continue
			}
		}

		filteredBanners = append(filteredBanners, banner)
	}

	return filteredBanners
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
