package domain

import "time"

// Product represents synced product from GTD API
type Product struct {
	ID           string    `db:"id" json:"id"`
	SKUCode      string    `db:"sku_code" json:"skuCode"` // Primary identifier from GTD
	Name         string    `db:"name" json:"name"`
	Category     string    `db:"category" json:"category"`     // Pulsa, Data, PLN, etc
	Brand        string    `db:"brand" json:"brand"`           // TELKOMSEL, INDOSAT, PLN, etc
	Type         string    `db:"type" json:"type"`             // prepaid, postpaid
	Price        int64     `db:"price" json:"price"`           // Selling price (from GTD, includes profit)
	Admin        int64     `db:"admin" json:"admin"`           // Admin fee (for postpaid)
	Commission   int64     `db:"commission" json:"commission"` // Commission (info only)
	IsActive     bool      `db:"is_active" json:"isActive"`
	Description  string    `db:"description" json:"description"`
	GTDUpdatedAt time.Time `db:"gtd_updated_at" json:"gtdUpdatedAt"` // Timestamp from GTD
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}

// Product types
const (
	ProductTypePrepaid  = "prepaid"
	ProductTypePostpaid = "postpaid"
)

// Product categories
const (
	CategoryPulsa    = "Pulsa"
	CategoryData     = "Data"
	CategoryPLN      = "PLN"
	CategoryPLNToken = "PLN Token"
	CategoryGame     = "Game"
	CategoryEwallet  = "E-Wallet"
	CategoryBPJS     = "BPJS"
	CategoryPDAM     = "PDAM"
	CategoryTelkom   = "Telkom"
	CategoryTV       = "TV"
)

// Operator represents mobile operator (Telkomsel, Indosat, etc)
type Operator struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Prefixes []string `json:"prefixes"`
	Icon     string   `json:"icon"`
	IconURL  string   `json:"iconUrl"`
	Status   string   `json:"status"`
}

// EwalletProvider represents e-wallet provider (GoPay, OVO, etc)
type EwalletProvider struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Icon             string `json:"icon"`
	IconURL          string `json:"iconUrl"`
	InputLabel       string `json:"inputLabel"`
	InputPlaceholder string `json:"inputPlaceholder"`
	InputType        string `json:"inputType"`
	Status           string `json:"status"`
}

// PDAMRegion represents PDAM region
type PDAMRegion struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Province string `json:"province"`
	Status   string `json:"status"`
}

// Bank represents bank information (synced from Gerbang API)
type Bank struct {
	ID                   string     `db:"id" json:"id"`
	Code                 string     `db:"code" json:"code"`
	Name                 string     `db:"name" json:"name"`
	ShortName            string     `db:"short_name" json:"shortName"`
	SwiftCode            *string    `db:"swift_code" json:"swiftCode,omitempty"`
	Icon                 string     `db:"icon" json:"icon"`
	IconURL              string     `db:"icon_url" json:"iconUrl"`
	TransferFee          int64      `db:"transfer_fee" json:"transferFee"`
	TransferFeeFormatted string     `db:"transfer_fee_formatted" json:"transferFeeFormatted"`
	IsPopular            bool       `db:"is_popular" json:"isPopular"`
	Status               string     `db:"status" json:"status"`
	CreatedAt            time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt            time.Time  `db:"updated_at" json:"updatedAt"`
}

// TVProvider represents TV cable provider
type TVProvider struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	IconURL    string `json:"iconUrl"`
	InputLabel string `json:"inputLabel"`
	Status     string `json:"status"`
}

// Product status
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)
