package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func (r *AdminRepository) CreateEmailDispatchLog(ctx context.Context, payload map[string]interface{}) error {
	metadata, _ := json.Marshal(payload["metadata"])
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO email_dispatch_logs (
			id, category, mailbox_id, thread_id, message_id, recipient, sender_address, sender_name,
			provider, provider_message_id, status, error_message, metadata, created_at, updated_at, sent_at, delivered_at, failed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW(),NOW(),$14,$15,$16)
	`,
		payload["id"],
		payload["category"],
		payload["mailboxId"],
		payload["threadId"],
		payload["messageId"],
		payload["recipient"],
		payload["senderAddress"],
		payload["senderName"],
		payload["provider"],
		payload["providerMessageId"],
		payload["status"],
		payload["errorMessage"],
		metadata,
		payload["sentAt"],
		payload["deliveredAt"],
		payload["failedAt"],
	)
	return err
}

func (r *AdminRepository) UpdateEmailDispatchLogStatusByProviderMessageID(ctx context.Context, providerMessageID, status string, errorMessage *string, timestamps map[string]*time.Time, metadata map[string]interface{}) error {
	metadataBytes, _ := json.Marshal(metadata)
	_, err := r.db.ExecContext(ctx, `
		UPDATE email_dispatch_logs
		SET status = $2,
			error_message = $3,
			metadata = CASE
				WHEN $4::jsonb IS NULL THEN metadata
				ELSE COALESCE(metadata, '{}'::jsonb) || $4::jsonb
			END,
			sent_at = COALESCE($5, sent_at),
			delivered_at = COALESCE($6, delivered_at),
			failed_at = COALESCE($7, failed_at),
			updated_at = NOW()
		WHERE provider_message_id = $1
	`, providerMessageID, status, nullableStringPointer(errorMessage), nullableJSON(metadataBytes), timestamps["sentAt"], timestamps["deliveredAt"], timestamps["failedAt"])
	return err
}

func (r *AdminRepository) ListEmailDispatchLogs(ctx context.Context, search, status, category string, page, perPage int) ([]map[string]interface{}, int, error) {
	base := `
		FROM email_dispatch_logs edl
		LEFT JOIN admin_mailboxes am ON am.id = edl.mailbox_id
	`

	whereClauses := []string{"1=1"}
	args := make([]interface{}, 0, 4)
	argIdx := 1

	if strings.TrimSpace(search) != "" {
		searchValue := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf(`(
			LOWER(edl.recipient) LIKE $%d OR
			LOWER(edl.sender_address) LIKE $%d OR
			LOWER(COALESCE(am.display_name, '')) LIKE $%d
		)`, argIdx, argIdx, argIdx))
		args = append(args, searchValue)
		argIdx++
	}
	if strings.TrimSpace(status) != "" && status != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("edl.status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if strings.TrimSpace(category) != "" && category != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("edl.category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}

	where := " WHERE " + strings.Join(whereClauses, " AND ")
	total, err := r.count(ctx, `SELECT COUNT(*) `+base+where, args...)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			edl.id,
			edl.category,
			edl.recipient,
			edl.sender_address,
			COALESCE(edl.sender_name, '') AS sender_name,
			edl.provider,
			edl.provider_message_id,
			edl.status,
			COALESCE(edl.error_message, '') AS error_message,
			edl.metadata,
			edl.created_at,
			edl.updated_at,
			edl.sent_at,
			edl.delivered_at,
			edl.failed_at,
			COALESCE(am.display_name, '') AS mailbox_name,
			COALESCE(am.address, '') AS mailbox_address
	` + base + where + `
		ORDER BY edl.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)

	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) FindMailboxByID(ctx context.Context, mailboxID string) (*domain.AdminMailbox, error) {
	var mailbox domain.AdminMailbox
	err := r.db.GetContext(ctx, &mailbox, `
		SELECT id, type, address, display_name, owner_admin_id, is_active, created_at, updated_at
		FROM admin_mailboxes
		WHERE id = $1
		LIMIT 1
	`, mailboxID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &mailbox, nil
}

func (r *AdminRepository) FindMailboxByAddress(ctx context.Context, address string) (*domain.AdminMailbox, error) {
	var mailbox domain.AdminMailbox
	err := r.db.GetContext(ctx, &mailbox, `
		SELECT id, type, address, display_name, owner_admin_id, is_active, created_at, updated_at
		FROM admin_mailboxes
		WHERE LOWER(address) = LOWER($1)
		LIMIT 1
	`, address)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &mailbox, nil
}

func (r *AdminRepository) FindPersonalMailboxByOwnerAdminID(ctx context.Context, adminID string) (*domain.AdminMailbox, error) {
	var mailbox domain.AdminMailbox
	err := r.db.GetContext(ctx, &mailbox, `
		SELECT id, type, address, display_name, owner_admin_id, is_active, created_at, updated_at
		FROM admin_mailboxes
		WHERE owner_admin_id = $1 AND type = $2
		LIMIT 1
	`, adminID, domain.AdminMailboxTypePersonal)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &mailbox, nil
}

func (r *AdminRepository) EnsurePersonalMailboxForAdmin(ctx context.Context, adminID, fullName, email, domainName string) (*domain.AdminMailbox, error) {
	existing, err := r.FindPersonalMailboxByOwnerAdminID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	localPart := personalMailboxLocalPart(fullName)
	if localPart == "" {
		localPart = personalMailboxLocalPart(strings.Split(strings.TrimSpace(strings.ToLower(email)), "@")[0])
	}
	if localPart == "" {
		localPart = "admin." + shortID(adminID)
	}

	address := localPart + "@" + strings.TrimPrefix(strings.TrimSpace(domainName), "@")
	candidate := address
	index := 2
	for {
		mailbox, err := r.FindMailboxByAddress(ctx, candidate)
		if err != nil {
			return nil, err
		}
		if mailbox == nil {
			break
		}
		candidate = fmt.Sprintf("%s.%02d@%s", localPart, index, strings.TrimPrefix(strings.TrimSpace(domainName), "@"))
		index++
	}

	mailbox := &domain.AdminMailbox{
		ID:          "ambx_" + uuid.New().String()[:8],
		Type:        domain.AdminMailboxTypePersonal,
		Address:     candidate,
		DisplayName: firstNonEmpty(strings.TrimSpace(fullName), strings.TrimSpace(email)),
		OwnerAdminID: sql.NullString{
			String: adminID,
			Valid:  strings.TrimSpace(adminID) != "",
		},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.CreateMailbox(ctx, mailbox); err != nil {
		return nil, err
	}
	return mailbox, nil
}

func (r *AdminRepository) ListMailboxes(ctx context.Context) ([]map[string]interface{}, error) {
	return r.selectMaps(ctx, `
		SELECT
			am.id,
			am.type,
			am.address,
			am.display_name,
			am.owner_admin_id,
			am.is_active,
			am.created_at,
			am.updated_at,
			COALESCE(owner.full_name, owner.email, '') AS owner_name,
			COALESCE(stats.unread_threads, 0) AS unread_threads,
			COALESCE(stats.total_threads, 0) AS total_threads,
			stats.latest_message_at
		FROM admin_mailboxes am
		LEFT JOIN admin_users owner ON owner.id = am.owner_admin_id
		LEFT JOIN (
			SELECT
				mailbox_id,
				COUNT(*) FILTER (WHERE unread_count > 0) AS unread_threads,
				COUNT(*) AS total_threads,
				MAX(latest_message_at) AS latest_message_at
			FROM admin_email_threads
			GROUP BY mailbox_id
		) stats ON stats.mailbox_id = am.id
		ORDER BY
			CASE am.type
				WHEN 'personal' THEN 0
				WHEN 'shared' THEN 1
				ELSE 2
			END,
			am.display_name ASC
	`)
}

func (r *AdminRepository) ListMailboxMembers(ctx context.Context, mailboxID string) ([]map[string]interface{}, error) {
	return r.selectMaps(ctx, `
		SELECT
			amm.mailbox_id,
			amm.admin_user_id,
			amm.can_reply,
			amm.created_at,
			COALESCE(au.full_name, au.email, '') AS admin_name,
			au.email,
			COALESCE(au.avatar_url, '') AS avatar_url
		FROM admin_mailbox_members amm
		INNER JOIN admin_users au ON au.id = amm.admin_user_id
		WHERE amm.mailbox_id = $1
		ORDER BY admin_name ASC
	`, mailboxID)
}

func (r *AdminRepository) ReplaceMailboxMembers(ctx context.Context, mailboxID string, memberIDs []string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM admin_mailbox_members WHERE mailbox_id = $1`, mailboxID); err != nil {
		return err
	}

	for _, memberID := range memberIDs {
		memberID = strings.TrimSpace(memberID)
		if memberID == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO admin_mailbox_members (mailbox_id, admin_user_id, can_reply, created_at)
			VALUES ($1, $2, TRUE, NOW())
			ON CONFLICT (mailbox_id, admin_user_id) DO NOTHING
		`, mailboxID, memberID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *AdminRepository) CreateMailbox(ctx context.Context, mailbox *domain.AdminMailbox) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_mailboxes (id, type, address, display_name, owner_admin_id, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, mailbox.ID, mailbox.Type, strings.ToLower(strings.TrimSpace(mailbox.Address)), mailbox.DisplayName, mailbox.OwnerAdminID, mailbox.IsActive, mailbox.CreatedAt, mailbox.UpdatedAt)
	return err
}

func (r *AdminRepository) UpdateMailbox(ctx context.Context, mailbox *domain.AdminMailbox) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_mailboxes
		SET type = $2,
			address = $3,
			display_name = $4,
			owner_admin_id = $5,
			is_active = $6,
			updated_at = NOW()
		WHERE id = $1
	`, mailbox.ID, mailbox.Type, strings.ToLower(strings.TrimSpace(mailbox.Address)), mailbox.DisplayName, mailbox.OwnerAdminID, mailbox.IsActive)
	return err
}

