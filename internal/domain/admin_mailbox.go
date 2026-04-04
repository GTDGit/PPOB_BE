package domain

import (
	"database/sql"
	"time"
)

const (
	AdminMailboxTypeSystem   = "system"
	AdminMailboxTypeShared   = "shared"
	AdminMailboxTypePersonal = "personal"

	AdminEmailDirectionInbound  = "inbound"
	AdminEmailDirectionOutbound = "outbound"

	AdminEmailThreadStatusBelumDibalas = "belum_dibalas"
	AdminEmailThreadStatusDibalas      = "dibalas"
	AdminEmailThreadStatusSelesai      = "selesai"
	AdminEmailThreadStatusArsip        = "arsip"
)

type AdminMailbox struct {
	ID           string         `db:"id" json:"id"`
	Type         string         `db:"type" json:"type"`
	Address      string         `db:"address" json:"address"`
	DisplayName  string         `db:"display_name" json:"displayName"`
	OwnerAdminID sql.NullString `db:"owner_admin_id" json:"ownerAdminId"`
	IsActive     bool           `db:"is_active" json:"isActive"`
	CreatedAt    time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updatedAt"`
}

type AdminEmailThread struct {
	ID                 string         `db:"id" json:"id"`
	MailboxID          string         `db:"mailbox_id" json:"mailboxId"`
	ParticipantName    sql.NullString `db:"participant_name" json:"participantName"`
	ParticipantEmail   string         `db:"participant_email" json:"participantEmail"`
	Subject            string         `db:"subject" json:"subject"`
	NormalizedSubject  string         `db:"normalized_subject" json:"normalizedSubject"`
	Status             string         `db:"status" json:"status"`
	AssignedAdminID    sql.NullString `db:"assigned_admin_id" json:"assignedAdminId"`
	UnreadCount        int            `db:"unread_count" json:"unreadCount"`
	LastDirection      sql.NullString `db:"last_direction" json:"lastDirection"`
	LastMessagePreview sql.NullString `db:"last_message_preview" json:"lastMessagePreview"`
	LatestMessageAt    sql.NullTime   `db:"latest_message_at" json:"latestMessageAt"`
	LastInboundAt      sql.NullTime   `db:"last_inbound_at" json:"lastInboundAt"`
	LastOutboundAt     sql.NullTime   `db:"last_outbound_at" json:"lastOutboundAt"`
	Meta               interface{}    `db:"meta" json:"meta"`
	CreatedAt          time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt          time.Time      `db:"updated_at" json:"updatedAt"`
}

type AdminEmailMessage struct {
	ID                string         `db:"id" json:"id"`
	ThreadID          string         `db:"thread_id" json:"threadId"`
	MailboxID         string         `db:"mailbox_id" json:"mailboxId"`
	Direction         string         `db:"direction" json:"direction"`
	SenderName        sql.NullString `db:"sender_name" json:"senderName"`
	SenderAddress     string         `db:"sender_address" json:"senderAddress"`
	ToAddresses       interface{}    `db:"to_addresses" json:"toAddresses"`
	CcAddresses       interface{}    `db:"cc_addresses" json:"ccAddresses"`
	BccAddresses      interface{}    `db:"bcc_addresses" json:"bccAddresses"`
	Subject           string         `db:"subject" json:"subject"`
	TextBody          sql.NullString `db:"text_body" json:"textBody"`
	HTMLBody          sql.NullString `db:"html_body" json:"htmlBody"`
	ProviderMessageID sql.NullString `db:"provider_message_id" json:"providerMessageId"`
	MessageIDHeader   sql.NullString `db:"message_id_header" json:"messageIdHeader"`
	InReplyTo         sql.NullString `db:"in_reply_to" json:"inReplyTo"`
	ReferencesHeaders interface{}    `db:"references_headers" json:"referencesHeaders"`
	SentAt            sql.NullTime   `db:"sent_at" json:"sentAt"`
	ReceivedAt        sql.NullTime   `db:"received_at" json:"receivedAt"`
	AdminUserID       sql.NullString `db:"admin_user_id" json:"adminUserId"`
	Meta              interface{}    `db:"meta" json:"meta"`
	CreatedAt         time.Time      `db:"created_at" json:"createdAt"`
}

type AdminEmailAttachment struct {
	ID          string    `db:"id" json:"id"`
	MessageID   string    `db:"message_id" json:"messageId"`
	FileName    string    `db:"file_name" json:"fileName"`
	ContentType string    `db:"content_type" json:"contentType"`
	SizeBytes   int64     `db:"size_bytes" json:"sizeBytes"`
	StorageKey  string    `db:"storage_key" json:"storageKey"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
}

type EmailDispatchLog struct {
	ID                string         `db:"id" json:"id"`
	Category          string         `db:"category" json:"category"`
	MailboxID         sql.NullString `db:"mailbox_id" json:"mailboxId"`
	ThreadID          sql.NullString `db:"thread_id" json:"threadId"`
	MessageID         sql.NullString `db:"message_id" json:"messageId"`
	Recipient         string         `db:"recipient" json:"recipient"`
	SenderAddress     string         `db:"sender_address" json:"senderAddress"`
	SenderName        sql.NullString `db:"sender_name" json:"senderName"`
	Provider          string         `db:"provider" json:"provider"`
	ProviderMessageID sql.NullString `db:"provider_message_id" json:"providerMessageId"`
	Status            string         `db:"status" json:"status"`
	ErrorMessage      sql.NullString `db:"error_message" json:"errorMessage"`
	Metadata          interface{}    `db:"metadata" json:"metadata"`
	CreatedAt         time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt         time.Time      `db:"updated_at" json:"updatedAt"`
	SentAt            sql.NullTime   `db:"sent_at" json:"sentAt"`
	DeliveredAt       sql.NullTime   `db:"delivered_at" json:"deliveredAt"`
	FailedAt          sql.NullTime   `db:"failed_at" json:"failedAt"`
}
