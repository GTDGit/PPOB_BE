package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// HistoryHandler handles transaction history requests
type HistoryHandler struct {
	historyService *service.HistoryService
	depositService *service.DepositService
}

// NewHistoryHandler creates a new history handler
func NewHistoryHandler(historyService *service.HistoryService, depositService *service.DepositService) *HistoryHandler {
	return &HistoryHandler{
		historyService: historyService,
		depositService: depositService,
	}
}

// List handles GET /v1/history/transactions
// Returns paginated transaction history list
func (h *HistoryHandler) List(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get query params
	txType := c.DefaultQuery("type", "all")
	serviceType := c.Query("serviceType")
	status := c.DefaultQuery("status", "all")
	search := c.Query("search")
	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("limit", "20")
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")

	// Parse pagination
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = 20
	}

	// Parse dates
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Build filter
	filter := repository.TransactionFilter{
		Type:        txType,
		ServiceType: serviceType,
		Status:      status,
		StartDate:   startDate,
		EndDate:     endDate,
		Search:      search,
		Page:        page,
		PerPage:     perPage,
	}

	// Call service
	response, err := h.historyService.List(c.Request.Context(), userID, filter)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetDetail handles GET /v1/history/transactions/:transactionId
// Returns detailed transaction information
func (h *HistoryHandler) GetDetail(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get transaction ID from path
	transactionID := c.Param("transactionId")
	if transactionID == "" {
		respondWithError(c, domain.ErrValidationFailed("ID Transaksi wajib diisi"))
		return
	}

	// Call service
	response, err := h.historyService.GetDetail(c.Request.Context(), userID, transactionID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// ListDeposits handles GET /v1/history/deposits
// Returns deposit history (alias to deposit history endpoint)
func (h *HistoryHandler) ListDeposits(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Parse filter params
	pageStr := c.DefaultQuery("page", "1")
	perPageStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage < 1 {
		perPage = 20
	}

	filter := repository.DepositFilter{
		Page:    page,
		PerPage: perPage,
	}

	response, err := h.depositService.GetHistory(c.Request.Context(), userID, filter)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// ListQRISIncome handles GET /v1/history/qris
// Returns QRIS income history (pembayaran dari pelanggan)
func (h *HistoryHandler) ListQRISIncome(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// TODO: Implement QRIS income tracking
	// This requires new database table and service
	respondWithError(c, domain.ErrValidationFailed("Fitur belum tersedia"))
}

// GetReceipt handles GET /v1/history/transactions/:transactionId/receipt
// Returns receipt data for transaction
func (h *HistoryHandler) GetReceipt(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	transactionID := c.Param("transactionId")
	if transactionID == "" {
		respondWithError(c, domain.ErrValidationFailed("ID Transaksi wajib diisi"))
		return
	}

	response, err := h.historyService.GetReceipt(c.Request.Context(), userID, transactionID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// DownloadReceipt handles GET /v1/history/transactions/:transactionId/receipt/download
// Generates and downloads receipt as image or PDF
func (h *HistoryHandler) DownloadReceipt(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	transactionID := c.Param("transactionId")
	format := c.DefaultQuery("format", "png") // png or pdf

	// Generate receipt image/PDF
	data, contentType, err := h.historyService.GenerateReceipt(c.Request.Context(), userID, transactionID, format)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.Header("Content-Disposition", "attachment; filename=receipt-"+transactionID+"."+format)
	c.Data(http.StatusOK, contentType, data)
}

// ShareReceipt handles GET /v1/history/transactions/:transactionId/receipt/share
// Returns shareable receipt data (URL, text, etc)
func (h *HistoryHandler) ShareReceipt(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	transactionID := c.Param("transactionId")

	response, err := h.historyService.GetShareData(c.Request.Context(), userID, transactionID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// UpdateSellingPrice handles PUT /v1/history/transactions/:transactionId/selling-price
// Updates custom selling price for transaction (for merchants)
func (h *HistoryHandler) UpdateSellingPrice(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	transactionID := c.Param("transactionId")

	var req struct {
		SellingPrice int64 `json:"sellingPrice" binding:"required,min=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}

	err := h.historyService.UpdateSellingPrice(c.Request.Context(), userID, transactionID, req.SellingPrice)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Harga jual berhasil diperbarui"})
}
