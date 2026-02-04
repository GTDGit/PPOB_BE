-- Migration: 019_create_audit_logs_table
-- Description: Create audit logs table for security and compliance
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) REFERENCES users(id),
    action VARCHAR(50) NOT NULL,                         -- login, logout, pin_change, phone_change, transaction, etc
    resource_type VARCHAR(50),                           -- user, transaction, device, etc
    resource_id VARCHAR(36),
    old_value JSONB,                                     -- Previous value (for updates)
    new_value JSONB,                                     -- New value (for updates)
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_id VARCHAR(100),
    status VARCHAR(20) DEFAULT 'success',                -- success, failed
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Failed login attempts tracking
CREATE TABLE IF NOT EXISTS login_attempts (
    id VARCHAR(36) PRIMARY KEY,
    phone VARCHAR(15) NOT NULL,
    device_id VARCHAR(100),
    ip_address VARCHAR(45),
    attempt_type VARCHAR(20) NOT NULL,                   -- otp, pin
    is_success BOOLEAN DEFAULT FALSE,
    failure_reason VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- PIN lock tracking
CREATE TABLE IF NOT EXISTS pin_locks (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    attempts INTEGER DEFAULT 0,
    is_locked BOOLEAN DEFAULT FALSE,
    locked_at TIMESTAMP,
    unlock_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id)
);

-- Indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_login_attempts_phone ON login_attempts(phone);
CREATE INDEX idx_login_attempts_ip_address ON login_attempts(ip_address);
CREATE INDEX idx_login_attempts_created_at ON login_attempts(created_at);
CREATE INDEX idx_pin_locks_user_id ON pin_locks(user_id);
