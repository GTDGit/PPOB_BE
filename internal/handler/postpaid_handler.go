package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/external/gerbang"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// PostpaidHandler handles postpaid requests
type PostpaidHandler struct {
	postpaidService *service.PostpaidService
}

// NewPostpaidHandler creates a new postpaid handler
func NewPostpaidHandler(postpaidService *service.PostpaidService) *PostpaidHandler {
	return &PostpaidHandler{
		postpaidService: postpaidService,
	}
}

// Request structs

// PostpaidInquiryRequest represents inquiry request
type PostpaidInquiryRequest struct {
	ServiceType string  `json:"serviceType" binding:"required"`
	Target      string  `json:"target" binding:"required"`
	ProviderID  *string `json:"providerId"`
	Period      *string `json:"period"` // For BPJS
}

// PostpaidPayRequest represents pay request
type PostpaidPayRequest struct {
	InquiryID    string               `json:"inquiryId" binding:"required"`
	VoucherCodes []string             `json:"voucherCodes"`
	PIN          *string              `json:"pin"`
	Contact      *PostpaidContactInfo `json:"contact"`
}

// PostpaidContactInfo for saving contact
type PostpaidContactInfo struct {
	SaveAsContact bool   `json:"saveAsContact"`
	ContactName   string `json:"contactName"`
}

// Handlers

// Inquiry handles POST /v1/postpaid/inquiry
func (h *PostpaidHandler) Inquiry(c *gin.Context) {
	// Extract user ID from JWT
	userID := middleware.GetUserID(c)

	// Parse request
	var req PostpaidInquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	// Call service
	response, err := h.postpaidService.Inquiry(
		c.Request.Context(),
		userID,
		req.ServiceType,
		req.Target,
		req.ProviderID,
		req.Period,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// Pay handles POST /v1/postpaid/pay
func (h *PostpaidHandler) Pay(c *gin.Context) {
	// Extract user ID from JWT
	userID := middleware.GetUserID(c)

	// Parse request
	var req PostpaidPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	// Call service
	response, err := h.postpaidService.Pay(
		c.Request.Context(),
		userID,
		req.InquiryID,
		req.VoucherCodes,
		req.PIN,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// TODO: Save contact if requested
	// if req.Contact != nil && req.Contact.SaveAsContact {
	//     go h.saveContact(userID.(string), req.Contact, response)
	// }

	respondWithSuccess(c, http.StatusOK, response)
}

// HandleWebhook handles postpaid transaction webhook from Gerbang API
func (h *PostpaidHandler) HandleWebhook(c *gin.Context) {
	// Read raw body for signature verification
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Error("failed to read webhook body", slog.String("error", err.Error()))
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	// Get signature from header
	signature := c.GetHeader("X-Callback-Signature")
	if signature == "" {
		slog.Error("webhook signature missing")
		respondWithError(c, domain.NewError(domain.CodeUnauthorized, "Signature tidak valid", 401))
		return
	}

	// Parse webhook payload
	webhook, err := gerbang.ParseWebhook(rawBody)
	if err != nil {
		slog.Error("failed to parse webhook", slog.String("error", err.Error()))
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	// Validate event type
	if !webhook.IsTransactionEvent() {
		slog.Error("invalid transaction webhook event", slog.String("event", webhook.Event))
		respondWithError(c, domain.NewError(domain.CodeInvalidRequest, "Event tidak valid", 400))
		return
	}

	// Parse transaction data
	transactionData, err := webhook.ParseTransactionData()
	if err != nil {
		slog.Error("failed to parse transaction data", slog.String("error", err.Error()))
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	// Log webhook received
	slog.Info("postpaid webhook received",
		slog.String("event", webhook.Event),
		slog.String("transaction_id", transactionData.TransactionID),
		slog.String("reference_id", transactionData.ReferenceID),
		slog.String("status", transactionData.Status),
	)

	// Note: Postpaid service HandleWebhook method will be similar to prepaid
	// For now, just acknowledge webhook
	slog.Info("postpaid webhook processing skipped - service method not yet implemented")

	// Return 200 OK immediately to acknowledge webhook
	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Webhook received",
		"event":   webhook.Event,
	})
}
