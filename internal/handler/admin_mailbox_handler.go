package handler

import (
	"io"
	"log"
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
	contentType := c.ContentType()

	var req service.ThreadReplyRequest
	if strings.Contains(contentType, "multipart/form-data") {
		req.Body = c.PostForm("body")
		req.HTMLBody = c.PostForm("htmlBody")
		req.Cc = c.PostFormArray("cc")
		req.Bcc = c.PostFormArray("bcc")
		isImp := c.PostForm("isImportant") == "true"
		req.IsImportant = &isImp

		form, _ := c.MultipartForm()
		if form != nil && form.File["attachments"] != nil {
			for _, fh := range form.File["attachments"] {
				if fh.Size > 10*1024*1024 {
					respondWithError(c, domain.ErrValidationFailed("Ukuran file maksimal 10MB per lampiran"))
					return
				}
				f, err := fh.Open()
				if err != nil {
					continue
				}
				data, err := io.ReadAll(f)
				f.Close()
				if err != nil {
					continue
				}
				ct := fh.Header.Get("Content-Type")
				if ct == "" {
					ct = "application/octet-stream"
				}
				req.Attachments = append(req.Attachments, service.EmailAttachmentInput{
					Filename:    fh.Filename,
					ContentType: ct,
					Data:        data,
				})
			}
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			respondWithError(c, domain.ErrValidationFailed("Body request reply email tidak valid"))
			return
		}
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
	contentType := c.ContentType()

	var req service.ComposeEmailRequest
	if strings.Contains(contentType, "multipart/form-data") {
		req.MailboxID = c.PostForm("mailboxId")
		req.Subject = c.PostForm("subject")
		req.Body = c.PostForm("body")
		req.HTMLBody = c.PostForm("htmlBody")
		req.IsImportant = c.PostForm("isImportant") == "true"
		req.To = c.PostFormArray("to")
		req.Cc = c.PostFormArray("cc")
		req.Bcc = c.PostFormArray("bcc")

		form, _ := c.MultipartForm()
		if form != nil && form.File["attachments"] != nil {
			for _, fh := range form.File["attachments"] {
				if fh.Size > 10*1024*1024 {
					respondWithError(c, domain.ErrValidationFailed("Ukuran file maksimal 10MB per lampiran"))
					return
				}
				f, err := fh.Open()
				if err != nil {
					continue
				}
				data, err := io.ReadAll(f)
				f.Close()
				if err != nil {
					continue
				}
				ct := fh.Header.Get("Content-Type")
				if ct == "" {
					ct = "application/octet-stream"
				}
				req.Attachments = append(req.Attachments, service.EmailAttachmentInput{
					Filename:    fh.Filename,
					ContentType: ct,
					Data:        data,
				})
			}
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			respondWithError(c, domain.ErrValidationFailed("Body request compose email tidak valid"))
			return
		}
	}

	resp, err := h.mailboxService.ComposeEmail(c.Request.Context(), middleware.GetAdminID(c), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminMailboxHandler) ToggleThreadImportant(c *gin.Context) {
	var req struct {
		IsImportant bool `json:"isImportant"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request tidak valid"))
		return
	}
	if err := h.mailboxService.ToggleThreadImportant(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"), req.IsImportant); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Status penting berhasil diperbarui"})
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
	log.Printf("[SNS-INBOUND] content-type=%s body-len=%d body-preview=%.500s", c.ContentType(), len(body), string(body))
	resp, err := h.mailboxService.HandleInboundSNSEvent(c.Request.Context(), body)
	if err != nil {
		log.Printf("[SNS-INBOUND] error: %v", err)
		handleServiceError(c, err)
		return
	}
	log.Printf("[SNS-INBOUND] success: %v", resp)
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
