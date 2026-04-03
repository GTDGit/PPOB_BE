-- Migration: 034_create_admin_console_tables
-- Description: Admin console auth, RBAC, approval, audit, and settings tables
-- Created: 2026-04-04

CREATE TABLE IF NOT EXISTS admin_permissions (
    key VARCHAR(80) PRIMARY KEY,
    module VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_roles (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT TRUE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_role_permissions (
    role_id VARCHAR(64) NOT NULL REFERENCES admin_roles(id) ON DELETE CASCADE,
    permission_key VARCHAR(80) NOT NULL REFERENCES admin_permissions(key) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_key)
);

CREATE TABLE IF NOT EXISTS admin_users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    full_name VARCHAR(150),
    password_hash TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'invited', -- invited, pending_totp, active, disabled
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMP,
    invited_by VARCHAR(36),
    created_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_admin_users_invited_by FOREIGN KEY (invited_by) REFERENCES admin_users(id) ON DELETE SET NULL,
    CONSTRAINT fk_admin_users_created_by FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_admin_users_status ON admin_users(status);
CREATE INDEX IF NOT EXISTS idx_admin_users_is_active ON admin_users(is_active);

CREATE TABLE IF NOT EXISTS admin_user_roles (
    admin_user_id VARCHAR(36) NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    role_id VARCHAR(64) NOT NULL REFERENCES admin_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (admin_user_id, role_id)
);

CREATE TABLE IF NOT EXISTS admin_invites (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    full_name VARCHAR(150),
    role_id VARCHAR(64) NOT NULL REFERENCES admin_roles(id),
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    invited_by VARCHAR(36) REFERENCES admin_users(id) ON DELETE SET NULL,
    admin_user_id VARCHAR(36) REFERENCES admin_users(id) ON DELETE SET NULL,
    expires_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_invites_email ON admin_invites(email);
CREATE INDEX IF NOT EXISTS idx_admin_invites_expires_at ON admin_invites(expires_at);

CREATE TABLE IF NOT EXISTS admin_sessions (
    id VARCHAR(36) PRIMARY KEY,
    admin_user_id VARCHAR(36) NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_sessions_admin_user_id ON admin_sessions(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_expires_at ON admin_sessions(expires_at);

CREATE TABLE IF NOT EXISTS admin_totp_secrets (
    admin_user_id VARCHAR(36) PRIMARY KEY REFERENCES admin_users(id) ON DELETE CASCADE,
    secret VARCHAR(255) NOT NULL,
    confirmed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_recovery_codes (
    id VARCHAR(36) PRIMARY KEY,
    admin_user_id VARCHAR(36) NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_recovery_codes_admin_user_id ON admin_recovery_codes(admin_user_id);

CREATE TABLE IF NOT EXISTS admin_audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    admin_user_id VARCHAR(36) REFERENCES admin_users(id) ON DELETE SET NULL,
    action VARCHAR(80) NOT NULL,
    resource_type VARCHAR(80),
    resource_id VARCHAR(80),
    old_value JSONB,
    new_value JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    status VARCHAR(20) DEFAULT 'success',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_admin_user_id ON admin_audit_logs(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_action ON admin_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_resource ON admin_audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_logs_created_at ON admin_audit_logs(created_at DESC);

CREATE TABLE IF NOT EXISTS admin_approval_requests (
    id VARCHAR(36) PRIMARY KEY,
    requester_id VARCHAR(36) NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    approver_id VARCHAR(36) REFERENCES admin_users(id) ON DELETE SET NULL,
    request_type VARCHAR(50) NOT NULL, -- price_change, balance_adjustment, refund, role_escalation
    resource_type VARCHAR(80) NOT NULL,
    resource_id VARCHAR(80),
    reason TEXT,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected, applied
    rejection_reason TEXT,
    decided_at TIMESTAMP,
    executed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_approval_requests_status ON admin_approval_requests(status);
CREATE INDEX IF NOT EXISTS idx_admin_approval_requests_requester ON admin_approval_requests(requester_id);
CREATE INDEX IF NOT EXISTS idx_admin_approval_requests_approver ON admin_approval_requests(approver_id);
CREATE INDEX IF NOT EXISTS idx_admin_approval_requests_resource ON admin_approval_requests(resource_type, resource_id);

CREATE TABLE IF NOT EXISTS admin_approval_events (
    id VARCHAR(36) PRIMARY KEY,
    approval_request_id VARCHAR(36) NOT NULL REFERENCES admin_approval_requests(id) ON DELETE CASCADE,
    actor_id VARCHAR(36) REFERENCES admin_users(id) ON DELETE SET NULL,
    action VARCHAR(30) NOT NULL, -- created, approved, rejected, applied
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_approval_events_request_id ON admin_approval_events(approval_request_id);

CREATE TABLE IF NOT EXISTS admin_settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_by VARCHAR(36) REFERENCES admin_users(id) ON DELETE SET NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO admin_permissions (key, module, action, description) VALUES
('dashboard.view', 'dashboard', 'view', 'Lihat ringkasan dashboard'),
('admins.view', 'admins', 'view', 'Lihat daftar admin'),
('admins.invite', 'admins', 'invite', 'Buat undangan admin baru'),
('admins.manage', 'admins', 'manage', 'Kelola admin dan statusnya'),
('roles.view', 'roles', 'view', 'Lihat role dan permission'),
('roles.manage', 'roles', 'manage', 'Kelola role dan permission'),
('customers.view', 'customers', 'view', 'Lihat daftar pengguna'),
('customers.manage', 'customers', 'manage', 'Kelola status pengguna'),
('transactions.view', 'transactions', 'view', 'Lihat transaksi'),
('transactions.refund', 'transactions', 'refund', 'Ajukan atau proses refund transaksi'),
('transactions.override', 'transactions', 'override', 'Override status transaksi'),
('deposits.view', 'deposits', 'view', 'Lihat daftar deposit'),
('deposits.approve', 'deposits', 'approve', 'Approve atau reject deposit'),
('qris.view', 'qris', 'view', 'Lihat data QRIS'),
('vouchers.view', 'vouchers', 'view', 'Lihat voucher'),
('vouchers.manage', 'vouchers', 'manage', 'Kelola voucher'),
('catalog.view', 'catalog', 'view', 'Lihat katalog produk dan layanan'),
('catalog.manage', 'catalog', 'manage', 'Kelola produk dan layanan'),
('pricing.view', 'pricing', 'view', 'Lihat pricing'),
('pricing.request', 'pricing', 'request', 'Buat request perubahan pricing'),
('pricing.approve', 'pricing', 'approve', 'Approve perubahan pricing'),
('kyc.view', 'kyc', 'view', 'Lihat review KYC'),
('kyc.approve', 'kyc', 'approve', 'Approve atau reject KYC'),
('notifications.view', 'notifications', 'view', 'Lihat notifikasi'),
('notifications.manage', 'notifications', 'manage', 'Kirim notifikasi'),
('banners.view', 'banners', 'view', 'Lihat banner'),
('banners.manage', 'banners', 'manage', 'Kelola banner'),
('reference.view', 'reference', 'view', 'Lihat data referensi'),
('reference.manage', 'reference', 'manage', 'Kelola data referensi'),
('approvals.view', 'approvals', 'view', 'Lihat queue approval'),
('approvals.act', 'approvals', 'act', 'Approve atau reject request'),
('audit.view', 'audit', 'view', 'Lihat audit log'),
('security.view', 'security', 'view', 'Lihat security dan session'),
('security.manage', 'security', 'manage', 'Kelola security policy'),
('settings.view', 'settings', 'view', 'Lihat settings'),
('settings.manage', 'settings', 'manage', 'Kelola settings'),
('finance.adjust_balance', 'finance', 'adjust_balance', 'Buat atau approve koreksi saldo')
ON CONFLICT (key) DO NOTHING;

INSERT INTO admin_roles (id, name, description, is_system, is_active) VALUES
('super_admin', 'Super Admin', 'Akses seluruh fitur dan approval final', TRUE, TRUE),
('admin_operasional', 'Admin Operasional', 'Monitoring operasional harian', TRUE, TRUE),
('finance_staff', 'Finance Staff', 'Operasional finance dan request perubahan sensitif', TRUE, TRUE),
('finance_approver', 'Finance Approver', 'Approver akhir untuk aksi sensitif finance', TRUE, TRUE),
('customer_service', 'Customer Service', 'Bantu pengguna dan investigasi transaksi', TRUE, TRUE),
('compliance_kyc', 'Compliance KYC', 'Review dan approve KYC', TRUE, TRUE),
('product_content', 'Product & Content', 'Kelola produk, voucher, banner, notifikasi', TRUE, TRUE),
('auditor_viewer', 'Auditor / Viewer', 'Akses read-only untuk audit dan laporan', TRUE, TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO admin_role_permissions (role_id, permission_key)
SELECT 'super_admin', key FROM admin_permissions
ON CONFLICT DO NOTHING;

INSERT INTO admin_role_permissions (role_id, permission_key) VALUES
('admin_operasional', 'dashboard.view'),
('admin_operasional', 'customers.view'),
('admin_operasional', 'customers.manage'),
('admin_operasional', 'transactions.view'),
('admin_operasional', 'deposits.view'),
('admin_operasional', 'deposits.approve'),
('admin_operasional', 'qris.view'),
('admin_operasional', 'vouchers.view'),
('admin_operasional', 'catalog.view'),
('admin_operasional', 'kyc.view'),
('admin_operasional', 'kyc.approve'),
('admin_operasional', 'notifications.view'),
('admin_operasional', 'banners.view'),
('admin_operasional', 'approvals.view'),
('finance_staff', 'dashboard.view'),
('finance_staff', 'transactions.view'),
('finance_staff', 'transactions.refund'),
('finance_staff', 'deposits.view'),
('finance_staff', 'qris.view'),
('finance_staff', 'pricing.view'),
('finance_staff', 'pricing.request'),
('finance_staff', 'approvals.view'),
('finance_staff', 'finance.adjust_balance'),
('finance_approver', 'dashboard.view'),
('finance_approver', 'transactions.view'),
('finance_approver', 'transactions.refund'),
('finance_approver', 'transactions.override'),
('finance_approver', 'deposits.view'),
('finance_approver', 'deposits.approve'),
('finance_approver', 'qris.view'),
('finance_approver', 'pricing.view'),
('finance_approver', 'pricing.request'),
('finance_approver', 'pricing.approve'),
('finance_approver', 'approvals.view'),
('finance_approver', 'approvals.act'),
('finance_approver', 'audit.view'),
('finance_approver', 'finance.adjust_balance'),
('customer_service', 'dashboard.view'),
('customer_service', 'customers.view'),
('customer_service', 'customers.manage'),
('customer_service', 'transactions.view'),
('customer_service', 'deposits.view'),
('customer_service', 'qris.view'),
('customer_service', 'kyc.view'),
('customer_service', 'security.view'),
('compliance_kyc', 'dashboard.view'),
('compliance_kyc', 'customers.view'),
('compliance_kyc', 'kyc.view'),
('compliance_kyc', 'kyc.approve'),
('compliance_kyc', 'audit.view'),
('product_content', 'dashboard.view'),
('product_content', 'vouchers.view'),
('product_content', 'vouchers.manage'),
('product_content', 'catalog.view'),
('product_content', 'catalog.manage'),
('product_content', 'pricing.view'),
('product_content', 'pricing.request'),
('product_content', 'notifications.view'),
('product_content', 'notifications.manage'),
('product_content', 'banners.view'),
('product_content', 'banners.manage'),
('product_content', 'reference.view'),
('auditor_viewer', 'dashboard.view'),
('auditor_viewer', 'customers.view'),
('auditor_viewer', 'transactions.view'),
('auditor_viewer', 'deposits.view'),
('auditor_viewer', 'qris.view'),
('auditor_viewer', 'vouchers.view'),
('auditor_viewer', 'catalog.view'),
('auditor_viewer', 'pricing.view'),
('auditor_viewer', 'kyc.view'),
('auditor_viewer', 'notifications.view'),
('auditor_viewer', 'banners.view'),
('auditor_viewer', 'reference.view'),
('auditor_viewer', 'approvals.view'),
('auditor_viewer', 'audit.view'),
('auditor_viewer', 'settings.view')
ON CONFLICT DO NOTHING;

INSERT INTO admin_settings (key, value, description)
VALUES
('support_contact', '{"email":"support@ppob.id","phone":"+6280000000000"}', 'Kontak dukungan admin console'),
('ops_notice', '{"enabled":false,"message":""}', 'Pesan operasional internal'),
('dashboard_defaults', '{"currency":"IDR","timezone":"Asia/Jakarta"}', 'Default dashboard rendering')
ON CONFLICT (key) DO NOTHING;
