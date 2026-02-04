package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// KYCHandler handles KYC verification HTTP requests
type KYCHandler struct {
	service *service.KYCService
}

// NewKYCHandler creates a new KYC handler
func NewKYCHandler(service *service.KYCService) *KYCHandler {
	return &KYCHandler{service: service}
}

// GetStatus handles GET /v1/kyc/status
func (h *KYCHandler) GetStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)

	verification, err := h.service.GetStatus(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	if verification == nil {
		respondWithSuccess(c, http.StatusOK, gin.H{
			"message":  "Belum ada verifikasi KYC",
			"verified": false,
			"data":     nil,
		})
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message":  "Berhasil mengambil status KYC",
		"verified": true,
		"data":     verification,
	})
}

// GetSession handles GET /v1/kyc/session
func (h *KYCHandler) GetSession(c *gin.Context) {
	userID := middleware.GetUserID(c)

	session, err := h.service.GetActiveSession(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	if session == nil {
		respondWithSuccess(c, http.StatusOK, gin.H{
			"message": "Tidak ada sesi verifikasi aktif",
			"data":    nil,
		})
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Berhasil mengambil sesi verifikasi",
		"data":    session,
	})
}

// StartVerification handles POST /v1/kyc/start
func (h *KYCHandler) StartVerification(c *gin.Context) {
	userID := middleware.GetUserID(c)

	session, err := h.service.StartVerification(c.Request.Context(), userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Sesi verifikasi KYC berhasil dibuat",
		"data":    session,
	})
}

// CancelVerification handles POST /v1/kyc/cancel
func (h *KYCHandler) CancelVerification(c *gin.Context) {
	userID := middleware.GetUserID(c)

	if err := h.service.CancelVerification(c.Request.Context(), userID); err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Verifikasi KYC berhasil dibatalkan",
	})
}

// UploadKTP handles POST /v1/kyc/ktp
func (h *KYCHandler) UploadKTP(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// Get session ID from form
	sessionID := c.PostForm("sessionId")
	if sessionID == "" {
		respondWithError(c, domain.ErrValidationFailed("Session ID wajib diisi"))
		return
	}

	// Get KTP file from multipart form
	file, err := c.FormFile("ktp")
	if err != nil {
		respondWithError(c, domain.ErrValidationFailed("File KTP wajib diupload"))
		return
	}

	// Open file
	fileContent, err := file.Open()
	if err != nil {
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}
	defer fileContent.Close()

	// Read file bytes
	fileBytes := make([]byte, file.Size)
	if _, err := fileContent.Read(fileBytes); err != nil {
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	// Call service
	if err := h.service.UploadKTP(c.Request.Context(), userID, sessionID, fileBytes, file.Filename); err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "KTP berhasil diupload dan diproses",
	})
}

// UploadFacePhotos handles POST /v1/kyc/face
func (h *KYCHandler) UploadFacePhotos(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// Get session ID from form
	sessionID := c.PostForm("sessionId")
	if sessionID == "" {
		respondWithError(c, domain.ErrValidationFailed("Session ID wajib diisi"))
		return
	}

	// Get face file (selfie)
	faceFile, err := c.FormFile("face")
	if err != nil {
		respondWithError(c, domain.ErrValidationFailed("File wajah wajib diupload"))
		return
	}

	// Get full image file (face + KTP)
	fullImageFile, err := c.FormFile("fullImage")
	if err != nil {
		respondWithError(c, domain.ErrValidationFailed("File gambar lengkap wajib diupload"))
		return
	}

	// Read face file
	faceContent, err := faceFile.Open()
	if err != nil {
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}
	defer faceContent.Close()

	faceBytes := make([]byte, faceFile.Size)
	if _, err := faceContent.Read(faceBytes); err != nil {
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	// Read full image file
	fullImageContent, err := fullImageFile.Open()
	if err != nil {
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}
	defer fullImageContent.Close()

	fullImageBytes := make([]byte, fullImageFile.Size)
	if _, err := fullImageContent.Read(fullImageBytes); err != nil {
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	// Call service
	if err := h.service.UploadFacePhotos(c.Request.Context(), userID, sessionID, faceBytes, faceFile.Filename, fullImageBytes, fullImageFile.Filename); err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Foto wajah berhasil diupload",
	})
}

// RunLiveness handles POST /v1/kyc/liveness (DEPRECATED)
// Use CreateLivenessSession and VerifyLiveness instead
func (h *KYCHandler) RunLiveness(c *gin.Context) {
	respondWithError(c, domain.NewError("DEPRECATED", "Gunakan POST /v1/kyc/liveness/session dan POST /v1/kyc/liveness/verify", 400))
}

// CreateLivenessSession handles POST /v1/kyc/liveness/session
// Creates liveness session for frontend FaceLivenessDetector SDK
func (h *KYCHandler) CreateLivenessSession(c *gin.Context) {
	userID := middleware.GetUserID(c)

	type Request struct {
		SessionID string `json:"sessionId" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	result, err := h.service.CreateLivenessSession(c.Request.Context(), userID, req.SessionID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Session liveness berhasil dibuat",
		"data":    result,
	})
}

// VerifyLiveness handles POST /v1/kyc/liveness/verify
// Verifies liveness after frontend completes FaceLivenessDetector
func (h *KYCHandler) VerifyLiveness(c *gin.Context) {
	userID := middleware.GetUserID(c)

	type Request struct {
		SessionID string `json:"sessionId" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	result, err := h.service.VerifyLiveness(c.Request.Context(), userID, req.SessionID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Verifikasi KYC berhasil",
		"data":    result,
	})
}

// Submit handles POST /v1/kyc/submit
func (h *KYCHandler) Submit(c *gin.Context) {
	userID := middleware.GetUserID(c)

	type Request struct {
		SessionID string `json:"sessionId" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrInvalidRequestError)
		return
	}

	if err := h.service.SubmitForReview(c.Request.Context(), userID, req.SessionID); err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Verifikasi KYC berhasil disubmit",
	})
}