func (r *AdminRepository) FindThreadByID(ctx context.Context, threadID string) (*domain.AdminEmailThread, error) {
	var thread domain.AdminEmailThread
	err := r.db.GetContext(ctx, &thread, `
		SELECT id, mailbox_id, participant_name, participant_email, subject, normalized_subject, status,
		       assigned_admin_id, unread_count, last_direction, last_message_preview, latest_message_at,
		       last_inbound_at, last_outbound_at, meta, created_at, updated_at
		FROM admin_email_threads
		WHERE id = $1
		LIMIT 1
	`, threadID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *AdminRepository) FindThreadByMessageReferences(ctx context.Context, mailboxID string, references []string) (*domain.AdminEmailThread, error) {
	cleanRefs := make([]string, 0, len(references))
	for _, ref := range references {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		cleanRefs = append(cleanRefs, ref)
	}
	if len(cleanRefs) == 0 {
		return nil, nil
	}

	query, args, err := sqlx.In(`
		SELECT DISTINCT
			t.id, t.mailbox_id, t.participant_name, t.participant_email, t.subject, t.normalized_subject, t.status,
			t.assigned_admin_id, t.unread_count, t.last_direction, t.last_message_preview, t.latest_message_at,
			t.last_inbound_at, t.last_outbound_at, t.meta, t.created_at, t.updated_at
		FROM admin_email_threads t
		INNER JOIN admin_email_messages m ON m.thread_id = t.id
		WHERE t.mailbox_id = ? AND m.message_id_header IN (?)
		ORDER BY t.latest_message_at DESC NULLS LAST, t.created_at DESC
		LIMIT 1
	`, mailboxID, cleanRefs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var thread domain.AdminEmailThread
	err = r.db.GetContext(ctx, &thread, query, args...)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *AdminRepository) FindThreadByParticipant(ctx context.Context, mailboxID, normalizedSubject, participantEmail string) (*domain.AdminEmailThread, error) {
	var thread domain.AdminEmailThread
	err := r.db.GetContext(ctx, &thread, `
		SELECT id, mailbox_id, participant_name, participant_email, subject, normalized_subject, status,
		       assigned_admin_id, unread_count, last_direction, last_message_preview, latest_message_at,
		       last_inbound_at, last_outbound_at, meta, created_at, updated_at
		FROM admin_email_threads
		WHERE mailbox_id = $1
		  AND normalized_subject = $2
		  AND LOWER(participant_email) = LOWER($3)
		ORDER BY latest_message_at DESC NULLS LAST, created_at DESC
		LIMIT 1
	`, mailboxID, normalizedSubject, participantEmail)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *AdminRepository) CreateThread(ctx context.Context, payload map[string]interface{}) error {
	meta, _ := json.Marshal(defaultJSONObject(payload["meta"]))
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_email_threads (
			id, mailbox_id, participant_name, participant_email, subject, normalized_subject,
			status, assigned_admin_id, unread_count, last_direction, last_message_preview,
			latest_message_at, last_inbound_at, last_outbound_at, meta, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
	`,
		payload["id"],
		payload["mailboxId"],
		payload["participantName"],
		payload["participantEmail"],
		payload["subject"],
		payload["normalizedSubject"],
		payload["status"],
		payload["assignedAdminId"],
		payload["unreadCount"],
		payload["lastDirection"],
		payload["lastMessagePreview"],
		payload["latestMessageAt"],
		payload["lastInboundAt"],
		payload["lastOutboundAt"],
		meta,
		payload["createdAt"],
		payload["updatedAt"],
	)
	return err
}

func (r *AdminRepository) UpdateThreadAfterInbound(ctx context.Context, threadID, participantName, participantEmail, subject, normalizedSubject, preview string, receivedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_email_threads
		SET participant_name = $2,
			participant_email = $3,
			subject = $4,
			normalized_subject = $5,
			status = $6,
			unread_count = unread_count + 1,
			last_direction = $7,
			last_message_preview = $8,
			latest_message_at = $9,
			last_inbound_at = $9,
			updated_at = NOW()
		WHERE id = $1
	`, threadID, nullableString(participantName), participantEmail, subject, normalizedSubject, domain.AdminEmailThreadStatusBelumDibalas, domain.AdminEmailDirectionInbound, preview, receivedAt)
	return err
}

func (r *AdminRepository) UpdateThreadAfterOutbound(ctx context.Context, threadID, preview string, sentAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_email_threads
		SET status = $2,
			unread_count = 0,
			last_direction = $3,
			last_message_preview = $4,
			latest_message_at = $5,
			last_outbound_at = $5,
			updated_at = NOW()
		WHERE id = $1
	`, threadID, domain.AdminEmailThreadStatusDibalas, domain.AdminEmailDirectionOutbound, preview, sentAt)
	return err
}

func (r *AdminRepository) UpdateThreadStatus(ctx context.Context, threadID, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_email_threads
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`, threadID, status)
	return err
}

func (r *AdminRepository) UpdateThreadImportant(ctx context.Context, threadID string, isImportant bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_email_threads
		SET is_important = $2, updated_at = NOW()
		WHERE id = $1
	`, threadID, isImportant)
	return err
}

func (r *AdminRepository) AssignThread(ctx context.Context, threadID string, assignedAdminID *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE admin_email_threads
		SET assigned_admin_id = $2, updated_at = NOW()
		WHERE id = $1
	`, threadID, nullableStringPointer(assignedAdminID))
	return err
}

