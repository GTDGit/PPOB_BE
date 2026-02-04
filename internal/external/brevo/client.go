package brevo

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

// Client handles Brevo email API communication
type Client struct {
	cfg        config.BrevoConfig
	httpClient *http.Client
}

// NewClient creates a new Brevo client
func NewClient(cfg config.BrevoConfig) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: httplog.NewTransport(nil, nil),
		},
	}
}

// EmailType represents the type of email to send
type EmailType string

const (
	EmailTypeVerification EmailType = "EMAIL_VERIFICATION"
	EmailTypeNewLogin     EmailType = "SECURITY_NEW_LOGIN"
	EmailTypePINChanged   EmailType = "SECURITY_PIN_CHANGED"
	EmailTypeEmailChanged EmailType = "SECURITY_EMAIL_CHANGED"
	EmailTypePhoneChanged EmailType = "SECURITY_PHONE_NUMBER_CHANGED"
)

// SendEmailRequest represents an email send request
type SendEmailRequest struct {
	To     string
	ToName string
	Type   EmailType
	Params map[string]string
}

// SendEmailResponse represents the API response
type SendEmailResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// brevoRequest represents the Brevo API request structure
type brevoRequest struct {
	Sender     brevoContact      `json:"sender"`
	To         []brevoContact    `json:"to"`
	TemplateID int64             `json:"templateId"`
	Params     map[string]string `json:"params,omitempty"`
}

type brevoContact struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// brevoResponse represents the Brevo API response
type brevoResponse struct {
	MessageID string `json:"messageId,omitempty"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}

// getTemplateID returns the template ID for the given email type from config
func (c *Client) getTemplateID(emailType EmailType) (int64, bool) {
	switch emailType {
	case EmailTypeVerification:
		return c.cfg.TemplateVerification, c.cfg.TemplateVerification > 0
	case EmailTypeNewLogin:
		return c.cfg.TemplateNewLogin, c.cfg.TemplateNewLogin > 0
	case EmailTypePINChanged:
		return c.cfg.TemplatePINChanged, c.cfg.TemplatePINChanged > 0
	case EmailTypeEmailChanged:
		return c.cfg.TemplateEmailChanged, c.cfg.TemplateEmailChanged > 0
	case EmailTypePhoneChanged:
		return c.cfg.TemplatePhoneChanged, c.cfg.TemplatePhoneChanged > 0
	default:
		return 0, false
	}
}

// SendEmail sends an email via Brevo
func (c *Client) SendEmail(ctx context.Context, req SendEmailRequest) (*SendEmailResponse, error) {
	// Get template ID from config
	templateID, ok := c.getTemplateID(req.Type)
	if !ok {
		return nil, fmt.Errorf("unknown email type: %s", req.Type)
	}

	// Build request
	brevoReq := brevoRequest{
		Sender: brevoContact{
			Email: c.cfg.SenderEmail,
			Name:  c.cfg.SenderName,
		},
		To: []brevoContact{
			{
				Email: req.To,
				Name:  req.ToName,
			},
		},
		TemplateID: templateID,
		Params:     req.Params,
	}

	// Marshal request body
	body, err := json.Marshal(brevoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build API URL
	url := fmt.Sprintf("%s/smtp/email", c.cfg.APIURL)

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-key", c.cfg.APIKey)

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var apiResp brevoResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check response
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &SendEmailResponse{
			Success:   true,
			MessageID: apiResp.MessageID,
		}, nil
	}

	return &SendEmailResponse{
		Success: false,
		Error:   apiResp.Message,
	}, nil
}

// SendVerificationEmail sends email verification link
func (c *Client) SendVerificationEmail(ctx context.Context, email, name, verificationLink string) (*SendEmailResponse, error) {
	return c.SendEmail(ctx, SendEmailRequest{
		To:     email,
		ToName: name,
		Type:   EmailTypeVerification,
		Params: map[string]string{
			"name": name,
			"link": verificationLink,
		},
	})
}

// SendNewLoginAlert sends new device login notification
func (c *Client) SendNewLoginAlert(ctx context.Context, email, name, deviceName, location, loginTime string) (*SendEmailResponse, error) {
	return c.SendEmail(ctx, SendEmailRequest{
		To:     email,
		ToName: name,
		Type:   EmailTypeNewLogin,
		Params: map[string]string{
			"name":       name,
			"device":     deviceName,
			"location":   location,
			"login_time": loginTime,
		},
	})
}

// SendPINChangedAlert sends PIN change notification
func (c *Client) SendPINChangedAlert(ctx context.Context, email, name, changeTime string) (*SendEmailResponse, error) {
	return c.SendEmail(ctx, SendEmailRequest{
		To:     email,
		ToName: name,
		Type:   EmailTypePINChanged,
		Params: map[string]string{
			"name":        name,
			"change_time": changeTime,
		},
	})
}

// SendEmailChangedAlert sends email change notification to old email
func (c *Client) SendEmailChangedAlert(ctx context.Context, oldEmail, name, newEmail, changeTime string) (*SendEmailResponse, error) {
	return c.SendEmail(ctx, SendEmailRequest{
		To:     oldEmail,
		ToName: name,
		Type:   EmailTypeEmailChanged,
		Params: map[string]string{
			"name":        name,
			"new_email":   newEmail,
			"change_time": changeTime,
		},
	})
}

// SendPhoneChangedAlert sends phone number change notification
func (c *Client) SendPhoneChangedAlert(ctx context.Context, email, name, oldPhone, newPhone, changeTime string) (*SendEmailResponse, error) {
	return c.SendEmail(ctx, SendEmailRequest{
		To:     email,
		ToName: name,
		Type:   EmailTypePhoneChanged,
		Params: map[string]string{
			"USER_NAME":        name,
			"OLD_PHONE_NUMBER": oldPhone,
			"NEW_PHONE_NUMBER": newPhone,
			"CHANGE_TIME":      changeTime,
		},
	})
}

// IsEnabled returns whether Brevo is configured
func (c *Client) IsEnabled() bool {
	return c.cfg.APIKey != ""
}
