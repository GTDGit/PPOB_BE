-- migrations/021_create_prepaid_postpaid_transfer_tables.sql
-- Create missing tables for Prepaid, Postpaid, and Transfer operations

-- =====================================================
-- PREPAID TABLES
-- =====================================================
CREATE TABLE IF NOT EXISTS prepaid_inquiries (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_type VARCHAR(30) NOT NULL,
    target VARCHAR(50) NOT NULL,
    operator_id VARCHAR(36),
    operator_name VARCHAR(50),
    product_id VARCHAR(36),
    product_name VARCHAR(100),
    product_price BIGINT,
    admin_fee BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_prepaid_inquiries_user ON prepaid_inquiries(user_id);
CREATE INDEX idx_prepaid_inquiries_expires ON prepaid_inquiries(expires_at);

CREATE TABLE IF NOT EXISTS prepaid_orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    inquiry_id VARCHAR(36) REFERENCES prepaid_inquiries(id),
    service_type VARCHAR(30) NOT NULL,
    target VARCHAR(50) NOT NULL,
    operator_id VARCHAR(36),
    operator_name VARCHAR(50),
    product_id VARCHAR(36) NOT NULL,
    product_name VARCHAR(100) NOT NULL,
    product_price BIGINT NOT NULL,
    admin_fee BIGINT DEFAULT 0,
    voucher_id VARCHAR(36),
    voucher_discount BIGINT DEFAULT 0,
    total_payment BIGINT NOT NULL,
    pin_verified BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_prepaid_orders_user ON prepaid_orders(user_id);
CREATE INDEX idx_prepaid_orders_status ON prepaid_orders(status);

CREATE TABLE IF NOT EXISTS prepaid_transactions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id VARCHAR(36) REFERENCES prepaid_orders(id),
    service_type VARCHAR(30) NOT NULL,
    target VARCHAR(50) NOT NULL,
    operator_id VARCHAR(36),
    operator_name VARCHAR(50),
    product_id VARCHAR(36) NOT NULL,
    product_name VARCHAR(100) NOT NULL,
    product_price BIGINT NOT NULL,
    admin_fee BIGINT DEFAULT 0,
    voucher_discount BIGINT DEFAULT 0,
    total_payment BIGINT NOT NULL,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    reference_number VARCHAR(50),
    serial_number VARCHAR(100),
    external_id VARCHAR(100),
    status VARCHAR(20) DEFAULT 'processing',
    failed_reason TEXT,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_prepaid_transactions_user ON prepaid_transactions(user_id);
CREATE INDEX idx_prepaid_transactions_status ON prepaid_transactions(status);
CREATE INDEX idx_prepaid_transactions_created ON prepaid_transactions(created_at DESC);

-- =====================================================
-- POSTPAID TABLES
-- =====================================================
CREATE TABLE IF NOT EXISTS postpaid_inquiries (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_type VARCHAR(30) NOT NULL,
    target VARCHAR(50) NOT NULL,
    provider_id VARCHAR(36),
    customer_id VARCHAR(100),
    customer_name VARCHAR(100),
    period VARCHAR(50),
    bill_amount BIGINT,
    admin_fee BIGINT DEFAULT 0,
    penalty BIGINT DEFAULT 0,
    total_payment BIGINT,
    has_bill BOOLEAN DEFAULT FALSE,
    external_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_postpaid_inquiries_user ON postpaid_inquiries(user_id);
CREATE INDEX idx_postpaid_inquiries_expires ON postpaid_inquiries(expires_at);

CREATE TABLE IF NOT EXISTS postpaid_transactions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    inquiry_id VARCHAR(36) REFERENCES postpaid_inquiries(id),
    service_type VARCHAR(30) NOT NULL,
    target VARCHAR(50) NOT NULL,
    provider_id VARCHAR(36),
    customer_id VARCHAR(100),
    customer_name VARCHAR(100),
    period VARCHAR(50),
    bill_amount BIGINT NOT NULL,
    admin_fee BIGINT DEFAULT 0,
    penalty BIGINT DEFAULT 0,
    voucher_discount BIGINT DEFAULT 0,
    total_payment BIGINT NOT NULL,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    reference_number VARCHAR(50),
    serial_number VARCHAR(100),
    external_id VARCHAR(100),
    status VARCHAR(20) DEFAULT 'processing',
    failed_reason TEXT,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_postpaid_transactions_user ON postpaid_transactions(user_id);
CREATE INDEX idx_postpaid_transactions_status ON postpaid_transactions(status);
CREATE INDEX idx_postpaid_transactions_created ON postpaid_transactions(created_at DESC);

-- =====================================================
-- TRANSFER TABLES
-- =====================================================
CREATE TABLE IF NOT EXISTS transfer_inquiries (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bank_code VARCHAR(10) NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    account_number VARCHAR(30) NOT NULL,
    account_name VARCHAR(100),
    amount BIGINT NOT NULL,
    admin_fee BIGINT DEFAULT 0,
    total_payment BIGINT NOT NULL,
    validated BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_transfer_inquiries_user ON transfer_inquiries(user_id);
CREATE INDEX idx_transfer_inquiries_expires ON transfer_inquiries(expires_at);

CREATE TABLE IF NOT EXISTS transfer_transactions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    inquiry_id VARCHAR(36) REFERENCES transfer_inquiries(id),
    bank_code VARCHAR(10) NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    account_number VARCHAR(30) NOT NULL,
    account_name VARCHAR(100) NOT NULL,
    amount BIGINT NOT NULL,
    admin_fee BIGINT DEFAULT 0,
    voucher_discount BIGINT DEFAULT 0,
    total_payment BIGINT NOT NULL,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    reference_number VARCHAR(50),
    external_id VARCHAR(100),
    note TEXT,
    status VARCHAR(20) DEFAULT 'processing',
    failed_reason TEXT,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transfer_transactions_user ON transfer_transactions(user_id);
CREATE INDEX idx_transfer_transactions_status ON transfer_transactions(status);
CREATE INDEX idx_transfer_transactions_created ON transfer_transactions(created_at DESC);

-- =====================================================
-- COMMENTS
-- =====================================================
COMMENT ON TABLE prepaid_inquiries IS 'Stores prepaid product inquiry results';
COMMENT ON TABLE prepaid_orders IS 'Stores prepaid orders awaiting payment';
COMMENT ON TABLE prepaid_transactions IS 'Stores completed prepaid transactions';
COMMENT ON TABLE postpaid_inquiries IS 'Stores postpaid bill inquiry results';
COMMENT ON TABLE postpaid_transactions IS 'Stores completed postpaid bill payments';
COMMENT ON TABLE transfer_inquiries IS 'Stores bank transfer inquiry results';
COMMENT ON TABLE transfer_transactions IS 'Stores completed bank transfer transactions';
