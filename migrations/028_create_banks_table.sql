-- Migration: 028_create_banks_table
-- Description: Add missing columns to banks table (already created in 011)
-- Note: banks table was created in 011_create_providers_tables.sql
--       This migration adds columns that may be missing from older versions

-- Add swift_code column if not exists
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'banks' AND column_name = 'swift_code') THEN
        ALTER TABLE banks ADD COLUMN swift_code VARCHAR(20);
    END IF;
END $$;

-- Add transfer_fee_formatted column if not exists
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'banks' AND column_name = 'transfer_fee_formatted') THEN
        ALTER TABLE banks ADD COLUMN transfer_fee_formatted VARCHAR(50) DEFAULT '';
    END IF;
END $$;

-- Alter existing columns to match new schema (safe operations)
ALTER TABLE banks ALTER COLUMN short_name TYPE VARCHAR(50);
ALTER TABLE banks ALTER COLUMN icon TYPE VARCHAR(255);
ALTER TABLE banks ALTER COLUMN icon SET DEFAULT '';
ALTER TABLE banks ALTER COLUMN icon_url TYPE VARCHAR(255);
ALTER TABLE banks ALTER COLUMN icon_url SET DEFAULT '';

-- Index for searching by name (if not exists)
CREATE INDEX IF NOT EXISTS idx_banks_name ON banks(name);

-- Index for popular banks filter (if not exists)
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
