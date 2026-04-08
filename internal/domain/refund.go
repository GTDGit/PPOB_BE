package domain

import "time"

type Refund struct {
	ID                  string     `db:"id" json:"id"`
	PublicID            string     `db:"public_id" json:"publicId"`
	SourceTransactionID string     `db:"source_transaction_id" json:"sourceTransactionId"`
	SourceType          string     `db:"source_type" json:"sourceType"`
	UserID              string     `db:"user_id" json:"userId"`
	Amount              int64      `db:"amount" json:"amount"`
	Reason              *string    `db:"reason" json:"reason,omitempty"`
	Status              string     `db:"status" json:"status"`
	RefundedAt          *time.Time `db:"refunded_at" json:"refundedAt,omitempty"`
	CreatedAt           time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt           time.Time  `db:"updated_at" json:"updatedAt"`
}

const (
	RefundStatusSuccess = "success"
	RefundStatusFailed  = "failed"
)
