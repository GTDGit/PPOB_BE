-- Migration: 001_create_users_table
-- Description: Create users table
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,                          -- UUID: usr_xxx
    mic VARCHAR(20) UNIQUE NOT NULL,                     -- Merchant Identification Code: PID12345
    phone VARCHAR(15) UNIQUE NOT NULL,                   -- Format: 08xxxxxxxxxx
    full_name VARCHAR(100) NOT NULL,
    email VARCHAR(100),
    gender VARCHAR(10),                                  -- MALE, FEMALE
    tier VARCHAR(20) DEFAULT 'BRONZE',                   -- BRONZE, SILVER, GOLD, PLATINUM
    avatar_url TEXT,
    kyc_status VARCHAR(20) DEFAULT 'unverified',         -- unverified, pending, verified, rejected
    business_type VARCHAR(30),                           -- SHOP, COUNTER, MONEY_AGENT, ONLINE_SELLER, NONE, OTHER
    source VARCHAR(30),                                  -- SALES, FRIEND, GOOGLE, PLAYSTORE_APPSTORE, ADS, SOCIAL_MEDIA, OUTDOOR, OTHER
    referred_by VARCHAR(10),                             -- USER_ID, SALES_ID, null
    referral_code VARCHAR(20) UNIQUE,                    -- User's own referral code
    used_referral_code VARCHAR(20),                      -- Referral code used during registration
    pin_hash VARCHAR(255),                               -- Bcrypt hash
    is_active BOOLEAN DEFAULT TRUE,
    is_locked BOOLEAN DEFAULT FALSE,
    locked_until TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_mic ON users(mic);
CREATE INDEX idx_users_referral_code ON users(referral_code);
CREATE INDEX idx_users_tier ON users(tier);
CREATE INDEX idx_users_kyc_status ON users(kyc_status);
CREATE INDEX idx_users_created_at ON users(created_at);
