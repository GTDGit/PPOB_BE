package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// NotificationHandler handles notification requests
type NotificationHandler struct {
	notificationService *service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// MarkAllAsReadRequest represents mark all as read request
type MarkAllAsReadRequest struct {
	Category *string `json:"category"`
}

// List handles GET /v1/notifications
// Returns paginated notification list
func (h *NotificationHandler) List(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get query params
	category := c.DefaultQuery("category", "all")
	isReadStr := c.Query("isRead")
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

	// Parse isRead filter
	var isRead *bool
	if isReadStr != "" {
		val := isReadStr == "true"
		isRead = &val
	}

	// Build filter
	filter := repository.NotificationFilter{
		Category: category,
		IsRead:   isRead,
		Page:     page,
		PerPage:  perPage,
	}

	// Call service
	response, err := h.notificationService.List(c.Request.Context(), userID, filter)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetDetail handles GET /v1/notifications/:id
// Returns detailed notification and auto marks as read
func (h *NotificationHandler) GetDetail(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get notification ID from path
	notificationID := c.Param("id")
	if notificationID == "" {
		respondWithError(c, domain.ErrValidationFailed("ID Notifikasi wajib diisi"))
		return
	}

	// Call service
	response, err := h.notificationService.GetDetail(c.Request.Context(), userID, notificationID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// MarkAsRead handles PUT /v1/notifications/:id/read
// Marks notification as read
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get notification ID from path
	notificationID := c.Param("id")
	if notificationID == "" {
		respondWithError(c, domain.ErrValidationFailed("ID Notifikasi wajib diisi"))
		return
	}

	// Call service
	response, err := h.notificationService.MarkAsRead(c.Request.Context(), userID, notificationID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// MarkAllAsRead handles PUT /v1/notifications/read-all
// Marks all notifications as read
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Parse request body (optional)
	var req MarkAllAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body or invalid JSON, just proceed with nil category
		req.Category = nil
	}

	// Call service
	response, err := h.notificationService.MarkAllAsRead(c.Request.Context(), userID, req.Category)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetUnreadCount handles GET /v1/notifications/unread-count
// Returns unread notification count
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Call service
	response, err := h.notificationService.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// Delete handles DELETE /v1/notifications/:id
// Deletes a notification
func (h *NotificationHandler) Delete(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get notification ID from path
	notificationID := c.Param("id")
	if notificationID == "" {
		respondWithError(c, domain.ErrValidationFailed("ID Notifikasi wajib diisi"))
		return
	}

	// Call service
	response, err := h.notificationService.Delete(c.Request.Context(), userID, notificationID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}
