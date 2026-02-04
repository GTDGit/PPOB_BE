package whatsapp

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

// Client handles WhatsApp Business API communication
type Client struct {
	cfg        config.WhatsAppConfig
	httpClient *http.Client
}

// NewClient creates a new WhatsApp client
func NewClient(cfg config.WhatsAppConfig) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: httplog.NewTransport(nil, nil),
		},
	}
}

// SendOTPRequest represents the OTP message request
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

// templateMessage represents WhatsApp template message structure
type templateMessage struct {
	MessagingProduct string          `json:"messaging_product"`
	To               string          `json:"to"`
	Type             string          `json:"type"`
	Template         templateContent `json:"template"`
}

type templateContent struct {
	Name       string              `json:"name"`
	Language   templateLanguage    `json:"language"`
	Components []templateComponent `json:"components,omitempty"`
}

type templateLanguage struct {
	Code string `json:"code"`
}

type templateComponent struct {
	Type       string              `json:"type"`
	Parameters []templateParameter `json:"parameters,omitempty"`
}

type templateParameter struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// whatsappAPIResponse represents the WhatsApp API response
type whatsappAPIResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

// SendOTP sends an OTP via WhatsApp
func (c *Client) SendOTP(ctx context.Context, req SendOTPRequest) (*SendOTPResponse, error) {
	// Format phone number for WhatsApp (must be in international format without +)
	phone := formatPhoneForWhatsApp(req.Phone)

	// Build template message
	msg := templateMessage{
		MessagingProduct: "whatsapp",
		To:               phone,
		Type:             "template",
		Template: templateContent{
			Name: c.cfg.OTPTemplateName,
			Language: templateLanguage{
				Code: "id", // Indonesian
			},
			Components: []templateComponent{
				{
					Type: "body",
					Parameters: []templateParameter{
						{
							Type: "text",
							Text: req.OTP,
						},
					},
				},
			},
		},
	}

	// Marshal request body
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build API URL
	url := fmt.Sprintf("%s/%s/messages", c.cfg.APIURL, c.cfg.PhoneNumberID)

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var apiResp whatsappAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for API error
	if apiResp.Error != nil {
		return &SendOTPResponse{
			Success: false,
			Error:   apiResp.Error.Message,
		}, nil
	}

	// Check for success
	if len(apiResp.Messages) > 0 {
		return &SendOTPResponse{
			Success:   true,
			MessageID: apiResp.Messages[0].ID,
		}, nil
	}

	return &SendOTPResponse{
		Success: false,
		Error:   "unknown error",
	}, nil
}

// formatPhoneForWhatsApp formats phone number for WhatsApp API
// WhatsApp requires international format without + sign (e.g., 6281234567890)
func formatPhoneForWhatsApp(phone string) string {
	// Remove any spaces or dashes
	cleaned := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			cleaned += string(c)
		}
	}

	// Convert 08xx to 628xx
	if len(cleaned) > 1 && cleaned[0] == '0' {
		cleaned = "62" + cleaned[1:]
	}

	// Remove + if present (already handled by digit-only extraction)
	return cleaned
}

// IsEnabled returns whether WhatsApp is configured
func (c *Client) IsEnabled() bool {
	return c.cfg.AccessToken != "" && c.cfg.PhoneNumberID != ""
}
