-- Add avatar_url column to admin_users table for profile photos
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS avatar_url TEXT DEFAULT NULL;
