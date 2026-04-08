package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// HomeRepository defines the interface for home screen data operations
type HomeRepository interface {
	// Services
	GetServicesVersion(ctx context.Context) string
	GetFeaturedServices(ctx context.Context) []*domain.ServiceMenu
	GetServiceCategories(ctx context.Context) []*domain.ServiceCategory
	GetAllServices(ctx context.Context) (*domain.ServicesResponse, error)

	// Banners
	GetBannersVersion(ctx context.Context) string
	GetBanners(ctx context.Context, placement string, userTier string) []*domain.Banner
}

// homeRepository implements HomeRepository
type homeRepository struct {
	db *sqlx.DB
}

// NewHomeRepository creates a new home repository
func NewHomeRepository(db *sqlx.DB) HomeRepository {
	return &homeRepository{db: db}
}

// ========== Service DB row types ==========

type serviceRow struct {
	ID         string  `db:"id"`
	CategoryID *string `db:"category_id"`
	Name       string  `db:"name"`
	Icon       string  `db:"icon"`
	IconURL    *string `db:"icon_url"`
	Route      string  `db:"route"`
	Status     string  `db:"status"`
	Badge      *string `db:"badge"`
	IsFeatured bool    `db:"is_featured"`
	SortOrder  int     `db:"sort_order"`
}

type serviceCategoryRow struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Slug      string `db:"slug"`
	SortOrder int    `db:"sort_order"`
	IsActive  bool   `db:"is_active"`
}

type bannerRow struct {
	ID              string    `db:"id"`
	Title           string    `db:"title"`
	Subtitle        *string   `db:"subtitle"`
	ImageURL        string    `db:"image_url"`
	ThumbnailURL    *string   `db:"thumbnail_url"`
	ActionType      *string   `db:"action_type"`
	ActionValue     *string   `db:"action_value"`
	BackgroundColor *string   `db:"background_color"`
	TextColor       *string   `db:"text_color"`
	Placement       string    `db:"placement"`
	StartDate       time.Time `db:"start_date"`
	EndDate         time.Time `db:"end_date"`
	Priority        int       `db:"priority"`
	TargetTiers     *string   `db:"target_tiers"`
	IsNewUserOnly   bool      `db:"is_new_user_only"`
	IsActive        bool      `db:"is_active"`
}

// ========== Services ==========

// GetServicesVersion returns current services version based on last update
func (r *homeRepository) GetServicesVersion(ctx context.Context) string {
	var lastUpdated sql.NullTime
	err := r.db.GetContext(ctx, &lastUpdated, `SELECT MAX(updated_at) FROM services`)
	if err != nil || !lastUpdated.Valid {
		return time.Now().Format("2006010215")
	}
	return lastUpdated.Time.Format("2006010215")
}

// GetFeaturedServices returns featured services for home screen
func (r *homeRepository) GetFeaturedServices(ctx context.Context) []*domain.ServiceMenu {
	query := `
		SELECT s.id, s.category_id, s.name, s.icon, s.icon_url, s.route,
		       s.status, s.badge, s.is_featured, s.sort_order
		FROM services s
		WHERE s.is_featured = true AND s.status != 'hidden'
		ORDER BY s.sort_order ASC
	`

	var rows []serviceRow
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return []*domain.ServiceMenu{}
	}

	return r.serviceRowsToMenus(rows)
}

// GetServiceCategories returns service categories with their services
func (r *homeRepository) GetServiceCategories(ctx context.Context) []*domain.ServiceCategory {
	// Get active categories
	catQuery := `
		SELECT id, name, slug, sort_order, is_active
		FROM service_categories
		WHERE is_active = true
		ORDER BY sort_order ASC
	`
	var catRows []serviceCategoryRow
	if err := r.db.SelectContext(ctx, &catRows, catQuery); err != nil {
		return []*domain.ServiceCategory{}
	}

	// Get all visible services
	svcQuery := `
		SELECT id, category_id, name, icon, icon_url, route, status, badge, is_featured, sort_order
		FROM services
		WHERE status != 'hidden'
		ORDER BY sort_order ASC
	`
	var svcRows []serviceRow
	if err := r.db.SelectContext(ctx, &svcRows, svcQuery); err != nil {
		return []*domain.ServiceCategory{}
	}

	// Group services by category_id
	svcByCat := make(map[string][]serviceRow)
	for _, s := range svcRows {
		catID := ""
		if s.CategoryID != nil {
			catID = *s.CategoryID
		}
		svcByCat[catID] = append(svcByCat[catID], s)
	}

	// Build category response
	var categories []*domain.ServiceCategory
	for _, cat := range catRows {
		services := r.serviceRowsToMenus(svcByCat[cat.ID])
		if len(services) == 0 {
			continue
		}
		categories = append(categories, &domain.ServiceCategory{
			ID:       cat.ID,
			Name:     cat.Name,
			Order:    cat.SortOrder,
			Services: services,
		})
	}

	return categories
}

