package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
)

// respondWithSuccess sends a success response
func respondWithSuccess(c *gin.Context, statusCode int, data interface{}) {
	requestID := middleware.GetRequestID(c)
	c.JSON(statusCode, domain.SuccessResponse(data, requestID))
}

// respondWithError sends an error response
func respondWithError(c *gin.Context, appErr *domain.AppError) {
	requestID := middleware.GetRequestID(c)
	c.JSON(appErr.HTTPStatus, domain.ErrorResponse(appErr, requestID))
}

// handleServiceError handles errors from service layer
func handleServiceError(c *gin.Context, err error) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		respondWithError(c, appErr)
		return
	}

	// Log the error for debugging
	c.Error(err)

	// Return generic internal error
	respondWithError(c, domain.ErrInternalServerError)
}

// Pagination helper
type PaginationQuery struct {
	Page    int `form:"page" binding:"min=1"`
	PerPage int `form:"perPage" binding:"min=1,max=100"`
}
