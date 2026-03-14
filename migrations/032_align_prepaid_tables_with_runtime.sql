-- Migration: 032_align_prepaid_tables_with_runtime
-- Description: Align prepaid tables with fields used by current service/runtime code

ALTER TABLE prepaid_inquiries
ADD COLUMN IF NOT EXISTS target_valid BOOLEAN DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS customer_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS customer_name VARCHAR(100);

ALTER TABLE prepaid_orders
ADD COLUMN IF NOT EXISTS subtotal BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_discount BIGINT DEFAULT 0,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE prepaid_transactions
ADD COLUMN IF NOT EXISTS token VARCHAR(255),
ADD COLUMN IF NOT EXISTS kwh VARCHAR(100);

UPDATE prepaid_orders
SET updated_at = COALESCE(updated_at, created_at)
WHERE updated_at IS NULL;
