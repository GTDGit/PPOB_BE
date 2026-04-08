-- Migration: 037_standardize_internal_ids_and_public_codes
-- Description: Add public reference IDs and refund records for core business entities
-- Created: 2026-04-06

CREATE SEQUENCE IF NOT EXISTS seq_user_mic START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE IF NOT EXISTS seq_public_order START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE IF NOT EXISTS seq_public_transaction START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE IF NOT EXISTS seq_public_deposit START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE IF NOT EXISTS seq_public_refund START WITH 1 INCREMENT BY 1;
CREATE SEQUENCE IF NOT EXISTS seq_public_qris START WITH 1 INCREMENT BY 1;

ALTER TABLE deposits ADD COLUMN IF NOT EXISTS public_id VARCHAR(32);
ALTER TABLE prepaid_orders ADD COLUMN IF NOT EXISTS public_id VARCHAR(32);
ALTER TABLE prepaid_transactions ADD COLUMN IF NOT EXISTS public_id VARCHAR(32);
ALTER TABLE postpaid_transactions ADD COLUMN IF NOT EXISTS public_id VARCHAR(32);
ALTER TABLE transfer_transactions ADD COLUMN IF NOT EXISTS public_id VARCHAR(32);
ALTER TABLE qris_incomes ADD COLUMN IF NOT EXISTS public_id VARCHAR(32);
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS public_id VARCHAR(32);

CREATE UNIQUE INDEX IF NOT EXISTS idx_deposits_public_id_unique ON deposits(public_id) WHERE public_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_prepaid_orders_public_id_unique ON prepaid_orders(public_id) WHERE public_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_prepaid_transactions_public_id_unique ON prepaid_transactions(public_id) WHERE public_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_postpaid_transactions_public_id_unique ON postpaid_transactions(public_id) WHERE public_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_transfer_transactions_public_id_unique ON transfer_transactions(public_id) WHERE public_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_qris_incomes_public_id_unique ON qris_incomes(public_id) WHERE public_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_public_id_unique ON transactions(public_id) WHERE public_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_deposits_public_id ON deposits(public_id);
CREATE INDEX IF NOT EXISTS idx_prepaid_orders_public_id ON prepaid_orders(public_id);
CREATE INDEX IF NOT EXISTS idx_prepaid_transactions_public_id ON prepaid_transactions(public_id);
CREATE INDEX IF NOT EXISTS idx_postpaid_transactions_public_id ON postpaid_transactions(public_id);
CREATE INDEX IF NOT EXISTS idx_transfer_transactions_public_id ON transfer_transactions(public_id);
CREATE INDEX IF NOT EXISTS idx_qris_incomes_public_id ON qris_incomes(public_id);
CREATE INDEX IF NOT EXISTS idx_transactions_public_id ON transactions(public_id);

CREATE TABLE IF NOT EXISTS refunds (
    id VARCHAR(36) PRIMARY KEY,
    public_id VARCHAR(32) UNIQUE NOT NULL,
    source_transaction_id VARCHAR(36) NOT NULL,
    source_type VARCHAR(30) NOT NULL,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    amount BIGINT NOT NULL,
    reason TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    refunded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_refunds_public_id ON refunds(public_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_refunds_source_transaction_id_unique ON refunds(source_transaction_id);
CREATE INDEX IF NOT EXISTS idx_refunds_source_transaction_id ON refunds(source_transaction_id);
CREATE INDEX IF NOT EXISTS idx_refunds_user_id ON refunds(user_id);
CREATE INDEX IF NOT EXISTS idx_refunds_status ON refunds(status);
