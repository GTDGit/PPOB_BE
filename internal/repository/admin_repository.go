package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type AdminRepository struct {
	db *sqlx.DB
}

func NewAdminRepository(db *sqlx.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) DB() *sqlx.DB {
	return r.db
}

func (r *AdminRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

const adminUserColumns = `
	id, email, phone, full_name, password_hash, status, is_active,
	last_login_at, invited_by, created_by, created_at, updated_at
`

func (r *AdminRepository) FindAdminByEmail(ctx context.Context, email string) (*domain.AdminUser, error) {
	var admin domain.AdminUser
	err := r.db.GetContext(ctx, &admin, `SELECT `+adminUserColumns+` FROM admin_users WHERE LOWER(email) = LOWER($1) LIMIT 1`, email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := r.loadAdminRolesAndPermissions(ctx, &admin); err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) FindAdminByID(ctx context.Context, id string) (*domain.AdminUser, error) {
	var admin domain.AdminUser
	err := r.db.GetContext(ctx, &admin, `SELECT `+adminUserColumns+` FROM admin_users WHERE id = $1 LIMIT 1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := r.loadAdminRolesAndPermissions(ctx, &admin); err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepository) CountAdmins(ctx context.Context) (int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM admin_users`); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *AdminRepository) CreateAdminUser(ctx context.Context, admin *domain.AdminUser, roleID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO admin_users (
			id, email, phone, full_name, password_hash, status, is_active,
			invited_by, created_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`,
		admin.ID,
		admin.Email,
		admin.Phone,
		admin.FullName,
		admin.PasswordHash,
		admin.Status,
		admin.IsActive,
		admin.InvitedBy,
		admin.CreatedBy,
		admin.CreatedAt,
		admin.UpdatedAt,
	); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO admin_user_roles (admin_user_id, role_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, admin.ID, roleID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AdminRepository) UpdateAdminPasswordAndStatus(ctx context.Context, adminID, passwordHash, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_users
		SET password_hash = $2, status = $3, updated_at = NOW()
		WHERE id = $1
	`, adminID, passwordHash, status)
	return err
}

func (r *AdminRepository) UpdateAdminStatus(ctx context.Context, adminID, status string, isActive bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_users
		SET status = $2, is_active = $3, updated_at = NOW()
		WHERE id = $1
	`, adminID, status, isActive)
	return err
}

func (r *AdminRepository) UpdateLastLogin(ctx context.Context, adminID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_users SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1
	`, adminID)
	return err
}

func (r *AdminRepository) ListRoles(ctx context.Context) ([]domain.AdminRole, error) {
	var roles []domain.AdminRole
	if err := r.db.SelectContext(ctx, &roles, `
		SELECT id, name, description, is_system, is_active, created_at, updated_at
		FROM admin_roles
		ORDER BY name ASC
	`); err != nil {
		return nil, err
	}

	for i := range roles {
		permissions, err := r.GetRolePermissionKeys(ctx, roles[i].ID)
		if err != nil {
			return nil, err
		}
		roles[i].Permissions = permissions
	}

	return roles, nil
}

func (r *AdminRepository) FindRoleByID(ctx context.Context, id string) (*domain.AdminRole, error) {
	var role domain.AdminRole
	err := r.db.GetContext(ctx, &role, `
		SELECT id, name, description, is_system, is_active, created_at, updated_at
		FROM admin_roles WHERE id = $1
	`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	role.Permissions, err = r.GetRolePermissionKeys(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *AdminRepository) ListPermissions(ctx context.Context) ([]domain.AdminPermission, error) {
	var permissions []domain.AdminPermission
	err := r.db.SelectContext(ctx, &permissions, `
		SELECT key, module, action, description, created_at
		FROM admin_permissions
		ORDER BY module ASC, action ASC
	`)
	return permissions, err
}

func (r *AdminRepository) GetRolePermissionKeys(ctx context.Context, roleID string) ([]string, error) {
	var permissions []string
	err := r.db.SelectContext(ctx, &permissions, `
		SELECT permission_key
		FROM admin_role_permissions
		WHERE role_id = $1
		ORDER BY permission_key ASC
	`, roleID)
	return permissions, err
}

func (r *AdminRepository) loadAdminRolesAndPermissions(ctx context.Context, admin *domain.AdminUser) error {
	var roles []domain.AdminRole
	if err := r.db.SelectContext(ctx, &roles, `
		SELECT ar.id, ar.name, ar.description, ar.is_system, ar.is_active, ar.created_at, ar.updated_at
		FROM admin_roles ar
		INNER JOIN admin_user_roles aur ON aur.role_id = ar.id
		WHERE aur.admin_user_id = $1
		ORDER BY ar.name ASC
	`, admin.ID); err != nil {
		return err
	}

	var permissions []string
	if err := r.db.SelectContext(ctx, &permissions, `
		SELECT DISTINCT arp.permission_key
		FROM admin_role_permissions arp
		INNER JOIN admin_user_roles aur ON aur.role_id = arp.role_id
		WHERE aur.admin_user_id = $1
		ORDER BY arp.permission_key ASC
	`, admin.ID); err != nil {
		return err
	}

	admin.Roles = roles
	admin.Permissions = permissions
	return nil
}

func (r *AdminRepository) CreateInvite(ctx context.Context, invite *domain.AdminInvite) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_invites (
			id, email, phone, full_name, role_id, token_hash, invited_by, admin_user_id, expires_at, accepted_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`,
		invite.ID,
		invite.Email,
		invite.Phone,
		invite.FullName,
		invite.RoleID,
		invite.TokenHash,
		invite.InvitedBy,
		invite.AdminUserID,
		invite.ExpiresAt,
		invite.AcceptedAt,
		invite.CreatedAt,
	)
	return err
}

func (r *AdminRepository) FindInviteByTokenHash(ctx context.Context, tokenHash string) (*domain.AdminInvite, error) {
	var invite domain.AdminInvite
	err := r.db.GetContext(ctx, &invite, `
		SELECT id, email, phone, full_name, role_id, token_hash, invited_by, admin_user_id, expires_at, accepted_at, created_at
		FROM admin_invites
		WHERE token_hash = $1
		LIMIT 1
	`, tokenHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *AdminRepository) LinkInviteToAdmin(ctx context.Context, inviteID, adminUserID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_invites
		SET admin_user_id = $2
		WHERE id = $1
	`, inviteID, adminUserID)
	return err
}

func (r *AdminRepository) MarkInviteAccepted(ctx context.Context, inviteID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_invites
		SET accepted_at = NOW()
		WHERE id = $1
	`, inviteID)
	return err
}

func (r *AdminRepository) UpsertTOTPSecret(ctx context.Context, adminUserID, secret string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_totp_secrets (admin_user_id, secret, confirmed_at, created_at, updated_at)
		VALUES ($1, $2, NULL, NOW(), NOW())
		ON CONFLICT (admin_user_id)
		DO UPDATE SET secret = EXCLUDED.secret, updated_at = NOW()
	`, adminUserID, secret)
	return err
}

func (r *AdminRepository) FindTOTPSecretByAdminID(ctx context.Context, adminUserID string) (*domain.AdminTOTPSecret, error) {
	var secret domain.AdminTOTPSecret
	err := r.db.GetContext(ctx, &secret, `
		SELECT admin_user_id, secret, confirmed_at, created_at, updated_at
		FROM admin_totp_secrets
		WHERE admin_user_id = $1
	`, adminUserID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &secret, nil
}

func (r *AdminRepository) ConfirmTOTP(ctx context.Context, adminUserID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE admin_totp_secrets
		SET confirmed_at = NOW(), updated_at = NOW()
		WHERE admin_user_id = $1
	`, adminUserID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE admin_users
		SET status = $2, is_active = TRUE, updated_at = NOW()
		WHERE id = $1
	`, adminUserID, domain.AdminStatusActive); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *AdminRepository) ReplaceRecoveryCodes(ctx context.Context, adminUserID string, hashedCodes []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM admin_recovery_codes WHERE admin_user_id = $1`, adminUserID); err != nil {
		return err
	}

	for _, codeHash := range hashedCodes {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO admin_recovery_codes (id, admin_user_id, code_hash, created_at)
			VALUES ($1, $2, $3, NOW())
		`, "arc_"+uuid.New().String()[:8], adminUserID, codeHash); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *AdminRepository) CreateSession(ctx context.Context, session *domain.AdminSession) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_sessions (
			id, admin_user_id, refresh_token_hash, ip_address, user_agent, expires_at, created_at, last_used_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`,
		session.ID,
		session.AdminUserID,
		session.RefreshTokenHash,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.CreatedAt,
		session.LastUsedAt,
	)
	return err
}

func (r *AdminRepository) FindSessionByID(ctx context.Context, sessionID string) (*domain.AdminSession, error) {
	var session domain.AdminSession
	err := r.db.GetContext(ctx, &session, `
		SELECT id, admin_user_id, refresh_token_hash, ip_address, user_agent, expires_at, created_at, last_used_at
		FROM admin_sessions
		WHERE id = $1
	`, sessionID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *AdminRepository) UpdateSessionRefreshToken(ctx context.Context, sessionID, refreshTokenHash string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_sessions
		SET refresh_token_hash = $2, expires_at = $3, last_used_at = NOW()
		WHERE id = $1
	`, sessionID, refreshTokenHash, expiresAt)
	return err
}

