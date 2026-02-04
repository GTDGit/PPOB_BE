-- Migration: 011_create_providers_tables
-- Description: Create provider tables for various services
-- Created: 2025-01-31

-- E-Wallet Providers
CREATE TABLE IF NOT EXISTS ewallet_providers (
    id VARCHAR(36) PRIMARY KEY,                          -- gopay, ovo, dana, shopeepay, linkaja
    name VARCHAR(50) NOT NULL,
    icon VARCHAR(50) NOT NULL,
    icon_url TEXT,
    input_label VARCHAR(50) NOT NULL,                    -- "Nomor HP GoPay"
    input_placeholder VARCHAR(50),                       -- "08xxxxxxxxxx"
    input_type VARCHAR(20) DEFAULT 'phone',              -- phone, account
    status VARCHAR(20) DEFAULT 'active',
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- PDAM Regions
CREATE TABLE IF NOT EXISTS pdam_regions (
    id VARCHAR(36) PRIMARY KEY,                          -- pdam_jakarta, pdam_bandung
    name VARCHAR(100) NOT NULL,                          -- PDAM DKI Jakarta
    province VARCHAR(50) NOT NULL,                       -- DKI Jakarta
    status VARCHAR(20) DEFAULT 'active',
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- TV Cable Providers
CREATE TABLE IF NOT EXISTS tv_providers (
    id VARCHAR(36) PRIMARY KEY,                          -- indovision, transvision, topas, firstmedia
    name VARCHAR(50) NOT NULL,
    icon VARCHAR(50) NOT NULL,
    icon_url TEXT,
    input_label VARCHAR(50) DEFAULT 'Nomor Pelanggan',
    status VARCHAR(20) DEFAULT 'active',
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Banks
CREATE TABLE IF NOT EXISTS banks (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(10) UNIQUE NOT NULL,                    -- 014, 009, 002, 008
    name VARCHAR(100) NOT NULL,                          -- Bank Central Asia
    short_name VARCHAR(20) NOT NULL,                     -- BCA
    icon VARCHAR(50) NOT NULL,
    icon_url TEXT,
    transfer_fee BIGINT DEFAULT 0,                       -- 6500
    is_popular BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'active',
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_ewallet_providers_status ON ewallet_providers(status);
CREATE INDEX idx_pdam_regions_province ON pdam_regions(province);
CREATE INDEX idx_pdam_regions_status ON pdam_regions(status);
CREATE INDEX idx_tv_providers_status ON tv_providers(status);
CREATE INDEX idx_banks_code ON banks(code);
CREATE INDEX idx_banks_is_popular ON banks(is_popular);
CREATE INDEX idx_banks_status ON banks(status);
