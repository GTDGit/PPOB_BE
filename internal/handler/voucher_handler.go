package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// VoucherHandler handles voucher-related requests
type VoucherHandler struct {
	voucherService *service.VoucherService
}

// NewVoucherHandler creates a new voucher handler
func NewVoucherHandler(voucherService *service.VoucherService) *VoucherHandler {
	return &VoucherHandler{
		voucherService: voucherService,
	}
}

// List handles GET /v1/vouchers
func (h *VoucherHandler) List(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get status filter from query param
	status := c.DefaultQuery("status", "")

	// Call service
	response, err := h.voucherService.List(c.Request.Context(), userID, status)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetApplicable handles GET /v1/vouchers/applicable
func (h *VoucherHandler) GetApplicable(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get query params
	serviceType := c.Query("serviceType")
	amountStr := c.Query("amount")

	// Validate required params
	if serviceType == "" || amountStr == "" {
		respondWithError(c, domain.ErrValidationFailed("serviceType and amount are required"))
		return
	}

	// Parse amount
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amount <= 0 {
		respondWithError(c, domain.ErrValidationFailed("Invalid amount"))
		return
	}

	// Call service
	response, err := h.voucherService.GetApplicable(c.Request.Context(), userID, serviceType, amount)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// ValidateRequest represents validate voucher request body
type ValidateRequest struct {
	Code        string `json:"code" binding:"required"`
	ServiceType string `json:"serviceType" binding:"required"`
	Amount      int64  `json:"amount" binding:"required,min=1"`
}

// Validate handles POST /v1/vouchers/validate
func (h *VoucherHandler) Validate(c *gin.Context) {
	var req ValidateRequest
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
	response, err := h.voucherService.Validate(c.Request.Context(), userID, req.Code, req.ServiceType, req.Amount)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}