func (r *AdminRepository) TouchSession(ctx context.Context, sessionID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE admin_sessions SET last_used_at = NOW() WHERE id = $1`, sessionID)
	return err
}

func (r *AdminRepository) DeleteSession(ctx context.Context, sessionID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM admin_sessions WHERE id = $1`, sessionID)
	return err
}

func (r *AdminRepository) CreateAuditLog(ctx context.Context, log *domain.AdminAuditLog) error {
	oldValue, _ := json.Marshal(log.OldValue)
	newValue, _ := json.Marshal(log.NewValue)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_audit_logs (
			id, admin_user_id, action, resource_type, resource_id, old_value, new_value,
			ip_address, user_agent, status, error_message, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`,
		log.ID,
		log.AdminUserID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		oldValue,
		newValue,
		log.IPAddress,
		log.UserAgent,
		log.Status,
		log.ErrorMessage,
		log.CreatedAt,
	)
	return err
}

func (r *AdminRepository) count(ctx context.Context, query string, args ...interface{}) (int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, query, args...); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *AdminRepository) selectMaps(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		item := map[string]interface{}{}
		if err := rows.MapScan(item); err != nil {
			return nil, err
		}
		for key, value := range item {
			item[key] = normalizeDBValue(value)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func normalizeDBValue(value interface{}) interface{} {
	switch v := value.(type) {
	case []byte:
		text := string(v)
		trimmed := strings.TrimSpace(text)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			var decoded interface{}
			if err := json.Unmarshal(v, &decoded); err == nil {
				return decoded
			}
		}
		return text
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return value
	}
}

func buildSearchWhere(search string, startArg int, columns ...string) (string, []interface{}) {
	if strings.TrimSpace(search) == "" {
		return "", nil
	}

	searchValue := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
	conditions := make([]string, 0, len(columns))
	for _, col := range columns {
		conditions = append(conditions, fmt.Sprintf("LOWER(%s) LIKE $%d", col, startArg))
	}
	return " WHERE (" + strings.Join(conditions, " OR ") + ")", []interface{}{searchValue}
}

func sanitizePageSize(perPage int) int {
	if perPage <= 0 {
		return 20
	}
	if perPage > 100 {
		return 100
	}
	return perPage
}

func calculateOffset(page, perPage int) int {
	if page <= 1 {
		return 0
	}
	return (page - 1) * sanitizePageSize(perPage)
}

func (r *AdminRepository) GetDashboardSummary(ctx context.Context) (*domain.AdminDashboardResponse, error) {
	var resp domain.AdminDashboardResponse

	if err := r.db.GetContext(ctx, &resp.TotalUsers, `SELECT COUNT(*) FROM users`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &resp.ActiveAdmins, `SELECT COUNT(*) FROM admin_users WHERE status = 'active' AND is_active = true`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &resp.TransactionsToday, `
		SELECT COUNT(*) FROM transactions WHERE DATE(created_at AT TIME ZONE 'Asia/Jakarta') = CURRENT_DATE
	`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &resp.DepositsPending, `SELECT COUNT(*) FROM deposits WHERE status = 'pending'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &resp.PendingKYC, `SELECT COUNT(*) FROM users WHERE kyc_status = 'pending'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &resp.PendingApprovals, `SELECT COUNT(*) FROM admin_approval_requests WHERE status = 'pending'`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &resp.RevenueToday, `
		SELECT COALESCE(SUM(total_payment), 0) FROM transactions
		WHERE status = 'success' AND DATE(created_at AT TIME ZONE 'Asia/Jakarta') = CURRENT_DATE
	`); err != nil {
		return nil, err
	}
	if err := r.db.GetContext(ctx, &resp.DepositAmountToday, `
		SELECT COALESCE(SUM(total_amount), 0) FROM deposits
		WHERE status = 'success' AND DATE(COALESCE(paid_at, created_at) AT TIME ZONE 'Asia/Jakarta') = CURRENT_DATE
	`); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (r *AdminRepository) ListAdmins(ctx context.Context, search string, page, perPage int) ([]map[string]interface{}, int, error) {
	baseFrom := `
		FROM admin_users au
		LEFT JOIN admin_user_roles aur ON aur.admin_user_id = au.id
		LEFT JOIN admin_roles ar ON ar.id = aur.role_id
	`
	where, args := buildSearchWhere(search, 1, "COALESCE(au.full_name, '')", "au.email", "au.phone")
	countQuery := `SELECT COUNT(*) FROM admin_users au ` + where
	total, err := r.count(ctx, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	query := `
		SELECT
			au.id,
			au.email,
			au.phone,
			COALESCE(au.full_name, '') AS full_name,
			au.status,
			au.is_active,
			au.last_login_at,
			au.created_at,
			COALESCE(STRING_AGG(DISTINCT ar.name, ', '), '') AS roles
	` + baseFrom + where + `
		GROUP BY au.id
		ORDER BY au.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) ListCustomers(ctx context.Context, search string, page, perPage int) ([]map[string]interface{}, int, error) {
	base := `
		FROM users u
		LEFT JOIN balances b ON b.user_id = u.id
	`
	where, args := buildSearchWhere(search, 1, "COALESCE(u.full_name, '')", "u.phone", "COALESCE(u.email::text, '')", "u.mic")
	total, err := r.count(ctx, `SELECT COUNT(*) `+base+where, args...)
	if err != nil {
		return nil, 0, err
	}
	query := `
		SELECT
			u.id,
			u.mic,
			u.phone,
			COALESCE(u.full_name, '') AS full_name,
			COALESCE(u.email::text, '') AS email,
			u.tier,
			u.kyc_status,
			u.is_active,
			COALESCE(b.amount, 0) AS balance,
			u.created_at
	` + base + where + `
		ORDER BY u.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) ListTransactions(ctx context.Context, search, status string, page, perPage int) ([]map[string]interface{}, int, error) {
	base := `
		FROM transactions t
		INNER JOIN users u ON u.id = t.user_id
	`
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(`(
			LOWER(COALESCE(u.full_name, '')) LIKE $%d OR
			LOWER(u.phone) LIKE $%d OR
			LOWER(COALESCE(t.product_name, '')) LIKE $%d OR
			LOWER(t.target) LIKE $%d OR
			LOWER(t.id) LIKE $%d
		)`, argIdx, argIdx, argIdx, argIdx, argIdx))
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(search))+"%")
		argIdx++
	}
	if status != "" && status != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("t.status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	where := " WHERE " + strings.Join(whereClauses, " AND ")
	total, err := r.count(ctx, `SELECT COUNT(*) `+base+where, args...)
	if err != nil {
		return nil, 0, err
	}
	query := `
		SELECT
			t.id,
			t.type,
			t.service_type,
			t.target,
			COALESCE(t.product_name, '') AS product_name,
			t.total_payment,
			t.price,
			t.admin_fee,
			t.selling_price,
			t.selling_payment_type,
			t.status,
			COALESCE(t.status_message, '') AS status_message,
			t.reference_number,
			t.created_at,
			COALESCE(u.full_name, '') AS user_name,
			u.phone AS user_phone
	` + base + where + `
		ORDER BY t.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) ListDeposits(ctx context.Context, search, status string, page, perPage int) ([]map[string]interface{}, int, error) {
	base := `
		FROM deposits d
		INNER JOIN users u ON u.id = d.user_id
	`
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(`(
			LOWER(COALESCE(u.full_name, '')) LIKE $%d OR
			LOWER(u.phone) LIKE $%d OR
			LOWER(d.id) LIKE $%d OR
			LOWER(COALESCE(d.reference_number, '')) LIKE $%d
		)`, argIdx, argIdx, argIdx, argIdx))
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(search))+"%")
		argIdx++
	}
	if status != "" && status != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("d.status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	where := " WHERE " + strings.Join(whereClauses, " AND ")
	total, err := r.count(ctx, `SELECT COUNT(*) `+base+where, args...)
	if err != nil {
		return nil, 0, err
	}
	query := `
		SELECT
			d.id,
			d.user_id,
			d.method,
			d.amount,
			d.admin_fee,
			d.total_amount,
			d.status,
			d.reference_number,
			d.external_id,
			d.expires_at,
			d.paid_at,
			d.created_at,
			COALESCE(u.full_name, '') AS user_name,
			u.phone AS user_phone
	` + base + where + `
		ORDER BY d.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) FindDepositRow(ctx context.Context, depositID string) (map[string]interface{}, error) {
	items, err := r.selectMaps(ctx, `
		SELECT id, user_id, total_amount, status, method, reference_number, external_id
		FROM deposits WHERE id = $1 LIMIT 1
	`, depositID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}

func (r *AdminRepository) ListQrisIncomes(ctx context.Context, search string, page, perPage int) ([]map[string]interface{}, int, error) {
	base := `
		FROM qris_incomes qi
		INNER JOIN users u ON u.id = qi.user_id
	`
	where, args := buildSearchWhere(search, 1, "COALESCE(qi.merchant_name, '')", "COALESCE(qi.payer_name, '')", "u.phone", "qi.id")
	total, err := r.count(ctx, `SELECT COUNT(*) `+base+where, args...)
	if err != nil {
		return nil, 0, err
	}
	query := `
		SELECT
			qi.id,
			qi.user_id,
			qi.merchant_name,
			qi.payer_name,
			qi.payer_bank,
			qi.amount,
			qi.fee,
			qi.net_amount,
			qi.status,
			qi.rrn,
			qi.reference_number,
			qi.completed_at,
			qi.created_at,
			COALESCE(u.full_name, '') AS user_name,
			u.phone AS user_phone
	` + base + where + `
		ORDER BY qi.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) ListVouchers(ctx context.Context, search string, page, perPage int) ([]map[string]interface{}, int, error) {
	base := ` FROM vouchers v `
	where, args := buildSearchWhere(search, 1, "v.code", "v.name", "COALESCE(v.description, '')")
	total, err := r.count(ctx, `SELECT COUNT(*) `+base+where, args...)
	if err != nil {
		return nil, 0, err
	}
	query := `
		SELECT
			v.id, v.code, v.name, COALESCE(v.description, '') AS description,
			v.discount_type, v.discount_value, v.min_transaction, v.max_discount,
			v.max_usage, v.current_usage, v.max_usage_per_user, v.is_active,
			v.starts_at, v.expires_at, v.created_at, v.updated_at
	` + base + where + `
		ORDER BY v.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) CreateVoucher(ctx context.Context, payload map[string]interface{}) error {
	applicableServices, _ := json.Marshal(payload["applicableServices"])
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO vouchers (
			id, code, name, description, discount_type, discount_value,
			min_transaction, max_discount, applicable_services, max_usage,
			max_usage_per_user, current_usage, terms_url, starts_at, expires_at,
			is_active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,0,$12,$13,$14,$15,NOW(),NOW())
	`,
		payload["id"], payload["code"], payload["name"], payload["description"],
		payload["discountType"], payload["discountValue"], payload["minTransaction"], payload["maxDiscount"],
		string(applicableServices), payload["maxUsage"], payload["maxUsagePerUser"], payload["termsUrl"],
		payload["startsAt"], payload["expiresAt"], payload["isActive"],
	)
	return err
}