func (r *AdminRepository) CreateEmailMessage(ctx context.Context, payload map[string]interface{}) error {
	toAddresses, _ := json.Marshal(defaultJSONArray(payload["toAddresses"]))
	ccAddresses, _ := json.Marshal(defaultJSONArray(payload["ccAddresses"]))
	bccAddresses, _ := json.Marshal(defaultJSONArray(payload["bccAddresses"]))
	referencesHeaders, _ := json.Marshal(defaultJSONArray(payload["referencesHeaders"]))
	meta, _ := json.Marshal(defaultJSONObject(payload["meta"]))
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_email_messages (
			id, thread_id, mailbox_id, direction, sender_name, sender_address, to_addresses, cc_addresses,
			bcc_addresses, subject, text_body, html_body, provider_message_id, message_id_header, in_reply_to,
			references_headers, sent_at, received_at, admin_user_id, meta, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
	`,
		payload["id"],
		payload["threadId"],
		payload["mailboxId"],
		payload["direction"],
		payload["senderName"],
		payload["senderAddress"],
		toAddresses,
		ccAddresses,
		bccAddresses,
		payload["subject"],
		payload["textBody"],
		payload["htmlBody"],
		payload["providerMessageId"],
		payload["messageIdHeader"],
		payload["inReplyTo"],
		referencesHeaders,
		payload["sentAt"],
		payload["receivedAt"],
		payload["adminUserId"],
		meta,
		payload["createdAt"],
	)
	return err
}

// ExistsMessageByHeader checks if an email message with the given message_id_header already exists.
func (r *AdminRepository) ExistsMessageByHeader(ctx context.Context, messageIDHeader string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM admin_email_messages WHERE message_id_header = $1)`, messageIDHeader)
	return exists, err
}

