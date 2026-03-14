package domain

// SandboxCheckoutResponse is returned after a dummy checkout succeeds.
type SandboxCheckoutResponse struct {
	TransactionID string               `json:"transactionId"`
	Status        string               `json:"status"`
	Title         string               `json:"title"`
	Message       string               `json:"message"`
	Balance       *SandboxBalanceDelta `json:"balance"`
}

// SandboxDepositCompleteResponse is returned after simulating a deposit payment.
type SandboxDepositCompleteResponse struct {
	DepositID string               `json:"depositId"`
	Status    string               `json:"status"`
	Message   string               `json:"message"`
	Balance   *SandboxBalanceDelta `json:"balance"`
}

// SandboxBalanceDelta shows the effect of a dummy action on the user's balance.
type SandboxBalanceDelta struct {
	Before          int64  `json:"before"`
	BeforeFormatted string `json:"beforeFormatted"`
	After           int64  `json:"after"`
	AfterFormatted  string `json:"afterFormatted"`
}
