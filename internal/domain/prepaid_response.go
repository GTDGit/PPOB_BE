package domain

// PrepaidInquiryResponse represents the inquiry response
type PrepaidInquiryResponse struct {
	Inquiry  *InquiryInfo   `json:"inquiry"`
	Products []*ProductInfo `json:"products"`
	Notices  []*NoticeInfo  `json:"notices"`
}

// InquiryInfo represents inquiry information
type InquiryInfo struct {
	InquiryID    *string       `json:"inquiryId"`
	ServiceType  string        `json:"serviceType"`
	Target       string        `json:"target"`
	TargetValid  bool          `json:"targetValid"`
	Operator     *OperatorInfo `json:"operator,omitempty"`
	Customer     *CustomerInfo `json:"customer,omitempty"`
	ErrorMessage *string       `json:"errorMessage,omitempty"`
	ExpiresAt    *string       `json:"expiresAt,omitempty"`
}

// OperatorInfo represents operator information
type OperatorInfo struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Icon    string  `json:"icon"`
	IconURL *string `json:"iconUrl,omitempty"`
}

// CustomerInfo represents customer information
type CustomerInfo struct {
	CustomerID   string  `json:"customerId"`
	Name         string  `json:"name"`
	SegmentPower *string `json:"segmentPower,omitempty"`
	Address      *string `json:"address,omitempty"`
}

// ProductInfo represents product information
type ProductInfo struct {
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	Description         string        `json:"description"`
	Category            string        `json:"category"`
	Nominal             int64         `json:"nominal"`
	Price               int64         `json:"price"`
	PriceFormatted      string        `json:"priceFormatted"`
	AdminFee            int64         `json:"adminFee"`
	AdminFeeFormatted   *string       `json:"adminFeeFormatted,omitempty"`
	TotalPrice          *int64        `json:"totalPrice,omitempty"`
	TotalPriceFormatted *string       `json:"totalPriceFormatted,omitempty"`
	Discount            *DiscountInfo `json:"discount"`
	Status              string        `json:"status"`
	Stock               string        `json:"stock"`
}

// DiscountInfo represents discount information
type DiscountInfo struct {
	Type                        string `json:"type"`
	Value                       int64  `json:"value"`
	PriceAfterDiscount          int64  `json:"priceAfterDiscount"`
	PriceAfterDiscountFormatted string `json:"priceAfterDiscountFormatted"`
	OriginalPrice               int64  `json:"originalPrice"`
	OriginalPriceFormatted      string `json:"originalPriceFormatted"`
}

// NoticeInfo represents notice information
type NoticeInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// PrepaidOrderResponse represents the order response
type PrepaidOrderResponse struct {
	Order       *OrderInfo        `json:"order"`
	Product     *OrderProductInfo `json:"product"`
	Target      *OrderTargetInfo  `json:"target"`
	Pricing     *PricingInfo      `json:"pricing"`
	Payment     *PaymentInfo      `json:"payment"`
	PINRequired bool              `json:"pinRequired"`
}

// OrderInfo represents order information
type OrderInfo struct {
	OrderID     string `json:"orderId"`
	Status      string `json:"status"`
	ServiceType string `json:"serviceType"`
	CreatedAt   string `json:"createdAt"`
	ExpiresAt   string `json:"expiresAt"`
}

// OrderProductInfo represents product in order
type OrderProductInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Nominal     int64  `json:"nominal"`
}

// OrderTargetInfo represents target information in order
type OrderTargetInfo struct {
	Number       string        `json:"number"`
	Operator     *OperatorInfo `json:"operator,omitempty"`
	CustomerName *string       `json:"customerName,omitempty"`
}

// PricingInfo represents pricing breakdown
type PricingInfo struct {
	ProductPrice           int64          `json:"productPrice"`
	ProductPriceFormatted  string         `json:"productPriceFormatted"`
	AdminFee               int64          `json:"adminFee"`
	AdminFeeFormatted      string         `json:"adminFeeFormatted"`
	Subtotal               int64          `json:"subtotal"`
	SubtotalFormatted      string         `json:"subtotalFormatted"`
	Vouchers               []*VoucherInfo `json:"vouchers"`
	TotalDiscount          int64          `json:"totalDiscount"`
	TotalDiscountFormatted string         `json:"totalDiscountFormatted"`
	TotalPayment           int64          `json:"totalPayment"`
	TotalPaymentFormatted  string         `json:"totalPaymentFormatted"`
}

// VoucherInfo represents voucher information
type VoucherInfo struct {
	Code              string `json:"code"`
	Name              string `json:"name"`
	Discount          int64  `json:"discount"`
	DiscountFormatted string `json:"discountFormatted"`
}

// PaymentInfo represents payment information
type PaymentInfo struct {
	Method                    string  `json:"method"`
	BalanceAvailable          int64   `json:"balanceAvailable"`
	BalanceAvailableFormatted string  `json:"balanceAvailableFormatted"`
	BalanceSufficient         bool    `json:"balanceSufficient"`
	Shortfall                 *int64  `json:"shortfall,omitempty"`
	ShortfallFormatted        *string `json:"shortfallFormatted,omitempty"`
}

// PrepaidPayResponse represents the pay response
type PrepaidPayResponse struct {
	Transaction *TransactionInfo        `json:"transaction"`
	Product     *TransactionProductInfo `json:"product"`
	Target      *TransactionTargetInfo  `json:"target"`
	Payment     *TransactionPaymentInfo `json:"payment"`
	Receipt     *ReceiptInfo            `json:"receipt"`
	Message     *MessageInfo            `json:"message"`
}

// TransactionInfo represents transaction information
type TransactionInfo struct {
	TransactionID       string  `json:"transactionId"`
	OrderID             string  `json:"orderId"`
	Status              string  `json:"status"`
	ServiceType         string  `json:"serviceType"`
	CompletedAt         *string `json:"completedAt,omitempty"`
	EstimatedCompletion *string `json:"estimatedCompletion,omitempty"`
}

// TransactionProductInfo represents product in transaction
type TransactionProductInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Nominal int64  `json:"nominal"`
}

// TransactionTargetInfo represents target in transaction
type TransactionTargetInfo struct {
	Number       string  `json:"number"`
	Operator     *string `json:"operator,omitempty"`
	CustomerName *string `json:"customerName,omitempty"`
}

// TransactionPaymentInfo represents payment in transaction
type TransactionPaymentInfo struct {
	TotalPayment          int64  `json:"totalPayment"`
	TotalPaymentFormatted string `json:"totalPaymentFormatted"`
	BalanceBefore         int64  `json:"balanceBefore"`
	BalanceAfter          int64  `json:"balanceAfter"`
	BalanceAfterFormatted string `json:"balanceAfterFormatted"`
}

// ReceiptInfo represents receipt information
type ReceiptInfo struct {
	SerialNumber    *string `json:"serialNumber,omitempty"`
	ReferenceNumber *string `json:"referenceNumber,omitempty"`
	Token           *string `json:"token,omitempty"`
	KWH             *string `json:"kwh,omitempty"`
}

// MessageInfo represents message information
type MessageInfo struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}
