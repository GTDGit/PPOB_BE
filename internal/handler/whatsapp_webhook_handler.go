package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/GTDGit/PPOB_BE/internal/config"
	"github.com/GTDGit/PPOB_BE/internal/external/whatsapp"
	"github.com/GTDGit/PPOB_BE/pkg/validator"
)

// WhatsAppWebhookHandler handles Meta WhatsApp webhook verification and message status events.
type WhatsAppWebhookHandler struct {
	verifyToken string
	appSecret   string
}

// NewWhatsAppWebhookHandler creates a new WhatsApp webhook handler.
func NewWhatsAppWebhookHandler(cfg config.WhatsAppConfig) *WhatsAppWebhookHandler {
	return &WhatsAppWebhookHandler{
		verifyToken: cfg.WebhookVerifyToken,
		appSecret:   cfg.AppSecret,
	}
}

// HandleVerify handles Meta's GET webhook verification challenge.
func (h *WhatsAppWebhookHandler) HandleVerify(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if h.verifyToken == "" {
		slog.Error("whatsapp webhook verify token not configured")
		c.String(http.StatusInternalServerError, "webhook verify token not configured")
		return
	}

	if mode == "subscribe" && token == h.verifyToken && challenge != "" {
		slog.Info("whatsapp webhook verified")
		c.String(http.StatusOK, challenge)
		return
	}

	slog.Warn("whatsapp webhook verification failed",
		slog.String("mode", mode),
	)
	c.String(http.StatusForbidden, "forbidden")
}

// HandleWebhook handles Meta's POST webhook payloads for inbound messages and delivery statuses.
func (h *WhatsAppWebhookHandler) HandleWebhook(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Error("failed to read whatsapp webhook body", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	signature := getWhatsAppSignature(c)
	if h.appSecret != "" {
		if signature == "" || !whatsapp.VerifySignature(rawBody, signature, h.appSecret) {
			slog.Warn("whatsapp webhook signature mismatch")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}

	webhook, err := whatsapp.ParseWebhook(rawBody)
	if err != nil {
		slog.Error("failed to parse whatsapp webhook", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	statusEvents := 0
	inboundMessages := 0

	for _, entry := range webhook.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				slog.Info("whatsapp webhook field ignored",
					slog.String("field", change.Field),
					slog.String("entry_id", entry.ID),
				)
				continue
			}

			for _, status := range change.Value.Statuses {
				statusEvents++
				attrs := []any{
					slog.String("entry_id", entry.ID),
					slog.String("phone_number_id", change.Value.Metadata.PhoneNumberID),
					slog.String("display_phone_number", change.Value.Metadata.DisplayPhoneNumber),
					slog.String("message_id", status.ID),
					slog.String("status", status.Status),
					slog.String("recipient_id", validator.MaskPhone(status.RecipientID)),
					slog.String("timestamp", status.Timestamp),
				}

				if status.Conversation != nil {
					attrs = append(attrs, slog.String("conversation_id", status.Conversation.ID))
					if status.Conversation.Origin != nil {
						attrs = append(attrs, slog.String("conversation_origin", status.Conversation.Origin.Type))
					}
				}

				if status.Pricing != nil {
					attrs = append(attrs,
						slog.Bool("billable", status.Pricing.Billable),
						slog.String("pricing_model", status.Pricing.PricingModel),
						slog.String("pricing_category", status.Pricing.Category),
					)
				}

				if len(status.Errors) > 0 {
					attrs = append(attrs, slog.String("errors", marshalWhatsAppErrors(status.Errors)))
					slog.Warn("whatsapp message status received", attrs...)
					continue
				}

				slog.Info("whatsapp message status received", attrs...)
			}

			for _, message := range change.Value.Messages {
				inboundMessages++
				attrs := []any{
					slog.String("entry_id", entry.ID),
					slog.String("phone_number_id", change.Value.Metadata.PhoneNumberID),
					slog.String("display_phone_number", change.Value.Metadata.DisplayPhoneNumber),
					slog.String("message_id", message.ID),
					slog.String("from", validator.MaskPhone(message.From)),
					slog.String("type", message.Type),
					slog.String("timestamp", message.Timestamp),
				}

				if message.Text != nil {
					attrs = append(attrs, slog.Int("text_length", len(message.Text.Body)))
				}

				slog.Info("whatsapp inbound message received", attrs...)
			}

			if len(change.Value.Errors) > 0 {
				slog.Warn("whatsapp webhook top-level errors",
					slog.String("entry_id", entry.ID),
					slog.String("errors", marshalWhatsAppErrors(change.Value.Errors)),
				)
			}
		}
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message":         "Webhook received",
		"statusEvents":    statusEvents,
		"inboundMessages": inboundMessages,
		"object":          webhook.Object,
	})
}

func marshalWhatsAppErrors(errors []whatsapp.WebhookError) string {
	if len(errors) == 0 {
		return ""
	}

	body, err := json.Marshal(errors)
	if err != nil {
		return "failed to encode webhook errors"
	}

	return string(body)
}
