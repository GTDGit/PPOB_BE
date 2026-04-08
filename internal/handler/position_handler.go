package handler

import (
	"net/http"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/service"
	"github.com/gin-gonic/gin"
)

type PositionHandler struct {
	positionService *service.PositionService
}

func NewPositionHandler(positionService *service.PositionService) *PositionHandler {
	return &PositionHandler{positionService: positionService}
}

func (h *PositionHandler) ListPositions(c *gin.Context) {
	items, err := h.positionService.ListPositions(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, items)
}

func (h *PositionHandler) CreatePosition(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Nama posisi wajib diisi"))
		return
	}
	pos, err := h.positionService.CreatePosition(c.Request.Context(), req.Name)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusCreated, pos)
}

func (h *PositionHandler) UpdatePosition(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Nama posisi wajib diisi"))
		return
	}
	pos, err := h.positionService.UpdatePosition(c.Request.Context(), id, req.Name)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, pos)
}

func (h *PositionHandler) DeletePosition(c *gin.Context) {
	id := c.Param("id")
	if err := h.positionService.DeletePosition(c.Request.Context(), id); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Posisi berhasil dihapus"})
}

func (h *PositionHandler) GetPositionAdmins(c *gin.Context) {
	id := c.Param("id")
	items, err := h.positionService.GetPositionAdmins(c.Request.Context(), id)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, items)
}

func (h *PositionHandler) AssignPosition(c *gin.Context) {
	positionID := c.Param("id")
	var req struct {
		AdminID string `json:"adminId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Admin ID wajib diisi"))
		return
	}
	if err := h.positionService.AssignPosition(c.Request.Context(), positionID, req.AdminID); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Posisi berhasil ditetapkan"})
}

func (h *PositionHandler) RemoveAdminPosition(c *gin.Context) {
	adminID := c.Param("id")
	if err := h.positionService.RemoveAdminPosition(c.Request.Context(), adminID); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Posisi admin berhasil dihapus"})
}
