-- Migration: 018_create_referrals_table
-- Description: Create referrals table for referral program
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS referrals (
    id VARCHAR(36) PRIMARY KEY,                          -- ref_xxx
    referrer_id VARCHAR(36) NOT NULL REFERENCES users(id),  -- User who shared the code
    referee_id VARCHAR(36) NOT NULL REFERENCES users(id),   -- User who used the code
    referral_code VARCHAR(20) NOT NULL,                  -- The code that was used
    status VARCHAR(20) DEFAULT 'pending',                -- pending, completed, expired

    -- Bonus amounts
    referrer_bonus BIGINT DEFAULT 0,                     -- Bonus for referrer
    referee_bonus BIGINT DEFAULT 0,                      -- Bonus for referee

    -- Status tracking
    is_bonus_paid BOOLEAN DEFAULT FALSE,
    bonus_paid_at TIMESTAMP,

    -- Timestamps
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,   -- When referee registered
    first_transaction_at TIMESTAMP,                      -- When referee made first transaction
    completed_at TIMESTAMP,                              -- When bonus was credited
    expires_at TIMESTAMP,                                -- Deadline for first transaction

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(referee_id)                                   -- Each user can only be referred once
);

-- Referral settings (global)
CREATE TABLE IF NOT EXISTS referral_settings (
    id VARCHAR(36) PRIMARY KEY,
    referrer_bonus BIGINT DEFAULT 10000,                 -- Default Rp10.000
    referee_bonus BIGINT DEFAULT 5000,                   -- Default Rp5.000
    expiry_days INTEGER DEFAULT 30,                      -- Days to make first transaction
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_referrals_referrer_id ON referrals(referrer_id);
CREATE INDEX idx_referrals_referee_id ON referrals(referee_id);
CREATE INDEX idx_referrals_referral_code ON referrals(referral_code);
CREATE INDEX idx_referrals_status ON referrals(status);
CREATE INDEX idx_referrals_created_at ON referrals(created_at);
