-- Migration: 031_create_kyc_tables
-- Description: Tables for KYC verification flow (OCR + Face Comparison + Liveness)
-- Note: users.kyc_status already exists in 001_create_users_table.sql

-- KYC Sessions (temporary, expires in 24h)
CREATE TABLE IF NOT EXISTS kyc_sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    nik VARCHAR(16),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending, processing, completed, failed
    current_step INT NOT NULL DEFAULT 1,            -- 1=ktp, 2=face, 3=liveness
    ocr_data JSONB,
    face_urls JSONB,                                -- {ktp, face, fullImage}
    liveness_data JSONB,                            -- {sessionId, confidence, imageUrl}
    face_comparison JSONB,                          -- {matched, similarity, threshold}
    error_message TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_kyc_sessions_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_kyc_sessions_user ON kyc_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_kyc_sessions_status ON kyc_sessions(status);
CREATE INDEX IF NOT EXISTS idx_kyc_sessions_nik ON kyc_sessions(nik);
CREATE INDEX IF NOT EXISTS idx_kyc_sessions_expires ON kyc_sessions(expires_at);

COMMENT ON TABLE kyc_sessions IS 'Temporary KYC verification sessions (expires in 24h)';
COMMENT ON COLUMN kyc_sessions.status IS 'Session status: pending, processing, completed, failed';
COMMENT ON COLUMN kyc_sessions.current_step IS 'Current step: 1=KTP OCR, 2=Face capture, 3=Liveness check';

-- KYC Verified Data (permanent, 1 per user, 1 per NIK)
CREATE TABLE IF NOT EXISTS kyc_verifications (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) UNIQUE NOT NULL,
    nik VARCHAR(16) UNIQUE NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    place_of_birth VARCHAR(100) NOT NULL,
    date_of_birth DATE NOT NULL,
    gender VARCHAR(10) NOT NULL,
    religion VARCHAR(20),
    address_street VARCHAR(255) NOT NULL,
    address_rt VARCHAR(5),
    address_rw VARCHAR(5),
    address_sub_district VARCHAR(100) NOT NULL,
    address_district VARCHAR(100) NOT NULL,
    address_city VARCHAR(100) NOT NULL,
    address_province VARCHAR(100) NOT NULL,
    administrative_code JSONB NOT NULL,             -- {province, city, district, subDistrict}
    ktp_url VARCHAR(500),
    face_url VARCHAR(500),
    face_with_ktp_url VARCHAR(500),
    liveness_url VARCHAR(500),
    face_similarity DECIMAL(5,2),
    liveness_confidence DECIMAL(5,2),
    verified_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_kyc_verifications_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_kyc_verifications_user ON kyc_verifications(user_id);
CREATE INDEX IF NOT EXISTS idx_kyc_verifications_nik ON kyc_verifications(nik);
CREATE INDEX IF NOT EXISTS idx_kyc_verifications_date ON kyc_verifications(verified_at);

COMMENT ON TABLE kyc_verifications IS 'Permanent KYC verification data (1 per user, 1 per NIK)';
COMMENT ON COLUMN kyc_verifications.nik IS 'National ID number (16 digits, unique)';
COMMENT ON COLUMN kyc_verifications.administrative_code IS 'Location codes: {province, city, district, subDistrict}';
COMMENT ON COLUMN kyc_verifications.face_similarity IS 'Face comparison similarity score (0-100)';
COMMENT ON COLUMN kyc_verifications.liveness_confidence IS 'Liveness detection confidence (0-100)';

-- KYC History (audit log)
CREATE TABLE IF NOT EXISTS kyc_history (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    session_id VARCHAR(36),
    action VARCHAR(50) NOT NULL,                    -- session_created, ocr_completed, face_captured, liveness_checked, verification_approved, verification_rejected
    status VARCHAR(20) NOT NULL,                    -- success, failed
    metadata JSONB,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_kyc_history_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_kyc_history_user ON kyc_history(user_id);
CREATE INDEX IF NOT EXISTS idx_kyc_history_session ON kyc_history(session_id);
CREATE INDEX IF NOT EXISTS idx_kyc_history_action ON kyc_history(action);
CREATE INDEX IF NOT EXISTS idx_kyc_history_created ON kyc_history(created_at);

COMMENT ON TABLE kyc_history IS 'Audit log for all KYC verification activities';
COMMENT ON COLUMN kyc_history.action IS 'Type of action performed in verification flow';
