package fazpass

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/config"
	"github.com/GTDGit/PPOB_BE/pkg/httplog"
)

// Client handles Fazpass SMS API communication
type Client struct {
	cfg        config.FazpassConfig
	httpClient *http.Client
}

// NewClient creates a new Fazpass client
func NewClient(cfg config.FazpassConfig) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: httplog.NewTransport(nil, nil),
		},
	}
}

// SendOTPRequest represents the OTP SMS request
type SendOTPRequest struct {
	Phone string
	OTP   string
}

// SendOTPResponse represents the API response
type SendOTPResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// fazpassRequest represents the Fazpass API request structure for OTP
type fazpassRequest struct {
	Phone      string `json:"phone"`
	OTP        string `json:"otp"`
	GatewayKey string `json:"gateway_key"`
}

// fazpassResponse represents the Fazpass API response
type fazpassResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
	Data    *struct {
		ID      string `json:"id"`
		OTP     string `json:"otp"`
		Channel string `json:"channel"`
	} `json:"data,omitempty"`
}

// SendOTP sends an OTP via SMS using Fazpass
func (c *Client) SendOTP(ctx context.Context, req SendOTPRequest) (*SendOTPResponse, error) {
	// Format phone number (Fazpass accepts 08xx or 628xx format)
	phone := formatPhoneForFazpass(req.Phone)

	// Build request as per Fazpass API docs
	fazReq := fazpassRequest{
		Phone:      phone,
		OTP:        req.OTP,
		GatewayKey: c.cfg.GatewayKey,
	}

	// Marshal request body
	body, err := json.Marshal(fazReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build API URL - use /v1/otp/send endpoint
	url := fmt.Sprintf("%s/v1/otp/send", c.cfg.APIURL)

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.MerchantKey)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var apiResp fazpassResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check response status (status is boolean in Fazpass API)
	if apiResp.Status {
		messageID := ""
		if apiResp.Data != nil {
			messageID = apiResp.Data.ID
		}
		return &SendOTPResponse{
			Success:   true,
			MessageID: messageID,
		}, nil
	}

	return &SendOTPResponse{
		Success: false,
		Error:   fmt.Sprintf("[%s] %s", apiResp.Code, apiResp.Message),
	}, nil
}

// formatPhoneForFazpass formats phone number for Fazpass API
func formatPhoneForFazpass(phone string) string {
	// Remove any spaces or dashes
	cleaned := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			cleaned += string(c)
		}
	}

	// Convert 08xx to 628xx (Fazpass prefers international format)
	if len(cleaned) > 1 && cleaned[0] == '0' {
		cleaned = "62" + cleaned[1:]
	}

	return cleaned
}

// IsEnabled returns whether Fazpass is configured
func (c *Client) IsEnabled() bool {
	return c.cfg.MerchantKey != ""
}
