package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/google/uuid"
)

func (s *AdminService) DashboardSummary(ctx context.Context) (*domain.AdminDashboardResponse, error) {
	return s.repo.GetDashboardSummary(ctx)
}

func (s *AdminService) ListAdmins(ctx context.Context, search string, page, perPage int) (*domain.AdminListResponse, error) {
	items, total, err := s.repo.ListAdmins(ctx, search, page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(items, total, page, perPage), nil
}

func (s *AdminService) GetAdminDetail(ctx context.Context, adminID string) (map[string]interface{}, error) {
	item, err := s.repo.GetAdminDetail(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, domain.ErrNotFound("Admin")
	}
	return item, nil
}

func (s *AdminService) SetAdminStatus(ctx context.Context, actorID, adminID, status string, isActive bool) error {
	if err := s.repo.UpdateAdminStatus(ctx, adminID, status, isActive); err != nil {
		return fmt.Errorf("failed to update admin status: %w", err)
	}
	_ = s.logAudit(ctx, actorID, "admin.status.update", "admin_user", adminID, nil, map[string]interface{}{
		"status":   status,
		"isActive": isActive,
	}, "", "", "success", nil)
	return nil
}

func (s *AdminService) ListCustomers(ctx context.Context, search string, page, perPage int) (*domain.AdminListResponse, error) {
	items, total, err := s.repo.ListCustomers(ctx, search, page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(items, total, page, perPage), nil
}

func (s *AdminService) GetCustomerDetail(ctx context.Context, userID string) (map[string]interface{}, error) {
	item, err := s.repo.GetCustomerDetail(ctx, userID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, domain.ErrNotFound("Pelanggan")
	}
	return item, nil
}

func (s *AdminService) ListTransactions(ctx context.Context, search, status string, page, perPage int) (*domain.AdminListResponse, error) {
	items, total, err := s.repo.ListTransactions(ctx, search, status, page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(items, total, page, perPage), nil
}

func (s *AdminService) GetTransactionDetail(ctx context.Context, transactionID string) (map[string]interface{}, error) {
	item, err := s.repo.GetTransactionDetail(ctx, transactionID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, domain.ErrNotFound("Transaksi")
	}
	return item, nil
}

func (s *AdminService) ListDeposits(ctx context.Context, search, status string, page, perPage int) (*domain.AdminListResponse, error) {
	items, total, err := s.repo.ListDeposits(ctx, search, status, page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(items, total, page, perPage), nil
}

func (s *AdminService) GetDepositDetail(ctx context.Context, depositID string) (map[string]interface{}, error) {
	item, err := s.repo.GetDepositDetail(ctx, depositID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, domain.ErrNotFound("Deposit")
	}
	return item, nil
}

func (s *AdminService) ApproveDeposit(ctx context.Context, actorID, depositID string) error {
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var deposit struct {
		ID          string `db:"id"`
		UserID      string `db:"user_id"`
		Amount      int64  `db:"amount"`
		Status      string `db:"status"`
		TotalAmount int64  `db:"total_amount"`
	}
	if err := tx.GetContext(ctx, &deposit, `
		SELECT id, user_id, amount, status, total_amount
		FROM deposits
		WHERE id = $1
		FOR UPDATE
	`, depositID); err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrDepositNotFound
		}
		return fmt.Errorf("failed to load deposit: %w", err)
	}
	if deposit.Status == domain.DepositStatusSuccess {
		return domain.NewError("DEPOSIT_ALREADY_APPROVED", "Deposit sudah berhasil", 409)
	}
	if deposit.Status != domain.DepositStatusPending {
		return domain.NewError("DEPOSIT_INVALID_STATUS", "Hanya deposit pending yang dapat di-approve", 409)
	}

	var balance struct {
		ID     string `db:"id"`
		UserID string `db:"user_id"`
		Amount int64  `db:"amount"`
	}
	if err := tx.GetContext(ctx, &balance, `
		SELECT id, user_id, amount FROM balances WHERE user_id = $1 FOR UPDATE
	`, deposit.UserID); err != nil {
		return fmt.Errorf("failed to load balance: %w", err)
	}

	before := balance.Amount
	after := before + deposit.Amount

	if _, err := tx.ExecContext(ctx, `
		UPDATE deposits SET status = 'success', paid_at = NOW(), updated_at = NOW() WHERE id = $1
	`, depositID); err != nil {
		return fmt.Errorf("failed to update deposit: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE balances SET amount = $2, updated_at = NOW() WHERE id = $1
	`, balance.ID, after); err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO balance_history (
			id, user_id, type, category, amount, balance_before, balance_after, reference_type, reference_id, description, created_at
		) VALUES ($1,$2,'credit','deposit',$3,$4,$5,'deposit',$6,$7,NOW())
	`, "bh_"+uuid.New().String()[:8], deposit.UserID, deposit.Amount, before, after, deposit.ID, "Approval deposit admin"); err != nil {
		return fmt.Errorf("failed to insert balance history: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit deposit approval: %w", err)
	}

	_ = s.logAudit(ctx, actorID, "deposit.approve", "deposit", depositID, map[string]interface{}{
		"status": domain.DepositStatusPending,
	}, map[string]interface{}{
		"status": domain.DepositStatusSuccess,
	}, "", "", "success", nil)
	return nil
}

func (s *AdminService) RejectDeposit(ctx context.Context, actorID, depositID string) error {
	if _, err := s.repo.DB().ExecContext(ctx, `
		UPDATE deposits SET status = 'failed', updated_at = NOW() WHERE id = $1 AND status = 'pending'
	`, depositID); err != nil {
		return fmt.Errorf("failed to reject deposit: %w", err)
	}
	_ = s.logAudit(ctx, actorID, "deposit.reject", "deposit", depositID, nil, map[string]interface{}{
		"status": domain.DepositStatusFailed,
	}, "", "", "success", nil)
	return nil
}

func (s *AdminService) ListQris(ctx context.Context, search string, page, perPage int) (*domain.AdminListResponse, error) {
	items, total, err := s.repo.ListQrisIncomes(ctx, search, page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(items, total, page, perPage), nil
}

func (s *AdminService) GetQrisDetail(ctx context.Context, qrisID string) (map[string]interface{}, error) {
	item, err := s.repo.GetQrisIncomeDetail(ctx, qrisID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, domain.ErrNotFound("QRIS")
	}
	return item, nil
}

func (s *AdminService) ListVouchers(ctx context.Context, search string, page, perPage int) (*domain.AdminListResponse, error) {
	items, total, err := s.repo.ListVouchers(ctx, search, page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(items, total, page, perPage), nil
}

func (s *AdminService) GetVoucherDetail(ctx context.Context, voucherID string) (map[string]interface{}, error) {
	item, err := s.repo.GetVoucherDetail(ctx, voucherID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, domain.ErrNotFound("Voucher")
	}
	return item, nil
}

func (s *AdminService) CreateVoucher(ctx context.Context, actorID string, payload map[string]interface{}) error {
	payload["id"] = "vch_" + uuid.New().String()[:8]
	if err := s.repo.CreateVoucher(ctx, payload); err != nil {
		return fmt.Errorf("failed to create voucher: %w", err)
	}
	_ = s.logAudit(ctx, actorID, "voucher.create", "voucher", fmt.Sprintf("%v", payload["id"]), nil, payload, "", "", "success", nil)
	return nil
}

func (s *AdminService) UpdateVoucher(ctx context.Context, actorID, voucherID string, payload map[string]interface{}) error {
	if err := s.repo.UpdateVoucher(ctx, voucherID, payload); err != nil {
		return fmt.Errorf("failed to update voucher: %w", err)
	}
	_ = s.logAudit(ctx, actorID, "voucher.update", "voucher", voucherID, nil, payload, "", "", "success", nil)
	return nil
}

func (s *AdminService) UpdateVoucherStatus(ctx context.Context, actorID, voucherID string, isActive bool) error {
	if err := s.repo.UpdateVoucherStatus(ctx, voucherID, isActive); err != nil {
		return fmt.Errorf("failed to update voucher status: %w", err)
	}
	_ = s.logAudit(ctx, actorID, "voucher.status.update", "voucher", voucherID, nil, map[string]interface{}{
		"isActive": isActive,
	}, "", "", "success", nil)
	return nil
}

func (s *AdminService) ListCatalog(ctx context.Context, search string, page, perPage int) (map[string]interface{}, error) {
	services, err := s.repo.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	products, total, err := s.repo.ListProducts(ctx, search, page, perPage)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"services": services,
		"products": paginated(products, total, page, perPage),
	}, nil
}

func (s *AdminService) CreatePricingRequest(ctx context.Context, actorID string, payload map[string]interface{}, reason string) error {
	productID := stringValue(payload["productId"])
	if productID == "" {
		return domain.ErrValidationFailed("productId wajib diisi")
	}
	product, err := s.repo.FindProductByID(ctx, productID)
	if err != nil {
		return err
	}
	if product == nil {
		return domain.ErrNotFound("Produk")
	}

	req := &domain.AdminApprovalRequest{
		ID:           "apr_" + uuid.New().String()[:8],
		RequesterID:  actorID,
		RequestType:  "price_change",
		ResourceType: "product",
		ResourceID:   sqlNullString(productID),
		Reason:       sqlNullString(reason),
		Payload: map[string]interface{}{
			"productId":   productID,
			"productName": product["name"],
			"oldPrice":    product["price"],
			"oldAdminFee": product["admin_fee"],
			"newPrice":    payload["newPrice"],
			"newAdminFee": payload["newAdminFee"],
		},
		Status:    domain.ApprovalStatusPending,
		CreatedAt: time.Now(),
	}
	if err := s.repo.CreateApprovalRequest(ctx, req); err != nil {
		return err
	}
	_ = s.repo.AddApprovalEvent(ctx, req.ID, actorID, "created", reason)
	_ = s.logAudit(ctx, actorID, "pricing.request.create", "approval_request", req.ID, nil, req.Payload, "", "", "success", nil)
	return nil
}

func (s *AdminService) CreateBalanceAdjustmentRequest(ctx context.Context, actorID string, payload map[string]interface{}, reason string) error {
	userID := stringValue(payload["userId"])
	if userID == "" {
		return domain.ErrValidationFailed("userId wajib diisi")
	}
	req := &domain.AdminApprovalRequest{
		ID:           "apr_" + uuid.New().String()[:8],
		RequesterID:  actorID,
		RequestType:  "balance_adjustment",
		ResourceType: "balance",
		ResourceID:   sqlNullString(userID),
		Reason:       sqlNullString(reason),
		Payload:      payload,
		Status:       domain.ApprovalStatusPending,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.CreateApprovalRequest(ctx, req); err != nil {
		return err
	}
	_ = s.repo.AddApprovalEvent(ctx, req.ID, actorID, "created", reason)
	_ = s.logAudit(ctx, actorID, "balance.adjustment.request", "approval_request", req.ID, nil, payload, "", "", "success", nil)
	return nil
}

func (s *AdminService) ListKYC(ctx context.Context, search, status string, page, perPage int) (*domain.AdminListResponse, error) {
	items, total, err := s.repo.ListKYC(ctx, search, status, page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(items, total, page, perPage), nil
}

func (s *AdminService) GetKYCDetail(ctx context.Context, userID string) (map[string]interface{}, error) {
	item, err := s.repo.GetKYCDetail(ctx, userID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, domain.ErrNotFound("KYC")
	}
	return item, nil
}

func (s *AdminService) ApproveKYC(ctx context.Context, actorID, userID string) error {
	ok, err := s.repo.HasKYCVerification(ctx, userID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrValidationFailed("Data verifikasi KYC belum tersedia")
	}
	if err := s.repo.UpdateUserKYCStatus(ctx, userID, domain.KYCStatusVerified); err != nil {
		return fmt.Errorf("failed to approve kyc: %w", err)
	}
	_ = s.logAudit(ctx, actorID, "kyc.approve", "user", userID, nil, map[string]interface{}{"kycStatus": domain.KYCStatusVerified}, "", "", "success", nil)
	return nil
}

func (s *AdminService) RejectKYC(ctx context.Context, actorID, userID string) error {
	if err := s.repo.UpdateUserKYCStatus(ctx, userID, domain.KYCStatusRejected); err != nil {
		return fmt.Errorf("failed to reject kyc: %w", err)
	}
	_ = s.logAudit(ctx, actorID, "kyc.reject", "user", userID, nil, map[string]interface{}{"kycStatus": domain.KYCStatusRejected}, "", "", "success", nil)
	return nil
}

func (s *AdminService) ListBanners(ctx context.Context) ([]map[string]interface{}, error) {
	return s.repo.ListBanners(ctx)
}

func (s *AdminService) CreateBanner(ctx context.Context, actorID string, payload map[string]interface{}) error {
	payload["id"] = "banner_" + uuid.New().String()[:8]
	if err := s.repo.CreateBanner(ctx, payload); err != nil {
		return err
	}
	_ = s.logAudit(ctx, actorID, "banner.create", "banner", fmt.Sprintf("%v", payload["id"]), nil, payload, "", "", "success", nil)
	return nil
}

func (s *AdminService) UpdateBanner(ctx context.Context, actorID, bannerID string, payload map[string]interface{}) error {
	if err := s.repo.UpdateBanner(ctx, bannerID, payload); err != nil {
		return err
	}
	_ = s.logAudit(ctx, actorID, "banner.update", "banner", bannerID, nil, payload, "", "", "success", nil)
	return nil
}

func (s *AdminService) DeleteBanner(ctx context.Context, actorID, bannerID string) error {
	if err := s.repo.DeleteBanner(ctx, bannerID); err != nil {
		return err
	}
	_ = s.logAudit(ctx, actorID, "banner.delete", "banner", bannerID, nil, nil, "", "", "success", nil)
	return nil
}

func (s *AdminService) ListCatalogServices(ctx context.Context) ([]map[string]interface{}, error) {
	return s.repo.ListServices(ctx)
}

func (s *AdminService) UpdateCatalogService(ctx context.Context, actorID, serviceID string, payload map[string]interface{}) error {
	if err := s.repo.UpdateService(ctx, serviceID, payload); err != nil {
		return err
	}
	_ = s.logAudit(ctx, actorID, "service.update", "service", serviceID, nil, payload, "", "", "success", nil)
	return nil
}

func (s *AdminService) UploadServiceIcon(ctx context.Context, actorID, serviceID string, file *multipart.FileHeader) (string, error) {
	if s.publicS3 == nil {
		return "", domain.NewError("UPLOAD_DISABLED", "Public S3 not configured", 500)
	}
	iconURL, err := s.publicS3.UploadFile(ctx, file, fmt.Sprintf("services/%s", serviceID))
	if err != nil {
		return "", fmt.Errorf("failed to upload service icon: %w", err)
	}
	if err := s.repo.UpdateServiceIconURL(ctx, serviceID, iconURL); err != nil {
		return "", fmt.Errorf("failed to update icon url: %w", err)
	}
	_ = s.logAudit(ctx, actorID, "service.upload_icon", "service", serviceID, nil, map[string]interface{}{"iconUrl": iconURL}, "", "", "success", nil)
	return iconURL, nil
}

func (s *AdminService) ListNotifications(ctx context.Context, search string, page, perPage int) (*domain.AdminListResponse, error) {
	items, total, err := s.repo.ListNotifications(ctx, search, page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(items, total, page, perPage), nil
}

func (s *AdminService) BroadcastNotification(ctx context.Context, actorID string, payload map[string]interface{}) error {
	targetMode := firstNonEmpty(stringValue(payload["targetMode"]), "all")
	targetIDs := stringSliceValue(payload["targetUserIds"])
	targets, err := s.repo.FindNotificationTargetUserIDs(ctx, targetMode, targetIDs)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return domain.ErrValidationFailed("Target notifikasi tidak ditemukan")
	}
	if err := s.repo.BroadcastNotification(ctx, targets, payload); err != nil {
		return err
	}
	_ = s.logAudit(ctx, actorID, "notification.broadcast", "notification", "", nil, map[string]interface{}{
		"targetMode": targetMode,
		"count":      len(targets),
		"title":      payload["title"],
	}, "", "", "success", nil)
	return nil
}

func (s *AdminService) ListApprovals(ctx context.Context, status string, page, perPage int) (*domain.AdminListResponse, error) {
	items, total, err := s.repo.ListApprovalRequests(ctx, status, page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(items, total, page, perPage), nil
}

func (s *AdminService) ApproveApproval(ctx context.Context, actorID, approvalID string) error {
	req, err := s.repo.FindApprovalRequestByID(ctx, approvalID)
	if err != nil {
		return err
	}
	if req == nil {
		return domain.ErrNotFound("Approval")
	}
	if req.Status != domain.ApprovalStatusPending {
		return domain.ErrValidationFailed("Request approval tidak lagi pending")
	}
	if req.RequesterID == actorID {
		return domain.NewError("APPROVAL_SELF_FORBIDDEN", "Maker tidak boleh approve request sendiri", 403)
	}

	if err := s.repo.UpdateApprovalDecision(ctx, approvalID, actorID, domain.ApprovalStatusApproved, nil); err != nil {
		return err
	}
	_ = s.repo.AddApprovalEvent(ctx, approvalID, actorID, "approved", "")

	switch req.RequestType {
	case "price_change":
		payload := mapValue(req.Payload)
		productID := stringValue(payload["productId"])
		if err := s.repo.UpdateProductPricing(ctx, productID, int64Value(payload["newPrice"]), int64Value(payload["newAdminFee"])); err != nil {
			return err
		}
	case "balance_adjustment":
		if err := s.applyBalanceAdjustment(ctx, approvalID, actorID, mapValue(req.Payload)); err != nil {
			return err
		}
	}

	if err := s.repo.MarkApprovalApplied(ctx, approvalID); err != nil {
		return err
	}
	_ = s.repo.AddApprovalEvent(ctx, approvalID, actorID, "applied", "")
	_ = s.logAudit(ctx, actorID, "approval.approve", "approval_request", approvalID, nil, map[string]interface{}{
		"requestType": req.RequestType,
	}, "", "", "success", nil)
	return nil
}

func (s *AdminService) RejectApproval(ctx context.Context, actorID, approvalID, reason string) error {
	req, err := s.repo.FindApprovalRequestByID(ctx, approvalID)
	if err != nil {
		return err
	}
	if req == nil {
		return domain.ErrNotFound("Approval")
	}
	if req.Status != domain.ApprovalStatusPending {
		return domain.ErrValidationFailed("Request approval tidak lagi pending")
	}
	if req.RequesterID == actorID {
		return domain.NewError("APPROVAL_SELF_FORBIDDEN", "Maker tidak boleh reject request sendiri", 403)
	}
	if err := s.repo.UpdateApprovalDecision(ctx, approvalID, actorID, domain.ApprovalStatusRejected, &reason); err != nil {
		return err
	}
	_ = s.repo.AddApprovalEvent(ctx, approvalID, actorID, "rejected", reason)
	_ = s.logAudit(ctx, actorID, "approval.reject", "approval_request", approvalID, nil, map[string]interface{}{
		"reason": reason,
	}, "", "", "success", nil)
	return nil
}

func (s *AdminService) ListAuditLogs(ctx context.Context, page, perPage int) (*domain.AdminListResponse, error) {
	items, total, err := s.repo.ListAuditLogs(ctx, page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(items, total, page, perPage), nil
}

func (s *AdminService) ListSettings(ctx context.Context) ([]map[string]interface{}, error) {
	return s.repo.ListSettings(ctx)
}

func (s *AdminService) UpsertSetting(ctx context.Context, actorID, key, description string, value interface{}) error {
	var updatedBy *string
	if actorID != "" {
		updatedBy = &actorID
	}
	if err := s.repo.UpsertSetting(ctx, key, value, description, updatedBy); err != nil {
		return err
	}
	_ = s.logAudit(ctx, actorID, "setting.upsert", "admin_setting", key, nil, value, "", "", "success", nil)
	return nil
}

func (s *AdminService) ListReferenceData(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.ListReferenceData(ctx)
}

func (s *AdminService) applyBalanceAdjustment(ctx context.Context, approvalID, actorID string, payload map[string]interface{}) error {
	userID := stringValue(payload["userId"])
	amountDelta := int64Value(payload["amountDelta"])
	description := firstNonEmpty(stringValue(payload["description"]), "Manual balance adjustment")
	if userID == "" || amountDelta == 0 {
		return domain.ErrValidationFailed("Payload koreksi saldo tidak valid")
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var balance struct {
		ID     string `db:"id"`
		UserID string `db:"user_id"`
		Amount int64  `db:"amount"`
	}
	if err := tx.GetContext(ctx, &balance, `SELECT id, user_id, amount FROM balances WHERE user_id = $1 FOR UPDATE`, userID); err != nil {
		return err
	}

	before := balance.Amount
	after := before + amountDelta
	if after < 0 {
		return domain.ErrValidationFailed("Koreksi saldo tidak boleh membuat saldo negatif")
	}

	if _, err := tx.ExecContext(ctx, `UPDATE balances SET amount = $2, updated_at = NOW() WHERE id = $1`, balance.ID, after); err != nil {
		return err
	}

	entryType := "credit"
	if amountDelta < 0 {
		entryType = "debit"
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO balance_history (
			id, user_id, type, category, amount, balance_before, balance_after, reference_type, reference_id, description, created_at
		) VALUES ($1,$2,$3,'bonus',$4,$5,$6,'approval',$7,$8,NOW())
	`, "bh_"+uuid.New().String()[:8], userID, entryType, absInt64(amountDelta), before, after, approvalID, description); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	_ = s.logAudit(ctx, actorID, "balance.adjustment.applied", "approval_request", approvalID, nil, payload, "", "", "success", nil)
	return nil
}

func paginated(items interface{}, total, page, perPage int) *domain.AdminListResponse {
	safePerPage := sanitizePerPage(perPage)
	safePage := page
	if safePage <= 0 {
		safePage = 1
	}
	return &domain.AdminListResponse{
		Items:   items,
		Page:    safePage,
		PerPage: safePerPage,
		Total:   total,
		HasNext: safePage*safePerPage < total,
	}
}

func sanitizePerPage(perPage int) int {
	if perPage <= 0 {
		return 20
	}
	if perPage > 100 {
		return 100
	}
	return perPage
}

func stringValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func int64Value(value interface{}) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case json.Number:
		n, _ := v.Int64()
		return n
	case string:
		var n int64
		fmt.Sscan(v, &n)
		return n
	default:
		return 0
	}
}

func stringSliceValue(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if text := stringValue(item); text != "" {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

func mapValue(value interface{}) map[string]interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return v
	default:
		return map[string]interface{}{}
	}
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
