-- Migration: 013_create_vouchers_tables
-- Description: Create vouchers and user_vouchers tables
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS vouchers (
    id VARCHAR(36) PRIMARY KEY,                          -- vch_xxx
    code VARCHAR(30) UNIQUE NOT NULL,                    -- DISKON10, CASHBACK5
    name VARCHAR(100) NOT NULL,                          -- Discount Rp12.000
    description TEXT,
    discount_type VARCHAR(20) NOT NULL,                  -- fixed, percentage
    discount_value INTEGER NOT NULL,                     -- Amount or percentage
    min_transaction BIGINT DEFAULT 0,                    -- Minimum transaction amount
    max_discount BIGINT,                                 -- Maximum discount (for percentage)
    applicable_services TEXT,                            -- JSON array: ["pulsa", "data", "pln_prepaid"] or ["all"]
    max_usage INTEGER,                                   -- Total usage limit
    max_usage_per_user INTEGER DEFAULT 1,                -- Per user limit
    current_usage INTEGER DEFAULT 0,
    terms_url TEXT,
    starts_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_vouchers (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    voucher_id VARCHAR(36) NOT NULL REFERENCES vouchers(id),
    status VARCHAR(20) DEFAULT 'active',                 -- active, used, expired
    usage_count INTEGER DEFAULT 0,
    transaction_id VARCHAR(36),                          -- Transaction where voucher was used
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, voucher_id)
);

-- Voucher usage log
CREATE TABLE IF NOT EXISTS voucher_usages (
    id VARCHAR(36) PRIMARY KEY,
    voucher_id VARCHAR(36) NOT NULL REFERENCES vouchers(id),
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    transaction_id VARCHAR(36) NOT NULL REFERENCES transactions(id),
    discount_amount BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_vouchers_code ON vouchers(code);
CREATE INDEX idx_vouchers_expires_at ON vouchers(expires_at);
CREATE INDEX idx_vouchers_is_active ON vouchers(is_active);
CREATE INDEX idx_user_vouchers_user_id ON user_vouchers(user_id);
CREATE INDEX idx_user_vouchers_voucher_id ON user_vouchers(voucher_id);
CREATE INDEX idx_user_vouchers_status ON user_vouchers(status);
CREATE INDEX idx_voucher_usages_voucher_id ON voucher_usages(voucher_id);
CREATE INDEX idx_voucher_usages_user_id ON voucher_usages(user_id);
