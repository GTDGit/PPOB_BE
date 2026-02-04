package domain

// DepositMethodsResponse represents list of deposit methods
type DepositMethodsResponse struct {
	Methods []*DepositMethod `json:"methods"`
}

// BankTransferResponse represents bank transfer deposit creation response
type BankTransferResponse struct {
	Deposit      *DepositInfo          `json:"deposit"`
	PaymentInfo  *BankTransferPayment  `json:"paymentInfo"`
	BankAccounts []*CompanyBankAccount `json:"bankAccounts"`
	Instructions []string              `json:"instructions"`
}

// QRISResponse represents QRIS deposit creation response
type QRISResponse struct {
	Deposit      *DepositInfo `json:"deposit"`
	PaymentInfo  *QRISPayment `json:"paymentInfo"`
	Instructions []string     `json:"instructions"`
}

// RetailProvidersResponse represents list of retail providers
type RetailProvidersResponse struct {
	Providers []*RetailProvider `json:"providers"`
}

// RetailResponse represents retail deposit creation response
type RetailResponse struct {
	Deposit      *DepositInfo   `json:"deposit"`
	PaymentInfo  *RetailPayment `json:"paymentInfo"`
	Instructions []string       `json:"instructions"`
}

// VABanksResponse represents list of VA banks
type VABanksResponse struct {
	Banks []*VABank `json:"banks"`
}

// VAResponse represents VA deposit creation response
type VAResponse struct {
	Deposit      *DepositInfo `json:"deposit"`
	PaymentInfo  *VAPayment   `json:"paymentInfo"`
	Instructions []string     `json:"instructions"`
}

// DepositStatusResponse represents deposit status check response
type DepositStatusResponse struct {
	Deposit *DepositDetail `json:"deposit"`
}

// DepositHistoryResponse represents deposit history list response
type DepositHistoryResponse struct {
	Deposits   []*DepositSummary `json:"deposits"`
	Pagination *Pagination       `json:"pagination"`
}

// DepositInfo represents basic deposit information
type DepositInfo struct {
	DepositID            string `json:"depositId"`
	Method               string `json:"method"`
	MethodName           string `json:"methodName"`
	Amount               int64  `json:"amount"`
	AmountFormatted      string `json:"amountFormatted"`
	AdminFee             int64  `json:"adminFee"`
	AdminFeeFormatted    string `json:"adminFeeFormatted"`
	UniqueCode           int    `json:"uniqueCode,omitempty"`
	TotalAmount          int64  `json:"totalAmount"`
	TotalAmountFormatted string `json:"totalAmountFormatted"`
	Status               string `json:"status"`
	StatusLabel          string `json:"statusLabel"`
	ExpiresAt            string `json:"expiresAt"`
	ExpiresAtFormatted   string `json:"expiresAtFormatted"`
	CreatedAt            string `json:"createdAt"`
}

// DepositDetail represents detailed deposit information
type DepositDetail struct {
	DepositID            string      `json:"depositId"`
	Method               string      `json:"method"`
	MethodName           string      `json:"methodName"`
	Amount               int64       `json:"amount"`
	AmountFormatted      string      `json:"amountFormatted"`
	AdminFee             int64       `json:"adminFee"`
	AdminFeeFormatted    string      `json:"adminFeeFormatted"`
	UniqueCode           int         `json:"uniqueCode,omitempty"`
	TotalAmount          int64       `json:"totalAmount"`
	TotalAmountFormatted string      `json:"totalAmountFormatted"`
	Status               string      `json:"status"`
	StatusLabel          string      `json:"statusLabel"`
	PaymentInfo          interface{} `json:"paymentInfo,omitempty"`
	ExpiresAt            string      `json:"expiresAt"`
	ExpiresAtFormatted   string      `json:"expiresAtFormatted"`
	PaidAt               *string     `json:"paidAt,omitempty"`
	CreatedAt            string      `json:"createdAt"`
}

// DepositSummary represents summary deposit info for list
type DepositSummary struct {
	DepositID            string `json:"depositId"`
	Method               string `json:"method"`
	MethodName           string `json:"methodName"`
	Amount               int64  `json:"amount"`
	AmountFormatted      string `json:"amountFormatted"`
	TotalAmount          int64  `json:"totalAmount"`
	TotalAmountFormatted string `json:"totalAmountFormatted"`
	Status               string `json:"status"`
	StatusLabel          string `json:"statusLabel"`
	CreatedAt            string `json:"createdAt"`
	CreatedAtFormatted   string `json:"createdAtFormatted"`
}

// BankTransferPayment represents bank transfer payment info
type BankTransferPayment struct {
	TotalTransfer          int64  `json:"totalTransfer"`
	TotalTransferFormatted string `json:"totalTransferFormatted"`
	UniqueCode             int    `json:"uniqueCode"`
	ValidUntil             string `json:"validUntil"`
	ValidUntilFormatted    string `json:"validUntilFormatted"`
}

// QRISPayment represents QRIS payment info
type QRISPayment struct {
	QRISString          string `json:"qrisString"`
	QRISImageURL        string `json:"qrisImageUrl"`
	ValidUntil          string `json:"validUntil"`
	ValidUntilFormatted string `json:"validUntilFormatted"`
}

// RetailPayment represents retail payment info
type RetailPayment struct {
	ProviderCode        string `json:"providerCode"`
	ProviderName        string `json:"providerName"`
	PaymentCode         string `json:"paymentCode"`
	ValidUntil          string `json:"validUntil"`
	ValidUntilFormatted string `json:"validUntilFormatted"`
}

// VAPayment represents virtual account payment info
type VAPayment struct {
	BankCode            string `json:"bankCode"`
	BankName            string `json:"bankName"`
	VANumber            string `json:"vaNumber"`
	ValidUntil          string `json:"validUntil"`
	ValidUntilFormatted string `json:"validUntilFormatted"`
}