func (r *AdminRepository) UpdateVoucher(ctx context.Context, voucherID string, payload map[string]interface{}) error {
	applicableServices, _ := json.Marshal(payload["applicableServices"])
	_, err := r.db.ExecContext(ctx, `
		UPDATE vouchers
		SET code = $2, name = $3, description = $4, discount_type = $5, discount_value = $6,
			min_transaction = $7, max_discount = $8, applicable_services = $9, max_usage = $10,
			max_usage_per_user = $11, terms_url = $12, starts_at = $13, expires_at = $14,
			is_active = $15, updated_at = NOW()
		WHERE id = $1
	`,
		voucherID, payload["code"], payload["name"], payload["description"],
		payload["discountType"], payload["discountValue"], payload["minTransaction"], payload["maxDiscount"],
		string(applicableServices), payload["maxUsage"], payload["maxUsagePerUser"], payload["termsUrl"],
		payload["startsAt"], payload["expiresAt"], payload["isActive"],
	)
	return err
}

func (r *AdminRepository) UpdateVoucherStatus(ctx context.Context, voucherID string, isActive bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE vouchers SET is_active = $2, updated_at = NOW() WHERE id = $1`, voucherID, isActive)
	return err
}

func (r *AdminRepository) ListServices(ctx context.Context) ([]map[string]interface{}, error) {
	return r.selectMaps(ctx, `
		SELECT s.id, s.name, s.route, s.status, s.badge, s.is_featured, s.sort_order,
		       COALESCE(sc.name, '') AS category_name
		FROM services s
		LEFT JOIN service_categories sc ON sc.id = s.category_id
		ORDER BY s.sort_order ASC, s.name ASC
	`)
}

func (r *AdminRepository) ListProducts(ctx context.Context, search string, page, perPage int) ([]map[string]interface{}, int, error) {
	base := ` FROM products p `
	where, args := buildSearchWhere(search, 1, "p.name", "COALESCE(p.description, '')", "p.service_type", "COALESCE(p.category, '')")
	total, err := r.count(ctx, `SELECT COUNT(*) `+base+where, args...)
	if err != nil {
		return nil, 0, err
	}
	query := `
		SELECT
			p.id, p.service_type, p.name, COALESCE(p.description, '') AS description,
			COALESCE(p.category, '') AS category, p.nominal, p.price, p.admin_fee,
			p.status, p.stock, p.sort_order, p.created_at, p.updated_at
	` + base + where + `
		ORDER BY p.updated_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) FindProductByID(ctx context.Context, productID string) (map[string]interface{}, error) {
	items, err := r.selectMaps(ctx, `
		SELECT id, name, service_type, price, admin_fee, status
		FROM products
		WHERE id = $1 LIMIT 1
	`, productID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}

func (r *AdminRepository) UpdateProductPricing(ctx context.Context, productID string, price, adminFee int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE products
		SET price = $2, admin_fee = $3, updated_at = NOW()
		WHERE id = $1
	`, productID, price, adminFee)
	return err
}

func (r *AdminRepository) ListKYC(ctx context.Context, search, status string, page, perPage int) ([]map[string]interface{}, int, error) {
	base := `
		FROM users u
		LEFT JOIN kyc_verifications kv ON kv.user_id = u.id
	`
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	argIdx := 1
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(`(
			LOWER(COALESCE(u.full_name, '')) LIKE $%d OR
			LOWER(u.phone) LIKE $%d OR
			LOWER(COALESCE(u.email::text, '')) LIKE $%d OR
			LOWER(COALESCE(kv.nik, '')) LIKE $%d
		)`, argIdx, argIdx, argIdx, argIdx))
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(search))+"%")
		argIdx++
	}
	if status != "" && status != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("u.kyc_status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	where := " WHERE " + strings.Join(whereClauses, " AND ")
	total, err := r.count(ctx, `SELECT COUNT(*) `+base+where, args...)
	if err != nil {
		return nil, 0, err
	}
	query := `
		SELECT
			u.id AS user_id,
			COALESCE(u.full_name, '') AS full_name,
			u.phone,
			COALESCE(u.email::text, '') AS email,
			u.kyc_status,
			COALESCE(kv.nik, '') AS nik,
			COALESCE(kv.full_name, '') AS kyc_full_name,
			kv.ktp_url,
			kv.face_url,
			kv.face_with_ktp_url,
			kv.liveness_url,
			kv.verified_at,
			u.updated_at
	` + base + where + `
		ORDER BY u.updated_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) UpdateUserKYCStatus(ctx context.Context, userID, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET kyc_status = $2, updated_at = NOW() WHERE id = $1
	`, userID, status)
	return err
}

func (r *AdminRepository) HasKYCVerification(ctx context.Context, userID string) (bool, error) {
	var count int
	if err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM kyc_verifications WHERE user_id = $1`, userID); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AdminRepository) ListBanners(ctx context.Context) ([]map[string]interface{}, error) {
	return r.selectMaps(ctx, `
		SELECT id, title, subtitle, image_url, thumbnail_url, action_type, action_value,
		       background_color, text_color, placement, start_date, end_date, priority,
		       target_tiers, is_new_user_only, is_active, created_at, updated_at
		FROM banners
		ORDER BY priority ASC, created_at DESC
	`)
}

