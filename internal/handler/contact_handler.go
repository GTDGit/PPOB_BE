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

// ContactHandler handles contact-related requests
type ContactHandler struct {
	contactService *service.ContactService
}

// NewContactHandler creates a new contact handler
func NewContactHandler(contactService *service.ContactService) *ContactHandler {
	return &ContactHandler{
		contactService: contactService,
	}
}

// List handles GET /v1/contacts
func (h *ContactHandler) List(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get query params
	contactType := c.DefaultQuery("type", "")
	search := c.DefaultQuery("search", "")
	limitStr := c.DefaultQuery("limit", "20")

	// Parse limit
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	// Build filter
	filter := repository.ContactFilter{
		Type:   contactType,
		Search: search,
		Limit:  limit,
	}

	// Call service
	response, err := h.contactService.List(c.Request.Context(), userID, filter)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// CreateContactRequest represents create contact request body
type CreateContactRequest struct {
	Name     string                 `json:"name" binding:"required,min=1,max=50"`
	Type     string                 `json:"type" binding:"required,oneof=phone pln pdam bpjs telkom bank"`
	Value    string                 `json:"value" binding:"required"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Create handles POST /v1/contacts
func (h *ContactHandler) Create(c *gin.Context) {
	var req CreateContactRequest
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
	response, err := h.contactService.Create(c.Request.Context(), userID, req.Name, req.Type, req.Value, req.Metadata)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusCreated, response)
}

// UpdateContactRequest represents update contact request body
type UpdateContactRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// Update handles PUT /v1/contacts/:contactId
func (h *ContactHandler) Update(c *gin.Context) {
	var req UpdateContactRequest
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

	// Get contact ID from path
	contactID := c.Param("contactId")
	if contactID == "" {
		respondWithError(c, domain.ErrValidationFailed("ID Kontak wajib diisi"))
		return
	}

	// Call service
	response, err := h.contactService.Update(c.Request.Context(), userID, contactID, req.Name)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// Delete handles DELETE /v1/contacts/:contactId
func (h *ContactHandler) Delete(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get contact ID from path
	contactID := c.Param("contactId")
	if contactID == "" {
		respondWithError(c, domain.ErrValidationFailed("ID Kontak wajib diisi"))
		return
	}

	// Call service
	response, err := h.contactService.Delete(c.Request.Context(), userID, contactID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}
