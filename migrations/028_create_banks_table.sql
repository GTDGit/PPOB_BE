-- Migration: 028_create_banks_table
-- Description: Create banks table for storing bank information from Gerbang API

CREATE TABLE IF NOT EXISTS banks (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    short_name VARCHAR(50) NOT NULL,
    swift_code VARCHAR(20),
    icon VARCHAR(255) DEFAULT '',
    icon_url VARCHAR(255) DEFAULT '',
    transfer_fee BIGINT DEFAULT 0,
    transfer_fee_formatted VARCHAR(50) DEFAULT '',
    is_popular BOOLEAN DEFAULT false,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for fast lookup by code (most common query)
CREATE INDEX IF NOT EXISTS idx_banks_code ON banks(code);

-- Index for searching by name
CREATE INDEX IF NOT EXISTS idx_banks_name ON banks(name);

-- Index for filtering by status
CREATE INDEX IF NOT EXISTS idx_banks_status ON banks(status);

-- Index for popular banks filter
CREATE INDEX IF NOT EXISTS idx_banks_popular ON banks(is_popular) WHERE is_popular = true;

-- Comment on table and columns
COMMENT ON TABLE banks IS 'Stores bank information synced from Gerbang API';
COMMENT ON COLUMN banks.id IS 'Unique identifier (UUID)';
COMMENT ON COLUMN banks.code IS 'Bank code (3 digits, unique)';
COMMENT ON COLUMN banks.name IS 'Full bank name (e.g., Bank BRI)';
COMMENT ON COLUMN banks.short_name IS 'Short bank name (e.g., BRI)';
COMMENT ON COLUMN banks.swift_code IS 'SWIFT/BIC code (optional)';
COMMENT ON COLUMN banks.icon IS 'Icon filename';
COMMENT ON COLUMN banks.icon_url IS 'Icon URL';
COMMENT ON COLUMN banks.transfer_fee IS 'Transfer fee in smallest currency unit';
COMMENT ON COLUMN banks.transfer_fee_formatted IS 'Formatted transfer fee (e.g., Rp 2.500)';
COMMENT ON COLUMN banks.is_popular IS 'Whether bank is marked as popular';
COMMENT ON COLUMN banks.status IS 'Bank status (active/inactive)';
COMMENT ON COLUMN banks.created_at IS 'Record creation timestamp';
COMMENT ON COLUMN banks.updated_at IS 'Last update timestamp (from sync)';
