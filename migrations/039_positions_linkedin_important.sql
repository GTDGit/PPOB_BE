-- Position management table
CREATE TABLE IF NOT EXISTS admin_positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add position and linkedin fields to admin_users
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS position_id UUID REFERENCES admin_positions(id) ON DELETE SET NULL;
ALTER TABLE admin_users ADD COLUMN IF NOT EXISTS linkedin_url TEXT DEFAULT NULL;

-- Add is_important flag to email threads
ALTER TABLE admin_email_threads ADD COLUMN IF NOT EXISTS is_important BOOLEAN NOT NULL DEFAULT FALSE;
