package handler

import (
	"net/http"
	"strconv"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/middleware"
	"github.com/GTDGit/PPOB_BE/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

type adminLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	TOTPCode string `json:"totpCode" binding:"required"`
}

type adminRefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type adminAcceptInviteRequest struct {
	Token    string `json:"token" binding:"required"`
	FullName string `json:"fullName" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type adminConfirmInviteRequest struct {
	Token string `json:"token" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type adminCreateInviteRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required"`
	FullName string `json:"fullName"`
	RoleID   string `json:"roleId" binding:"required"`
}

type adminSetStatusRequest struct {
	Status   string `json:"status" binding:"required"`
	IsActive bool   `json:"isActive"`
}

type approvalRejectRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type upsertSettingRequest struct {
	Key         string      `json:"key" binding:"required"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
}

func (h *AdminHandler) Bootstrap(c *gin.Context) {
	var req struct {
		Secret   string `json:"secret" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Phone    string `json:"phone" binding:"required"`
		FullName string `json:"fullName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request bootstrap admin tidak valid"))
		return
	}
	resp, err := h.adminService.BootstrapFirstAdmin(c.Request.Context(), req.Secret, req.Email, req.Phone, req.FullName)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req adminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request login admin tidak valid"))
		return
	}
	resp, err := h.adminService.Login(c.Request.Context(), req.Email, req.Password, req.TOTPCode, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) Refresh(c *gin.Context) {
	var req adminRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request refresh admin tidak valid"))
		return
	}
	resp, err := h.adminService.Refresh(c.Request.Context(), req.RefreshToken, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) Logout(c *gin.Context) {
	if err := h.adminService.Logout(c.Request.Context(), middleware.GetAdminID(c), middleware.GetAdminSessionID(c)); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Logout admin berhasil"})
}

func (h *AdminHandler) Me(c *gin.Context) {
	user, permissions, err := h.adminService.GetMe(c.Request.Context(), middleware.GetAdminID(c))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{
		"user":        user,
		"permissions": permissions,
	})
}

func (h *AdminHandler) GetInvitePreview(c *gin.Context) {
	resp, err := h.adminService.GetInvitePreview(c.Request.Context(), c.Param("token"))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) AcceptInvite(c *gin.Context) {
	var req adminAcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request aktivasi admin tidak valid"))
		return
	}
	resp, err := h.adminService.AcceptInvite(c.Request.Context(), req.Token, req.FullName, req.Password)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) ConfirmInviteTOTP(c *gin.Context) {
	var req adminConfirmInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request konfirmasi authenticator tidak valid"))
		return
	}
	resp, err := h.adminService.ConfirmInviteTOTP(c.Request.Context(), req.Token, req.Code, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) ListRoles(c *gin.Context) {
	resp, err := h.adminService.ListRoles(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) ListPermissions(c *gin.Context) {
	resp, err := h.adminService.ListPermissions(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) DashboardSummary(c *gin.Context) {
	resp, err := h.adminService.DashboardSummary(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) ListAdmins(c *gin.Context) {
	resp, err := h.adminService.ListAdmins(c.Request.Context(), c.Query("search"), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) CreateInvite(c *gin.Context) {
	var req adminCreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request undangan admin tidak valid"))
		return
	}
	resp, err := h.adminService.CreateInvite(c.Request.Context(), middleware.GetAdminID(c), service.CreateAdminInviteRequest{
		Email:    req.Email,
		Phone:    req.Phone,
		FullName: req.FullName,
		RoleID:   req.RoleID,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) SetAdminStatus(c *gin.Context) {
	var req adminSetStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request status admin tidak valid"))
		return
	}
	if err := h.adminService.SetAdminStatus(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"), req.Status, req.IsActive); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Status admin diperbarui"})
}

func (h *AdminHandler) ListCustomers(c *gin.Context) {
	resp, err := h.adminService.ListCustomers(c.Request.Context(), c.Query("search"), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) ListTransactions(c *gin.Context) {
	resp, err := h.adminService.ListTransactions(c.Request.Context(), c.Query("search"), c.Query("status"), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) ListDeposits(c *gin.Context) {
	resp, err := h.adminService.ListDeposits(c.Request.Context(), c.Query("search"), c.Query("status"), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) ApproveDeposit(c *gin.Context) {
	if err := h.adminService.ApproveDeposit(c.Request.Context(), middleware.GetAdminID(c), c.Param("id")); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Deposit berhasil di-approve"})
}

func (h *AdminHandler) RejectDeposit(c *gin.Context) {
	if err := h.adminService.RejectDeposit(c.Request.Context(), middleware.GetAdminID(c), c.Param("id")); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Deposit berhasil di-reject"})
}

func (h *AdminHandler) ListQris(c *gin.Context) {
	resp, err := h.adminService.ListQris(c.Request.Context(), c.Query("search"), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func (h *AdminHandler) ListVouchers(c *gin.Context) {
	resp, err := h.adminService.ListVouchers(c.Request.Context(), c.Query("search"), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) CreateVoucher(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request voucher tidak valid"))
		return
	}
	if err := h.adminService.CreateVoucher(c.Request.Context(), middleware.GetAdminID(c), payload); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Voucher berhasil dibuat"})
}

func (h *AdminHandler) UpdateVoucher(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request voucher tidak valid"))
		return
	}
	if err := h.adminService.UpdateVoucher(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"), payload); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Voucher berhasil diperbarui"})
}

func (h *AdminHandler) UpdateVoucherStatus(c *gin.Context) {
	var req struct {
		IsActive bool `json:"isActive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request status voucher tidak valid"))
		return
	}
	if err := h.adminService.UpdateVoucherStatus(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"), req.IsActive); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Status voucher berhasil diperbarui"})
}

