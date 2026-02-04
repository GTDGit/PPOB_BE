package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// ProductHandler handles product-related requests
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// GetOperators handles GET /v1/products/operators
func (h *ProductHandler) GetOperators(c *gin.Context) {
	operators, err := h.productService.GetOperators(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Build response with version for cache control
	response := map[string]interface{}{
		"operators": operators,
		"version":   time.Now().Format("2006010215"), // YYYYMMDDHH format
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetEwalletProviders handles GET /v1/products/ewallet/providers
func (h *ProductHandler) GetEwalletProviders(c *gin.Context) {
	providers, err := h.productService.GetEwalletProviders(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Build response with version for cache control
	response := map[string]interface{}{
		"providers": providers,
		"version":   time.Now().Format("2006010215"),
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetPDAMRegions handles GET /v1/products/pdam/regions
func (h *ProductHandler) GetPDAMRegions(c *gin.Context) {
	regions, err := h.productService.GetPDAMRegions(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Build response with version for cache control
	response := map[string]interface{}{
		"regions": regions,
		"version": time.Now().Format("2006010215"),
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetBanks handles GET /v1/products/banks
func (h *ProductHandler) GetBanks(c *gin.Context) {
	// Get filter type from query param
	filterType := c.DefaultQuery("type", "all")

	banks, err := h.productService.GetBanks(c.Request.Context(), filterType)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Build response with version for cache control
	response := map[string]interface{}{
		"banks":   banks,
		"version": time.Now().Format("2006010215"),
	}

	respondWithSuccess(c, http.StatusOK, response)
}

// GetTVProviders handles GET /v1/products/tv/providers
func (h *ProductHandler) GetTVProviders(c *gin.Context) {
	providers, err := h.productService.GetTVProviders(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// Build response with version for cache control
	response := map[string]interface{}{
		"providers": providers,
		"version":   time.Now().Format("2006010215"),
	}

	respondWithSuccess(c, http.StatusOK, response)
}
