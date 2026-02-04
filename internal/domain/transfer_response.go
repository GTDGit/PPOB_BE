package domain

// TransferInquiryResponse represents the inquiry response
type TransferInquiryResponse struct {
	Inquiry     *TransferInquiryInfo     `json:"inquiry"`
	Destination *TransferDestinationInfo `json:"destination"`
	Transfer    *TransferAmountInfo      `json:"transfer"`
	Payment     *PaymentInfo             `json:"payment,omitempty"`
	PINRequired bool                     `json:"pinRequired"`
	Notices     []*NoticeInfo            `json:"notices"`
}

// TransferInquiryInfo represents inquiry information
type TransferInquiryInfo struct {
	InquiryID string `json:"inquiryId"`
	ExpiresAt string `json:"expiresAt"`
}

// TransferDestinationInfo represents destination account information
type TransferDestinationInfo struct {
	BankCode      string  `json:"bankCode"`
	BankName      string  `json:"bankName"`
	BankShortName *string `json:"bankShortName,omitempty"`
	BankIcon      *string `json:"bankIcon,omitempty"`
	AccountNumber string  `json:"accountNumber"`
	AccountName   string  `json:"accountName"`
}

// TransferAmountInfo represents transfer amount breakdown
type TransferAmountInfo struct {
	Amount                int64  `json:"amount"`
	AmountFormatted       string `json:"amountFormatted"`
	AdminFee              int64  `json:"adminFee"`
	AdminFeeFormatted     string `json:"adminFeeFormatted"`
	TotalPayment          int64  `json:"totalPayment"`
	TotalPaymentFormatted string `json:"totalPaymentFormatted"`
}

// TransferExecuteResponse represents the execute response
type TransferExecuteResponse struct {
	Transaction *TransferTransactionInfo `json:"transaction"`
	Destination *TransferDestInfo        `json:"destination"`
	Transfer    *TransferExecInfo        `json:"transfer"`
	Payment     *TransferPaymentInfo     `json:"payment"`
	Receipt     *ReceiptInfo             `json:"receipt"`
	Message     *MessageInfo             `json:"message"`
}

// TransferTransactionInfo represents transaction information
type TransferTransactionInfo struct {
	TransactionID string  `json:"transactionId"`
	InquiryID     string  `json:"inquiryId"`
	Status        string  `json:"status"`
	CompletedAt   *string `json:"completedAt,omitempty"`
}

// TransferDestInfo represents destination in execute response
type TransferDestInfo struct {
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
}

// TransferExecInfo represents transfer info in execute response
type TransferExecInfo struct {
	Amount          int64   `json:"amount"`
	AmountFormatted string  `json:"amountFormatted"`
	Note            *string `json:"note,omitempty"`
}

// TransferPaymentInfo represents payment information in execute response
type TransferPaymentInfo struct {
	TotalPayment          int64  `json:"totalPayment"`
	TotalPaymentFormatted string `json:"totalPaymentFormatted"`
	BalanceBefore         int64  `json:"balanceBefore"`
	BalanceAfter          int64  `json:"balanceAfter"`
	BalanceAfterFormatted string `json:"balanceAfterFormatted"`
}

// KYCInfo represents KYC information in error response
type KYCInfo struct {
	Status         string `json:"status"`
	VerifyURL      string `json:"verifyUrl"`
	VerifyDeeplink string `json:"verifyDeeplink"`
}
