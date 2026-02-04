package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// UserHandler handles user profile and settings requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile handles GET /v1/user/profile
// Returns user profile with stats and tier information
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Call service
	response, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// UpdateProfileRequest represents update profile request body
type UpdateProfileRequest struct {
	FullName     *string `json:"fullName"`
	Email        *string `json:"email"`
	Gender       *string `json:"gender"`
	BusinessType *string `json:"businessType"`
}

// UpdateProfile handles PUT /v1/user/profile
// Updates user profile information
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
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
	response, err := h.userService.UpdateProfile(c.Request.Context(), userID, req.FullName, req.Email, req.Gender, req.BusinessType)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// UploadAvatar handles POST /v1/user/avatar
// Uploads user avatar image
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// TODO: Handle multipart form file upload
	// file, err := c.FormFile("avatar")
	// For now, just call service with mock

	// Call service
	response, err := h.userService.UploadAvatar(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// DeleteAvatar handles DELETE /v1/user/avatar
// Deletes user avatar image
func (h *UserHandler) DeleteAvatar(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Call service
	response, err := h.userService.DeleteAvatar(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetSettings handles GET /v1/user/settings
// Returns user settings
func (h *UserHandler) GetSettings(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Call service
	response, err := h.userService.GetSettings(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// UpdateSettings handles PUT /v1/user/settings
// Updates user settings
func (h *UserHandler) UpdateSettings(c *gin.Context) {
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
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
	response, err := h.userService.UpdateSettings(c.Request.Context(), userID, updates)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetReferralInfo handles GET /v1/user/referral
// Returns referral information and statistics
func (h *UserHandler) GetReferralInfo(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Call service
	response, err := h.userService.GetReferralInfo(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetReferralHistory handles GET /v1/user/referral/history
// Returns referral history list
func (h *UserHandler) GetReferralHistory(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Call service
	response, err := h.userService.GetReferralHistory(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}
