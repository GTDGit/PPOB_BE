-- Migration: 006_create_user_settings_table
-- Description: Create user settings table
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS user_settings (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Security settings
    pin_required_for_transaction BOOLEAN DEFAULT TRUE,
    pin_required_min_amount BIGINT DEFAULT 0,            -- Min amount requiring PIN (if pin_required_for_transaction = false)
    biometric_enabled BOOLEAN DEFAULT FALSE,

    -- Transaction settings
    default_selling_price_markup INTEGER DEFAULT 0,      -- Default markup for selling price
    auto_save_contact BOOLEAN DEFAULT TRUE,
    show_profit_on_receipt BOOLEAN DEFAULT TRUE,

    -- Display settings
    language VARCHAR(5) DEFAULT 'id',                    -- id, en
    currency VARCHAR(5) DEFAULT 'IDR',
    theme VARCHAR(10) DEFAULT 'light',                   -- light, dark, system

    -- Privacy settings
    show_phone_on_qris BOOLEAN DEFAULT FALSE,
    show_name_on_qris BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_user_settings_user_id ON user_settings(user_id);