func (r *AdminRepository) CreateBanner(ctx context.Context, payload map[string]interface{}) error {
	targetTiers, _ := json.Marshal(payload["targetTiers"])
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO banners (
			id, title, subtitle, image_url, thumbnail_url, action_type, action_value, background_color,
			text_color, placement, start_date, end_date, priority, target_tiers, is_new_user_only, is_active, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,NOW(),NOW())
	`,
		payload["id"], payload["title"], payload["subtitle"], payload["imageUrl"], payload["thumbnailUrl"],
		payload["actionType"], payload["actionValue"], payload["backgroundColor"], payload["textColor"],
		payload["placement"], payload["startDate"], payload["endDate"], payload["priority"], string(targetTiers),
		payload["isNewUserOnly"], payload["isActive"],
	)
	return err
}

func (r *AdminRepository) UpdateBanner(ctx context.Context, bannerID string, payload map[string]interface{}) error {
	targetTiers, _ := json.Marshal(payload["targetTiers"])
	_, err := r.db.ExecContext(ctx, `
		UPDATE banners
		SET title = $2, subtitle = $3, image_url = $4, thumbnail_url = $5, action_type = $6,
			action_value = $7, background_color = $8, text_color = $9, placement = $10,
			start_date = $11, end_date = $12, priority = $13, target_tiers = $14,
			is_new_user_only = $15, is_active = $16, updated_at = NOW()
		WHERE id = $1
	`,
		bannerID, payload["title"], payload["subtitle"], payload["imageUrl"], payload["thumbnailUrl"],
		payload["actionType"], payload["actionValue"], payload["backgroundColor"], payload["textColor"],
		payload["placement"], payload["startDate"], payload["endDate"], payload["priority"], string(targetTiers),
		payload["isNewUserOnly"], payload["isActive"],
	)
	return err
}

func (r *AdminRepository) DeleteBanner(ctx context.Context, bannerID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM banners WHERE id = $1`, bannerID)
	return err
}

