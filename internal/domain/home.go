package domain

import "time"

// Banner represents promotional banner
type Banner struct {
	ID              string        `json:"id"`
	Title           string        `json:"title"`
	Subtitle        *string       `json:"subtitle,omitempty"`
	ImageURL        string        `json:"imageUrl"`
	ThumbnailURL    *string       `json:"thumbnailUrl,omitempty"`
	Action          *BannerAction `json:"action,omitempty"`
	BackgroundColor string        `json:"backgroundColor"`
	TextColor       *string       `json:"textColor,omitempty"`
	StartDate       time.Time     `json:"startDate"`
	EndDate         time.Time     `json:"endDate"`
	Priority        int           `json:"priority"`
	Placement       string        `json:"placement"` // home, services, checkout, profile
	TargetAudience  *Audience     `json:"targetAudience,omitempty"`
}

// BannerAction represents banner action
type BannerAction struct {
	Type  string `json:"type"`  // deeplink, webview, external_url, none
	Value string `json:"value"` // route or URL
}

// Audience represents target audience for banner
type Audience struct {
	Tiers           []string `json:"tiers,omitempty"`           // BRONZE, SILVER, GOLD, PLATINUM
	IsNewUser       *bool    `json:"isNewUser,omitempty"`       // target only new users
	MinTransactions *int     `json:"minTransactions,omitempty"` // minimum transaction count
}

// ServiceMenu represents service menu item
type ServiceMenu struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Icon     string  `json:"icon"`
	IconURL  string  `json:"iconUrl"`
	Route    string  `json:"route"`
	Status   string  `json:"status"`          // active, maintenance, coming_soon, hidden
	Badge    *string `json:"badge,omitempty"` // PROMO, NEW, HOT
	Category string  `json:"category,omitempty"`
	Position int     `json:"position"`
}

// ServiceCategory represents service category
type ServiceCategory struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Order    int            `json:"order"`
	Services []*ServiceMenu `json:"services"`
}

// Banner placement types
const (
	PlacementHome     = "home"
	PlacementServices = "services"
	PlacementCheckout = "checkout"
	PlacementProfile  = "profile"
)

// Service status
const (
	ServiceStatusActive      = "active"
	ServiceStatusMaintenance = "maintenance"
	ServiceStatusComingSoon  = "coming_soon"
	ServiceStatusHidden      = "hidden"
)

// Service badges
const (
	BadgePromo = "PROMO"
	BadgeNew   = "NEW"
	BadgeHot   = "HOT"
)

// Banner action types
const (
	ActionTypeDeeplink    = "deeplink"
	ActionTypeWebview     = "webview"
	ActionTypeExternalURL = "external_url"
	ActionTypeNone        = "none"
)