func (h *AdminHandler) GetCatalog(c *gin.Context) {
	resp, err := h.adminService.ListCatalog(c.Request.Context(), c.Query("search"), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) CreatePricingRequest(c *gin.Context) {
	var payload struct {
		ProductID   string `json:"productId" binding:"required"`
		NewPrice    int64  `json:"newPrice" binding:"required"`
		NewAdminFee int64  `json:"newAdminFee"`
		Reason      string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request pricing tidak valid"))
		return
	}
	err := h.adminService.CreatePricingRequest(c.Request.Context(), middleware.GetAdminID(c), map[string]interface{}{
		"productId":   payload.ProductID,
		"newPrice":    payload.NewPrice,
		"newAdminFee": payload.NewAdminFee,
	}, payload.Reason)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Request perubahan pricing berhasil dibuat"})
}

func (h *AdminHandler) CreateBalanceAdjustmentRequest(c *gin.Context) {
	var payload struct {
		UserID      string `json:"userId" binding:"required"`
		AmountDelta int64  `json:"amountDelta" binding:"required"`
		Description string `json:"description"`
		Reason      string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request koreksi saldo tidak valid"))
		return
	}
	err := h.adminService.CreateBalanceAdjustmentRequest(c.Request.Context(), middleware.GetAdminID(c), map[string]interface{}{
		"userId":      payload.UserID,
		"amountDelta": payload.AmountDelta,
		"description": payload.Description,
	}, payload.Reason)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Request koreksi saldo berhasil dibuat"})
}

func (h *AdminHandler) ListKYC(c *gin.Context) {
	resp, err := h.adminService.ListKYC(c.Request.Context(), c.Query("search"), c.Query("status"), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) ApproveKYC(c *gin.Context) {
	if err := h.adminService.ApproveKYC(c.Request.Context(), middleware.GetAdminID(c), c.Param("userId")); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "KYC berhasil di-approve"})
}

func (h *AdminHandler) RejectKYC(c *gin.Context) {
	if err := h.adminService.RejectKYC(c.Request.Context(), middleware.GetAdminID(c), c.Param("userId")); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "KYC berhasil di-reject"})
}

func (h *AdminHandler) ListBanners(c *gin.Context) {
	resp, err := h.adminService.ListBanners(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) CreateBanner(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request banner tidak valid"))
		return
	}
	if err := h.adminService.CreateBanner(c.Request.Context(), middleware.GetAdminID(c), payload); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Banner berhasil dibuat"})
}

func (h *AdminHandler) UpdateBanner(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request banner tidak valid"))
		return
	}
	if err := h.adminService.UpdateBanner(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"), payload); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Banner berhasil diperbarui"})
}

func (h *AdminHandler) DeleteBanner(c *gin.Context) {
	if err := h.adminService.DeleteBanner(c.Request.Context(), middleware.GetAdminID(c), c.Param("id")); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Banner berhasil dihapus"})
}

func (h *AdminHandler) ListNotifications(c *gin.Context) {
	resp, err := h.adminService.ListNotifications(c.Request.Context(), c.Query("search"), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) BroadcastNotification(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request broadcast notification tidak valid"))
		return
	}
	if err := h.adminService.BroadcastNotification(c.Request.Context(), middleware.GetAdminID(c), payload); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Notifikasi berhasil dikirim"})
}

func (h *AdminHandler) ListApprovals(c *gin.Context) {
	resp, err := h.adminService.ListApprovals(c.Request.Context(), c.Query("status"), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) ApproveApproval(c *gin.Context) {
	if err := h.adminService.ApproveApproval(c.Request.Context(), middleware.GetAdminID(c), c.Param("id")); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Approval berhasil diproses"})
}

func (h *AdminHandler) RejectApproval(c *gin.Context) {
	var req approvalRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request reject approval tidak valid"))
		return
	}
	if err := h.adminService.RejectApproval(c.Request.Context(), middleware.GetAdminID(c), c.Param("id"), req.Reason); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Approval berhasil ditolak"})
}

func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	resp, err := h.adminService.ListAuditLogs(c.Request.Context(), queryInt(c, "page", 1), queryInt(c, "perPage", 20))
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) ListSettings(c *gin.Context) {
	resp, err := h.adminService.ListSettings(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}

func (h *AdminHandler) UpsertSetting(c *gin.Context) {
	var req upsertSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondWithError(c, domain.ErrValidationFailed("Body request setting tidak valid"))
		return
	}
	if err := h.adminService.UpsertSetting(c.Request.Context(), middleware.GetAdminID(c), req.Key, req.Description, req.Value); err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, gin.H{"message": "Setting berhasil diperbarui"})
}

func (h *AdminHandler) ListReferenceData(c *gin.Context) {
	resp, err := h.adminService.ListReferenceData(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}
	respondWithSuccess(c, http.StatusOK, resp)
}
