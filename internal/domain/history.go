package domain

import "time"

// Transaction represents a unified transaction record
type Transaction struct {
	ID            string     `db:"id" json:"id"`
	UserID        string     `db:"user_id" json:"userId"`
	OrderID       *string    `db:"order_id" json:"orderId"`
	InquiryID     *string    `db:"inquiry_id" json:"inquiryId"`
	Type          string     `db:"type" json:"type"` // prepaid, postpaid, transfer
	ServiceType   string     `db:"service_type" json:"serviceType"`
	Target        string     `db:"target" json:"target"`
	ProductID     *string    `db:"product_id" json:"productId"`
	ProductName   string     `db:"product_name" json:"productName"`
	Amount        int64      `db:"amount" json:"amount"`
	AdminFee      int64      `db:"admin_fee" json:"adminFee"`
	Discount      int64      `db:"discount" json:"discount"`
	TotalPayment  int64      `db:"total_payment" json:"totalPayment"`
	Status        string     `db:"status" json:"status"`
	ProviderRef   *string    `db:"provider_ref" json:"providerRef"`
	SerialNumber  *string    `db:"serial_number" json:"serialNumber"`
	Token         *string    `db:"token" json:"token"`              // for PLN
	ReceiptData   *string    `db:"receipt_data" json:"receiptData"` // JSON
	FailureReason *string    `db:"failure_reason" json:"failureReason"`
	CompletedAt   *time.Time `db:"completed_at" json:"completedAt"`
	CreatedAt     time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updatedAt"`
}

// Transaction types
const (
	TransactionTypePrepaid  = "prepaid"
	TransactionTypePostpaid = "postpaid"
	TransactionTypeTransfer = "transfer"
)

// Transaction status
const (
	TransactionStatusPending    = "pending"
	TransactionStatusProcessing = "processing"
	TransactionStatusSuccess    = "success"
	TransactionStatusFailed     = "failed"
	TransactionStatusCancelled  = "cancelled"
	TransactionStatusRefunded   = "refunded"
	TransactionStatusExpired    = "expired"
)
