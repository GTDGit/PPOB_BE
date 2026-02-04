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

// TransferHandler handles transfer transaction requests
type TransferHandler struct {
	transferService *service.TransferService
}

// NewTransferHandler creates a new transfer handler
func NewTransferHandler(transferService *service.TransferService) *TransferHandler {
	return &TransferHandler{
		transferService: transferService,
	}
}

// TransferInquiryRequest represents inquiry request body
type TransferInquiryRequest struct {
	BankCode      string `json:"bankCode" binding:"required"`
	AccountNumber string `json:"accountNumber" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,min=10000"`
}

// Inquiry handles POST /v1/transfer/inquiry
func (h *TransferHandler) Inquiry(c *gin.Context) {
	var req TransferInquiryRequest
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
	resp, err := h.transferService.Inquiry(c.Request.Context(), service.TransferInquiryRequest{
		UserID:        userID,
		BankCode:      req.BankCode,
		AccountNumber: req.AccountNumber,
		Amount:        req.Amount,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// TransferExecuteRequest represents execute request body
type TransferExecuteRequest struct {
	InquiryID string       `json:"inquiryId" binding:"required"`
	Purpose   *string      `json:"purpose"` // Optional, default "99" (Lainnya)
	Note      *string      `json:"note"`
	PIN       *string      `json:"pin"`
	Contact   *ContactInfo `json:"contact"`
}

// Execute handles POST /v1/transfer/execute
func (h *TransferHandler) Execute(c *gin.Context) {
	var req TransferExecuteRequest
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
	resp, err := h.transferService.Execute(c.Request.Context(), service.TransferExecuteRequest{
		UserID:    userID,
		InquiryID: req.InquiryID,
		Purpose:   req.Purpose,
		Note:      req.Note,
		PIN:       req.PIN,
	})

	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, resp)
}

// HandleWebhook handles transfer webhook from Gerbang API
func (h *TransferHandler) HandleWebhook(c *gin.Context) {
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
	if !webhook.IsTransferEvent() {
		slog.Error("invalid transfer webhook event", slog.String("event", webhook.Event))
		respondWithError(c, domain.NewError(domain.CodeInvalidRequest, "Event tidak valid", 400))
		return
	}

	// Parse transfer data
	transferData, err := webhook.ParseTransferData()
	if err != nil {
		slog.Error("failed to parse transfer data", slog.String("error", err.Error()))
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	// Log webhook received
	slog.Info("transfer webhook received",
		slog.String("event", webhook.Event),
		slog.String("transfer_id", transferData.TransferID),
		slog.String("reference_id", transferData.ReferenceID),
		slog.String("status", transferData.Status),
	)

	// Handle webhook - SYNCHRONOUS processing for proper error handling
	// Note: Could be made async with Redis distributed lock for better performance,
	// but synchronous ensures proper error responses to Gerbang
	if err := h.transferService.HandleWebhook(c.Request.Context(), transferData); err != nil {
		slog.Error("failed to process transfer webhook",
			slog.String("reference_id", transferData.ReferenceID),
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
