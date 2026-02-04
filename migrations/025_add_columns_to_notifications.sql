-- migrations/025_add_columns_to_notifications.sql
-- Add missing columns to notifications table

ALTER TABLE notifications ADD COLUMN IF NOT EXISTS short_body TEXT;
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS action_button_text VARCHAR(50);
ALTER TABLE notifications ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Add comments
COMMENT ON COLUMN notifications.short_body IS 'Short summary of notification body for list view';
COMMENT ON COLUMN notifications.action_button_text IS 'Custom text for action button';
COMMENT ON COLUMN notifications.updated_at IS 'Last update timestamp';
