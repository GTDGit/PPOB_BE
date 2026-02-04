package domain

// VoucherListResponse represents the list vouchers response
type VoucherListResponse struct {
	Vouchers      []*VoucherDetail `json:"vouchers"`
	TotalVouchers int              `json:"totalVouchers"`
}

// VoucherDetail represents detailed voucher information
type VoucherDetail struct {
	ID                      string   `json:"id"`
	Code                    string   `json:"code"`
	Name                    string   `json:"name"`
	Description             string   `json:"description"`
	DiscountType            string   `json:"discountType"`
	DiscountValue           int64    `json:"discountValue"`
	DiscountFormatted       string   `json:"discountFormatted"`
	MinTransaction          int64    `json:"minTransaction"`
	MinTransactionFormatted string   `json:"minTransactionFormatted"`
	MaxDiscount             int64    `json:"maxDiscount"`
	MaxDiscountFormatted    string   `json:"maxDiscountFormatted"`
	ApplicableServices      []string `json:"applicableServices"`
	ExpiresAt               string   `json:"expiresAt"`
	Status                  string   `json:"status"`
	TermsURL                string   `json:"termsUrl"`
}

// ApplicableVouchersResponse represents the applicable vouchers response
type ApplicableVouchersResponse struct {
	Vouchers                  []*ApplicableVoucher `json:"vouchers"`
	MaxVouchersPerTransaction int                  `json:"maxVouchersPerTransaction"`
}

// ApplicableVoucher represents a voucher with applicability info
type ApplicableVoucher struct {
	ID                         string  `json:"id"`
	Code                       string  `json:"code"`
	Name                       string  `json:"name"`
	Description                string  `json:"description"`
	DiscountType               string  `json:"discountType"`
	DiscountValue              int64   `json:"discountValue"`
	DiscountFormatted          *string `json:"discountFormatted,omitempty"`
	EstimatedDiscount          *int64  `json:"estimatedDiscount,omitempty"`
	EstimatedDiscountFormatted *string `json:"estimatedDiscountFormatted,omitempty"`
	Applicable                 bool    `json:"applicable"`
	NotApplicableReason        *string `json:"notApplicableReason,omitempty"`
	ExpiresAt                  string  `json:"expiresAt"`
}

// ValidateVoucherResponse represents the validate voucher response
type ValidateVoucherResponse struct {
	Valid   bool          `json:"valid"`
	Voucher *ValidVoucher `json:"voucher,omitempty"`
	Reason  *string       `json:"reason,omitempty"`
	Message *string       `json:"message,omitempty"`
}

// ValidVoucher represents valid voucher with discount calculation
type ValidVoucher struct {
	ID                         string `json:"id"`
	Code                       string `json:"code"`
	Name                       string `json:"name"`
	DiscountType               string `json:"discountType"`
	DiscountValue              int64  `json:"discountValue"`
	EstimatedDiscount          int64  `json:"estimatedDiscount"`
	EstimatedDiscountFormatted string `json:"estimatedDiscountFormatted"`
	MaxDiscount                int64  `json:"maxDiscount"`
	MaxDiscountFormatted       string `json:"maxDiscountFormatted"`
}
