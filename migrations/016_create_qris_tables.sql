-- Migration: 016_create_qris_tables
-- Description: Create QRIS related tables
-- Created: 2025-01-31

-- QRIS Static (Merchant QRIS)
CREATE TABLE IF NOT EXISTS qris_statics (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    merchant_name VARCHAR(100) NOT NULL,
    merchant_city VARCHAR(50),
    qris_code TEXT NOT NULL,                             -- QRIS string
    qris_image_url TEXT,                                 -- Generated QR image
    nmid VARCHAR(50),                                    -- National Merchant ID
    status VARCHAR(20) DEFAULT 'active',                 -- active, inactive, pending
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- QRIS Income (Payments received via QRIS)
CREATE TABLE IF NOT EXISTS qris_incomes (
    id VARCHAR(36) PRIMARY KEY,                          -- qris_inc_xxx
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    qris_static_id VARCHAR(36) REFERENCES qris_statics(id),

    -- Merchant info
    merchant_name VARCHAR(100) NOT NULL,

    -- Payer info
    payer_name VARCHAR(100),
    payer_bank VARCHAR(50),
    payer_account VARCHAR(30),

    -- Amount
    amount BIGINT NOT NULL,
    fee BIGINT DEFAULT 0,                                -- MDR fee
    net_amount BIGINT NOT NULL,

    -- Status
    status VARCHAR(20) DEFAULT 'pending',                -- pending, success, failed

    -- Reference
    rrn VARCHAR(20),                                     -- Retrieval Reference Number
    reference_number VARCHAR(50),

    -- Timestamps
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_qris_statics_user_id ON qris_statics(user_id);
CREATE INDEX idx_qris_statics_status ON qris_statics(status);
CREATE INDEX idx_qris_incomes_user_id ON qris_incomes(user_id);
CREATE INDEX idx_qris_incomes_status ON qris_incomes(status);
CREATE INDEX idx_qris_incomes_rrn ON qris_incomes(rrn);
CREATE INDEX idx_qris_incomes_created_at ON qris_incomes(created_at);
