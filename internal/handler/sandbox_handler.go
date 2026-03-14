package handler

import (
	"net/http"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/service"
	"github.com/gin-gonic/gin"
)

// SandboxHandler exposes dummy/testing endpoints used by the mobile app.
type SandboxHandler struct {
	sandboxService *service.SandboxService
}

// NewSandboxHandler creates a sandbox handler.
func NewSandboxHandler(sandboxService *service.SandboxService) *SandboxHandler {
	return &SandboxHandler{sandboxService: sandboxService}
}

// SandboxCheckoutRequest represents the public request body.
type SandboxCheckoutRequest struct {
	TransactionType string `json:"transactionType"`
	ServiceType     string `json:"serviceType"`
	ProductName     string `json:"productName" binding:"required"`
	Target          string `json:"target" binding:"required"`
	Description     string `json:"description"`
	Amount          int64  `json:"amount" binding:"required,min=1"`
	AdminFee        int64  `json:"adminFee"`
}

// Checkout handles POST /v1/sandbox/checkout.
func (h *SandboxHandler) Checkout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	var req SandboxCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	response, err := h.sandboxService.Checkout(c.Request.Context(), service.SandboxCheckoutRequest{
		UserID:          userID,
		TransactionType: req.TransactionType,
		ServiceType:     req.ServiceType,
		ProductName:     req.ProductName,
		Target:          req.Target,
		Description:     req.Description,
		Amount:          req.Amount,
		AdminFee:        req.AdminFee,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// CompleteDeposit handles POST /v1/sandbox/deposits/:depositId/complete.
func (h *SandboxHandler) CompleteDeposit(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	depositID := c.Param("depositId")
	if depositID == "" {
		respondWithError(c, domain.ErrValidationFailed("ID deposit wajib diisi"))
		return
	}

	response, err := h.sandboxService.CompleteDeposit(c.Request.Context(), userID, depositID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}