func (r *AdminRepository) ListNotifications(ctx context.Context, search string, page, perPage int) ([]map[string]interface{}, int, error) {
	base := `
		FROM notifications n
		LEFT JOIN users u ON u.id = n.user_id
	`
	where, args := buildSearchWhere(search, 1, "n.title", "n.body", "COALESCE(u.full_name, '')", "u.phone")
	total, err := r.count(ctx, `SELECT COUNT(*) `+base+where, args...)
	if err != nil {
		return nil, 0, err
	}
	query := `
		SELECT
			n.id, n.user_id, n.category, n.title, n.body, n.short_body, n.action_type,
			n.action_value, n.action_button_text, n.image_url, n.is_read, n.read_at, n.created_at,
			COALESCE(u.full_name, '') AS user_name, COALESCE(u.phone, '') AS user_phone
	` + base + where + `
		ORDER BY n.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) BroadcastNotification(ctx context.Context, targets []string, payload map[string]interface{}) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, userID := range targets {
		metadata, _ := json.Marshal(payload["metadata"])
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO notifications (
				id, user_id, category, title, body, short_body, image_url, action_type,
				action_value, action_button_text, metadata, is_read, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,false,NOW())
		`,
			"notif_"+uuid.New().String()[:8], userID, payload["category"], payload["title"], payload["body"],
			payload["shortBody"], payload["imageUrl"], payload["actionType"], payload["actionValue"],
			payload["actionButtonText"], metadata,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *AdminRepository) FindNotificationTargetUserIDs(ctx context.Context, target string, userIDs []string) ([]string, error) {
	var rows *sqlx.Rows
	var err error
	switch target {
	case "selected":
		if len(userIDs) == 0 {
			return []string{}, nil
		}
		query, args, inErr := sqlx.In(`SELECT id FROM users WHERE id IN (?) AND is_active = true`, userIDs)
		if inErr != nil {
			return nil, inErr
		}
		query = r.db.Rebind(query)
		rows, err = r.db.QueryxContext(ctx, query, args...)
	default:
		rows, err = r.db.QueryxContext(ctx, `SELECT id FROM users WHERE is_active = true`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		targets = append(targets, id)
	}
	return targets, rows.Err()
}

func (r *AdminRepository) ListApprovalRequests(ctx context.Context, status string, page, perPage int) ([]map[string]interface{}, int, error) {
	base := `
		FROM admin_approval_requests aar
		LEFT JOIN admin_users requester ON requester.id = aar.requester_id
		LEFT JOIN admin_users approver ON approver.id = aar.approver_id
	`
	where := " WHERE 1=1 "
	args := []interface{}{}
	if status != "" && status != "all" {
		where += " AND aar.status = $1"
		args = append(args, status)
	}
	total, err := r.count(ctx, `SELECT COUNT(*) `+base+where, args...)
	if err != nil {
		return nil, 0, err
	}
	argIdx := len(args) + 1
	query := `
		SELECT
			aar.id, aar.request_type, aar.resource_type, aar.resource_id, aar.reason,
			aar.payload, aar.status, aar.rejection_reason, aar.decided_at, aar.executed_at, aar.created_at,
			COALESCE(requester.full_name, requester.email) AS requester_name,
			COALESCE(approver.full_name, approver.email) AS approver_name
	` + base + where + `
		ORDER BY aar.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) CreateApprovalRequest(ctx context.Context, req *domain.AdminApprovalRequest) error {
	payload, _ := json.Marshal(req.Payload)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_approval_requests (
			id, requester_id, approver_id, request_type, resource_type, resource_id,
			reason, payload, status, rejection_reason, decided_at, executed_at, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`,
		req.ID, req.RequesterID, req.ApproverID, req.RequestType, req.ResourceType, req.ResourceID,
		req.Reason, payload, req.Status, req.RejectionReason, req.DecidedAt, req.ExecutedAt, req.CreatedAt,
	)
	return err
}

