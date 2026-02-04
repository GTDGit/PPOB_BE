-- Migration: 005_create_balances_table
-- Description: Create user balances table
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS balances (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount BIGINT DEFAULT 0,                             -- Main balance in Rupiah
    pending_amount BIGINT DEFAULT 0,                     -- Pending balance (refunds, etc)
    points INTEGER DEFAULT 0,                            -- Reward points
    points_expires_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Balance history for audit trail
CREATE TABLE IF NOT EXISTS balance_histories (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL,                           -- credit, debit
    category VARCHAR(30) NOT NULL,                       -- transaction, deposit, refund, bonus, points
    amount BIGINT NOT NULL,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    reference_type VARCHAR(30),                          -- transaction, deposit, adjustment
    reference_id VARCHAR(36),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_balances_user_id ON balances(user_id);
CREATE INDEX idx_balance_histories_user_id ON balance_histories(user_id);
CREATE INDEX idx_balance_histories_reference ON balance_histories(reference_type, reference_id);
CREATE INDEX idx_balance_histories_created_at ON balance_histories(created_at);
