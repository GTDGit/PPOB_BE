-- Migration: 012_create_transactions_table
-- Description: Create transactions table for all transaction types
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(36) PRIMARY KEY,                          -- trx_xxx
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    type VARCHAR(20) NOT NULL,                           -- prepaid, postpaid, transfer
    service_type VARCHAR(30) NOT NULL,                   -- pulsa, data, pln_prepaid, pln_postpaid, pdam, bpjs, etc

    -- Order/Inquiry info
    inquiry_id VARCHAR(50),
    order_id VARCHAR(50),

    -- Target info
    target VARCHAR(50) NOT NULL,                         -- Phone number, customer ID, account number
    target_name VARCHAR(100),                            -- Customer name
    target_operator VARCHAR(50),                         -- Telkomsel, BCA, etc

    -- Product info (for prepaid)
    product_id VARCHAR(36),
    product_name VARCHAR(100),
    nominal BIGINT,

    -- Billing info (for postpaid)
    bill_period VARCHAR(20),                             -- 202501
    bill_amount BIGINT,

    -- Transfer info
    bank_code VARCHAR(10),
    bank_name VARCHAR(50),
    account_number VARCHAR(30),
    account_name VARCHAR(100),
    transfer_note TEXT,

    -- Pricing
    price BIGINT NOT NULL,                               -- Product/transfer amount
    admin_fee BIGINT DEFAULT 0,
    voucher_discount BIGINT DEFAULT 0,
    total_payment BIGINT NOT NULL,

    -- Status
    status VARCHAR(20) DEFAULT 'pending',                -- pending, pending_payment, processing, success, failed, cancelled, refunded, expired
    status_message TEXT,

    -- Receipt
    serial_number VARCHAR(50),
    reference_number VARCHAR(50),
    token VARCHAR(50),                                   -- For PLN prepaid
    kwh VARCHAR(20),                                     -- For PLN prepaid

    -- Selling price (for agents)
    selling_price BIGINT,
    selling_payment_type VARCHAR(10),                    -- cash, credit

    -- Timestamps
    expires_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_service_type ON transactions(service_type);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_target ON transactions(target);
CREATE INDEX idx_transactions_inquiry_id ON transactions(inquiry_id);
CREATE INDEX idx_transactions_order_id ON transactions(order_id);
CREATE INDEX idx_transactions_reference_number ON transactions(reference_number);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_completed_at ON transactions(completed_at);
