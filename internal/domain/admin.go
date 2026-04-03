package domain

import (
	"database/sql"
	"time"
)

const (
	AdminStatusInvited     = "invited"
	AdminStatusPendingTOTP = "pending_totp"
	AdminStatusActive      = "active"
	AdminStatusDisabled    = "disabled"

	ApprovalStatusPending  = "pending"
	ApprovalStatusApproved = "approved"
	ApprovalStatusRejected = "rejected"
	ApprovalStatusApplied  = "applied"
)

type AdminUser struct {
	ID           string         `db:"id" json:"id"`
	Email        string         `db:"email" json:"email"`
	Phone        string         `db:"phone" json:"phone"`
	FullName     sql.NullString `db:"full_name" json:"-"`
	PasswordHash sql.NullString `db:"password_hash" json:"-"`
	Status       string         `db:"status" json:"status"`
	IsActive     bool           `db:"is_active" json:"isActive"`
	LastLoginAt  sql.NullTime   `db:"last_login_at" json:"-"`
	InvitedBy    sql.NullString `db:"invited_by" json:"-"`
	CreatedBy    sql.NullString `db:"created_by" json:"-"`
	CreatedAt    time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updatedAt"`

	Roles       []AdminRole `json:"roles,omitempty"`
	Permissions []string    `json:"permissions,omitempty"`
}

type AdminRole struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	IsSystem    bool      `db:"is_system" json:"isSystem"`
	IsActive    bool      `db:"is_active" json:"isActive"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`

	Permissions []string `json:"permissions,omitempty"`
}

type AdminPermission struct {
	Key         string    `db:"key" json:"key"`
	Module      string    `db:"module" json:"module"`
	Action      string    `db:"action" json:"action"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
}

type AdminInvite struct {
	ID          string         `db:"id" json:"id"`
	Email       string         `db:"email" json:"email"`
	Phone       string         `db:"phone" json:"phone"`
	FullName    sql.NullString `db:"full_name" json:"-"`
	RoleID      string         `db:"role_id" json:"roleId"`
	TokenHash   string         `db:"token_hash" json:"-"`
	InvitedBy   sql.NullString `db:"invited_by" json:"-"`
	AdminUserID sql.NullString `db:"admin_user_id" json:"-"`
	ExpiresAt   time.Time      `db:"expires_at" json:"expiresAt"`
	AcceptedAt  sql.NullTime   `db:"accepted_at" json:"-"`
	CreatedAt   time.Time      `db:"created_at" json:"createdAt"`
}

type AdminPasswordReset struct {
	ID          string         `db:"id" json:"id"`
	AdminUserID string         `db:"admin_user_id" json:"adminUserId"`
	Email       string         `db:"email" json:"email"`
	TokenHash   string         `db:"token_hash" json:"-"`
	ExpiresAt   time.Time      `db:"expires_at" json:"expiresAt"`
	UsedAt      sql.NullTime   `db:"used_at" json:"usedAt"`
	CreatedAt   time.Time      `db:"created_at" json:"createdAt"`
	RequestedBy sql.NullString `db:"requested_by" json:"requestedBy"`
}

type AdminSession struct {
	ID               string    `db:"id"`
	AdminUserID      string    `db:"admin_user_id"`
	RefreshTokenHash string    `db:"refresh_token_hash"`
	IPAddress        string    `db:"ip_address"`
	UserAgent        string    `db:"user_agent"`
	ExpiresAt        time.Time `db:"expires_at"`
	CreatedAt        time.Time `db:"created_at"`
	LastUsedAt       time.Time `db:"last_used_at"`
}

type AdminTOTPSecret struct {
	AdminUserID string       `db:"admin_user_id"`
	Secret      string       `db:"secret"`
	ConfirmedAt sql.NullTime `db:"confirmed_at"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}

type AdminRecoveryCode struct {
	ID          string       `db:"id" json:"id"`
	AdminUserID string       `db:"admin_user_id" json:"adminUserId"`
	CodeHash    string       `db:"code_hash" json:"-"`
	UsedAt      sql.NullTime `db:"used_at" json:"usedAt"`
	CreatedAt   time.Time    `db:"created_at" json:"createdAt"`
}

type AdminAuditLog struct {
	ID           string         `db:"id" json:"id"`
	AdminUserID  sql.NullString `db:"admin_user_id" json:"adminUserId"`
	Action       string         `db:"action" json:"action"`
	ResourceType sql.NullString `db:"resource_type" json:"resourceType"`
	ResourceID   sql.NullString `db:"resource_id" json:"resourceId"`
	OldValue     interface{}    `db:"old_value" json:"oldValue"`
	NewValue     interface{}    `db:"new_value" json:"newValue"`
	IPAddress    sql.NullString `db:"ip_address" json:"ipAddress"`
	UserAgent    sql.NullString `db:"user_agent" json:"userAgent"`
	Status       string         `db:"status" json:"status"`
	ErrorMessage sql.NullString `db:"error_message" json:"errorMessage"`
	CreatedAt    time.Time      `db:"created_at" json:"createdAt"`
}

