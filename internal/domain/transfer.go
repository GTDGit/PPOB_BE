package domain

import "time"

// TransferInquiry represents transfer inquiry
type TransferInquiry struct {
	ID               string    `db:"id" json:"id"`
	UserID           string    `db:"user_id" json:"userId"`
	BankCode         string    `db:"bank_code" json:"bankCode"`
	BankName         string    `db:"bank_name" json:"bankName"`
	AccountNumber    string    `db:"account_number" json:"accountNumber"`
	AccountName      string    `db:"account_name" json:"accountName"`
	Amount           int64     `db:"amount" json:"amount"`
	AdminFee         int64     `db:"admin_fee" json:"adminFee"`
	TotalPayment     int64     `db:"total_payment" json:"totalPayment"`
	GerbangInquiryID *string   `db:"gerbang_inquiry_id" json:"gerbangInquiryId"` // ID dari Gerbang
	Fee              int64     `db:"fee" json:"fee"`                               // Fee dari Gerbang response
	ExpiresAt        time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt        time.Time `db:"created_at" json:"createdAt"`
}

// TransferTransaction represents transfer transaction
type TransferTransaction struct {
	ID                string     `db:"id" json:"id"`
	UserID            string     `db:"user_id" json:"userId"`
	InquiryID         string     `db:"inquiry_id" json:"inquiryId"`
	Status            string     `db:"status" json:"status"`
	BankCode          string     `db:"bank_code" json:"bankCode"`
	BankName          string     `db:"bank_name" json:"bankName"`
	AccountNumber     string     `db:"account_number" json:"accountNumber"`
	AccountName       string     `db:"account_name" json:"accountName"`
	Amount            int64      `db:"amount" json:"amount"`
	AdminFee          int64      `db:"admin_fee" json:"adminFee"`
	TotalPayment      int64      `db:"total_payment" json:"totalPayment"`
	Note              *string    `db:"note" json:"note"`
	BalanceBefore     int64      `db:"balance_before" json:"balanceBefore"`
	BalanceAfter      int64      `db:"balance_after" json:"balanceAfter"`
	ReferenceNumber   *string    `db:"reference_number" json:"referenceNumber"`
	GerbangTransferID *string    `db:"gerbang_transfer_id" json:"gerbangTransferId"` // ID dari Gerbang
	Purpose           *string    `db:"purpose" json:"purpose"`                        // Purpose code (01, 02, 03, 99)
	Fee               int64      `db:"fee" json:"fee"`                                // Fee dari Gerbang (bukan admin fee hardcoded)
	CompletedAt       *time.Time `db:"completed_at" json:"completedAt"`
	CreatedAt         time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updatedAt"`
}
