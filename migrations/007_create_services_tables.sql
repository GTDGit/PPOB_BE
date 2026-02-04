-- Migration: 007_create_services_tables
-- Description: Create service categories and services tables
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS service_categories (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,                           -- Pra Bayar, Pasca Bayar, Keuangan
    slug VARCHAR(50) UNIQUE NOT NULL,                    -- prepaid, postpaid, finance
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS services (
    id VARCHAR(36) PRIMARY KEY,                          -- pulsa, paket_data, token_pln, etc
    category_id VARCHAR(36) REFERENCES service_categories(id),
    name VARCHAR(50) NOT NULL,                           -- Pulsa, Paket Data, Token PLN
    icon VARCHAR(50) NOT NULL,                           -- icon identifier
    icon_url TEXT,
    route VARCHAR(100) NOT NULL,                         -- /services/pulsa
    status VARCHAR(20) DEFAULT 'active',                 -- active, maintenance, coming_soon, hidden
    badge VARCHAR(20),                                   -- PROMO, NEW, HOT, null
    is_featured BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_services_category_id ON services(category_id);
CREATE INDEX idx_services_status ON services(status);
CREATE INDEX idx_services_is_featured ON services(is_featured);
