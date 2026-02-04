package domain

import "time"

// PrepaidInquiry represents inquiry data
type PrepaidInquiry struct {
	ID           string    `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"userId"`
	ServiceType  string    `db:"service_type" json:"serviceType"`
	Target       string    `db:"target" json:"target"`
	TargetValid  bool      `db:"target_valid" json:"targetValid"`
	OperatorID   *string   `db:"operator_id" json:"operatorId"`
	CustomerID   *string   `db:"customer_id" json:"customerId"`
	CustomerName *string   `db:"customer_name" json:"customerName"`
	ExpiresAt    time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}

// PrepaidOrder represents order data
type PrepaidOrder struct {
	ID            string    `db:"id" json:"id"`
	UserID        string    `db:"user_id" json:"userId"`
	InquiryID     string    `db:"inquiry_id" json:"inquiryId"`
	ProductID     string    `db:"product_id" json:"productId"`
	Status        string    `db:"status" json:"status"`
	ServiceType   string    `db:"service_type" json:"serviceType"`
	Target        string    `db:"target" json:"target"`
	ProductPrice  int64     `db:"product_price" json:"productPrice"`
	AdminFee      int64     `db:"admin_fee" json:"adminFee"`
	Subtotal      int64     `db:"subtotal" json:"subtotal"`
	TotalDiscount int64     `db:"total_discount" json:"totalDiscount"`
	TotalPayment  int64     `db:"total_payment" json:"totalPayment"`
	ExpiresAt     time.Time `db:"expires_at" json:"expiresAt"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt     time.Time `db:"updated_at" json:"updatedAt"`
}

// PrepaidTransaction represents transaction data
type PrepaidTransaction struct {
	ID              string     `db:"id" json:"id"`
	UserID          string     `db:"user_id" json:"userId"`
	OrderID         string     `db:"order_id" json:"orderId"`
	Status          string     `db:"status" json:"status"`
	ServiceType     string     `db:"service_type" json:"serviceType"`
	Target          string     `db:"target" json:"target"`
	ProductID       string     `db:"product_id" json:"productId"`
	TotalPayment    int64      `db:"total_payment" json:"totalPayment"`
	BalanceBefore   int64      `db:"balance_before" json:"balanceBefore"`
	BalanceAfter    int64      `db:"balance_after" json:"balanceAfter"`
	SerialNumber    *string    `db:"serial_number" json:"serialNumber"`
	ReferenceNumber *string    `db:"reference_number" json:"referenceNumber"`
	Token           *string    `db:"token" json:"token"`
	KWH             *string    `db:"kwh" json:"kwh"`
	CompletedAt     *time.Time `db:"completed_at" json:"completedAt"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updatedAt"`
}

// Service types - Prepaid
const (
	ServicePulsa      = "pulsa"
	ServiceData       = "data"
	ServicePLNPrepaid = "pln_prepaid"
	ServiceEwallet    = "ewallet"
	ServiceGame       = "game"
)

// Order status
const (
	OrderPendingPayment = "pending_payment"
	OrderProcessing     = "processing"
	OrderSuccess        = "success"
	OrderFailed         = "failed"
	OrderExpired        = "expired"
)

// Transaction status
const (
	TransactionPendingPayment = "pending_payment"
	TransactionProcessing     = "processing"
	TransactionSuccess        = "success"
	TransactionFailed         = "failed"
	TransactionRefunded       = "refunded"
	TransactionExpired        = "expired"
)
