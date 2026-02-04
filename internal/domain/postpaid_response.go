package domain

// PostpaidInquiryResponse represents inquiry response
type PostpaidInquiryResponse struct {
	Inquiry     *PostpaidInquiryInfo  `json:"inquiry"`
	Customer    *PostpaidCustomerInfo `json:"customer,omitempty"`
	Bill        *BillInfo             `json:"bill,omitempty"`
	Payment     *PostpaidPaymentInfo  `json:"payment,omitempty"`
	PinRequired bool                  `json:"pinRequired"`
	Notices     []string              `json:"notices"`
	Message     *string               `json:"message,omitempty"` // For no bill case
}

// PostpaidPayResponse represents payment response
type PostpaidPayResponse struct {
	Transaction *PostpaidTransactionInfo `json:"transaction"`
	Customer    *PostpaidCustomerInfo    `json:"customer"`
	Bill        *BillInfoSimple          `json:"bill"`
	Payment     *PaymentDetail           `json:"payment"`
	Receipt     *PostpaidReceiptInfo     `json:"receipt"`
	Message     *PostpaidMessageInfo     `json:"message"`
}

// PostpaidInquiryInfo for postpaid
type PostpaidInquiryInfo struct {
	InquiryID   *string `json:"inquiryId"` // null if no bill
	ServiceType string  `json:"serviceType"`
	Target      string  `json:"target"`
	TargetValid bool    `json:"targetValid"`
	NoBill      bool    `json:"noBill,omitempty"`
	ExpiresAt   *string `json:"expiresAt,omitempty"`
}

// PostpaidCustomerInfo for postpaid
type PostpaidCustomerInfo struct {
	CustomerID   string  `json:"customerId"`
	Name         string  `json:"name"`
	SegmentPower *string `json:"segmentPower,omitempty"` // PLN
	Address      *string `json:"address,omitempty"`      // PLN
	MemberCount  *int    `json:"memberCount,omitempty"`  // BPJS
	Branch       *string `json:"branch,omitempty"`       // BPJS
}

// BillInfo for postpaid inquiry
type BillInfo struct {
	Period                *string      `json:"period,omitempty"`
	PeriodCode            *string      `json:"periodCode,omitempty"`
	Periods               []BillPeriod `json:"periods,omitempty"` // For BPJS multiple periods
	TotalPeriods          *int         `json:"totalPeriods,omitempty"`
	StandMeter            *MeterInfo   `json:"standMeter,omitempty"` // PLN
	Amount                int64        `json:"amount"`
	AmountFormatted       string       `json:"amountFormatted"`
	AdminFee              int64        `json:"adminFee"`
	AdminFeeFormatted     string       `json:"adminFeeFormatted"`
	Penalty               int64        `json:"penalty,omitempty"`
	PenaltyFormatted      string       `json:"penaltyFormatted,omitempty"`
	TotalPayment          int64        `json:"totalPayment"`
	TotalPaymentFormatted string       `json:"totalPaymentFormatted"`
	DueDate               *string      `json:"dueDate,omitempty"`
}

// BillPeriod for BPJS multiple periods
type BillPeriod struct {
	Period     string `json:"period"`
	PeriodCode string `json:"periodCode"`
	Amount     int64  `json:"amount"`
}

// MeterInfo for PLN stand meter
type MeterInfo struct {
	Previous int `json:"previous"`
	Current  int `json:"current"`
	Usage    int `json:"usage"`
}

// PostpaidPaymentInfo for inquiry
type PostpaidPaymentInfo struct {
	Method                    string `json:"method"`
	BalanceAvailable          int64  `json:"balanceAvailable"`
	BalanceAvailableFormatted string `json:"balanceAvailableFormatted"`
	BalanceSufficient         bool   `json:"balanceSufficient"`
}

// BillInfoSimple for pay response
type BillInfoSimple struct {
	Period          string `json:"period"`
	Amount          int64  `json:"amount"`
	AmountFormatted string `json:"amountFormatted"`
}

// PostpaidTransactionInfo for postpaid
type PostpaidTransactionInfo struct {
	TransactionID string `json:"transactionId"`
	InquiryID     string `json:"inquiryId"`
	Status        string `json:"status"`
	ServiceType   string `json:"serviceType"`
	CompletedAt   string `json:"completedAt"`
}

// PaymentDetail for postpaid pay
type PaymentDetail struct {
	TotalPayment          int64  `json:"totalPayment"`
	TotalPaymentFormatted string `json:"totalPaymentFormatted"`
	VoucherDiscount       int64  `json:"voucherDiscount"`
	BalanceBefore         int64  `json:"balanceBefore"`
	BalanceAfter          int64  `json:"balanceAfter"`
	BalanceAfterFormatted string `json:"balanceAfterFormatted"`
}

// PostpaidReceiptInfo for postpaid
type PostpaidReceiptInfo struct {
	ReferenceNumber string  `json:"referenceNumber"`
	SerialNumber    *string `json:"serialNumber,omitempty"`
}

// PostpaidMessageInfo for postpaid
type PostpaidMessageInfo struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}
