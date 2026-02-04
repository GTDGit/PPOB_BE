-- Migration: 015_create_deposits_table
-- Description: Create deposits tables for top-up/tambah saldo feature
-- Created: 2025-01-31
-- Updated: 2025-02-01

-- =====================================================
-- 1. DEPOSIT METHODS CONFIG
-- =====================================================
CREATE TABLE IF NOT EXISTS deposit_methods (
    code VARCHAR(30) PRIMARY KEY,                       -- bank_transfer, qris, retail, virtual_account
    name VARCHAR(50) NOT NULL,                          -- "Transfer Bank", "QRIS", etc
    description VARCHAR(255),
    icon VARCHAR(255),
    icon_url VARCHAR(255),
    
    -- Fee configuration
    fee_type VARCHAR(20) NOT NULL,                      -- fixed, percentage
    fee_amount BIGINT NOT NULL DEFAULT 0,               -- amount or percentage * 100 (0.7% = 70)
    
    -- Limits
    min_amount BIGINT NOT NULL DEFAULT 10000,
    max_amount BIGINT NOT NULL DEFAULT 100000000,
    
    -- Availability
    schedule VARCHAR(50),                               -- "07:00-20:45" or "24 Jam"
    available_start TIME,                               -- 07:00:00
    available_end TIME,                                 -- 20:45:00
    
    -- Ordering & status
    position INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',       -- active, maintenance, inactive
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default deposit methods
INSERT INTO deposit_methods (code, name, description, icon, icon_url, fee_type, fee_amount, min_amount, max_amount, schedule, available_start, available_end, position, status) VALUES
('bank_transfer', 'Transfer Bank', 'Transfer ke rekening perusahaan', 'bank_transfer', 'https://cdn.ppob.id/deposit/bank_transfer.png', 'fixed', 0, 10000, 100000000, '07:00-20:45 WIB', '07:00:00', '20:45:00', 1, 'active'),
('qris', 'QRIS', 'Scan QR untuk bayar', 'qris', 'https://cdn.ppob.id/deposit/qris.png', 'percentage', 70, 10000, 10000000, '24 Jam', NULL, NULL, 2, 'active'),
('retail', 'Alfamart/Indomaret', 'Bayar di gerai retail', 'retail', 'https://cdn.ppob.id/deposit/retail.png', 'fixed', 5000, 10000, 5000000, '24 Jam', NULL, NULL, 3, 'active'),
('virtual_account', 'Virtual Account', 'Transfer ke nomor VA', 'virtual_account', 'https://cdn.ppob.id/deposit/va.png', 'fixed', 3500, 10000, 50000000, '24 Jam', NULL, NULL, 4, 'active')
ON CONFLICT (code) DO NOTHING;

-- =====================================================
-- 2. COMPANY BANK ACCOUNTS (for bank transfer)
-- =====================================================
CREATE TABLE IF NOT EXISTS company_bank_accounts (
    id SERIAL PRIMARY KEY,
    bank_code VARCHAR(10) NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    bank_short_name VARCHAR(20) NOT NULL,
    bank_icon VARCHAR(255),
    account_number VARCHAR(30) NOT NULL,
    account_name VARCHAR(100) NOT NULL,
    position INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',       -- active, inactive
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_company_bank_accounts_status ON company_bank_accounts(status);

-- Insert sample company bank accounts (update with real data)
INSERT INTO company_bank_accounts (bank_code, bank_name, bank_short_name, bank_icon, account_number, account_name, position, status) VALUES
('014', 'Bank Central Asia', 'BCA', 'https://cdn.ppob.id/banks/bca.png', '2988398238', 'PT PPOB INDONESIA', 1, 'active'),
('002', 'Bank Rakyat Indonesia', 'BRI', 'https://cdn.ppob.id/banks/bri.png', '2988398238', 'PT PPOB INDONESIA', 2, 'active'),
('008', 'Bank Mandiri', 'Mandiri', 'https://cdn.ppob.id/banks/mandiri.png', '2988398238', 'PT PPOB INDONESIA', 3, 'active'),
('009', 'Bank Negara Indonesia', 'BNI', 'https://cdn.ppob.id/banks/bni.png', '2988398238', 'PT PPOB INDONESIA', 4, 'active')
ON CONFLICT DO NOTHING;

-- =====================================================
-- 3. DEPOSIT VA BANKS CONFIG
-- =====================================================
CREATE TABLE IF NOT EXISTS deposit_va_banks (
    code VARCHAR(10) PRIMARY KEY,                       -- Bank code (014, 002, etc)
    name VARCHAR(100) NOT NULL,
    short_name VARCHAR(20) NOT NULL,
    icon VARCHAR(255),
    icon_url VARCHAR(255),
    fee BIGINT NOT NULL DEFAULT 3500,
    min_amount BIGINT NOT NULL DEFAULT 10000,
    max_amount BIGINT NOT NULL DEFAULT 50000000,
    position INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert VA banks
INSERT INTO deposit_va_banks (code, name, short_name, icon, icon_url, fee, min_amount, max_amount, position, status) VALUES
('014', 'Bank Central Asia', 'BCA', 'bca', 'https://cdn.ppob.id/banks/bca.png', 4500, 10000, 50000000, 1, 'active'),
('002', 'Bank Rakyat Indonesia', 'BRI', 'bri', 'https://cdn.ppob.id/banks/bri.png', 3500, 10000, 50000000, 2, 'active'),
('008', 'Bank Mandiri', 'Mandiri', 'mandiri', 'https://cdn.ppob.id/banks/mandiri.png', 3500, 10000, 50000000, 3, 'active'),
('009', 'Bank Negara Indonesia', 'BNI', 'bni', 'https://cdn.ppob.id/banks/bni.png', 3500, 10000, 50000000, 4, 'active'),
('013', 'Bank Permata', 'Permata', 'permata', 'https://cdn.ppob.id/banks/permata.png', 3500, 10000, 50000000, 5, 'active'),
('022', 'Bank CIMB Niaga', 'CIMB', 'cimb', 'https://cdn.ppob.id/banks/cimb.png', 3500, 10000, 50000000, 6, 'active'),
('426', 'Bank BSI', 'BSI', 'bsi', 'https://cdn.ppob.id/banks/bsi.png', 3500, 10000, 50000000, 7, 'active')
ON CONFLICT (code) DO NOTHING;

-- =====================================================
-- 4. DEPOSIT RETAIL PROVIDERS CONFIG
-- =====================================================
CREATE TABLE IF NOT EXISTS deposit_retail_providers (
    code VARCHAR(30) PRIMARY KEY,                       -- alfamart, indomaret
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(255),
    icon_url VARCHAR(255),
    fee BIGINT NOT NULL DEFAULT 5000,
    min_amount BIGINT NOT NULL DEFAULT 10000,
    max_amount BIGINT NOT NULL DEFAULT 5000000,
    position INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert retail providers
INSERT INTO deposit_retail_providers (code, name, icon, icon_url, fee, min_amount, max_amount, position, status) VALUES
('alfamart', 'Alfamart / Alfamidi / Dan+Dan', 'alfamart', 'https://cdn.ppob.id/retail/alfamart.png', 5000, 10000, 5000000, 1, 'active'),
('indomaret', 'Indomaret / Isaku', 'indomaret', 'https://cdn.ppob.id/retail/indomaret.png', 5000, 10000, 5000000, 2, 'active')
ON CONFLICT (code) DO NOTHING;

-- =====================================================
-- 5. MAIN DEPOSITS TABLE
-- =====================================================
CREATE TABLE IF NOT EXISTS deposits (
    id VARCHAR(36) PRIMARY KEY,                         -- dep_xxx
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    method VARCHAR(30) NOT NULL,                        -- bank_transfer, qris, retail, virtual_account
    
    -- Provider/Bank (depends on method)
    provider_code VARCHAR(30),                          -- alfamart, indomaret (for retail)
    bank_code VARCHAR(10),                              -- for VA
    
    -- Amount
    amount BIGINT NOT NULL,                             -- Nominal deposit
    admin_fee BIGINT DEFAULT 0,                         -- Biaya admin
    unique_code INT DEFAULT 0,                          -- Kode unik (for bank_transfer)
    total_amount BIGINT NOT NULL,                       -- Total yang harus dibayar
    
    -- Status
    status VARCHAR(20) DEFAULT 'pending',               -- pending, success, expired, failed
    
    -- Payment data (JSON) - stores method-specific data
    -- For bank_transfer: company accounts list
    -- For QRIS: qris_data, qris_image_url
    -- For retail: payment_code
    -- For VA: va_number, va_display
    payment_data JSONB,
    
    -- External reference (from Gerbang API)
    external_id VARCHAR(100),
    reference_number VARCHAR(50),
    
    -- Payer info (filled after payment)
    payer_name VARCHAR(100),
    payer_bank VARCHAR(50),
    payer_account VARCHAR(30),
    
    -- Timestamps
    expires_at TIMESTAMP NOT NULL,
    paid_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_deposits_user_id ON deposits(user_id);
CREATE INDEX IF NOT EXISTS idx_deposits_method ON deposits(method);
CREATE INDEX IF NOT EXISTS idx_deposits_status ON deposits(status);
CREATE INDEX IF NOT EXISTS idx_deposits_external_id ON deposits(external_id);
CREATE INDEX IF NOT EXISTS idx_deposits_reference_number ON deposits(reference_number);
CREATE INDEX IF NOT EXISTS idx_deposits_created_at ON deposits(created_at);
CREATE INDEX IF NOT EXISTS idx_deposits_user_status ON deposits(user_id, status);

-- =====================================================
-- 6. UNIQUE CODE TRACKING (for bank transfer)
-- =====================================================
-- Tracks used unique codes to avoid duplicates within a period
CREATE TABLE IF NOT EXISTS deposit_unique_codes (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id),
    unique_code INT NOT NULL,
    amount BIGINT NOT NULL,
    deposit_id VARCHAR(36) REFERENCES deposits(id),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_deposit_unique_codes_user ON deposit_unique_codes(user_id);
CREATE INDEX idx_deposit_unique_codes_expires ON deposit_unique_codes(expires_at);
CREATE UNIQUE INDEX idx_deposit_unique_codes_active ON deposit_unique_codes(user_id, unique_code, amount) 
    WHERE expires_at > CURRENT_TIMESTAMP;

-- =====================================================
-- HELPER: Function to generate unique code
-- =====================================================
CREATE OR REPLACE FUNCTION generate_deposit_unique_code(
    p_user_id VARCHAR(36),
    p_amount BIGINT,
    p_expires_at TIMESTAMP
) RETURNS INT AS $$
DECLARE
    v_code INT;
    v_attempts INT := 0;
    v_max_attempts INT := 100;
BEGIN
    -- Try to find an unused code
    LOOP
        v_code := floor(random() * 900 + 100)::INT; -- 100-999
        
        -- Check if code is available
        IF NOT EXISTS (
            SELECT 1 FROM deposit_unique_codes 
            WHERE user_id = p_user_id 
            AND unique_code = v_code 
            AND amount = p_amount
            AND expires_at > CURRENT_TIMESTAMP
        ) THEN
            -- Code is available, return it
            RETURN v_code;
        END IF;
        
        v_attempts := v_attempts + 1;
        IF v_attempts >= v_max_attempts THEN
            -- Fallback: use timestamp-based code
            RETURN (extract(epoch from now())::INT % 900) + 100;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
