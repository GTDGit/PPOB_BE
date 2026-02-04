-- Migration: 010_create_products_table
-- Description: Create products table for all service types
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) PRIMARY KEY,                          -- prd_pulsa_5k, prd_pln_50k
    operator_id VARCHAR(36) REFERENCES operators(id),    -- For pulsa/data
    provider_id VARCHAR(36),                             -- For ewallet, game, etc
    service_type VARCHAR(30) NOT NULL,                   -- pulsa, data, pln_prepaid, ewallet, game
    name VARCHAR(100) NOT NULL,                          -- Pulsa 5.000, Token 50.000
    description TEXT,                                    -- Masa Aktif 7 Hari
    category VARCHAR(50),                                -- pulsa, data, token, voucher
    nominal BIGINT NOT NULL,                             -- 5000, 10000, 50000
    price BIGINT NOT NULL,                               -- Base price
    admin_fee BIGINT DEFAULT 0,                          -- Admin fee
    discount_type VARCHAR(20),                           -- fixed, percentage, null
    discount_value INTEGER DEFAULT 0,                    -- Discount amount or percentage
    status VARCHAR(20) DEFAULT 'active',                 -- active, maintenance, out_of_stock
    stock VARCHAR(20) DEFAULT 'available',               -- available, limited, empty
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_products_operator_id ON products(operator_id);
CREATE INDEX idx_products_provider_id ON products(provider_id);
CREATE INDEX idx_products_service_type ON products(service_type);
CREATE INDEX idx_products_status ON products(status);