// GetAllServices returns all services with version
func (r *homeRepository) GetAllServices(ctx context.Context) (*domain.ServicesResponse, error) {
	return &domain.ServicesResponse{
		Version:    r.GetServicesVersion(ctx),
		Featured:   r.GetFeaturedServices(ctx),
		Categories: r.GetServiceCategories(ctx),
	}, nil
}

// ========== Banners ==========

// GetBannersVersion returns current banners version based on last update
func (r *homeRepository) GetBannersVersion(ctx context.Context) string {
	var lastUpdated sql.NullTime
	err := r.db.GetContext(ctx, &lastUpdated, `SELECT MAX(updated_at) FROM banners WHERE is_active = true`)
	if err != nil || !lastUpdated.Valid {
		return time.Now().Format("2006010215")
	}
	return lastUpdated.Time.Format("2006010215")
}

// GetBanners returns active banners filtered by placement and user tier
func (r *homeRepository) GetBanners(ctx context.Context, placement string, userTier string) []*domain.Banner {
	query := `
		SELECT id, title, subtitle, image_url, thumbnail_url,
		       action_type, action_value, background_color, text_color,
		       placement, start_date, end_date, priority,
		       target_tiers, is_new_user_only, is_active
		FROM banners
		WHERE is_active = true
		  AND start_date <= NOW()
		  AND end_date >= NOW()
	`
	args := []interface{}{}
	argIdx := 1

	if placement != "" {
		query += fmt.Sprintf(" AND placement = $%d", argIdx)
		args = append(args, placement)
		argIdx++
	}

	query += " ORDER BY priority ASC"

	var rows []bannerRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return []*domain.Banner{}
	}

	// Filter by user tier in Go (target_tiers is stored as JSON text)
	var banners []*domain.Banner
	for _, row := range rows {
		// Check tier targeting
		if row.TargetTiers != nil && *row.TargetTiers != "" {
			var tiers []string
			if err := json.Unmarshal([]byte(*row.TargetTiers), &tiers); err == nil && len(tiers) > 0 {
				tierMatch := false
				for _, t := range tiers {
					if strings.EqualFold(t, userTier) {
						tierMatch = true
						break
					}
				}
				if !tierMatch {
					continue
				}
			}
		}

		banner := &domain.Banner{
			ID:              row.ID,
			Title:           row.Title,
			Subtitle:        row.Subtitle,
			ImageURL:        row.ImageURL,
			ThumbnailURL:    row.ThumbnailURL,
			BackgroundColor: derefStr(row.BackgroundColor, ""),
			TextColor:       row.TextColor,
			StartDate:       row.StartDate,
			EndDate:         row.EndDate,
			Priority:        row.Priority,
			Placement:       row.Placement,
		}

		// Build action
		if row.ActionType != nil && *row.ActionType != "" && *row.ActionType != "none" {
			banner.Action = &domain.BannerAction{
				Type:  *row.ActionType,
				Value: derefStr(row.ActionValue, ""),
			}
		}

		// Build audience
		if row.TargetTiers != nil && *row.TargetTiers != "" {
			var tiers []string
			json.Unmarshal([]byte(*row.TargetTiers), &tiers)
			if len(tiers) > 0 {
				banner.TargetAudience = &domain.Audience{
					Tiers: tiers,
				}
			}
		}

		banners = append(banners, banner)
	}

	return banners
}

// ========== Helpers ==========

func (r *homeRepository) serviceRowsToMenus(rows []serviceRow) []*domain.ServiceMenu {
	menus := make([]*domain.ServiceMenu, 0, len(rows))
	for _, row := range rows {
		menu := &domain.ServiceMenu{
			ID:       row.ID,
			Name:     row.Name,
			Icon:     row.Icon,
			IconURL:  derefStr(row.IconURL, ""),
			Route:    row.Route,
			Status:   row.Status,
			Badge:    row.Badge,
			Position: row.SortOrder,
		}
		if row.CategoryID != nil {
			menu.Category = *row.CategoryID
		}
		menus = append(menus, menu)
	}
	return menus
}

func derefStr(s *string, def string) string {
	if s != nil {
		return *s
	}
	return def
}
