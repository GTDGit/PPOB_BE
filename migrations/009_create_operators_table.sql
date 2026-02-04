-- Migration: 009_create_operators_table
-- Description: Create operators table for pulsa/data
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS operators (
    id VARCHAR(36) PRIMARY KEY,                          -- telkomsel, indosat, xl, axis, three, smartfren, byu
    name VARCHAR(50) NOT NULL,                           -- Telkomsel, Indosat Ooredoo, XL Axiata
    prefixes TEXT NOT NULL,                              -- JSON array: ["0811", "0812", "0813"]
    icon VARCHAR(50) NOT NULL,
    icon_url TEXT,
    status VARCHAR(20) DEFAULT 'active',                 -- active, maintenance, inactive
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_operators_status ON operators(status);
