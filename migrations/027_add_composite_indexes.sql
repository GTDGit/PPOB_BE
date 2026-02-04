-- Migration: 027_add_composite_indexes
-- Description: Add composite indexes for common query patterns to improve performance

-- Sessions: Active session lookup by user+device
CREATE INDEX IF NOT EXISTS idx_sessions_user_device_active 
ON sessions(user_id, device_id, expires_at) 
WHERE is_revoked = false;

-- Notifications: User's unread notifications (most common query)
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread 
ON notifications(user_id, created_at DESC) 
WHERE is_read = false;

-- Prepaid Transactions: User's transaction history (pagination)
CREATE INDEX IF NOT EXISTS idx_prepaid_transactions_user_created 
ON prepaid_transactions(user_id, created_at DESC);

-- Postpaid Transactions: User's transaction history (pagination)
CREATE INDEX IF NOT EXISTS idx_postpaid_transactions_user_created 
ON postpaid_transactions(user_id, created_at DESC);

-- Transfer Transactions: User's transaction history (pagination)
CREATE INDEX IF NOT EXISTS idx_transfer_transactions_user_created 
ON transfer_transactions(user_id, created_at DESC);

-- Deposits: User's deposit history with created_at sort
CREATE INDEX IF NOT EXISTS idx_deposits_user_created 
ON deposits(user_id, created_at DESC);

-- Devices: User's active devices
CREATE INDEX IF NOT EXISTS idx_devices_user_active 
ON devices(user_id, last_active_at DESC) 
WHERE is_active = true;

-- Balances: User balance lookup (already has single column index, but this helps with joins)
CREATE INDEX IF NOT EXISTS idx_balances_user 
ON balances(user_id);

-- Prepaid Orders: User order lookup with status
CREATE INDEX IF NOT EXISTS idx_prepaid_orders_user_status 
ON prepaid_orders(user_id, status, created_at DESC);

-- Postpaid Inquiries: User inquiry lookup
CREATE INDEX IF NOT EXISTS idx_postpaid_inquiries_user 
ON postpaid_inquiries(user_id, created_at DESC);

-- Transfer Inquiries: User inquiry lookup
CREATE INDEX IF NOT EXISTS idx_transfer_inquiries_user 
ON transfer_inquiries(user_id, created_at DESC);

-- Contacts: User contacts lookup
CREATE INDEX IF NOT EXISTS idx_contacts_user 
ON contacts(user_id, created_at DESC);

-- User Settings: User settings lookup (unique per user)
CREATE INDEX IF NOT EXISTS idx_user_settings_user 
ON user_settings(user_id);

-- OTP Sessions: Phone lookup for active sessions
CREATE INDEX IF NOT EXISTS idx_otp_sessions_phone_active 
ON otp_sessions(phone, expires_at DESC) 
WHERE is_verified = false;

-- Comment on indexes
COMMENT ON INDEX idx_sessions_user_device_active IS 'Optimizes active session lookup by user and device';
COMMENT ON INDEX idx_notifications_user_unread IS 'Optimizes unread notification queries';
COMMENT ON INDEX idx_prepaid_transactions_user_created IS 'Optimizes prepaid transaction history pagination';
COMMENT ON INDEX idx_postpaid_transactions_user_created IS 'Optimizes postpaid transaction history pagination';
COMMENT ON INDEX idx_transfer_transactions_user_created IS 'Optimizes transfer transaction history pagination';
COMMENT ON INDEX idx_deposits_user_created IS 'Optimizes deposit history queries';
COMMENT ON INDEX idx_devices_user_active IS 'Optimizes active device lookups';
