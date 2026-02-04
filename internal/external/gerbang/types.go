package gerbang

import "time"

// ========== Common Response ==========

type Response struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    *Meta       `json:"meta"`
}

type Meta struct {
	RequestID  string      `json:"requestId"`
	Timestamp  string      `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

// ========== Payment Types (Deposit) ==========

type PaymentMethod struct {
	Type string `json:"type"` // VA, QRIS, RETAIL
	Code string `json:"code"` // bank code or provider code
}

type CustomerInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type Fee struct {
	Type   string  `json:"type"`   // flat, percent
	Amount float64 `json:"amount"` // amount or percentage
}

// Create Payment Request
type CreatePaymentRequest struct {
	ReferenceID   string            `json:"referenceId"`
	PaymentMethod PaymentMethod     `json:"paymentMethod"`
	Amount        int64             `json:"amount"`
	Customer      CustomerInfo      `json:"customer"`
	Description   string            `json:"description"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	ExpiredAt     *time.Time        `json:"expiredAt,omitempty"`
}

// Payment Response
type PaymentResponse struct {
	PaymentID          string                 `json:"paymentId"`
	ReferenceID        string                 `json:"referenceId"`
	PaymentMethod      PaymentMethod          `json:"paymentMethod"`
	Status             string                 `json:"status"`
	Amount             int64                  `json:"amount"`
	Fee                int64                  `json:"fee"`
	TotalAmount        int64                  `json:"totalAmount"`
	Customer           CustomerInfo           `json:"customer"`
	PaymentDetail      map[string]interface{} `json:"paymentDetail"`
	PaymentInstruction map[string]interface{} `json:"paymentInstruction"`
	Description        string                 `json:"description"`
	Metadata           map[string]string      `json:"metadata"`
	ExpiredAt          string                 `json:"expiredAt"`
	CreatedAt          string                 `json:"createdAt"`
	PaidAt             *string                `json:"paidAt"`
}

// VA Detail (parsed from PaymentDetail)
type VADetail struct {
	BankCode    string `json:"bankCode"`
	BankName    string `json:"bankName"`
	VANumber    string `json:"vaNumber"`
	AccountName string `json:"accountName"`
	SenderBank  string `json:"senderBank,omitempty"`
	SenderName  string `json:"senderName,omitempty"`
}

// QRIS Detail
type QRISDetail struct {
	QRString            string `json:"qrString"`
	QRImageURL          string `json:"qrImageUrl"`
	ProviderReferenceNo string `json:"providerReferenceNo"`
}

// Retail Detail
type RetailDetail struct {
	RetailName  string `json:"retailName"`
	PaymentCode string `json:"paymentCode"`
}

// ========== PPOB Types (Transaction) ==========

type Product struct {
	SKUCode     string `json:"skuCode"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Brand       string `json:"brand"`
	Type        string `json:"type"` // prepaid, postpaid
	Price       int64  `json:"price,omitempty"`
	Admin       int64  `json:"admin,omitempty"`
	Commission  int64  `json:"commission,omitempty"`
	IsActive    bool   `json:"isActive"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updatedAt"`
}

type TransactionRequest struct {
	ReferenceID   string `json:"referenceId"`
	SKUCode       string `json:"skuCode"`
	CustomerNo    string `json:"customerNo"`
	Type          string `json:"type"`                    // prepaid, inquiry, payment
	TransactionID string `json:"transactionId,omitempty"` // for payment type
}

type TransactionResponse struct {
	TransactionID string                 `json:"transactionId"`
	ReferenceID   string                 `json:"referenceId"`
	SKUCode       string                 `json:"skuCode"`
	CustomerNo    string                 `json:"customerNo"`
	CustomerName  string                 `json:"customerName,omitempty"`
	Type          string                 `json:"type"`
	Status        string                 `json:"status"`
	SerialNumber  *string                `json:"serialNumber"`
	Price         int64                  `json:"price,omitempty"`
	Admin         int64                  `json:"admin,omitempty"`
	Amount        int64                  `json:"amount,omitempty"`
	TotalAmount   int64                  `json:"totalAmount,omitempty"`
	Period        string                 `json:"period,omitempty"`
	Description   map[string]interface{} `json:"description,omitempty"`
	RetryCount    int                    `json:"retryCount,omitempty"`
	NextRetryAt   string                 `json:"nextRetryAt,omitempty"`
	ExpiredAt     string                 `json:"expiredAt,omitempty"`
	CreatedAt     string                 `json:"createdAt"`
	ProcessedAt   *string                `json:"processedAt"`
}

// Payment/Transaction Status
const (
	StatusPending    = "Pending"
	StatusProcessing = "Processing"
	StatusSuccess    = "Success"
	StatusPaid       = "Paid"
	StatusFailed     = "Failed"
	StatusExpired    = "Expired"
	StatusCancelled  = "Cancelled"
)

// ========== Bank Code Types ==========

type BankCodeItem struct {
	Name      string  `json:"name"`
	ShortName string  `json:"shortName"`
	Code      string  `json:"code"`
	SwiftCode *string `json:"swiftCode"`
}

// ========== Transfer Types ==========

// GerbangTransferInquiryRequest untuk Gerbang API
type GerbangTransferInquiryRequest struct {
	BankCode      string `json:"bankCode"`
	AccountNumber string `json:"accountNumber"`
}

// GerbangTransferInquiryResponse dari Gerbang API
type GerbangTransferInquiryResponse struct {
	BankCode      string `json:"bankCode"`
	BankShortName string `json:"bankShortName"`
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
	TransferType  string `json:"transferType"` // ABAIKAN field ini
	InquiryID     string `json:"inquiryId"`
	ExpiredAt     string `json:"expiredAt"`
}

// GerbangTransferExecuteRequest untuk Gerbang API
type GerbangTransferExecuteRequest struct {
	ReferenceID   string `json:"referenceId"`
	InquiryID     string `json:"inquiryId"`
	BankCode      string `json:"bankCode"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
	Amount        int64  `json:"amount"`
	Purpose       string `json:"purpose"`         // WAJIB, default "99" jika kosong
	Remark        string `json:"remark,omitempty"` // Optional
}

// GerbangTransferExecuteResponse dari Gerbang API
type GerbangTransferExecuteResponse struct {
	TransferID         string `json:"transferId"`
	ReferenceID        string `json:"referenceId"`
	Status             string `json:"status"`
	TransferType       string `json:"transferType"` // ABAIKAN
	BankCode           string `json:"bankCode"`
	BankShortName      string `json:"bankShortName"`
	BankName           string `json:"bankName"`
	AccountNumber      string `json:"accountNumber"`
	AccountName        string `json:"accountName"`
	Amount             int64  `json:"amount"`
	Fee                int64  `json:"fee"` // FEE INI YANG DIKURANGKAN DARI SALDO USER
	TotalAmount        int64  `json:"totalAmount"`
	Purpose            string `json:"purpose"`
	PurposeDescription string `json:"purposeDescription"`
	Remark             string `json:"remark"`
	CreatedAt          string `json:"createdAt"`
}
