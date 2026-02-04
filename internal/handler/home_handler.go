package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// HomeHandler handles home screen requests
type HomeHandler struct {
	homeService *service.HomeService
}

// NewHomeHandler creates a new home handler
func NewHomeHandler(homeService *service.HomeService) *HomeHandler {
	return &HomeHandler{
		homeService: homeService,
	}
}

// GetHome handles GET /v1/home
// Returns aggregated home screen data with optional version-based caching
func (h *HomeHandler) GetHome(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get version query params for caching
	servicesVersion := c.Query("servicesVersion")
	bannersVersion := c.Query("bannersVersion")

	// Call service
	response, err := h.homeService.GetHome(c.Request.Context(), userID, servicesVersion, bannersVersion)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetBalance handles GET /v1/user/balance
// Returns user balance information only (lightweight refresh)
func (h *HomeHandler) GetBalance(c *gin.Context) {
	// Get user ID from JWT
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Call service
	response, err := h.homeService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetServices handles GET /v1/services
// Returns services list with version support (304 Not Modified if cached)
func (h *HomeHandler) GetServices(c *gin.Context) {
	// Get user ID from JWT (for future user-specific filtering)
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get version query param
	version := c.Query("version")

	// Call service
	response, notModified, err := h.homeService.GetServices(c.Request.Context(), version)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Return 304 Not Modified if version matches
	if notModified {
		c.Status(http.StatusNotModified)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetBanners handles GET /v1/banners
// Returns banners list with placement filtering and version support
func (h *HomeHandler) GetBanners(c *gin.Context) {
	// Get user ID from JWT (for tier-based targeting)
	userID := middleware.GetUserID(c)
	if userID == "" {
		respondWithError(c, domain.ErrUnauthorizedError)
		return
	}

	// Get query params
	placement := c.DefaultQuery("placement", "home")
	version := c.Query("version")

	// Call service
	response, notModified, err := h.homeService.GetBanners(c.Request.Context(), userID, placement, version)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Return 304 Not Modified if version matches
	if notModified {
		c.Status(http.StatusNotModified)
		return
	}

	respondWithSuccess(c, http.StatusOK, response)
}
