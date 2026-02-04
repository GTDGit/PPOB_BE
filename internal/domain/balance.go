package domain

import (
	"time"
)

// Balance represents user balance
type Balance struct {
	ID              string     `db:"id" json:"id"`
	UserID          string     `db:"user_id" json:"userId"`
	Amount          int64      `db:"amount" json:"amount"`
	PendingAmount   int64      `db:"pending_amount" json:"pendingAmount"`
	Points          int        `db:"points" json:"points"`
	PointsExpiresAt *time.Time `db:"points_expires_at" json:"pointsExpiresAt"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updatedAt"`
}

// UserSettings represents user settings
type UserSettings struct {
	ID                        string    `db:"id" json:"id"`
	UserID                    string    `db:"user_id" json:"userId"`
	PINRequiredForTransaction bool      `db:"pin_required_for_transaction" json:"pinRequiredForTransaction"`
	PINRequiredMinAmount      int64     `db:"pin_required_min_amount" json:"pinRequiredMinAmount"`
	BiometricEnabled          bool      `db:"biometric_enabled" json:"biometricEnabled"`
	DefaultSellingPriceMarkup int       `db:"default_selling_price_markup" json:"defaultSellingPriceMarkup"`
	AutoSaveContact           bool      `db:"auto_save_contact" json:"autoSaveContact"`
	ShowProfitOnReceipt       bool      `db:"show_profit_on_receipt" json:"showProfitOnReceipt"`
	Language                  string    `db:"language" json:"language"`
	Currency                  string    `db:"currency" json:"currency"`
	Theme                     string    `db:"theme" json:"theme"`
	ShowPhoneOnQRIS           bool      `db:"show_phone_on_qris" json:"showPhoneOnQris"`
	ShowNameOnQRIS            bool      `db:"show_name_on_qris" json:"showNameOnQris"`
	CreatedAt                 time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt                 time.Time `db:"updated_at" json:"updatedAt"`
}

// Language enum
const (
	LanguageID = "id"
	LanguageEN = "en"
)

// Theme enum
const (
	ThemeLight  = "light"
	ThemeDark   = "dark"
	ThemeSystem = "system"
)

// BalanceHistory represents balance mutation history
type BalanceHistory struct {
	ID            string    `db:"id" json:"id"`
	UserID        string    `db:"user_id" json:"userId"`
	Type          string    `db:"type" json:"type"`         // credit, debit
	Category      string    `db:"category" json:"category"` // transaction, deposit, refund, bonus, points, fee
	Amount        int64     `db:"amount" json:"amount"`
	BalanceBefore int64     `db:"balance_before" json:"balanceBefore"`
	BalanceAfter  int64     `db:"balance_after" json:"balanceAfter"`
	ReferenceType *string   `db:"reference_type" json:"referenceType"` // deposit, prepaid, postpaid, transfer, voucher
	ReferenceID   *string   `db:"reference_id" json:"referenceId"`     // ID of related record
	Description   *string   `db:"description" json:"description"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
}

// Balance history types
const (
	BalanceHistoryCredit = "credit"
	BalanceHistoryDebit  = "debit"
)

// Balance history categories
const (
	BalanceCategoryTransaction = "transaction"
	BalanceCategoryDeposit     = "deposit"
	BalanceCategoryRefund      = "refund"
	BalanceCategoryBonus       = "bonus"
	BalanceCategoryPoints      = "points"
	BalanceCategoryFee         = "fee"
)
