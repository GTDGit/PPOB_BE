package domain

// HomeResponse represents the aggregated home screen data
type HomeResponse struct {
	User          *HomeUserInfo          `json:"user"`
	Balance       *HomeBalanceInfo       `json:"balance"`
	Services      *HomeServicesData      `json:"services,omitempty"`
	Banners       *HomeBannersData       `json:"banners,omitempty"`
	Notifications *HomeNotificationsInfo `json:"notifications"`
	Announcements []interface{}          `json:"announcements"` // Empty array for now
}

// HomeUserInfo represents user information for home screen
type HomeUserInfo struct {
	ID        string  `json:"id"`
	MIC       string  `json:"mic"` // Merchant Identification Code
	FullName  string  `json:"fullName"`
	Phone     string  `json:"phone"`
	Email     string  `json:"email"`
	Tier      string  `json:"tier"`
	AvatarURL *string `json:"avatarUrl"`
	KYCStatus string  `json:"kycStatus"`
}

// HomeBalanceInfo represents balance information for home screen
type HomeBalanceInfo struct {
	Amount         int64               `json:"amount"`
	Formatted      string              `json:"formatted"`
	LastUpdated    string              `json:"lastUpdated"`
	PendingBalance *HomePendingBalance `json:"pendingBalance,omitempty"`
	Points         *HomePointsInfo     `json:"points,omitempty"`
}

// HomePendingBalance represents pending balance
type HomePendingBalance struct {
	Amount    int64  `json:"amount"`
	Formatted string `json:"formatted"`
}

// HomePointsInfo represents points information
type HomePointsInfo struct {
	Amount    int64   `json:"amount"`
	Formatted string  `json:"formatted"`
	ExpiresAt *string `json:"expiresAt,omitempty"`
}

// HomeServicesData represents services data with versioning
type HomeServicesData struct {
	Version    string             `json:"version"`
	Featured   []*ServiceMenu     `json:"featured"`
	Categories []*ServiceCategory `json:"categories"`
}

// HomeBannersData represents banners data with versioning
type HomeBannersData struct {
	Version            string    `json:"version"`
	Placement          string    `json:"placement"`
	Items              []*Banner `json:"items"`
	AutoScrollInterval int       `json:"autoScrollInterval"` // in milliseconds
	TotalItems         int       `json:"totalItems"`
}

// HomeNotificationsInfo represents notification info
type HomeNotificationsInfo struct {
	UnreadCount int `json:"unreadCount"`
}

// BalanceResponse represents balance-only response
type BalanceResponse struct {
	Balance        *HomeBalanceInfo    `json:"balance"`
	PendingBalance *HomePendingBalance `json:"pendingBalance,omitempty"`
	Points         *HomePointsInfo     `json:"points,omitempty"`
}

// ServicesResponse represents services list response
type ServicesResponse struct {
	Version    string             `json:"version"`
	Featured   []*ServiceMenu     `json:"featured"`
	Categories []*ServiceCategory `json:"categories"`
}

// BannersResponse represents banners list response
type BannersResponse struct {
	Version            string    `json:"version"`
	Placement          string    `json:"placement"`
	Items              []*Banner `json:"items"`
	AutoScrollInterval int       `json:"autoScrollInterval"`
	TotalItems         int       `json:"totalItems"`
}
