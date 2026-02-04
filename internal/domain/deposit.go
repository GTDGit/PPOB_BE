package domain

import "time"

// Deposit represents a deposit transaction
type Deposit struct {
	ID              string     `db:"id" json:"id"`
	UserID          string     `db:"user_id" json:"userId"`
	Method          string     `db:"method" json:"method"`
	ProviderCode    *string    `db:"provider_code" json:"providerCode"`
	BankCode        *string    `db:"bank_code" json:"bankCode"`
	Amount          int64      `db:"amount" json:"amount"`
	AdminFee        int64      `db:"admin_fee" json:"adminFee"`
	UniqueCode      int        `db:"unique_code" json:"uniqueCode"`
	TotalAmount     int64      `db:"total_amount" json:"totalAmount"`
	Status          string     `db:"status" json:"status"`
	PaymentData     *string    `db:"payment_data" json:"paymentData"` // JSON
	ExternalID      *string    `db:"external_id" json:"externalId"`
	ReferenceNumber *string    `db:"reference_number" json:"referenceNumber"` // Payment reference from gateway
	PayerName       *string    `db:"payer_name" json:"payerName"`             // Name of payer
	PayerBank       *string    `db:"payer_bank" json:"payerBank"`             // Bank used by payer
	PayerAccount    *string    `db:"payer_account" json:"payerAccount"`       // Account number used by payer
	ExpiresAt       time.Time  `db:"expires_at" json:"expiresAt"`
	PaidAt          *time.Time `db:"paid_at" json:"paidAt"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updatedAt"`
}

// DepositMethod represents available deposit method
type DepositMethod struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	IconURL     string `json:"iconUrl"`
	FeeType     string `json:"feeType"`
	FeeAmount   int64  `json:"feeAmount"`
	MinAmount   int64  `json:"minAmount"`
	MaxAmount   int64  `json:"maxAmount"`
	Schedule    string `json:"schedule"`
	Position    int    `json:"position"`
	Status      string `json:"status"`
}

// RetailProvider represents retail store provider
type RetailProvider struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	IconURL   string `json:"iconUrl"`
	Fee       int64  `json:"fee"`
	MinAmount int64  `json:"minAmount"`
	MaxAmount int64  `json:"maxAmount"`
	Status    string `json:"status"`
}

// VABank represents virtual account bank
type VABank struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	Icon      string `json:"icon"`
	IconURL   string `json:"iconUrl"`
	Fee       int64  `json:"fee"`
	MinAmount int64  `json:"minAmount"`
	MaxAmount int64  `json:"maxAmount"`
	Status    string `json:"status"`
}

// CompanyBankAccount represents company's bank account for transfer
type CompanyBankAccount struct {
	BankCode      string `json:"bankCode"`
	BankName      string `json:"bankName"`
	BankShortName string `json:"bankShortName"`
	BankIcon      string `json:"bankIcon"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
	Status        string `json:"status"`
}

// Deposit methods
const (
	DepositMethodBankTransfer   = "bank_transfer"
	DepositMethodQRIS           = "qris"
	DepositMethodRetail         = "retail"
	DepositMethodVirtualAccount = "virtual_account"
)

// Deposit status
const (
	DepositStatusPending = "pending"
	DepositStatusSuccess = "success"
	DepositStatusExpired = "expired"
	DepositStatusFailed  = "failed"
)

// Retail providers
const (
	RetailAlfamart  = "alfamart"
	RetailIndomaret = "indomaret"
)

// Fee types
const (
	FeeTypeFixed      = "fixed"
	FeeTypePercentage = "percentage"
)
