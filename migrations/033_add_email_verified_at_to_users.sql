-- Migration: 033_add_email_verified_at_to_users
-- Description: Add email verification timestamp to users table
-- Created: 2026-03-15

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP;
