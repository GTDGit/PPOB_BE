-- migrations/023_add_phone_verified_at_to_users.sql
-- Add phone verification timestamp to users table

ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMP;

-- Add comment
COMMENT ON COLUMN users.phone_verified_at IS 'Timestamp when phone number was verified via OTP';

-- Optional: Set existing users with phone as verified (if migrating existing data)
-- UPDATE users SET phone_verified_at = created_at WHERE phone IS NOT NULL;
