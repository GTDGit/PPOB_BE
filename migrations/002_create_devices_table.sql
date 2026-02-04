-- Migration: 002_create_devices_table
-- Description: Create user devices table for multi-device management
-- Created: 2025-01-31

CREATE TABLE IF NOT EXISTS devices (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(100) NOT NULL,                     -- UUID from device
    device_name VARCHAR(100) NOT NULL,                   -- e.g., "Samsung A54"
    platform VARCHAR(20),                                -- android, ios, web
    last_active_at TIMESTAMP,
    location VARCHAR(100),                               -- e.g., "Jakarta, Indonesia"
    ip_address VARCHAR(45),                              -- IPv4/IPv6
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, device_id)
);

-- Indexes
CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_device_id ON devices(device_id);
CREATE INDEX idx_devices_last_active_at ON devices(last_active_at);
