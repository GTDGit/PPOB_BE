package whatsapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// WebhookPayload represents the payload sent by Meta to the WhatsApp webhook.
type WebhookPayload struct {
	Object string         `json:"object"`
	Entry  []WebhookEntry `json:"entry"`
}

type WebhookEntry struct {
	ID      string          `json:"id"`
	Changes []WebhookChange `json:"changes"`
}

type WebhookChange struct {
	Field string       `json:"field"`
	Value WebhookValue `json:"value"`
}

type WebhookValue struct {
	MessagingProduct string           `json:"messaging_product"`
	Metadata         WebhookMetadata  `json:"metadata"`
	Contacts         []WebhookContact `json:"contacts,omitempty"`
	Messages         []WebhookMessage `json:"messages,omitempty"`
	Statuses         []WebhookStatus  `json:"statuses,omitempty"`
	Errors           []WebhookError   `json:"errors,omitempty"`
}

type WebhookMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type WebhookContact struct {
	Profile struct {
		Name string `json:"name"`
	} `json:"profile,omitempty"`
	WaID string `json:"wa_id"`
}

type WebhookMessage struct {
	From      string       `json:"from"`
	ID        string       `json:"id"`
	Timestamp string       `json:"timestamp"`
	Type      string       `json:"type"`
	Text      *WebhookText `json:"text,omitempty"`
}

type WebhookText struct {
	Body string `json:"body"`
}

type WebhookStatus struct {
	ID           string               `json:"id"`
	Status       string               `json:"status"`
	Timestamp    string               `json:"timestamp"`
	RecipientID  string               `json:"recipient_id"`
	Conversation *WebhookConversation `json:"conversation,omitempty"`
	Pricing      *WebhookPricing      `json:"pricing,omitempty"`
	Errors       []WebhookError       `json:"errors,omitempty"`
}

type WebhookConversation struct {
	ID     string                     `json:"id"`
	Origin *WebhookConversationOrigin `json:"origin,omitempty"`
}

type WebhookConversationOrigin struct {
	Type string `json:"type"`
}

type WebhookPricing struct {
	Billable     bool   `json:"billable"`
	PricingModel string `json:"pricing_model"`
	Category     string `json:"category"`
}

type WebhookError struct {
	Code      int            `json:"code"`
	Title     string         `json:"title"`
	Message   string         `json:"message"`
	Href      string         `json:"href,omitempty"`
	ErrorData map[string]any `json:"error_data,omitempty"`
}

// ParseWebhook parses the WhatsApp webhook payload from raw bytes.
func ParseWebhook(payload []byte) (*WebhookPayload, error) {
	var webhook WebhookPayload
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return nil, fmt.Errorf("failed to parse whatsapp webhook: %w", err)
	}
	return &webhook, nil
}

// VerifySignature verifies Meta's X-Hub-Signature-256 header.
func VerifySignature(payload []byte, signatureHeader, appSecret string) bool {
	signatureHeader = strings.TrimSpace(strings.ToLower(signatureHeader))
	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return false
	}

	providedSignature, err := hex.DecodeString(strings.TrimPrefix(signatureHeader, "sha256="))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(payload)
	expectedSignature := mac.Sum(nil)

	return hmac.Equal(providedSignature, expectedSignature)
}
