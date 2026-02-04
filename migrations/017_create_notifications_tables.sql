-- Migration: 017_create_notifications_tables
-- Description: Create notifications related tables
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS notifications (
    id VARCHAR(36) PRIMARY KEY,                          -- notif_xxx
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(20) NOT NULL,                       -- security, transaction, deposit, promo, info, qris
    title VARCHAR(100) NOT NULL,
    body TEXT NOT NULL,
    short_body VARCHAR(200),
    image_url TEXT,
    action_type VARCHAR(20) DEFAULT 'none',              -- deeplink, webview, external_url, none
    action_value TEXT,
    action_button_text VARCHAR(50),
    metadata JSONB,                                      -- Additional data (transactionId, promoId, etc)
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Push tokens for FCM
CREATE TABLE IF NOT EXISTS push_tokens (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(100) NOT NULL,
    token TEXT NOT NULL,
    platform VARCHAR(20) NOT NULL,                       -- android, ios, web
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, device_id)
);

-- Notification settings per user
CREATE TABLE IF NOT EXISTS notification_settings (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    push_enabled BOOLEAN DEFAULT TRUE,

    -- Category settings (push, in_app)
    security_push BOOLEAN DEFAULT TRUE,
    security_in_app BOOLEAN DEFAULT TRUE,
    transaction_push BOOLEAN DEFAULT TRUE,
    transaction_in_app BOOLEAN DEFAULT TRUE,
    deposit_push BOOLEAN DEFAULT TRUE,
    deposit_in_app BOOLEAN DEFAULT TRUE,
    promo_push BOOLEAN DEFAULT TRUE,
    promo_in_app BOOLEAN DEFAULT TRUE,
    info_push BOOLEAN DEFAULT FALSE,
    info_in_app BOOLEAN DEFAULT TRUE,
    qris_push BOOLEAN DEFAULT TRUE,
    qris_in_app BOOLEAN DEFAULT TRUE,

    -- Quiet hours
    quiet_hours_enabled BOOLEAN DEFAULT FALSE,
    quiet_hours_start TIME DEFAULT '22:00',
    quiet_hours_end TIME DEFAULT '07:00',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_category ON notifications(category);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE INDEX idx_push_tokens_user_id ON push_tokens(user_id);
CREATE INDEX idx_push_tokens_device_id ON push_tokens(device_id);
CREATE INDEX idx_push_tokens_is_active ON push_tokens(is_active);
CREATE INDEX idx_notification_settings_user_id ON notification_settings(user_id);