func (r *AdminRepository) CreateEmailAttachment(ctx context.Context, payload map[string]interface{}) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_email_attachments (
			id, message_id, file_name, content_type, size_bytes, storage_key, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
	`,
		payload["id"],
		payload["messageId"],
		payload["fileName"],
		payload["contentType"],
		payload["sizeBytes"],
		payload["storageKey"],
		payload["createdAt"],
	)
	return err
}

func (r *AdminRepository) AddEmailThreadEvent(ctx context.Context, threadID, actorID, eventType, notes string, payload interface{}) error {
	payloadBytes, _ := json.Marshal(defaultJSONObject(payload))
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_email_thread_events (
			id, thread_id, actor_id, event_type, notes, payload, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,NOW())
	`, "aev_"+uuid.New().String()[:8], threadID, nullableString(actorID), eventType, nullableString(notes), payloadBytes)
	return err
}

func (r *AdminRepository) ListThreadsByMailbox(ctx context.Context, mailboxID, search, status, assigned string, unreadOnly bool, page, perPage int) ([]map[string]interface{}, int, error) {
	base := `
		FROM admin_email_threads aet
		LEFT JOIN admin_users assigned_admin ON assigned_admin.id = aet.assigned_admin_id
	`

	whereClauses := []string{"aet.mailbox_id = $1"}
	args := []interface{}{mailboxID}
	argIdx := 2

	if strings.TrimSpace(search) != "" {
		searchValue := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf(`(
			LOWER(aet.subject) LIKE $%d OR
			LOWER(COALESCE(aet.participant_name, '')) LIKE $%d OR
			LOWER(aet.participant_email) LIKE $%d OR
			LOWER(COALESCE(aet.last_message_preview, '')) LIKE $%d
		)`, argIdx, argIdx, argIdx, argIdx))
		args = append(args, searchValue)
		argIdx++
	}
	if strings.TrimSpace(status) != "" && status != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("aet.status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if strings.TrimSpace(assigned) != "" {
		switch assigned {
		case "unassigned":
			whereClauses = append(whereClauses, "aet.assigned_admin_id IS NULL")
		default:
			whereClauses = append(whereClauses, fmt.Sprintf("aet.assigned_admin_id = $%d", argIdx))
			args = append(args, assigned)
			argIdx++
		}
	}
	if unreadOnly {
		whereClauses = append(whereClauses, "aet.unread_count > 0")
	}

	where := " WHERE " + strings.Join(whereClauses, " AND ")
	total, err := r.count(ctx, `SELECT COUNT(*) `+base+where, args...)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			aet.id,
			aet.mailbox_id,
			COALESCE(aet.participant_name, '') AS participant_name,
			aet.participant_email,
			aet.subject,
			aet.status,
			aet.unread_count,
			COALESCE(aet.last_direction, '') AS last_direction,
			COALESCE(aet.last_message_preview, '') AS last_message_preview,
			aet.latest_message_at,
			aet.last_inbound_at,
			aet.last_outbound_at,
			aet.created_at,
			aet.updated_at,
			COALESCE(aet.assigned_admin_id, '') AS assigned_admin_id,
			COALESCE(assigned_admin.full_name, assigned_admin.email, '') AS assigned_admin_name,
			aet.is_important
	` + base + where + `
		ORDER BY aet.latest_message_at DESC NULLS LAST, aet.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)

	items, err := r.selectMaps(ctx, query, append(args, sanitizePageSize(perPage), calculateOffset(page, perPage))...)
	return items, total, err
}

