package handler

import (
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/external/gerbang"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// DepositHandler handles deposit requests
type DepositHandler struct {
	depositService *service.DepositService
	callbackSecret string
}

// NewDepositHandler creates a new deposit handler
func NewDepositHandler(depositService *service.DepositService, callbackSecret string) *DepositHandler {
	return &DepositHandler{
		depositService: depositService,
		callbackSecret: callbackSecret,
	}
}

// Request structs

// CreateBankTransferRequest represents bank transfer deposit request
type CreateBankTransferRequest struct {
	Amount int64 `json:"amount" binding:"required,min=10000"`
}

// CreateQRISRequest represents QRIS deposit request
type CreateQRISRequest struct {
	Amount int64 `json:"amount" binding:"required,min=10000"`
}

// CreateRetailRequest represents retail deposit request
type CreateRetailRequest struct {
	ProviderCode string `json:"providerCode" binding:"required,oneof=alfamart indomaret"`
	Amount       int64  `json:"amount" binding:"required,min=10000"`
}

// CreateVARequest represents VA deposit request
type CreateVARequest struct {
	BankCode string `json:"bankCode" binding:"required"`
	Amount   int64  `json:"amount" binding:"required,min=10000"`
}

// GetMethods handles GET /v1/deposit/methods
func (h *DepositHandler) GetMethods(c *gin.Context) {
	response, err := h.depositService.GetMethods(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// CreateBankTransfer handles POST /v1/deposit/bank-transfer
func (h *DepositHandler) CreateBankTransfer(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Parse request
	var req CreateBankTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	// Call service
	response, err := h.depositService.CreateBankTransfer(c.Request.Context(), userID, req.Amount)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// CreateQRIS handles POST /v1/deposit/qris
func (h *DepositHandler) CreateQRIS(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Parse request
	var req CreateQRISRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	// Call service
	response, err := h.depositService.CreateQRIS(c.Request.Context(), userID, req.Amount)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetRetailProviders handles GET /v1/deposit/retail/providers
func (h *DepositHandler) GetRetailProviders(c *gin.Context) {
	response, err := h.depositService.GetRetailProviders(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// CreateRetail handles POST /v1/deposit/retail
func (h *DepositHandler) CreateRetail(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Parse request
	var req CreateRetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	// Call service
	response, err := h.depositService.CreateRetail(c.Request.Context(), userID, req.ProviderCode, req.Amount)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetVABanks handles GET /v1/deposit/va/banks
func (h *DepositHandler) GetVABanks(c *gin.Context) {
	response, err := h.depositService.GetVABanks(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// CreateVA handles POST /v1/deposit/va
func (h *DepositHandler) CreateVA(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Parse request
	var req CreateVARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	// Call service
	response, err := h.depositService.CreateVA(c.Request.Context(), userID, req.BankCode, req.Amount)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetStatus handles GET /v1/deposit/:depositId
func (h *DepositHandler) GetStatus(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get deposit ID from path
	depositID := c.Param("depositId")
	if depositID == "" {
		respondWithError(c, domain.ErrValidationFailed("ID Deposit wajib diisi"))
		return
	}

	// Call service
	response, err := h.depositService.GetStatus(c.Request.Context(), userID, depositID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetHistory handles GET /v1/deposit/history
func (h *DepositHandler) GetHistory(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get query params
	status := c.DefaultQuery("status", "all")
	method := c.DefaultQuery("method", "all")
	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("limit", "20")

	// Parse pagination
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = 20
	}

	// Build filter
	filter := repository.DepositFilter{
		Status:  status,
		Method:  method,
		Page:    page,
		PerPage: perPage,
	}

	// Call service
	response, err := h.depositService.GetHistory(c.Request.Context(), userID, filter)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// HandleWebhook handles POST /internal/webhook/deposit
func (h *DepositHandler) HandleWebhook(c *gin.Context) {
	// Read raw body for signature verification
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondWithError(c, domain.ErrValidationFailed("Gagal membaca body request"))
		return
	}

	// Get signature from header
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Verify webhook signature
	if h.callbackSecret != "" && !gerbang.VerifySignature(bodyBytes, signature, h.callbackSecret) {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Parse webhook payload
	webhook, err := gerbang.ParseWebhook(bodyBytes)
	if err != nil {
		respondWithError(c, domain.ErrValidationFailed("Invalid webhook payload"))
		return
	}

	// Only handle payment.paid events
	if webhook.Event != gerbang.EventPaymentPaid {
		respondWithSuccess(c, http.StatusOK, gin.H{
			"success": true,
			"message": "Event ignored",
		})
		return
	}

	// Parse payment data
	paymentData, err := webhook.ParsePaymentData()
	if err != nil {
		respondWithError(c, domain.ErrValidationFailed("Data pembayaran tidak valid"))
		return
	}

	// Log webhook received
	slog.Info("deposit webhook received",
		slog.String("event", webhook.Event),
		slog.String("payment_id", paymentData.PaymentID),
		slog.String("reference_id", paymentData.ReferenceID),
		slog.String("status", paymentData.Status),
	)

	// Process webhook SYNCHRONOUSLY for proper idempotency handling
	// The service layer has built-in idempotency check (return nil if already processed)
	// This ensures proper error responses to Gerbang and prevents race conditions
	if err := h.depositService.HandleWebhook(c.Request.Context(), paymentData.ReferenceID); err != nil {
		slog.Error("webhook processing failed",
			slog.String("reference_id", paymentData.ReferenceID),
			slog.String("error", err.Error()),
		)
		handleServiceError(c, err)
		return
	}

	slog.Info("webhook processed successfully",
		slog.String("reference_id", paymentData.ReferenceID),
	)

	// Return 200 OK after successful processing
	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Webhook received",
		"event":   webhook.Event,
	})
}
