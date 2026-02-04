package domain

// TransactionListResponse represents transaction history list response
type TransactionListResponse struct {
	Transactions []*TransactionSummary `json:"transactions"`
	Pagination   *Pagination           `json:"pagination"`
}

// TransactionSummary represents summary transaction info for list
type TransactionSummary struct {
	ID                   string  `json:"id"`
	Type                 string  `json:"type"`
	ServiceType          *string `json:"serviceType,omitempty"`
	Title                string  `json:"title"`
	Description          string  `json:"description"`
	Amount               int64   `json:"amount"`
	AmountFormatted      string  `json:"amountFormatted"`
	AdminFee             int64   `json:"adminFee"`
	TotalAmount          int64   `json:"totalAmount"`
	TotalAmountFormatted string  `json:"totalAmountFormatted"`
	Status               string  `json:"status"`
	StatusLabel          string  `json:"statusLabel"`
	CreatedAt            string  `json:"createdAt"`
	CompletedAt          *string `json:"completedAt"`
	Icon                 string  `json:"icon"`
	IconURL              string  `json:"iconUrl"`
}

// Pagination represents pagination information
type Pagination struct {
	CurrentPage  int  `json:"currentPage"`
	TotalPages   int  `json:"totalPages"`
	TotalItems   int  `json:"totalItems"`
	ItemsPerPage int  `json:"itemsPerPage"`
	HasNextPage  bool `json:"hasNextPage"`
	HasPrevPage  bool `json:"hasPrevPage"`
}

// TransactionDetailResponse represents detailed transaction response
type TransactionDetailResponse struct {
	Transaction  *HistoryTransactionInfo `json:"transaction"`
	Product      *HistoryProductInfo     `json:"product,omitempty"`
	Target       *TargetInfo             `json:"target,omitempty"`
	Destination  *DestinationInfo        `json:"destination,omitempty"`
	Transfer     *TransferInfo           `json:"transfer,omitempty"`
	Pricing      *HistoryPricingInfo     `json:"pricing"`
	Payment      *HistoryPaymentInfo     `json:"payment"`
	Receipt      *HistoryReceiptInfo     `json:"receipt,omitempty"`
	SellingPrice *SellingPriceInfo       `json:"sellingPrice,omitempty"`
	Actions      *TransactionActions     `json:"actions"`
}

// HistoryTransactionInfo represents basic transaction information
type HistoryTransactionInfo struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	ServiceType *string `json:"serviceType,omitempty"`
	Title       string  `json:"title"`
	Status      string  `json:"status"`
	StatusLabel string  `json:"statusLabel"`
	CreatedAt   string  `json:"createdAt"`
	CompletedAt *string `json:"completedAt"`
}

// HistoryProductInfo represents product information
type HistoryProductInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Nominal     int64   `json:"nominal"`
	Description *string `json:"description,omitempty"`
}

// HistoryPricingInfo represents pricing breakdown for history
type HistoryPricingInfo struct {
	ProductPrice             *int64  `json:"productPrice,omitempty"`
	ProductPriceFormatted    *string `json:"productPriceFormatted,omitempty"`
	TransferAmount           *int64  `json:"transferAmount,omitempty"`
	TransferAmountFormatted  *string `json:"transferAmountFormatted,omitempty"`
	AdminFee                 int64   `json:"adminFee"`
	AdminFeeFormatted        string  `json:"adminFeeFormatted"`
	VoucherDiscount          *int64  `json:"voucherDiscount,omitempty"`
	VoucherDiscountFormatted *string `json:"voucherDiscountFormatted,omitempty"`
	TotalPayment             int64   `json:"totalPayment"`
	TotalPaymentFormatted    string  `json:"totalPaymentFormatted"`
}

// HistoryPaymentInfo represents payment method information
type HistoryPaymentInfo struct {
	Method      string `json:"method"`
	MethodLabel string `json:"methodLabel"`
}

// HistoryReceiptInfo represents receipt information
type HistoryReceiptInfo struct {
	SerialNumber    *string `json:"serialNumber,omitempty"`
	ReferenceNumber *string `json:"referenceNumber,omitempty"`
	Token           *string `json:"token,omitempty"`
	KWH             *string `json:"kwh,omitempty"`
}

// TargetInfo represents target information (phone, meter, etc)
type TargetInfo struct {
	Number       *string       `json:"number,omitempty"`
	Name         *string       `json:"name,omitempty"`
	Operator     *OperatorInfo `json:"operator,omitempty"`
	CustomerID   *string       `json:"customerId,omitempty"`
	CustomerName *string       `json:"customerName,omitempty"`
	SegmentPower *string       `json:"segmentPower,omitempty"`
	Address      *string       `json:"address,omitempty"`
}

// DestinationInfo represents destination information (for transfer)
type DestinationInfo struct {
	BankCode      string `json:"bankCode"`
	BankName      string `json:"bankName"`
	BankShortName string `json:"bankShortName"`
	BankIcon      string `json:"bankIcon"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
}

// TransferInfo represents transfer details
type TransferInfo struct {
	Amount          int64   `json:"amount"`
	AmountFormatted string  `json:"amountFormatted"`
	Note            *string `json:"note,omitempty"`
}

// SellingPriceInfo represents selling price information
type SellingPriceInfo struct {
	Enabled          bool   `json:"enabled"`
	Price            int64  `json:"price"`
	PriceFormatted   string `json:"priceFormatted"`
	PaymentType      string `json:"paymentType"`
	PaymentTypeLabel string `json:"paymentTypeLabel"`
}

// TransactionActions represents available actions
type TransactionActions struct {
	CanShare           bool `json:"canShare"`
	CanDownloadReceipt bool `json:"canDownloadReceipt"`
	CanRepeat          bool `json:"canRepeat"`
	CanCopyToken       bool `json:"canCopyToken,omitempty"`
}

// ReceiptResponse represents receipt data response
type ReceiptResponse struct {
	TransactionID string `json:"transactionId"`
	Title         string `json:"title"`
	Subtitle      string `json:"subtitle"`
	Amount        string `json:"amount"`
	Date          string `json:"date"`
	ReceiptURL    string `json:"receiptUrl"`
}

// ShareReceiptResponse represents shareable receipt data
type ShareReceiptResponse struct {
	ShareText string `json:"shareText"`
	ShareURL  string `json:"shareUrl"`
	ImageURL  string `json:"imageUrl"`
	DeepLink  string `json:"deepLink"`
	CanShare  bool   `json:"canShare"`
	CanCopy   bool   `json:"canCopy"`
	CanPrint  bool   `json:"canPrint"`
	CanEmail  bool   `json:"canEmail"`
}
