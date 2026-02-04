-- Migration: 004_create_otp_sessions_table
-- Description: Create OTP sessions table for verification
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS otp_sessions (
    id VARCHAR(36) PRIMARY KEY,                          -- otp_xxx
    phone VARCHAR(15) NOT NULL,
    otp_code VARCHAR(6) NOT NULL,                        -- 6 digit OTP
    otp_method VARCHAR(10) NOT NULL,                     -- wa, sms
    session_id VARCHAR(50) NOT NULL,                     -- otp_abc123xyz
    flow VARCHAR(20) NOT NULL,                           -- REGISTER, LOGIN, RESET_PIN, CHANGE_PHONE
    attempts INTEGER DEFAULT 0,                          -- Max 5
    resend_count INTEGER DEFAULT 0,                      -- Max 3
    last_resend_at TIMESTAMP,
    is_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_otp_sessions_phone ON otp_sessions(phone);
CREATE INDEX idx_otp_sessions_session_id ON otp_sessions(session_id);
CREATE INDEX idx_otp_sessions_expires_at ON otp_sessions(expires_at);