func (r *AdminRepository) FindApprovalRequestByID(ctx context.Context, approvalID string) (*domain.AdminApprovalRequest, error) {
	var result domain.AdminApprovalRequest
	var payloadBytes []byte
	err := r.db.QueryRowxContext(ctx, `
		SELECT id, requester_id, approver_id, request_type, resource_type, resource_id,
		       reason, payload, status, rejection_reason, decided_at, executed_at, created_at
		FROM admin_approval_requests
		WHERE id = $1
	`, approvalID).Scan(
		&result.ID, &result.RequesterID, &result.ApproverID, &result.RequestType, &result.ResourceType,
		&result.ResourceID, &result.Reason, &payloadBytes, &result.Status, &result.RejectionReason,
		&result.DecidedAt, &result.ExecutedAt, &result.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var payload map[string]interface{}
	if len(payloadBytes) > 0 {
		_ = json.Unmarshal(payloadBytes, &payload)
	}
	result.Payload = payload
	return &result, nil
}

func (r *AdminRepository) UpdateApprovalDecision(ctx context.Context, approvalID, approverID, status string, rejectionReason *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_approval_requests
		SET approver_id = $2, status = $3, rejection_reason = $4, decided_at = NOW()
		WHERE id = $1
	`, approvalID, approverID, status, rejectionReason)
	return err
}

func (r *AdminRepository) MarkApprovalApplied(ctx context.Context, approvalID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_approval_requests
		SET status = $2, executed_at = NOW()
		WHERE id = $1
	`, approvalID, domain.ApprovalStatusApplied)
	return err
}

