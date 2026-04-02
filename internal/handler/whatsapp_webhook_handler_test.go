package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/GTDGit/PPOB_BE/internal/config"
)

func TestWhatsAppWebhookVerifySuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewWhatsAppWebhookHandler(config.WhatsAppConfig{
		WebhookVerifyToken: "verify-me",
	})

	req := httptest.NewRequest(http.MethodGet, "/internal/webhook/whatsapp?hub.mode=subscribe&hub.verify_token=verify-me&hub.challenge=abc123", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.HandleVerify(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if body := strings.TrimSpace(w.Body.String()); body != "abc123" {
		t.Fatalf("unexpected challenge body: %q", body)
	}
}

func TestWhatsAppWebhookRejectsInvalidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewWhatsAppWebhookHandler(config.WhatsAppConfig{
		AppSecret: "super-secret",
	})

	req := httptest.NewRequest(http.MethodPost, "/internal/webhook/whatsapp", strings.NewReader(`{"object":"whatsapp_business_account","entry":[]}`))
	req.Header.Set("X-Hub-Signature-256", "sha256=deadbeef")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.HandleWebhook(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
