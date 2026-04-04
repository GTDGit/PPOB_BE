-- Migration: 036_create_admin_mailboxes_and_ses_logs
-- Description: Shared inbox, personal mailboxes, inbound/outbound email tracking, and mailbox permissions
-- Created: 2026-04-04

CREATE TABLE IF NOT EXISTS admin_mailboxes (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(20) NOT NULL, -- system, shared, personal
    address VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(150) NOT NULL,
    owner_admin_id VARCHAR(36) REFERENCES admin_users(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_mailboxes_type ON admin_mailboxes(type);
CREATE INDEX IF NOT EXISTS idx_admin_mailboxes_owner_admin_id ON admin_mailboxes(owner_admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_mailboxes_is_active ON admin_mailboxes(is_active);

CREATE TABLE IF NOT EXISTS admin_mailbox_members (
    mailbox_id VARCHAR(36) NOT NULL REFERENCES admin_mailboxes(id) ON DELETE CASCADE,
    admin_user_id VARCHAR(36) NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    can_reply BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (mailbox_id, admin_user_id)
);

CREATE INDEX IF NOT EXISTS idx_admin_mailbox_members_admin_user_id ON admin_mailbox_members(admin_user_id);

CREATE TABLE IF NOT EXISTS admin_email_threads (
    id VARCHAR(36) PRIMARY KEY,
    mailbox_id VARCHAR(36) NOT NULL REFERENCES admin_mailboxes(id) ON DELETE CASCADE,
    participant_name VARCHAR(150),
    participant_email VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    normalized_subject VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'belum_dibalas',
    assigned_admin_id VARCHAR(36) REFERENCES admin_users(id) ON DELETE SET NULL,
    unread_count INT NOT NULL DEFAULT 0,
    last_direction VARCHAR(20),
    last_message_preview TEXT,
    latest_message_at TIMESTAMP,
    last_inbound_at TIMESTAMP,
    last_outbound_at TIMESTAMP,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_email_threads_mailbox_id ON admin_email_threads(mailbox_id);
CREATE INDEX IF NOT EXISTS idx_admin_email_threads_status ON admin_email_threads(status);
CREATE INDEX IF NOT EXISTS idx_admin_email_threads_assigned_admin_id ON admin_email_threads(assigned_admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_email_threads_participant_email ON admin_email_threads(participant_email);
CREATE INDEX IF NOT EXISTS idx_admin_email_threads_latest_message_at ON admin_email_threads(latest_message_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_email_threads_normalized_subject ON admin_email_threads(normalized_subject);

CREATE TABLE IF NOT EXISTS admin_email_messages (
    id VARCHAR(36) PRIMARY KEY,
    thread_id VARCHAR(36) NOT NULL REFERENCES admin_email_threads(id) ON DELETE CASCADE,
    mailbox_id VARCHAR(36) NOT NULL REFERENCES admin_mailboxes(id) ON DELETE CASCADE,
    direction VARCHAR(20) NOT NULL, -- inbound, outbound
    sender_name VARCHAR(150),
    sender_address VARCHAR(255) NOT NULL,
    to_addresses JSONB NOT NULL DEFAULT '[]'::jsonb,
    cc_addresses JSONB NOT NULL DEFAULT '[]'::jsonb,
    bcc_addresses JSONB NOT NULL DEFAULT '[]'::jsonb,
    subject VARCHAR(255) NOT NULL,
    text_body TEXT,
    html_body TEXT,
    provider_message_id VARCHAR(255),
    message_id_header VARCHAR(255),
    in_reply_to VARCHAR(255),
    references_headers JSONB NOT NULL DEFAULT '[]'::jsonb,
    sent_at TIMESTAMP,
    received_at TIMESTAMP,
    admin_user_id VARCHAR(36) REFERENCES admin_users(id) ON DELETE SET NULL,
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_email_messages_thread_id ON admin_email_messages(thread_id);
CREATE INDEX IF NOT EXISTS idx_admin_email_messages_mailbox_id ON admin_email_messages(mailbox_id);
CREATE INDEX IF NOT EXISTS idx_admin_email_messages_direction ON admin_email_messages(direction);
CREATE INDEX IF NOT EXISTS idx_admin_email_messages_provider_message_id ON admin_email_messages(provider_message_id);
CREATE INDEX IF NOT EXISTS idx_admin_email_messages_message_id_header ON admin_email_messages(message_id_header);
CREATE INDEX IF NOT EXISTS idx_admin_email_messages_received_at ON admin_email_messages(received_at DESC);

CREATE TABLE IF NOT EXISTS admin_email_attachments (
    id VARCHAR(36) PRIMARY KEY,
    message_id VARCHAR(36) NOT NULL REFERENCES admin_email_messages(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(120),
    size_bytes BIGINT NOT NULL DEFAULT 0,
    storage_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_email_attachments_message_id ON admin_email_attachments(message_id);

CREATE TABLE IF NOT EXISTS admin_email_thread_events (
    id VARCHAR(36) PRIMARY KEY,
    thread_id VARCHAR(36) NOT NULL REFERENCES admin_email_threads(id) ON DELETE CASCADE,
    actor_id VARCHAR(36) REFERENCES admin_users(id) ON DELETE SET NULL,
    event_type VARCHAR(40) NOT NULL, -- assigned, replied, status_changed, reopened
    notes TEXT,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_admin_email_thread_events_thread_id ON admin_email_thread_events(thread_id);

CREATE TABLE IF NOT EXISTS email_dispatch_logs (
    id VARCHAR(36) PRIMARY KEY,
    category VARCHAR(50) NOT NULL,
    mailbox_id VARCHAR(36) REFERENCES admin_mailboxes(id) ON DELETE SET NULL,
    thread_id VARCHAR(36) REFERENCES admin_email_threads(id) ON DELETE SET NULL,
    message_id VARCHAR(36) REFERENCES admin_email_messages(id) ON DELETE SET NULL,
    recipient VARCHAR(255) NOT NULL,
    sender_address VARCHAR(255) NOT NULL,
    sender_name VARCHAR(150),
    provider VARCHAR(20) NOT NULL,
    provider_message_id VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'queued',
    error_message TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    failed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_email_dispatch_logs_status ON email_dispatch_logs(status);
CREATE INDEX IF NOT EXISTS idx_email_dispatch_logs_category ON email_dispatch_logs(category);
CREATE INDEX IF NOT EXISTS idx_email_dispatch_logs_provider_message_id ON email_dispatch_logs(provider_message_id);
CREATE INDEX IF NOT EXISTS idx_email_dispatch_logs_mailbox_id ON email_dispatch_logs(mailbox_id);
CREATE INDEX IF NOT EXISTS idx_email_dispatch_logs_thread_id ON email_dispatch_logs(thread_id);
CREATE INDEX IF NOT EXISTS idx_email_dispatch_logs_created_at ON email_dispatch_logs(created_at DESC);

INSERT INTO admin_permissions (key, module, action, description) VALUES
('mailboxes.view_assigned', 'mailboxes', 'view_assigned', 'Lihat mailbox yang ditugaskan atau dimiliki'),
('mailboxes.view_all', 'mailboxes', 'view_all', 'Lihat seluruh mailbox operasional'),
('mailboxes.reply', 'mailboxes', 'reply', 'Balas email dari mailbox admin'),
('mailboxes.assign', 'mailboxes', 'assign', 'Assign thread inbox ke admin lain'),
('mailboxes.status.manage', 'mailboxes', 'status_manage', 'Ubah status thread inbox'),
('mailboxes.manage', 'mailboxes', 'manage', 'Kelola mailbox dan membership'),
('email_logs.view', 'email_logs', 'view', 'Lihat log pengiriman email')
ON CONFLICT (key) DO NOTHING;

INSERT INTO admin_roles (id, name, description, is_system, is_active) VALUES
('director', 'Director', 'Akses eksekutif untuk mailbox dan log email lintas tim', TRUE, TRUE),
('commissioner', 'Commissioner', 'Akses review eksekutif untuk mailbox dan log email', TRUE, TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO admin_role_permissions (role_id, permission_key)
SELECT 'super_admin', key FROM admin_permissions
ON CONFLICT DO NOTHING;

INSERT INTO admin_role_permissions (role_id, permission_key) VALUES
('customer_service', 'mailboxes.view_assigned'),
('customer_service', 'mailboxes.reply'),
('customer_service', 'mailboxes.assign'),
('customer_service', 'mailboxes.status.manage'),
('compliance_kyc', 'mailboxes.view_assigned'),
('compliance_kyc', 'mailboxes.reply'),
('compliance_kyc', 'mailboxes.status.manage'),
('product_content', 'mailboxes.view_assigned'),
('product_content', 'mailboxes.reply'),
('product_content', 'mailboxes.status.manage'),
('director', 'dashboard.view'),
('director', 'approvals.view'),
('director', 'audit.view'),
('director', 'settings.view'),
('director', 'mailboxes.view_all'),
('director', 'mailboxes.reply'),
('director', 'mailboxes.assign'),
('director', 'mailboxes.status.manage'),
('director', 'mailboxes.manage'),
('director', 'email_logs.view'),
('commissioner', 'dashboard.view'),
('commissioner', 'approvals.view'),
('commissioner', 'audit.view'),
('commissioner', 'mailboxes.view_all'),
('commissioner', 'mailboxes.reply'),
('commissioner', 'mailboxes.status.manage'),
('commissioner', 'email_logs.view')
ON CONFLICT DO NOTHING;

INSERT INTO admin_mailboxes (id, type, address, display_name, owner_admin_id, is_active, created_at, updated_at) VALUES
('ambx_system_noreply', 'system', 'noreply@ppob.id', 'PPOB.ID No Reply', NULL, TRUE, NOW(), NOW()),
('ambx_shared_cs', 'shared', 'cs@ppob.id', 'Customer Support', NULL, TRUE, NOW(), NOW()),
('ambx_shared_partner', 'shared', 'partner@ppob.id', 'Partner', NULL, TRUE, NOW(), NOW()),
('ambx_shared_partnership', 'shared', 'partnership@ppob.id', 'Partnership', NULL, TRUE, NOW(), NOW()),
('ambx_shared_dpo', 'shared', 'dpo@ppob.id', 'Data Protection Officer', NULL, TRUE, NOW(), NOW()),
('ambx_shared_legal', 'shared', 'legal@ppob.id', 'Legal', NULL, TRUE, NOW(), NOW()),
('ambx_shared_unmapped', 'shared', 'unmapped@ppob.id', 'Unmapped Inbox', NULL, TRUE, NOW(), NOW())
ON CONFLICT (address) DO NOTHING;

WITH normalized_admins AS (
    SELECT
        au.id,
        COALESCE(NULLIF(BTRIM(au.full_name), ''), SPLIT_PART(au.email, '@', 1)) AS source_name,
        ROW_NUMBER() OVER (
            PARTITION BY
                NULLIF(
                    BTRIM(
                        REGEXP_REPLACE(
                            REGEXP_REPLACE(
                                LOWER(COALESCE(NULLIF(BTRIM(au.full_name), ''), SPLIT_PART(au.email, '@', 1))),
                                '[^a-z0-9\s]+',
                                '',
                                'g'
                            ),
                            '\s+',
                            '.',
                            'g'
                        )
                    ),
                    ''
                )
            ORDER BY au.created_at, au.id
        ) AS duplicate_order,
        NULLIF(
            BTRIM(
                REGEXP_REPLACE(
                    REGEXP_REPLACE(
                        LOWER(COALESCE(NULLIF(BTRIM(au.full_name), ''), SPLIT_PART(au.email, '@', 1))),
                        '[^a-z0-9\s]+',
                        '',
                        'g'
                    ),
                    '\s+',
                    '.',
                    'g'
                )
            ),
            ''
        ) AS local_part
    FROM admin_users au
),
resolved_mailboxes AS (
    SELECT
        id,
        CASE
            WHEN COALESCE(local_part, '') = '' THEN 'admin.' || SUBSTRING(REPLACE(id, '-', ''), 1, 8)
            WHEN duplicate_order = 1 THEN local_part
            ELSE local_part || '.' || LPAD(duplicate_order::text, 2, '0')
        END AS resolved_local_part,
        source_name
    FROM normalized_admins
)
INSERT INTO admin_mailboxes (id, type, address, display_name, owner_admin_id, is_active, created_at, updated_at)
SELECT
    'ambx_' || SUBSTRING(REPLACE(id, '-', ''), 1, 28),
    'personal',
    resolved_local_part || '@ppob.id',
    source_name,
    id,
    TRUE,
    NOW(),
    NOW()
FROM resolved_mailboxes
ON CONFLICT (address) DO NOTHING;
