-- Migration: 029_update_transfer_tables
-- Description: Add Gerbang API integration fields to transfer tables

-- Add new columns to transfer_inquiries
ALTER TABLE transfer_inquiries
ADD COLUMN IF NOT EXISTS gerbang_inquiry_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS fee BIGINT DEFAULT 0;

-- Add new columns to transfer_transactions
ALTER TABLE transfer_transactions
ADD COLUMN IF NOT EXISTS gerbang_transfer_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS purpose VARCHAR(10) DEFAULT '99',
ADD COLUMN IF NOT EXISTS fee BIGINT DEFAULT 0;

-- Index for gerbang IDs (for faster lookups when processing webhooks)
CREATE INDEX IF NOT EXISTS idx_transfer_inquiries_gerbang_id ON transfer_inquiries(gerbang_inquiry_id);
CREATE INDEX IF NOT EXISTS idx_transfer_transactions_gerbang_id ON transfer_transactions(gerbang_transfer_id);

-- Comments
COMMENT ON COLUMN transfer_inquiries.gerbang_inquiry_id IS 'Inquiry ID from Gerbang API response';
COMMENT ON COLUMN transfer_inquiries.fee IS 'Actual fee from Gerbang API (may differ from admin_fee estimate)';
COMMENT ON COLUMN transfer_transactions.gerbang_transfer_id IS 'Transfer ID from Gerbang API response';
COMMENT ON COLUMN transfer_transactions.purpose IS 'Transfer purpose code (01=Investasi, 02=Pemindahan Dana, 03=Pembelian, 99=Lainnya)';
COMMENT ON COLUMN transfer_transactions.fee IS 'Actual fee charged by Gerbang API';
