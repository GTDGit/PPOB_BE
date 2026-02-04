package gerbang

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// ========== Webhook Types ==========

// WebhookPayload represents the webhook payload from Gerbang API
type WebhookPayload struct {
	Event     string          `json:"event"`
	Timestamp string          `json:"timestamp"`
	Signature string          `json:"signature"`
	Data      json.RawMessage `json:"data"`
}

// PaymentWebhookData represents payment webhook event data
type PaymentWebhookData struct {
	PaymentID   string `json:"paymentId"`
	ReferenceID string `json:"referenceId"`
	Method      string `json:"method"`
	Amount      int64  `json:"amount"`
	Fee         int64  `json:"fee"`
	TotalAmount int64  `json:"totalAmount"`
	Status      string `json:"status"`
	PaidAt      string `json:"paidAt,omitempty"`
}

// TransactionWebhookData represents transaction webhook event data
type TransactionWebhookData struct {
	TransactionID string  `json:"transactionId"`
	ReferenceID   string  `json:"referenceId"`
	SKUCode       string  `json:"skuCode"`
	CustomerNo    string  `json:"customerNo"`
	Status        string  `json:"status"`
	SerialNumber  *string `json:"serialNumber"`
	ProcessedAt   string  `json:"processedAt,omitempty"`
}

// TransferWebhookData represents transfer webhook event data
type TransferWebhookData struct {
	TransferID         string  `json:"transferId"`
	ReferenceID        string  `json:"referenceId"`
	Status             string  `json:"status"`
	TransferType       string  `json:"transferType"`
	Route              *string `json:"route,omitempty"`
	BankCode           string  `json:"bankCode"`
	BankShortName      string  `json:"bankShortName"`
	BankName           string  `json:"bankName"`
	AccountNumber      string  `json:"accountNumber"`
	AccountName        string  `json:"accountName"`
	Amount             int64   `json:"amount"`
	Fee                int64   `json:"fee"`
	TotalAmount        int64   `json:"totalAmount"`
	Purpose            string  `json:"purpose"`
	PurposeDescription string  `json:"purposeDescription"`
	Remark             *string `json:"remark,omitempty"`
	ProviderRef        *string `json:"providerRef,omitempty"`
	FailedReason       *string `json:"failedReason,omitempty"`
	FailedCode         *string `json:"failedCode,omitempty"`
	CreatedAt          string  `json:"createdAt"`
	CompletedAt        *string `json:"completedAt,omitempty"`
	FailedAt           *string `json:"failedAt,omitempty"`
}

// Webhook events
const (
	EventPaymentPaid        = "payment.paid"
	EventPaymentExpired     = "payment.expired"
	EventPaymentCancelled   = "payment.cancelled"
	EventTransactionSuccess = "transaction.success"
	EventTransactionFailed  = "transaction.failed"
	EventTransactionPending = "transaction.pending"
	EventTransferSuccess    = "transfer.success"
	EventTransferFailed     = "transfer.failed"
)

// ========== Webhook Functions ==========

// VerifySignature verifies webhook signature using HMAC-SHA256
func VerifySignature(payload []byte, signature, secretKey string) bool {
	// Create HMAC-SHA256 hash
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// ParseWebhook parses webhook payload from raw bytes
func ParseWebhook(payload []byte) (*WebhookPayload, error) {
	var webhook WebhookPayload
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse webhook: %w", err)
	}
	return &webhook, nil
}

// ParsePaymentData parses payment webhook data
func (w *WebhookPayload) ParsePaymentData() (*PaymentWebhookData, error) {
	if w.Event != EventPaymentPaid &&
		w.Event != EventPaymentExpired &&
		w.Event != EventPaymentCancelled {
		return nil, errors.New("not a payment event")
	}

	var data PaymentWebhookData
	if err := json.Unmarshal(w.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse payment data: %w", err)
	}
	return &data, nil
}

// ParseTransactionData parses transaction webhook data
func (w *WebhookPayload) ParseTransactionData() (*TransactionWebhookData, error) {
	if w.Event != EventTransactionSuccess &&
		w.Event != EventTransactionFailed &&
		w.Event != EventTransactionPending {
		return nil, errors.New("not a transaction event")
	}

	var data TransactionWebhookData
	if err := json.Unmarshal(w.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse transaction data: %w", err)
	}
	return &data, nil
}

// IsPaymentEvent checks if event is a payment event
func (w *WebhookPayload) IsPaymentEvent() bool {
	return w.Event == EventPaymentPaid ||
		w.Event == EventPaymentExpired ||
		w.Event == EventPaymentCancelled
}

// IsTransactionEvent checks if event is a transaction event
func (w *WebhookPayload) IsTransactionEvent() bool {
	return w.Event == EventTransactionSuccess ||
		w.Event == EventTransactionFailed ||
		w.Event == EventTransactionPending
}

// ParseTransferData parses transfer webhook data
func (w *WebhookPayload) ParseTransferData() (*TransferWebhookData, error) {
	if w.Event != EventTransferSuccess &&
		w.Event != EventTransferFailed {
		return nil, errors.New("not a transfer event")
	}

	var data TransferWebhookData
	if err := json.Unmarshal(w.Data, &data); err != nil {
		return nil, fmt.Errorf("failed to parse transfer data: %w", err)
	}
	return &data, nil
}

// IsTransferEvent checks if event is a transfer event
func (w *WebhookPayload) IsTransferEvent() bool {
	return w.Event == EventTransferSuccess ||
		w.Event == EventTransferFailed
}
