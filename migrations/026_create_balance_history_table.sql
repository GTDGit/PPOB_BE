-- migrations/026_create_balance_history_table.sql
-- Create balance history table for tracking all balance mutations

CREATE TABLE IF NOT EXISTS balance_history (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL,                       -- credit, debit
    category VARCHAR(20) NOT NULL,                   -- transaction, deposit, refund, bonus, points
    amount BIGINT NOT NULL,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    reference_type VARCHAR(20),                      -- deposit, prepaid, postpaid, transfer, voucher
    reference_id VARCHAR(36),                        -- ID of related record
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_balance_history_user_id ON balance_history(user_id);
CREATE INDEX idx_balance_history_type ON balance_history(type);
CREATE INDEX idx_balance_history_category ON balance_history(category);
CREATE INDEX idx_balance_history_created ON balance_history(created_at DESC);
CREATE INDEX idx_balance_history_reference ON balance_history(reference_type, reference_id);

-- Add constraint to ensure valid types and categories
ALTER TABLE balance_history ADD CONSTRAINT chk_balance_history_type 
    CHECK (type IN ('credit', 'debit'));

ALTER TABLE balance_history ADD CONSTRAINT chk_balance_history_category 
    CHECK (category IN ('transaction', 'deposit', 'refund', 'bonus', 'points', 'fee'));

-- Comments
COMMENT ON TABLE balance_history IS 'Tracks all balance mutations for audit trail';
COMMENT ON COLUMN balance_history.type IS 'Type of mutation: credit (increase) or debit (decrease)';
COMMENT ON COLUMN balance_history.category IS 'Category: transaction, deposit, refund, bonus, points, fee';
COMMENT ON COLUMN balance_history.reference_type IS 'Type of related record: deposit, prepaid, postpaid, transfer, voucher';
COMMENT ON COLUMN balance_history.reference_id IS 'ID of the related record';