type AdminApprovalRequest struct {
	ID              string         `db:"id" json:"id"`
	RequesterID     string         `db:"requester_id" json:"requesterId"`
	ApproverID      sql.NullString `db:"approver_id" json:"approverId"`
	RequestType     string         `db:"request_type" json:"requestType"`
	ResourceType    string         `db:"resource_type" json:"resourceType"`
	ResourceID      sql.NullString `db:"resource_id" json:"resourceId"`
	Reason          sql.NullString `db:"reason" json:"reason"`
	Payload         interface{}    `db:"payload" json:"payload"`
	Status          string         `db:"status" json:"status"`
	RejectionReason sql.NullString `db:"rejection_reason" json:"rejectionReason"`
	DecidedAt       sql.NullTime   `db:"decided_at" json:"decidedAt"`
	ExecutedAt      sql.NullTime   `db:"executed_at" json:"executedAt"`
	CreatedAt       time.Time      `db:"created_at" json:"createdAt"`
}

type AdminSetting struct {
	Key         string      `db:"key" json:"key"`
	Value       interface{} `db:"value" json:"value"`
	Description string      `db:"description" json:"description"`
	UpdatedBy   *string     `json:"updatedBy,omitempty"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updatedAt"`
}

type AdminUserSummary struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	FullName    string   `json:"fullName"`
	Status      string   `json:"status"`
	IsActive    bool     `json:"isActive"`
	LastLoginAt *string  `json:"lastLoginAt,omitempty"`
	Roles       []string `json:"roles"`
	CreatedAt   string   `json:"createdAt"`
}

type AdminAuthResponse struct {
	AccessToken  string            `json:"accessToken"`
	RefreshToken string            `json:"refreshToken"`
	ExpiresIn    int               `json:"expiresIn"`
	User         *AdminUserSummary `json:"user"`
	Permissions  []string          `json:"permissions"`
}

type AdminInvitePreviewResponse struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	RoleID    string `json:"roleId"`
	RoleName  string `json:"roleName"`
	FullName  string `json:"fullName,omitempty"`
	ExpiresAt string `json:"expiresAt"`
}

type AdminInviteAcceptResponse struct {
	AdminID       string   `json:"adminId"`
	Secret        string   `json:"secret"`
	OTPAuthURL    string   `json:"otpauthUrl"`
	RecoveryCodes []string `json:"recoveryCodes"`
}

type AdminPasswordResetPreviewResponse struct {
	Email     string `json:"email"`
	ExpiresAt string `json:"expiresAt"`
}

type AdminDashboardResponse struct {
	TotalUsers         int   `json:"totalUsers"`
	ActiveAdmins       int   `json:"activeAdmins"`
	TransactionsToday  int   `json:"transactionsToday"`
	DepositsPending    int   `json:"depositsPending"`
	PendingKYC         int   `json:"pendingKYC"`
	PendingApprovals   int   `json:"pendingApprovals"`
	RevenueToday       int64 `json:"revenueToday"`
	DepositAmountToday int64 `json:"depositAmountToday"`
}

type AdminListResponse struct {
	Items   interface{} `json:"items"`
	Page    int         `json:"page"`
	PerPage int         `json:"perPage"`
	Total   int         `json:"total"`
	HasNext bool        `json:"hasNext"`
}

type AdminReferenceOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
}

func (u *AdminUser) DisplayName() string {
	if u.FullName.Valid && u.FullName.String != "" {
		return u.FullName.String
	}
	return u.Email
}

func (u *AdminUser) ToSummary() *AdminUserSummary {
	roles := make([]string, 0, len(u.Roles))
	for _, role := range u.Roles {
		roles = append(roles, role.Name)
	}

	var lastLogin *string
	if u.LastLoginAt.Valid {
		formatted := u.LastLoginAt.Time.Format(time.RFC3339)
		lastLogin = &formatted
	}

	return &AdminUserSummary{
		ID:          u.ID,
		Email:       u.Email,
		Phone:       u.Phone,
		FullName:    u.DisplayName(),
		Status:      u.Status,
		IsActive:    u.IsActive,
		LastLoginAt: lastLogin,
		Roles:       roles,
		CreatedAt:   u.CreatedAt.Format(time.RFC3339),
	}
}