func (r *AdminRepository) GetThreadDetail(ctx context.Context, threadID string) (map[string]interface{}, error) {
	items, err := r.selectMaps(ctx, `
		SELECT
			aet.id,
			aet.mailbox_id,
			aet.participant_name,
			aet.participant_email,
			aet.subject,
			aet.normalized_subject,
			aet.status,
			aet.assigned_admin_id,
			COALESCE(assigned_admin.full_name, assigned_admin.email, '') AS assigned_admin_name,
			aet.unread_count,
			aet.last_direction,
			aet.last_message_preview,
			aet.latest_message_at,
			aet.last_inbound_at,
			aet.last_outbound_at,
			aet.meta,
			aet.created_at,
			aet.updated_at,
			am.type AS mailbox_type,
			am.address AS mailbox_address,
			am.display_name AS mailbox_name,
			aet.is_important
		FROM admin_email_threads aet
		INNER JOIN admin_mailboxes am ON am.id = aet.mailbox_id
		LEFT JOIN admin_users assigned_admin ON assigned_admin.id = aet.assigned_admin_id
		WHERE aet.id = $1
		LIMIT 1
	`, threadID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}

func (r *AdminRepository) ListThreadMessages(ctx context.Context, threadID string) ([]map[string]interface{}, error) {
	return r.selectMaps(ctx, `
		SELECT
			aem.id,
			aem.thread_id,
			aem.mailbox_id,
			aem.direction,
			COALESCE(aem.sender_name, '') AS sender_name,
			aem.sender_address,
			aem.to_addresses,
			aem.cc_addresses,
			aem.bcc_addresses,
			aem.subject,
			COALESCE(aem.text_body, '') AS text_body,
			COALESCE(aem.html_body, '') AS html_body,
			COALESCE(aem.provider_message_id, '') AS provider_message_id,
			COALESCE(aem.message_id_header, '') AS message_id_header,
			COALESCE(aem.in_reply_to, '') AS in_reply_to,
			aem.references_headers,
			aem.sent_at,
			aem.received_at,
			aem.admin_user_id,
			COALESCE(admin_user.full_name, admin_user.email, '') AS admin_name,
			aem.meta,
			aem.created_at
		FROM admin_email_messages aem
		LEFT JOIN admin_users admin_user ON admin_user.id = aem.admin_user_id
		WHERE aem.thread_id = $1
		ORDER BY COALESCE(aem.received_at, aem.sent_at, aem.created_at) ASC, aem.created_at ASC
	`, threadID)
}

func (r *AdminRepository) ListThreadAttachments(ctx context.Context, threadID string) ([]map[string]interface{}, error) {
	return r.selectMaps(ctx, `
		SELECT
			aea.id,
			aea.message_id,
			aea.file_name,
			COALESCE(aea.content_type, '') AS content_type,
			aea.size_bytes,
			aea.storage_key,
			aea.created_at
		FROM admin_email_attachments aea
		INNER JOIN admin_email_messages aem ON aem.id = aea.message_id
		WHERE aem.thread_id = $1
		ORDER BY aea.created_at ASC
	`, threadID)
}

func (r *AdminRepository) FindAttachmentByID(ctx context.Context, attachmentID string) (*domain.AdminEmailAttachment, error) {
	var attachment domain.AdminEmailAttachment
	err := r.db.GetContext(ctx, &attachment, `
		SELECT id, message_id, file_name, content_type, size_bytes, storage_key, created_at
		FROM admin_email_attachments
		WHERE id = $1
		LIMIT 1
	`, attachmentID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func personalMailboxLocalPart(fullName string) string {
	fullName = strings.TrimSpace(strings.ToLower(fullName))
	if fullName == "" {
		return ""
	}

	builder := strings.Builder{}
	lastDot := false
	for _, r := range fullName {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastDot = false
		case unicode.IsSpace(r):
			if !lastDot && builder.Len() > 0 {
				builder.WriteRune('.')
				lastDot = true
			}
		}
	}

	value := strings.Trim(builder.String(), ".")
	value = strings.ReplaceAll(value, "..", ".")
	return value
}

func shortID(value string) string {
	value = strings.ReplaceAll(value, "-", "")
	if len(value) <= 8 {
		return value
	}
	return value[:8]
}

func defaultJSONArray(value interface{}) interface{} {
	if value == nil {
		return []interface{}{}
	}
	return value
}

func defaultJSONObject(value interface{}) interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	return value
}

func nullableStringPointer(value *string) interface{} {
	if value == nil {
		return nil
	}
	return nullableString(*value)
}

func nullableJSON(value []byte) interface{} {
	if len(value) == 0 || string(value) == "null" {
		return nil
	}
	return value
}

func nullableString(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
