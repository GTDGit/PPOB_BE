package gerbang

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GetBankCodes fetches bank codes from Gerbang API
func (c *Client) GetBankCodes(ctx context.Context) ([]BankCodeItem, error) {
	url := c.config.BaseURL + "/v1/bank-codes"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication headers as per documentation
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.ClientSecret)
	req.Header.Set("X-Client-Id", c.config.ClientID)
	req.Header.Set("User-Agent", "PPOB.ID/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle 5xx errors
	if resp.StatusCode >= 500 {
		return nil, &Error{
			Code:    resp.StatusCode,
			Message: fmt.Sprintf("server error: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		}
	}

	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response (status=%d): %w, body: %s",
			resp.StatusCode, err, truncateBody(respBody, 200))
	}

	// Check for API errors
	if !apiResp.Success {
		return nil, ParseError(&apiResp)
	}

	// Parse data as array of bank codes
	var bankCodes []BankCodeItem
	if err := c.parseData(&apiResp, &bankCodes); err != nil {
		return nil, fmt.Errorf("failed to parse bank codes: %w", err)
	}

	return bankCodes, nil
}
