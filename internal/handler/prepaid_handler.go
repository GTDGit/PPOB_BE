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

// PrepaidHandler handles prepaid transaction requests
type PrepaidHandler struct {
	prepaidService *service.PrepaidService
}

// NewPrepaidHandler creates a new prepaid handler
func NewPrepaidHandler(prepaidService *service.PrepaidService) *PrepaidHandler {
	return &PrepaidHandler{
		prepaidService: prepaidService,
	}
}

// InquiryRequest represents inquiry request body
type InquiryRequest struct {
	ServiceType string  `json:"serviceType" binding:"required,oneof=pulsa data pln_prepaid ewallet game"`
	Target      string  `json:"target" binding:"required"`
	ProviderID  *string `json:"providerId"`
}

// Inquiry handles POST /v1/prepaid/inquiry
func (h *PrepaidHandler) Inquiry(c *gin.Context) {
	var req InquiryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Call service
	resp, err := h.prepaidService.Inquiry(c.Request.Context(), service.InquiryRequest{
		UserID:      userID,
		ServiceType: req.ServiceType,
		Target:      req.Target,
		ProviderID:  req.ProviderID,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// CreateOrderRequest represents create order request body
type CreateOrderRequest struct {
	InquiryID    string       `json:"inquiryId" binding:"required"`
	ProductID    string       `json:"productId" binding:"required"`
	VoucherCodes []string     `json:"voucherCodes"`
	Contact      *ContactInfo `json:"contact"`
}

// ContactInfo represents contact information
type ContactInfo struct {
	SaveAsContact bool   `json:"saveAsContact"`
	ContactName   string `json:"contactName"`
}

// CreateOrder handles POST /v1/prepaid/order
func (h *PrepaidHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Call service
	resp, err := h.prepaidService.CreateOrder(c.Request.Context(), service.CreateOrderRequest{
		UserID:       userID,
		InquiryID:    req.InquiryID,
		ProductID:    req.ProductID,
		VoucherCodes: req.VoucherCodes,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// PayRequest represents payment request body
type PayRequest struct {
	OrderID string  `json:"orderId" binding:"required"`
	PIN     *string `json:"pin"`
}

// Pay handles POST /v1/prepaid/pay
func (h *PrepaidHandler) Pay(c *gin.Context) {
	var req PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Call service
	resp, err := h.prepaidService.Pay(c.Request.Context(), service.PayRequest{
		UserID:  userID,
		OrderID: req.OrderID,
		PIN:     req.PIN,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// HandleWebhook handles prepaid transaction webhook from Gerbang API
func (h *PrepaidHandler) HandleWebhook(c *gin.Context) {
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

	// Verify signature (using Gerbang callback secret from config)
	// TODO: Get callback secret from config
	// if !gerbang.VerifySignature(rawBody, signature, callbackSecret) {
	//     respondWithError(c, domain.NewError(domain.CodeUnauthorized, "Signature tidak valid", 401))
	//     return
	// }

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
	slog.Info("prepaid webhook received",
		slog.String("event", webhook.Event),
		slog.String("transaction_id", transactionData.TransactionID),
		slog.String("reference_id", transactionData.ReferenceID),
		slog.String("status", transactionData.Status),
	)

	// Handle webhook - SYNCHRONOUS processing for proper error handling
	if err := h.prepaidService.HandleWebhook(c.Request.Context(), transactionData); err != nil {
		slog.Error("failed to process prepaid webhook",
			slog.String("reference_id", transactionData.ReferenceID),
			slog.String("error", err.Error()),
		)
		handleServiceError(c, err)
		return
	}

	// Return 200 OK immediately to acknowledge webhook
	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Webhook received",
		"event":   webhook.Event,
	})
}
