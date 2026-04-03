-- Migration: 035_create_admin_password_resets_table
-- Description: Password reset tokens for admin console accounts
-- Created: 2026-04-04

CREATE TABLE IF NOT EXISTS admin_password_resets (
    id VARCHAR(36) PRIMARY KEY,
    admin_user_id VARCHAR(36) NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    requested_by VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_password_resets_admin_user_id
    ON admin_password_resets(admin_user_id);

CREATE INDEX IF NOT EXISTS idx_admin_password_resets_expires_at
    ON admin_password_resets(expires_at);
