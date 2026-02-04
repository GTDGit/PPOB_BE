package gerbang

import (
	"context"
	"fmt"
)

// TransferInquiry validates account number and gets account name
func (c *Client) TransferInquiry(ctx context.Context, bankCode, accountNumber string) (*GerbangTransferInquiryResponse, error) {
	req := GerbangTransferInquiryRequest{
		BankCode:      bankCode,
		AccountNumber: accountNumber,
	}

	resp, err := c.doRequestWithRetry(ctx, "POST", "/v1/transfer/inquiry", req)
	if err != nil {
		return nil, err
	}

	var inquiryResp GerbangTransferInquiryResponse
	if err := c.parseData(resp, &inquiryResp); err != nil {
		return nil, fmt.Errorf("failed to parse transfer inquiry response: %w", err)
	}

	return &inquiryResp, nil
}

// TransferExecute executes bank transfer
func (c *Client) TransferExecute(ctx context.Context, req GerbangTransferExecuteRequest) (*GerbangTransferExecuteResponse, error) {
	// Ensure purpose has default value if empty
	if req.Purpose == "" {
		req.Purpose = "99" // Default: Lainnya
	}

	resp, err := c.doRequestWithRetry(ctx, "POST", "/v1/transfer", req)
	if err != nil {
		return nil, err
	}

	var executeResp GerbangTransferExecuteResponse
	if err := c.parseData(resp, &executeResp); err != nil {
		return nil, fmt.Errorf("failed to parse transfer execute response: %w", err)
	}

	return &executeResp, nil
}
