package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminMailboxHandler struct {
	mailboxService *service.AdminMailboxService
}

func NewAdminMailboxHandler(mailboxService *service.AdminMailboxService) *AdminMailboxHandler {
	return &AdminMailboxHandler{mailboxService: mailboxService}
}

func (h *AdminMailboxHandler) ListMailboxes(c *gin.Context) {
	resp, err := h.mailboxService.ListMailboxes(c.Request.Context(), middleware.GetAdminID(c))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminMailboxHandler) CreateMailbox(c *gin.Context) {
	var req service.MailboxUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request mailbox tidak valid"))
		return
	}
	resp, err := h.mailboxService.CreateMailbox(c.Request.Context(), middleware.GetAdminID(c), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminMailboxHandler) UpdateMailbox(c *gin.Context) {
	var req service.MailboxUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request mailbox tidak valid"))
		return
	}
	resp, err := h.mailboxService.UpdateMailbox(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminMailboxHandler) ListMailboxThreads(c *gin.Context) {
	unreadOnly := strings.EqualFold(strings.TrimSpace(c.Query("unreadOnly")), "true")
	resp, err := h.mailboxService.ListMailboxThreads(
		c.Request.Context(),
		middleware.GetAdminID(c),
		c.Param("id"),
		c.Query("search"),
		c.Query("status"),
		c.Query("assigned"),
		unreadOnly,
		queryInt(c, "page", 1),
		queryInt(c, "perPage", 20),
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminMailboxHandler) GetThreadDetail(c *gin.Context) {
	resp, err := h.mailboxService.GetThreadDetail(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminMailboxHandler) ReplyThread(c *gin.Context) {
	var req service.ThreadReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request reply email tidak valid"))
		return
	}
	resp, err := h.mailboxService.ReplyThread(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminMailboxHandler) UpdateThreadStatus(c *gin.Context) {
	var req service.ThreadStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request status inbox tidak valid"))
		return
	}
	if err := h.mailboxService.UpdateThreadStatus(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"), req); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Status thread berhasil diperbarui"})
}

func (h *AdminMailboxHandler) AssignThread(c *gin.Context) {
	var req service.ThreadAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request assign inbox tidak valid"))
		return
	}
	if err := h.mailboxService.AssignThread(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"), req); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Thread inbox berhasil di-assign"})
}

func (h *AdminMailboxHandler) ComposeEmail(c *gin.Context) {
	var req service.ComposeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request compose email tidak valid"))
		return
	}
	resp, err := h.mailboxService.ComposeEmail(c.Request.Context(), middleware.GetAdminID(c), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminMailboxHandler) UpdateMailboxDisplayName(c *gin.Context) {
	var req struct {
		DisplayName string `json:"displayName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request display name tidak valid"))
		return
	}
	resp, err := h.mailboxService.UpdateMailboxDisplayName(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"), req.DisplayName)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminMailboxHandler) ListEmailLogs(c *gin.Context) {
	resp, err := h.mailboxService.ListEmailLogs(c.Request.Context(), middleware.GetAdminID(c), c.Query("search"), c.Query("status"), c.Query("category"), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminMailboxHandler) InboundSNS(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondWithError(c, domain.ErrValidationFailed("Gagal membaca payload SNS inbound email"))
		return
	}
	resp, err := h.mailboxService.HandleInboundSNSEvent(c.Request.Context(), body)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminMailboxHandler) DeliverySNS(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondWithError(c, domain.ErrValidationFailed("Gagal membaca payload SNS delivery email"))
		return
	}
	resp, err := h.mailboxService.HandleDeliveryEvent(c.Request.Context(), body)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}
