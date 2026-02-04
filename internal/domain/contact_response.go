package domain

// ContactListResponse represents the list contacts response
type ContactListResponse struct {
	Contacts      []*ContactDetail `json:"contacts"`
	TotalContacts int              `json:"totalContacts"`
}

// ContactDetail represents detailed contact information
type ContactDetail struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	Value        string        `json:"value"`
	Operator     *OperatorInfo `json:"operator,omitempty"`
	Bank         *BankInfo     `json:"bank,omitempty"`
	CustomerName *string       `json:"customerName,omitempty"`
	AccountName  *string       `json:"accountName,omitempty"`
	LastUsedAt   *string       `json:"lastUsedAt,omitempty"`
	UsageCount   int           `json:"usageCount"`
	CreatedAt    string        `json:"createdAt"`
	UpdatedAt    *string       `json:"updatedAt,omitempty"`
}

// Note: OperatorInfo is already defined in prepaid_response.go and can be reused

// BankInfo represents bank information (for bank contacts)
type BankInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// ContactResponse represents single contact response
type ContactResponse struct {
	Contact *ContactDetail `json:"contact"`
}

// DeleteContactResponse represents delete contact response
type DeleteContactResponse struct {
	Deleted   bool   `json:"deleted"`
	ContactID string `json:"contactId"`
}
