-- migrations/024_add_columns_to_deposits.sql
-- Add missing columns to deposits table for payment tracking

ALTER TABLE deposits ADD COLUMN IF NOT EXISTS reference_number VARCHAR(50);
ALTER TABLE deposits ADD COLUMN IF NOT EXISTS payer_name VARCHAR(100);
ALTER TABLE deposits ADD COLUMN IF NOT EXISTS payer_bank VARCHAR(50);
ALTER TABLE deposits ADD COLUMN IF NOT EXISTS payer_account VARCHAR(30);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_deposits_reference_number ON deposits(reference_number);

-- Add comments
COMMENT ON COLUMN deposits.reference_number IS 'Payment reference number from payment gateway';
COMMENT ON COLUMN deposits.payer_name IS 'Name of payer (from payment notification)';
COMMENT ON COLUMN deposits.payer_bank IS 'Bank used by payer';
COMMENT ON COLUMN deposits.payer_account IS 'Account number used by payer';
