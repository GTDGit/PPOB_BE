package gerbang

import (
	"context"
	"fmt"
	"net/http"
)

// ========== Payment Methods (for Deposit) ==========

// CreatePayment creates a new payment (VA, QRIS, Retail)
func (c *Client) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResponse, error) {
	path := "/v1/payment"

	resp, err := c.doRequestWithRetry(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	var paymentResp PaymentResponse
	if err := c.parseData(resp, &paymentResp); err != nil {
		return nil, fmt.Errorf("failed to parse payment response: %w", err)
	}

	return &paymentResp, nil
}

// GetPaymentStatus gets payment status by payment ID
func (c *Client) GetPaymentStatus(ctx context.Context, paymentID string) (*PaymentResponse, error) {
	path := fmt.Sprintf("/v1/payment/%s", paymentID)

	resp, err := c.doRequestWithRetry(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment status: %w", err)
	}

	var paymentResp PaymentResponse
	if err := c.parseData(resp, &paymentResp); err != nil {
		return nil, fmt.Errorf("failed to parse payment status: %w", err)
	}

	return &paymentResp, nil
}

// CancelPayment cancels a pending payment
func (c *Client) CancelPayment(ctx context.Context, paymentID string) error {
	path := fmt.Sprintf("/v1/payment/%s/cancel", paymentID)

	_, err := c.doRequestWithRetry(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel payment: %w", err)
	}

	return nil
}

// ========== Helper Methods for Payment Types ==========

// CreateVAPayment creates a Virtual Account payment
func (c *Client) CreateVAPayment(ctx context.Context, referenceID, bankCode string, amount int64, customer CustomerInfo) (*PaymentResponse, error) {
	req := CreatePaymentRequest{
		ReferenceID: referenceID,
		PaymentMethod: PaymentMethod{
			Type: "VA",
			Code: bankCode,
		},
		Amount:      amount,
		Customer:    customer,
		Description: "Deposit PPOB.ID",
	}

	return c.CreatePayment(ctx, req)
}

// CreateQRISPayment creates a QRIS payment
func (c *Client) CreateQRISPayment(ctx context.Context, referenceID string, amount int64, customer CustomerInfo) (*PaymentResponse, error) {
	req := CreatePaymentRequest{
		ReferenceID: referenceID,
		PaymentMethod: PaymentMethod{
			Type: "QRIS",
		},
		Amount:      amount,
		Customer:    customer,
		Description: "Deposit PPOB.ID",
	}

	return c.CreatePayment(ctx, req)
}

// CreateRetailPayment creates a Retail payment
func (c *Client) CreateRetailPayment(ctx context.Context, referenceID, retailCode string, amount int64, customer CustomerInfo) (*PaymentResponse, error) {
	req := CreatePaymentRequest{
		ReferenceID: referenceID,
		PaymentMethod: PaymentMethod{
			Type: "RETAIL",
			Code: retailCode,
		},
		Amount:      amount,
		Customer:    customer,
		Description: "Deposit PPOB.ID",
	}

	return c.CreatePayment(ctx, req)
}
