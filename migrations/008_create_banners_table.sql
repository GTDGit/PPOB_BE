-- Migration: 008_create_banners_table
-- Description: Create banners table for promotions
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS banners (
    id VARCHAR(36) PRIMARY KEY,                          -- banner_001
    title VARCHAR(100) NOT NULL,
    subtitle VARCHAR(200),
    image_url TEXT NOT NULL,
    thumbnail_url TEXT,
    action_type VARCHAR(20) DEFAULT 'none',              -- deeplink, webview, external_url, none
    action_value TEXT,                                   -- route or URL
    background_color VARCHAR(10),                        -- #1E3A8A
    text_color VARCHAR(10) DEFAULT '#FFFFFF',
    placement VARCHAR(20) DEFAULT 'home',                -- home, services, checkout, profile
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    priority INTEGER DEFAULT 0,                          -- Lower = higher priority
    target_tiers TEXT,                                   -- JSON array: ["BRONZE", "SILVER", "GOLD", "PLATINUM"]
    is_new_user_only BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_banners_placement ON banners(placement);
CREATE INDEX idx_banners_dates ON banners(start_date, end_date);
CREATE INDEX idx_banners_is_active ON banners(is_active);
CREATE INDEX idx_banners_priority ON banners(priority);
