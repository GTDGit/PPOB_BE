-- Migration: 014_create_contacts_table
-- Description: Create contacts table for favorite contacts
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS contacts (
    id VARCHAR(36) PRIMARY KEY,                          -- cnt_xxx
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,                          -- Ayah, Rumah, Budi BCA
    type VARCHAR(20) NOT NULL,                           -- phone, pln, pdam, bpjs, telkom, bank
    value VARCHAR(50) NOT NULL,                          -- Phone number, customer ID, account number

    -- Additional metadata based on type
    operator_id VARCHAR(36),                             -- For phone type
    operator_name VARCHAR(50),
    bank_code VARCHAR(10),                               -- For bank type
    bank_name VARCHAR(50),
    account_name VARCHAR(100),                           -- For bank type
    customer_name VARCHAR(100),                          -- For PLN, PDAM, etc

    last_used_at TIMESTAMP,
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, type, value)
);

-- Indexes
CREATE INDEX idx_contacts_user_id ON contacts(user_id);
CREATE INDEX idx_contacts_type ON contacts(type);
CREATE INDEX idx_contacts_value ON contacts(value);
CREATE INDEX idx_contacts_last_used_at ON contacts(last_used_at);