func (r *AdminRepository) AddApprovalEvent(ctx context.Context, approvalRequestID, actorID, action, notes string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_approval_events (id, approval_request_id, actor_id, action, notes, created_at)
		VALUES ($1,$2,$3,$4,$5,NOW())
	`, "ape_"+uuid.New().String()[:8], approvalRequestID, actorID, action, notes)
	return err
}

func (r *AdminRepository) ListAuditLogs(ctx context.Context, page, perPage int) ([]map[string]interface{}, int, error) {
	base := `
		FROM admin_audit_logs aal
		LEFT JOIN admin_users au ON au.id = aal.admin_user_id
	`
	total, err := r.count(ctx, `SELECT COUNT(*) `+base)
	if err != nil {
		return nil, 0, err
	}
	query := `
		SELECT
			aal.id, aal.action, aal.resource_type, aal.resource_id, aal.status,
			aal.error_message, aal.ip_address, aal.user_agent, aal.created_at,
			COALESCE(au.full_name, au.email, '') AS admin_name
	` + base + `
		ORDER BY aal.created_at DESC
		LIMIT $1 OFFSET $2
	`
	items, err := r.selectMaps(ctx, query, sanitizePageSize(perPage), calculateOffset(page, perPage))
	return items, total, err
}

func (r *AdminRepository) ListSettings(ctx context.Context) ([]map[string]interface{}, error) {
	return r.selectMaps(ctx, `
		SELECT key, value, description, updated_at
		FROM admin_settings
		ORDER BY key ASC
	`)
}

func (r *AdminRepository) UpsertSetting(ctx context.Context, key string, value interface{}, description string, updatedBy *string) error {
	valueBytes, _ := json.Marshal(value)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_settings (key, value, description, updated_by, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (key)
		DO UPDATE SET value = EXCLUDED.value, description = EXCLUDED.description, updated_by = EXCLUDED.updated_by, updated_at = NOW()
	`, key, valueBytes, description, updatedBy)
	return err
}

func (r *AdminRepository) ListReferenceData(ctx context.Context) (map[string]interface{}, error) {
	operators, err := r.selectMaps(ctx, `SELECT id, name, status, sort_order FROM operators ORDER BY sort_order ASC, name ASC`)
	if err != nil {
		return nil, err
	}
	ewallets, err := r.selectMaps(ctx, `SELECT id, name, status, sort_order FROM ewallet_providers ORDER BY sort_order ASC, name ASC`)
	if err != nil {
		return nil, err
	}
	tvProviders, err := r.selectMaps(ctx, `SELECT id, name, status, sort_order FROM tv_providers ORDER BY sort_order ASC, name ASC`)
	if err != nil {
		return nil, err
	}
	banks, err := r.selectMaps(ctx, `SELECT code, name, short_name, status, is_popular FROM banks ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	provinces, err := r.selectMaps(ctx, `SELECT code, name FROM provinces ORDER BY name ASC LIMIT 100`)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"operators":   operators,
		"ewallets":    ewallets,
		"tvProviders": tvProviders,
		"banks":       banks,
		"provinces":   provinces,
	}, nil
}
