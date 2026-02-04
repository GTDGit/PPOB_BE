package domain

import "time"

// PostpaidInquiry represents bill inquiry
type PostpaidInquiry struct {
	ID           string  `db:"id" json:"id"`
	UserID       string  `db:"user_id" json:"userId"`
	ServiceType  string  `db:"service_type" json:"serviceType"`
	Target       string  `db:"target" json:"target"`
	ProviderID   *string `db:"provider_id" json:"providerId,omitempty"`
	CustomerID   string  `db:"customer_id" json:"customerId"`
	CustomerName string  `db:"customer_name" json:"customerName"`
	// Bill info
	Period       string `db:"period" json:"period"`
	BillAmount   int64  `db:"bill_amount" json:"billAmount"`
	AdminFee     int64  `db:"admin_fee" json:"adminFee"`
	Penalty      int64  `db:"penalty" json:"penalty"`
	TotalPayment int64  `db:"total_payment" json:"totalPayment"`
	// Status
	HasBill    bool      `db:"has_bill" json:"hasBill"`
	ExternalID *string   `db:"external_id" json:"externalId,omitempty"` // Gerbang inquiry ID
	ExpiresAt  time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
}

// PostpaidTransaction represents completed bill payment
type PostpaidTransaction struct {
	ID           string  `db:"id" json:"id"`
	UserID       string  `db:"user_id" json:"userId"`
	InquiryID    string  `db:"inquiry_id" json:"inquiryId"`
	ServiceType  string  `db:"service_type" json:"serviceType"`
	Target       string  `db:"target" json:"target"`
	ProviderID   *string `db:"provider_id" json:"providerId,omitempty"`
	CustomerID   string  `db:"customer_id" json:"customerId"`
	CustomerName string  `db:"customer_name" json:"customerName"`
	// Bill info
	Period     string `db:"period" json:"period"`
	BillAmount int64  `db:"bill_amount" json:"billAmount"`
	AdminFee   int64  `db:"admin_fee" json:"adminFee"`
	Penalty    int64  `db:"penalty" json:"penalty"`
	// Payment
	VoucherDiscount int64 `db:"voucher_discount" json:"voucherDiscount"`
	TotalPayment    int64 `db:"total_payment" json:"totalPayment"`
	BalanceBefore   int64 `db:"balance_before" json:"balanceBefore"`
	BalanceAfter    int64 `db:"balance_after" json:"balanceAfter"`
	// Receipt
	ReferenceNumber string  `db:"reference_number" json:"referenceNumber"`
	SerialNumber    *string `db:"serial_number" json:"serialNumber,omitempty"`
	ExternalID      *string `db:"external_id" json:"externalId,omitempty"` // Gerbang transaction ID
	// Status
	Status       string     `db:"status" json:"status"`
	FailedReason *string    `db:"failed_reason" json:"failedReason,omitempty"`
	CompletedAt  *time.Time `db:"completed_at" json:"completedAt,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updatedAt"`
}

// Postpaid service types
const (
	ServicePLNPostpaid   = "pln_postpaid"
	ServicePhonePostpaid = "phone_postpaid"
	ServicePDAM          = "pdam"
	ServiceBPJS          = "bpjs"
	ServiceTelkom        = "telkom"
	ServicePGN           = "pgn"
	ServicePBB           = "pbb"
	ServiceTVCable       = "tv_cable"
)

// Transaction status (reuse from prepaid)
const (
	PostpaidStatusProcessing = "processing"
	PostpaidStatusSuccess    = "success"
	PostpaidStatusFailed     = "failed"
)
