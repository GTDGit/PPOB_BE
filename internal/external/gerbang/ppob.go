package gerbang

import (
	"context"
	"fmt"
	"net/http"
)

// ========== PPOB Methods ==========

// GetProducts fetches all products from Gerbang API
// NOTE: This should be called by background job every 15 minutes, NOT on every user request
func (c *Client) GetProducts(ctx context.Context, productType, category, brand, search string, page, limit int) ([]Product, *Pagination, error) {
	path := "/v1/ppob/products"

	// Build query params
	query := fmt.Sprintf("?type=%s", productType)
	if category != "" {
		query += fmt.Sprintf("&category=%s", category)
	}
	if brand != "" {
		query += fmt.Sprintf("&brand=%s", brand)
	}
	if search != "" {
		query += fmt.Sprintf("&search=%s", search)
	}
	if page > 0 {
		query += fmt.Sprintf("&page=%d", page)
	}
	if limit > 0 {
		query += fmt.Sprintf("&limit=%d", limit)
	}

	resp, err := c.doRequestWithRetry(ctx, http.MethodGet, path+query, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get products: %w", err)
	}

	var result struct {
		Products []Product `json:"products"`
	}

	if err := c.parseData(resp, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to parse products: %w", err)
	}

	return result.Products, resp.Meta.Pagination, nil
}

// CreateTransaction creates a new PPOB transaction (prepaid, inquiry, or payment)
func (c *Client) CreateTransaction(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	path := "/v1/ppob/transaction"

	resp, err := c.doRequestWithRetry(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	var txResp TransactionResponse
	if err := c.parseData(resp, &txResp); err != nil {
		return nil, fmt.Errorf("failed to parse transaction response: %w", err)
	}

	return &txResp, nil
}

// GetTransactionStatus gets transaction status by transaction ID
func (c *Client) GetTransactionStatus(ctx context.Context, transactionID string) (*TransactionResponse, error) {
	path := fmt.Sprintf("/v1/ppob/transaction/%s", transactionID)

	resp, err := c.doRequestWithRetry(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction status: %w", err)
	}

	var txResp TransactionResponse
	if err := c.parseData(resp, &txResp); err != nil {
		return nil, fmt.Errorf("failed to parse transaction status: %w", err)
	}

	return &txResp, nil
}

// ========== Helper Methods for Transaction Types ==========

// CreatePrepaidTransaction creates a prepaid transaction
func (c *Client) CreatePrepaidTransaction(ctx context.Context, referenceID, skuCode, customerNo string) (*TransactionResponse, error) {
	req := TransactionRequest{
		ReferenceID: referenceID,
		SKUCode:     skuCode,
		CustomerNo:  customerNo,
		Type:        "prepaid",
	}

	return c.CreateTransaction(ctx, req)
}

// CreateInquiry creates a postpaid inquiry
func (c *Client) CreateInquiry(ctx context.Context, referenceID, skuCode, customerNo string) (*TransactionResponse, error) {
	req := TransactionRequest{
		ReferenceID: referenceID,
		SKUCode:     skuCode,
		CustomerNo:  customerNo,
		Type:        "inquiry",
	}

	return c.CreateTransaction(ctx, req)
}

// CreatePostpaidPayment creates a postpaid payment
func (c *Client) CreatePostpaidPayment(ctx context.Context, referenceID, transactionID string) (*TransactionResponse, error) {
	req := TransactionRequest{
		ReferenceID:   referenceID,
		Type:          "payment",
		TransactionID: transactionID,
	}

	return c.CreateTransaction(ctx, req)
}
