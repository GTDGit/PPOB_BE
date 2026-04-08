package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/mail"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/config"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	internals3 "github.com/GTDGit/PPOB_BE/internal/external/s3"
	internalsmtp "github.com/GTDGit/PPOB_BE/internal/external/smtp"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/google/uuid"
	"slices"
)

type AdminMailboxService struct {
	repo          *repository.AdminRepository
	emailService  *EmailService
	emailStorage  *internals3.Client
	emailCfg      config.EmailConfig
	httpClient    *http.Client
	unmappedInbox string
}

type MailboxUpsertRequest struct {
	Type         string   `json:"type"`
	Address      string   `json:"address"`
	DisplayName  string   `json:"displayName"`
	OwnerAdminID string   `json:"ownerAdminId"`
	IsActive     bool     `json:"isActive"`
	MemberIDs    []string `json:"memberIds"`
}

type ThreadReplyRequest struct {
	Body        string   `json:"body"`
	HTMLBody    string   `json:"htmlBody"`
	Cc          []string `json:"cc"`
	Bcc         []string `json:"bcc"`
	IsImportant *bool    `json:"isImportant"`
	Attachments []EmailAttachmentInput `json:"-"`
}

type EmailAttachmentInput struct {
	Filename    string
	ContentType string
	Data        []byte
}

type ThreadStatusUpdateRequest struct {
	Status string `json:"status"`
}

type ThreadAssignRequest struct {
	AdminUserID string `json:"adminUserId"`
}

type snsEnvelope struct {
	Type         string `json:"Type"`
	Message      string `json:"Message"`
	SubscribeURL string `json:"SubscribeURL"`
	TopicARN     string `json:"TopicArn"`
}

type sesEventMail struct {
	MessageID     string `json:"messageId"`
	Timestamp     string `json:"timestamp"`
	Source        string `json:"source"`
	CommonHeaders struct {
		Subject string   `json:"subject"`
		From    []string `json:"from"`
		To      []string `json:"to"`
		Cc      []string `json:"cc"`
	} `json:"commonHeaders"`
}

type sesReceiptAction struct {
	Type       string `json:"type"`
	BucketName string `json:"bucketName"`
	ObjectKey  string `json:"objectKey"`
}

type sesReceiptNotification struct {
	Mail    sesEventMail `json:"mail"`
	Receipt struct {
		Recipients []string         `json:"recipients"`
		Action     sesReceiptAction `json:"action"`
	} `json:"receipt"`
}

type parsedInboundEmail struct {
	MessageIDHeader string
	InReplyTo       string
	References      []string
	Subject         string
	Normalized      string
	SenderName      string
	SenderAddress   string
	To              []string
	Cc              []string
	TextBody        string
	HTMLBody        string
	Preview         string
	ReceivedAt      time.Time
	Attachments     []parsedInboundAttachment
}

type parsedInboundAttachment struct {
	FileName    string
	ContentType string
	Data        []byte
}

func NewAdminMailboxService(repo *repository.AdminRepository, emailService *EmailService, emailStorage *internals3.Client, emailCfg config.EmailConfig) *AdminMailboxService {
	return &AdminMailboxService{
		repo:          repo,
		emailService:  emailService,
		emailStorage:  emailStorage,
		emailCfg:      emailCfg,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		unmappedInbox: "unmapped@ppob.id",
	}
}

func (s *AdminMailboxService) ListMailboxes(ctx context.Context, adminID string) (map[string]interface{}, error) {
	admin, err := s.requireAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.ListMailboxes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list mailboxes: %w", err)
	}

	filtered := make([]map[string]interface{}, 0, len(items))
	myMailboxes := make([]map[string]interface{}, 0)
	sharedMailboxes := make([]map[string]interface{}, 0)
	systemMailboxes := make([]map[string]interface{}, 0)

	for _, item := range items {
		mailboxID := stringValue(item["id"])
		if mailboxID == "" {
			continue
		}

		mailbox, err := s.repo.FindMailboxByID(ctx, mailboxID)
		if err != nil {
			return nil, fmt.Errorf("failed to get mailbox: %w", err)
		}
		if mailbox == nil || !s.canViewMailbox(ctx, admin, mailbox) {
			continue
		}

		enriched := item
		enriched["section"] = s.mailboxSection(admin, mailbox)
		if mailbox.Type == domain.AdminMailboxTypeShared || mailbox.Type == domain.AdminMailboxTypePersonal {
			members, err := s.repo.ListMailboxMembers(ctx, mailbox.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get mailbox members: %w", err)
			}
			enriched["members"] = members
		}

		filtered = append(filtered, enriched)
		switch mailbox.Type {
		case domain.AdminMailboxTypePersonal:
			myMailboxes = append(myMailboxes, enriched)
		case domain.AdminMailboxTypeShared:
			sharedMailboxes = append(sharedMailboxes, enriched)
		default:
			systemMailboxes = append(systemMailboxes, enriched)
		}
	}

	return map[string]interface{}{
		"items":           filtered,
		"myMailboxes":     myMailboxes,
		"sharedMailboxes": sharedMailboxes,
		"systemMailboxes": systemMailboxes,
	}, nil
}

func (s *AdminMailboxService) CreateMailbox(ctx context.Context, actorID string, req MailboxUpsertRequest) (map[string]interface{}, error) {
	admin, err := s.requireAdmin(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !hasPermission(admin, "mailboxes.manage") {
		return nil, domain.NewError("ADMIN_FORBIDDEN", "Anda tidak memiliki akses untuk mengelola mailbox", 403)
	}

	mailboxType := normalizeMailboxType(req.Type)
	if mailboxType == "" {
		return nil, domain.ErrValidationFailed("Tipe mailbox tidak valid")
	}

	address := normalizeMailboxAddress(req.Address)
	ownerAdminID := strings.TrimSpace(req.OwnerAdminID)
	displayName := strings.TrimSpace(req.DisplayName)

	if mailboxType == domain.AdminMailboxTypePersonal && ownerAdminID == "" {
		return nil, domain.ErrValidationFailed("Owner mailbox personal wajib dipilih")
	}

	if mailboxType == domain.AdminMailboxTypePersonal && address == "" {
		owner, err := s.repo.FindAdminByID(ctx, ownerAdminID)
		if err != nil {
			return nil, fmt.Errorf("failed to get mailbox owner: %w", err)
		}
		if owner == nil {
			return nil, domain.NewError("ADMIN_NOT_FOUND", "Owner mailbox tidak ditemukan", 404)
		}
		mailbox, err := s.repo.EnsurePersonalMailboxForAdmin(ctx, owner.ID, owner.DisplayName(), owner.Email, s.mailboxDomain())
		if err != nil {
			return nil, fmt.Errorf("failed to ensure personal mailbox: %w", err)
		}
		if displayName != "" || !req.IsActive {
			mailbox.DisplayName = firstNonEmpty(displayName, mailbox.DisplayName)
			mailbox.IsActive = req.IsActive
			if err := s.repo.UpdateMailbox(ctx, mailbox); err != nil {
				return nil, fmt.Errorf("failed to update personal mailbox: %w", err)
			}
		}
		return s.buildMailboxPayload(ctx, mailbox.ID)
	}

	if address == "" {
		return nil, domain.ErrValidationFailed("Alamat email mailbox wajib diisi")
	}
	if !strings.HasSuffix(address, "@"+s.mailboxDomain()) {
		return nil, domain.ErrValidationFailed("Mailbox harus menggunakan domain @ppob.id")
	}

	existing, err := s.repo.FindMailboxByAddress(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to check mailbox address: %w", err)
	}
	if existing != nil {
		return nil, domain.NewError("MAILBOX_EXISTS", "Alamat mailbox sudah digunakan", 409)
	}

	mailbox := &domain.AdminMailbox{
		ID:          "ambx_" + uuid.New().String()[:8],
		Type:        mailboxType,
		Address:     address,
		DisplayName: firstNonEmpty(displayName, address),
		OwnerAdminID: sql.NullString{
			String: ownerAdminID,
			Valid:  ownerAdminID != "",
		},
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.CreateMailbox(ctx, mailbox); err != nil {
		return nil, fmt.Errorf("failed to create mailbox: %w", err)
	}
	if mailbox.Type == domain.AdminMailboxTypeShared && len(req.MemberIDs) > 0 {
		if err := s.repo.ReplaceMailboxMembers(ctx, mailbox.ID, req.MemberIDs); err != nil {
			return nil, fmt.Errorf("failed to update mailbox members: %w", err)
		}
	}

	return s.buildMailboxPayload(ctx, mailbox.ID)
}

func (s *AdminMailboxService) UpdateMailbox(ctx context.Context, actorID, mailboxID string, req MailboxUpsertRequest) (map[string]interface{}, error) {
	admin, err := s.requireAdmin(ctx, actorID)
	if err != nil {
		return nil, err
	}
	if !hasPermission(admin, "mailboxes.manage") {
		return nil, domain.NewError("ADMIN_FORBIDDEN", "Anda tidak memiliki akses untuk mengelola mailbox", 403)
	}

	mailbox, err := s.repo.FindMailboxByID(ctx, mailboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailbox: %w", err)
	}
	if mailbox == nil {
		return nil, domain.NewError("MAILBOX_NOT_FOUND", "Mailbox tidak ditemukan", 404)
	}

	if req.Type != "" {
		mailbox.Type = normalizeMailboxType(req.Type)
	}
	if req.Address != "" {
		mailbox.Address = normalizeMailboxAddress(req.Address)
	}
	if req.DisplayName != "" {
		mailbox.DisplayName = strings.TrimSpace(req.DisplayName)
	}
	mailbox.IsActive = req.IsActive
	if mailbox.Type == domain.AdminMailboxTypePersonal {
		mailbox.OwnerAdminID = sql.NullString{
			String: strings.TrimSpace(req.OwnerAdminID),
			Valid:  strings.TrimSpace(req.OwnerAdminID) != "",
		}
	}

	if err := s.repo.UpdateMailbox(ctx, mailbox); err != nil {
		return nil, fmt.Errorf("failed to update mailbox: %w", err)
	}
	if mailbox.Type == domain.AdminMailboxTypeShared {
		if err := s.repo.ReplaceMailboxMembers(ctx, mailbox.ID, req.MemberIDs); err != nil {
			return nil, fmt.Errorf("failed to replace mailbox members: %w", err)
		}
	}

	return s.buildMailboxPayload(ctx, mailbox.ID)
}

func (s *AdminMailboxService) ListMailboxThreads(ctx context.Context, adminID, mailboxID, search, status, assigned string, unreadOnly bool, page, perPage int) (map[string]interface{}, error) {
	admin, mailbox, err := s.requireMailboxAccess(ctx, adminID, mailboxID)
	if err != nil {
		return nil, err
	}

	items, total, err := s.repo.ListThreadsByMailbox(ctx, mailbox.ID, search, status, assigned, unreadOnly, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to list mailbox threads: %w", err)
	}

	return map[string]interface{}{
		"mailbox": map[string]interface{}{
			"id":          mailbox.ID,
			"type":        mailbox.Type,
			"address":     mailbox.Address,
			"displayName": mailbox.DisplayName,
		},
		"list": domain.AdminListResponse{
			Items:   items,
			Page:    page,
			PerPage: perPage,
			Total:   total,
			HasNext: mailboxOffset(page, perPage)+len(items) < total,
		},
		"canReply":        hasPermission(admin, "mailboxes.reply"),
		"canAssign":       hasPermission(admin, "mailboxes.assign"),
		"canManageStatus": hasPermission(admin, "mailboxes.status.manage"),
	}, nil
}

func (s *AdminMailboxService) GetThreadDetail(ctx context.Context, adminID, threadID string) (map[string]interface{}, error) {
	admin, err := s.requireAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}

	thread, err := s.repo.GetThreadDetail(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread detail: %w", err)
	}
	if thread == nil {
		return nil, domain.NewError("THREAD_NOT_FOUND", "Thread email tidak ditemukan", 404)
	}

	mailbox, err := s.repo.FindMailboxByID(ctx, stringValue(thread["mailbox_id"]))
	if err != nil {
		return nil, fmt.Errorf("failed to get mailbox: %w", err)
	}
	if mailbox == nil || !s.canViewMailbox(ctx, admin, mailbox) {
		return nil, domain.NewError("MAILBOX_FORBIDDEN", "Anda tidak memiliki akses ke thread ini", 403)
	}

	messages, err := s.repo.ListThreadMessages(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread messages: %w", err)
	}
	attachments, err := s.repo.ListThreadAttachments(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread attachments: %w", err)
	}
	members, err := s.repo.ListMailboxMembers(ctx, mailbox.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailbox members: %w", err)
	}

	return map[string]interface{}{
		"thread":       thread,
		"mailbox":      mailbox,
		"messages":     messages,
		"attachments":  attachments,
		"members":      members,
		"canReply":     hasPermission(admin, "mailboxes.reply") && s.canReplyFromMailbox(ctx, admin, mailbox),
		"canAssign":    hasPermission(admin, "mailboxes.assign"),
		"canSetStatus": hasPermission(admin, "mailboxes.status.manage"),
	}, nil
}

func (s *AdminMailboxService) ReplyThread(ctx context.Context, adminID, threadID string, req ThreadReplyRequest) (map[string]interface{}, error) {
	admin, err := s.requireAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if !hasPermission(admin, "mailboxes.reply") {
		return nil, domain.NewError("ADMIN_FORBIDDEN", "Anda tidak memiliki akses untuk membalas email", 403)
	}

	thread, err := s.repo.FindThreadByID(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	if thread == nil {
		return nil, domain.NewError("THREAD_NOT_FOUND", "Thread email tidak ditemukan", 404)
	}

	mailbox, err := s.repo.FindMailboxByID(ctx, thread.MailboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailbox: %w", err)
	}
	if mailbox == nil || !s.canReplyFromMailbox(ctx, admin, mailbox) {
		return nil, domain.NewError("MAILBOX_FORBIDDEN", "Anda tidak memiliki akses untuk membalas dari mailbox ini", 403)
	}

	body := strings.TrimSpace(req.Body)
	htmlBody := strings.TrimSpace(req.HTMLBody)
	if body == "" && htmlBody == "" {
		return nil, domain.ErrValidationFailed("Isi balasan email tidak boleh kosong")
	}
	if htmlBody == "" {
		htmlBody = fmt.Sprintf("<p style=\"white-space:pre-wrap;line-height:1.6;color:#0f172a;\">%s</p>", strings.ReplaceAll(html.EscapeString(body), "\n", "<br/>"))
	}

	// Append email signature
	htmlBody += s.buildSignature(ctx, adminID)

	messages, err := s.repo.ListThreadMessages(ctx, thread.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread messages: %w", err)
	}
	lastHeader, references := collectThreadReferences(messages)
	localHeader := generateMessageHeader(mailbox.Address)
	subject := ensureReplySubject(thread.Subject)
	preview := buildPreview(firstNonEmpty(body, stripHTML(htmlBody)))
	now := time.Now()
	messageID := "aem_" + uuid.New().String()[:8]

	replyHeaders := map[string]string{
		"Message-ID":  localHeader,
		"In-Reply-To": lastHeader,
		"References":  strings.Join(references, " "),
	}
	if req.IsImportant != nil && *req.IsImportant {
		replyHeaders["X-Priority"] = "1"
		replyHeaders["Importance"] = "high"
	}

	providerMessageID, err := s.emailService.SendMailboxReply(ctx, MailReplyRequest{
		Category:         "mailbox_reply",
		MailboxID:        mailbox.ID,
		ThreadID:         thread.ID,
		MessageID:        messageID,
		FromAddress:      mailbox.Address,
		FromName:         mailbox.DisplayName,
		ToAddresses:      []string{thread.ParticipantEmail},
		CcAddresses:      sanitizeEmailList(req.Cc),
		BccAddresses:     sanitizeEmailList(req.Bcc),
		ReplyToAddresses: []string{mailbox.Address},
		Subject:          subject,
		HTMLBody:         htmlBody,
		TextBody:         firstNonEmpty(body, stripHTML(htmlBody)),
		Headers:          replyHeaders,
		ConfigurationSet: s.emailCfg.SES.ConfigurationSetOperations,
		Tags: map[string]string{
			"category": "mailbox_reply",
			"mailbox":  mailbox.Address,
		},
		Attachments: toSMTPAttachments(req.Attachments),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send mailbox reply: %w", err)
	}

	if err := s.repo.CreateEmailMessage(ctx, map[string]interface{}{
		"id":                messageID,
		"threadId":          thread.ID,
		"mailboxId":         mailbox.ID,
		"direction":         domain.AdminEmailDirectionOutbound,
		"senderName":        nullableString(mailbox.DisplayName),
		"senderAddress":     mailbox.Address,
		"toAddresses":       []string{thread.ParticipantEmail},
		"ccAddresses":       sanitizeEmailList(req.Cc),
		"bccAddresses":      sanitizeEmailList(req.Bcc),
		"subject":           subject,
		"textBody":          nullableString(firstNonEmpty(body, stripHTML(htmlBody))),
		"htmlBody":          nullableString(htmlBody),
		"providerMessageId": nullableString(providerMessageID),
		"messageIdHeader":   nullableString(localHeader),
		"inReplyTo":         nullableString(lastHeader),
		"referencesHeaders": references,
		"sentAt":            now,
		"adminUserId":       nullableString(admin.ID),
		"meta": map[string]interface{}{
			"mailboxAddress": mailbox.Address,
		},
		"createdAt": now,
	}); err != nil {
		return nil, fmt.Errorf("failed to save outbound message: %w", err)
	}

	if err := s.repo.UpdateThreadAfterOutbound(ctx, thread.ID, preview, now); err != nil {
		return nil, fmt.Errorf("failed to update thread state: %w", err)
	}
	if err := s.repo.AddEmailThreadEvent(ctx, thread.ID, admin.ID, "replied", fmt.Sprintf("Balasan dikirim dari %s", mailbox.Address), map[string]interface{}{
		"messageId": messageID,
	}); err != nil {
		return nil, fmt.Errorf("failed to add thread event: %w", err)
	}
	if err := s.repo.CreateEmailDispatchLog(ctx, map[string]interface{}{
		"id":                "edl_" + uuid.New().String()[:8],
		"category":          "mailbox_reply",
		"mailboxId":         mailbox.ID,
		"threadId":          thread.ID,
		"messageId":         messageID,
		"recipient":         thread.ParticipantEmail,
		"senderAddress":     mailbox.Address,
		"senderName":        mailbox.DisplayName,
		"provider":          "SES",
		"providerMessageId": nullableString(providerMessageID),
		"status":            firstNonEmpty(statusFromProvider("ses"), "queued"),
		"metadata": map[string]interface{}{
			"threadId": thread.ID,
		},
		"sentAt": now,
	}); err != nil {
		return nil, fmt.Errorf("failed to create email dispatch log: %w", err)
	}

	return map[string]interface{}{
		"message":           "Balasan email berhasil dikirim",
		"threadId":          thread.ID,
		"messageId":         messageID,
		"providerMessageId": providerMessageID,
	}, nil
}

type ComposeEmailRequest struct {
	MailboxID   string   `json:"mailboxId"`
	To          []string `json:"to"`
	Cc          []string `json:"cc"`
	Bcc         []string `json:"bcc"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	HTMLBody    string   `json:"htmlBody"`
	IsImportant bool     `json:"isImportant"`
	Attachments []EmailAttachmentInput `json:"-"`
}

func (s *AdminMailboxService) ComposeEmail(ctx context.Context, adminID string, req ComposeEmailRequest) (map[string]interface{}, error) {
	admin, err := s.requireAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if !hasPermission(admin, "mailboxes.reply") {
		return nil, domain.NewError("ADMIN_FORBIDDEN", "Anda tidak memiliki akses untuk mengirim email", 403)
	}

	mailbox, err := s.repo.FindMailboxByID(ctx, strings.TrimSpace(req.MailboxID))
	if err != nil {
		return nil, fmt.Errorf("failed to get mailbox: %w", err)
	}
	if mailbox == nil || !s.canReplyFromMailbox(ctx, admin, mailbox) {
		return nil, domain.NewError("MAILBOX_FORBIDDEN", "Anda tidak memiliki akses untuk mengirim dari mailbox ini", 403)
	}

	toAddresses := sanitizeEmailList(req.To)
	if len(toAddresses) == 0 {
		return nil, domain.ErrValidationFailed("Minimal satu alamat email penerima harus diisi")
	}
	subject := strings.TrimSpace(req.Subject)
	if subject == "" {
		return nil, domain.ErrValidationFailed("Subjek email tidak boleh kosong")
	}
	body := strings.TrimSpace(req.Body)
	htmlBody := strings.TrimSpace(req.HTMLBody)
	if body == "" && htmlBody == "" {
		return nil, domain.ErrValidationFailed("Isi email tidak boleh kosong")
	}
	if htmlBody == "" {
		htmlBody = fmt.Sprintf("<p style=\"white-space:pre-wrap;line-height:1.6;color:#0f172a;\">%s</p>", strings.ReplaceAll(html.EscapeString(body), "\n", "<br/>"))
	}

	// Append email signature
	htmlBody += s.buildSignature(ctx, adminID)

	localHeader := generateMessageHeader(mailbox.Address)
	preview := buildPreview(firstNonEmpty(body, stripHTML(htmlBody)))
	now := time.Now()
	messageID := "aem_" + uuid.New().String()[:8]

	thread := &domain.AdminEmailThread{
		ID:                "aet_" + uuid.New().String()[:8],
		MailboxID:         mailbox.ID,
		ParticipantEmail:  toAddresses[0],
		Subject:           subject,
		NormalizedSubject: normalizeEmailSubject(subject),
		Status:            domain.AdminEmailThreadStatusDibalas,
		UnreadCount:       0,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := s.repo.CreateThread(ctx, map[string]interface{}{
		"id":                 thread.ID,
		"mailboxId":          thread.MailboxID,
		"participantName":    nil,
		"participantEmail":   thread.ParticipantEmail,
		"subject":            subject,
		"normalizedSubject":  thread.NormalizedSubject,
		"status":             thread.Status,
		"unreadCount":        0,
		"lastDirection":      domain.AdminEmailDirectionOutbound,
		"lastMessagePreview": preview,
		"latestMessageAt":    now,
		"meta": map[string]interface{}{
			"mailboxAddress": mailbox.Address,
			"composed":       true,
		},
		"createdAt": now,
		"updatedAt": now,
	}); err != nil {
		return nil, fmt.Errorf("failed to create email thread: %w", err)
	}

	composeHeaders := map[string]string{
		"Message-ID": localHeader,
	}
	if req.IsImportant {
		composeHeaders["X-Priority"] = "1"
		composeHeaders["Importance"] = "high"
	}

	providerMessageID, err := s.emailService.SendMailboxReply(ctx, MailReplyRequest{
		Category:         "mailbox_compose",
		MailboxID:        mailbox.ID,
		ThreadID:         thread.ID,
		MessageID:        messageID,
		FromAddress:      mailbox.Address,
		FromName:         mailbox.DisplayName,
		ToAddresses:      toAddresses,
		CcAddresses:      sanitizeEmailList(req.Cc),
		BccAddresses:     sanitizeEmailList(req.Bcc),
		ReplyToAddresses: []string{mailbox.Address},
		Subject:          subject,
		HTMLBody:         htmlBody,
		TextBody:         firstNonEmpty(body, stripHTML(htmlBody)),
		Headers:          composeHeaders,
		ConfigurationSet: s.emailCfg.SES.ConfigurationSetOperations,
		Tags: map[string]string{
			"category": "mailbox_compose",
			"mailbox":  mailbox.Address,
		},
		Attachments: toSMTPAttachments(req.Attachments),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send composed email: %w", err)
	}

	if err := s.repo.CreateEmailMessage(ctx, map[string]interface{}{
		"id":                messageID,
		"threadId":          thread.ID,
		"mailboxId":         mailbox.ID,
		"direction":         domain.AdminEmailDirectionOutbound,
		"senderName":        nullableString(mailbox.DisplayName),
		"senderAddress":     mailbox.Address,
		"toAddresses":       toAddresses,
		"ccAddresses":       sanitizeEmailList(req.Cc),
		"bccAddresses":      sanitizeEmailList(req.Bcc),
		"subject":           subject,
		"textBody":          nullableString(firstNonEmpty(body, stripHTML(htmlBody))),
		"htmlBody":          nullableString(htmlBody),
		"providerMessageId": nullableString(providerMessageID),
		"messageIdHeader":   nullableString(localHeader),
		"sentAt":            now,
		"adminUserId":       nullableString(admin.ID),
		"meta": map[string]interface{}{
			"mailboxAddress": mailbox.Address,
		},
		"createdAt": now,
	}); err != nil {
		return nil, fmt.Errorf("failed to save composed message: %w", err)
	}

	if err := s.repo.UpdateThreadAfterOutbound(ctx, thread.ID, preview, now); err != nil {
		return nil, fmt.Errorf("failed to update thread state: %w", err)
	}
	_ = s.repo.AddEmailThreadEvent(ctx, thread.ID, admin.ID, "composed", fmt.Sprintf("Email baru dikirim dari %s", mailbox.Address), map[string]interface{}{
		"messageId": messageID,
	})
	_ = s.repo.CreateEmailDispatchLog(ctx, map[string]interface{}{
		"id":                "edl_" + uuid.New().String()[:8],
		"category":          "mailbox_compose",
		"mailboxId":         mailbox.ID,
		"threadId":          thread.ID,
		"messageId":         messageID,
		"recipient":         toAddresses[0],
		"senderAddress":     mailbox.Address,
		"senderName":        mailbox.DisplayName,
		"provider":          "SMTP",
		"providerMessageId": nullableString(providerMessageID),
		"status":            "queued",
		"metadata": map[string]interface{}{
			"threadId":    thread.ID,
			"toAddresses": toAddresses,
		},
		"sentAt": now,
	})

	return map[string]interface{}{
		"message":           "Email berhasil dikirim",
		"threadId":          thread.ID,
		"messageId":         messageID,
		"providerMessageId": providerMessageID,
	}, nil
}

func (s *AdminMailboxService) UpdateMailboxDisplayName(ctx context.Context, actorID, mailboxID, displayName string) (map[string]interface{}, error) {
	admin, err := s.requireAdmin(ctx, actorID)
	if err != nil {
		return nil, err
	}

	mailbox, err := s.repo.FindMailboxByID(ctx, mailboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailbox: %w", err)
	}
	if mailbox == nil {
		return nil, domain.NewError("MAILBOX_NOT_FOUND", "Mailbox tidak ditemukan", 404)
	}

	canEdit := hasPermission(admin, "mailboxes.manage") || isExecutiveAdmin(admin)
	if !canEdit && mailbox.Type == domain.AdminMailboxTypePersonal && mailbox.OwnerAdminID.Valid && mailbox.OwnerAdminID.String == admin.ID {
		canEdit = true
	}
	if !canEdit {
		return nil, domain.NewError("ADMIN_FORBIDDEN", "Anda tidak memiliki akses untuk mengubah nama pengirim mailbox ini", 403)
	}

	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return nil, domain.ErrValidationFailed("Nama pengirim tidak boleh kosong")
	}
	if len(displayName) > 150 {
		return nil, domain.ErrValidationFailed("Nama pengirim maksimal 150 karakter")
	}

	mailbox.DisplayName = displayName
	if err := s.repo.UpdateMailbox(ctx, mailbox); err != nil {
		return nil, fmt.Errorf("failed to update mailbox display name: %w", err)
	}

	return s.buildMailboxPayload(ctx, mailbox.ID)
}

func (s *AdminMailboxService) UpdateThreadStatus(ctx context.Context, adminID, threadID string, req ThreadStatusUpdateRequest) error {
	admin, err := s.requireAdmin(ctx, adminID)
	if err != nil {
		return err
	}
	if !hasPermission(admin, "mailboxes.status.manage") {
		return domain.NewError("ADMIN_FORBIDDEN", "Anda tidak memiliki akses untuk mengubah status inbox", 403)
	}

	thread, err := s.repo.FindThreadByID(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to get thread: %w", err)
	}
	if thread == nil {
		return domain.NewError("THREAD_NOT_FOUND", "Thread email tidak ditemukan", 404)
	}
	mailbox, err := s.repo.FindMailboxByID(ctx, thread.MailboxID)
	if err != nil {
		return fmt.Errorf("failed to get mailbox: %w", err)
	}
	if mailbox == nil || !s.canViewMailbox(ctx, admin, mailbox) {
		return domain.NewError("MAILBOX_FORBIDDEN", "Anda tidak memiliki akses ke thread ini", 403)
	}

	status := normalizeThreadStatus(req.Status)
	if status == "" {
		return domain.ErrValidationFailed("Status thread email tidak valid")
	}
	if err := s.repo.UpdateThreadStatus(ctx, threadID, status); err != nil {
		return fmt.Errorf("failed to update thread status: %w", err)
	}
	return s.repo.AddEmailThreadEvent(ctx, threadID, admin.ID, "status_changed", fmt.Sprintf("Status thread diubah menjadi %s", status), map[string]interface{}{
		"status": status,
	})
}

func (s *AdminMailboxService) ToggleThreadImportant(ctx context.Context, adminID, threadID string, isImportant bool) error {
	admin, err := s.requireAdmin(ctx, adminID)
	if err != nil {
		return err
	}
	if !hasPermission(admin, "mailboxes.reply") {
		return domain.NewError("ADMIN_FORBIDDEN", "Anda tidak memiliki akses", 403)
	}
	thread, err := s.repo.FindThreadByID(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to get thread: %w", err)
	}
	if thread == nil {
		return domain.NewError("THREAD_NOT_FOUND", "Thread email tidak ditemukan", 404)
	}
	return s.repo.UpdateThreadImportant(ctx, threadID, isImportant)
}

func (s *AdminMailboxService) AssignThread(ctx context.Context, adminID, threadID string, req ThreadAssignRequest) error {
	admin, err := s.requireAdmin(ctx, adminID)
	if err != nil {
		return err
	}
	if !hasPermission(admin, "mailboxes.assign") {
		return domain.NewError("ADMIN_FORBIDDEN", "Anda tidak memiliki akses untuk assign thread", 403)
	}

	thread, err := s.repo.FindThreadByID(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to get thread: %w", err)
	}
	if thread == nil {
		return domain.NewError("THREAD_NOT_FOUND", "Thread email tidak ditemukan", 404)
	}
	mailbox, err := s.repo.FindMailboxByID(ctx, thread.MailboxID)
	if err != nil {
		return fmt.Errorf("failed to get mailbox: %w", err)
	}
	if mailbox == nil || !s.canViewMailbox(ctx, admin, mailbox) {
		return domain.NewError("MAILBOX_FORBIDDEN", "Anda tidak memiliki akses ke thread ini", 403)
	}

	var target *string
	if strings.TrimSpace(req.AdminUserID) != "" {
		assignee, err := s.repo.FindAdminByID(ctx, strings.TrimSpace(req.AdminUserID))
		if err != nil {
			return fmt.Errorf("failed to get assignee: %w", err)
		}
		if assignee == nil {
			return domain.NewError("ADMIN_NOT_FOUND", "Admin tujuan tidak ditemukan", 404)
		}
		if !s.canViewMailbox(ctx, assignee, mailbox) {
			return domain.NewError("MAILBOX_ASSIGN_FORBIDDEN", "Admin tujuan tidak memiliki akses ke mailbox ini", 403)
		}
		target = &assignee.ID
	}

	if err := s.repo.AssignThread(ctx, threadID, target); err != nil {
		return fmt.Errorf("failed to assign thread: %w", err)
	}
	return s.repo.AddEmailThreadEvent(ctx, threadID, admin.ID, "assigned", "Assignment thread diperbarui", map[string]interface{}{
		"assignedAdminId": stringPointerValue(target),
	})
}

func (s *AdminMailboxService) ListEmailLogs(ctx context.Context, adminID, search, status, category string, page, perPage int) (*domain.AdminListResponse, error) {
	admin, err := s.requireAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if !hasPermission(admin, "email_logs.view") {
		return nil, domain.NewError("ADMIN_FORBIDDEN", "Anda tidak memiliki akses ke log email", 403)
	}

	items, total, err := s.repo.ListEmailDispatchLogs(ctx, search, status, category, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to list email logs: %w", err)
	}

	return &domain.AdminListResponse{
		Items:   items,
		Page:    page,
		PerPage: perPage,
		Total:   total,
		HasNext: mailboxOffset(page, perPage)+len(items) < total,
	}, nil
}

func (s *AdminMailboxService) HandleInboundSNSEvent(ctx context.Context, rawBody []byte) (map[string]interface{}, error) {
	envelope, err := parseSNSEnvelope(rawBody)
	if err != nil {
		return nil, domain.ErrValidationFailed("Payload SNS inbound email tidak valid")
	}
	if err := s.validateTopic(envelope.TopicARN, s.emailCfg.SES.InboundTopicARN); err != nil {
		return nil, err
	}
	if response, err := s.handleSNSHandshake(ctx, envelope); response != nil || err != nil {
		return response, err
	}

	var notification sesReceiptNotification
	if err := json.Unmarshal([]byte(envelope.Message), &notification); err != nil {
		return nil, domain.ErrValidationFailed("Pesan SNS inbound email tidak valid")
	}
	if s.emailStorage == nil {
		return nil, fmt.Errorf("email inbound storage is not configured")
	}

	objectKey := strings.TrimSpace(notification.Receipt.Action.ObjectKey)
	if objectKey == "" {
		return nil, domain.ErrValidationFailed("Object key email inbound tidak tersedia")
	}

	rawEmail, _, err := s.emailStorage.GetObjectBytes(ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch inbound email object: %w", err)
	}
	parsed, err := parseInboundEmail(rawEmail, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inbound email: %w", err)
	}

	mailbox, err := s.resolveInboundMailbox(ctx, append(notification.Receipt.Recipients, parsed.To...))
	if err != nil {
		return nil, err
	}
	thread, err := s.findOrCreateThreadForInbound(ctx, mailbox, parsed)
	if err != nil {
		return nil, err
	}

	messageID := "aem_" + uuid.New().String()[:8]
	if err := s.repo.CreateEmailMessage(ctx, map[string]interface{}{
		"id":                messageID,
		"threadId":          thread.ID,
		"mailboxId":         mailbox.ID,
		"direction":         domain.AdminEmailDirectionInbound,
		"senderName":        nullableString(parsed.SenderName),
		"senderAddress":     parsed.SenderAddress,
		"toAddresses":       parsed.To,
		"ccAddresses":       parsed.Cc,
		"subject":           parsed.Subject,
		"textBody":          nullableString(parsed.TextBody),
		"htmlBody":          nullableString(parsed.HTMLBody),
		"providerMessageId": nullableString(notification.Mail.MessageID),
		"messageIdHeader":   nullableString(parsed.MessageIDHeader),
		"inReplyTo":         nullableString(parsed.InReplyTo),
		"referencesHeaders": parsed.References,
		"receivedAt":        parsed.ReceivedAt,
		"meta": map[string]interface{}{
			"source": notification.Mail.Source,
		},
		"createdAt": parsed.ReceivedAt,
	}); err != nil {
		return nil, fmt.Errorf("failed to save inbound email: %w", err)
	}

	for _, attachment := range parsed.Attachments {
		storageKey := fmt.Sprintf("email-inbox/attachments/%s/%s/%s", thread.ID, messageID, sanitizeFileName(attachment.FileName))
		if err := s.emailStorage.PutBytes(ctx, attachment.Data, storageKey, attachment.ContentType); err != nil {
			return nil, fmt.Errorf("failed to store attachment: %w", err)
		}
		if err := s.repo.CreateEmailAttachment(ctx, map[string]interface{}{
			"id":          "aea_" + uuid.New().String()[:8],
			"messageId":   messageID,
			"fileName":    attachment.FileName,
			"contentType": attachment.ContentType,
			"sizeBytes":   len(attachment.Data),
			"storageKey":  storageKey,
			"createdAt":   time.Now(),
		}); err != nil {
			return nil, fmt.Errorf("failed to save attachment metadata: %w", err)
		}
	}

	if err := s.repo.UpdateThreadAfterInbound(ctx, thread.ID, parsed.SenderName, parsed.SenderAddress, parsed.Subject, parsed.Normalized, parsed.Preview, parsed.ReceivedAt); err != nil {
		return nil, fmt.Errorf("failed to update inbox thread: %w", err)
	}

	eventType := "inbound_received"
	if thread.Status == domain.AdminEmailThreadStatusSelesai {
		eventType = "reopened"
	}
	if err := s.repo.AddEmailThreadEvent(ctx, thread.ID, "", eventType, "Email inbound diterima", map[string]interface{}{
		"messageId": messageID,
		"mailbox":   mailbox.Address,
	}); err != nil {
		return nil, fmt.Errorf("failed to add inbox event: %w", err)
	}

	return map[string]interface{}{
		"threadId":    thread.ID,
		"messageId":   messageID,
		"mailboxId":   mailbox.ID,
		"mailbox":     mailbox.Address,
		"attachments": len(parsed.Attachments),
	}, nil
}

func (s *AdminMailboxService) HandleDeliveryEvent(ctx context.Context, rawBody []byte) (map[string]interface{}, error) {
	envelope, err := parseSNSEnvelope(rawBody)
	if err != nil {
		return nil, domain.ErrValidationFailed("Payload SNS delivery email tidak valid")
	}
	if err := s.validateTopic(envelope.TopicARN, s.emailCfg.SES.DeliveryTopicARN); err != nil {
		return nil, err
	}
	if response, err := s.handleSNSHandshake(ctx, envelope); response != nil || err != nil {
		return response, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(envelope.Message), &payload); err != nil {
		return nil, domain.ErrValidationFailed("Pesan SNS delivery email tidak valid")
	}

	mailData := mapValue(payload["mail"])
	providerMessageID := stringValue(mailData["messageId"])
	if providerMessageID == "" {
		return nil, domain.ErrValidationFailed("Provider message id email tidak tersedia")
	}

	eventType := strings.ToUpper(firstNonEmpty(stringValue(payload["eventType"]), stringValue(payload["notificationType"])))
	status := mapSESEventStatus(eventType)
	timestamps := map[string]*time.Time{"sentAt": nil, "deliveredAt": nil, "failedAt": nil}
	if ts := parseEventTimestamp(firstNonEmpty(stringValue(mailData["timestamp"]), stringValue(mapValue(payload["delivery"])["timestamp"]), stringValue(mapValue(payload["bounce"])["timestamp"]))); ts != nil {
		switch status {
		case "sent":
			timestamps["sentAt"] = ts
		case "delivered":
			timestamps["deliveredAt"] = ts
		default:
			timestamps["failedAt"] = ts
		}
	}

	errorMessage := extractDeliveryErrorMessage(payload, eventType)
	if err := s.repo.UpdateEmailDispatchLogStatusByProviderMessageID(ctx, providerMessageID, status, errorMessage, timestamps, payload); err != nil {
		return nil, fmt.Errorf("failed to update email dispatch status: %w", err)
	}

	return map[string]interface{}{
		"providerMessageId": providerMessageID,
		"status":            status,
		"eventType":         eventType,
	}, nil
}

func (s *AdminMailboxService) requireAdmin(ctx context.Context, adminID string) (*domain.AdminUser, error) {
	admin, err := s.repo.FindAdminByID(ctx, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	if admin == nil {
		return nil, domain.NewError("ADMIN_NOT_FOUND", "Akun admin tidak ditemukan", 404)
	}
	return admin, nil
}

func (s *AdminMailboxService) requireMailboxAccess(ctx context.Context, adminID, mailboxID string) (*domain.AdminUser, *domain.AdminMailbox, error) {
	admin, err := s.requireAdmin(ctx, adminID)
	if err != nil {
		return nil, nil, err
	}
	mailbox, err := s.repo.FindMailboxByID(ctx, mailboxID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get mailbox: %w", err)
	}
	if mailbox == nil {
		return nil, nil, domain.NewError("MAILBOX_NOT_FOUND", "Mailbox tidak ditemukan", 404)
	}
	if !s.canViewMailbox(ctx, admin, mailbox) {
		return nil, nil, domain.NewError("MAILBOX_FORBIDDEN", "Anda tidak memiliki akses ke mailbox ini", 403)
	}
	return admin, mailbox, nil
}

func (s *AdminMailboxService) canViewMailbox(ctx context.Context, admin *domain.AdminUser, mailbox *domain.AdminMailbox) bool {
	if admin == nil || mailbox == nil {
		return false
	}
	if hasPermission(admin, "mailboxes.view_all") || isExecutiveAdmin(admin) {
		return true
	}
	if mailbox.Type == domain.AdminMailboxTypeSystem {
		return false
	}
	if mailbox.Type == domain.AdminMailboxTypePersonal {
		return mailbox.OwnerAdminID.Valid && mailbox.OwnerAdminID.String == admin.ID
	}
	if !hasPermission(admin, "mailboxes.view_assigned") {
		return false
	}
	if s.isMailboxMember(ctx, mailbox.ID, admin.ID) {
		return true
	}
	return defaultRoleMailboxAccess(admin, mailbox.Address)
}

func (s *AdminMailboxService) canReplyFromMailbox(ctx context.Context, admin *domain.AdminUser, mailbox *domain.AdminMailbox) bool {
	if !hasPermission(admin, "mailboxes.reply") {
		return false
	}
	if hasPermission(admin, "mailboxes.view_all") || isExecutiveAdmin(admin) {
		return true
	}
	switch mailbox.Type {
	case domain.AdminMailboxTypePersonal:
		return mailbox.OwnerAdminID.Valid && mailbox.OwnerAdminID.String == admin.ID
	case domain.AdminMailboxTypeShared:
		if s.isMailboxMember(ctx, mailbox.ID, admin.ID) {
			return true
		}
		return defaultRoleMailboxAccess(admin, mailbox.Address)
	default:
		return false
	}
}

func (s *AdminMailboxService) isMailboxMember(ctx context.Context, mailboxID, adminID string) bool {
	members, err := s.repo.ListMailboxMembers(ctx, mailboxID)
	if err != nil {
		return false
	}
	for _, member := range members {
		if stringValue(member["admin_user_id"]) == adminID {
			return true
		}
	}
	return false
}

func (s *AdminMailboxService) mailboxSection(admin *domain.AdminUser, mailbox *domain.AdminMailbox) string {
	switch mailbox.Type {
	case domain.AdminMailboxTypePersonal:
		if mailbox.OwnerAdminID.Valid && mailbox.OwnerAdminID.String == admin.ID {
			return "my"
		}
		return "personal"
	case domain.AdminMailboxTypeSystem:
		return "system"
	default:
		return "shared"
	}
}

func (s *AdminMailboxService) buildMailboxPayload(ctx context.Context, mailboxID string) (map[string]interface{}, error) {
	mailbox, err := s.repo.FindMailboxByID(ctx, mailboxID)
	if err != nil {
		return nil, err
	}
	if mailbox == nil {
		return nil, domain.NewError("MAILBOX_NOT_FOUND", "Mailbox tidak ditemukan", 404)
	}
	members, err := s.repo.ListMailboxMembers(ctx, mailboxID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"mailbox": mailbox,
		"members": members,
	}, nil
}

func (s *AdminMailboxService) resolveInboundMailbox(ctx context.Context, candidates []string) (*domain.AdminMailbox, error) {
	for _, candidate := range candidates {
		address := normalizeMailboxAddress(candidate)
		if address == "" {
			continue
		}
		mailbox, err := s.repo.FindMailboxByAddress(ctx, address)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve mailbox: %w", err)
		}
		if mailbox != nil {
			return mailbox, nil
		}
	}

	mailbox, err := s.repo.FindMailboxByAddress(ctx, s.unmappedInbox)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve unmapped inbox: %w", err)
	}
	if mailbox == nil {
		return nil, domain.NewError("MAILBOX_NOT_FOUND", "Mailbox fallback tidak ditemukan", 404)
	}
	return mailbox, nil
}

func (s *AdminMailboxService) findOrCreateThreadForInbound(ctx context.Context, mailbox *domain.AdminMailbox, parsed *parsedInboundEmail) (*domain.AdminEmailThread, error) {
	references := make([]string, 0, len(parsed.References)+1)
	if parsed.InReplyTo != "" {
		references = append(references, parsed.InReplyTo)
	}
	references = append(references, parsed.References...)

	thread, err := s.repo.FindThreadByMessageReferences(ctx, mailbox.ID, references)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup email thread by references: %w", err)
	}
	if thread != nil {
		return thread, nil
	}

	thread, err = s.repo.FindThreadByParticipant(ctx, mailbox.ID, parsed.Normalized, parsed.SenderAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup email thread by participant: %w", err)
	}
	if thread != nil {
		return thread, nil
	}

	thread = &domain.AdminEmailThread{
		ID:                "aet_" + uuid.New().String()[:8],
		MailboxID:         mailbox.ID,
		ParticipantEmail:  parsed.SenderAddress,
		Subject:           parsed.Subject,
		NormalizedSubject: parsed.Normalized,
		Status:            domain.AdminEmailThreadStatusBelumDibalas,
		UnreadCount:       1,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.repo.CreateThread(ctx, map[string]interface{}{
		"id":                 thread.ID,
		"mailboxId":          thread.MailboxID,
		"participantName":    nullableString(parsed.SenderName),
		"participantEmail":   parsed.SenderAddress,
		"subject":            parsed.Subject,
		"normalizedSubject":  parsed.Normalized,
		"status":             thread.Status,
		"unreadCount":        0,
		"lastDirection":      domain.AdminEmailDirectionInbound,
		"lastMessagePreview": parsed.Preview,
		"latestMessageAt":    parsed.ReceivedAt,
		"lastInboundAt":      parsed.ReceivedAt,
		"meta": map[string]interface{}{
			"mailboxAddress": mailbox.Address,
		},
		"createdAt": time.Now(),
		"updatedAt": time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to create email thread: %w", err)
	}
	return thread, nil
}

func (s *AdminMailboxService) handleSNSHandshake(ctx context.Context, envelope *snsEnvelope) (map[string]interface{}, error) {
	if envelope == nil {
		return nil, nil
	}
	switch strings.TrimSpace(envelope.Type) {
	case "SubscriptionConfirmation", "UnsubscribeConfirmation":
		if strings.TrimSpace(envelope.SubscribeURL) == "" {
			return nil, nil
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, envelope.SubscribeURL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return map[string]interface{}{"message": "SNS subscription confirmed", "status": resp.StatusCode}, nil
	default:
		return nil, nil
	}
}

func (s *AdminMailboxService) validateTopic(received, expected string) error {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return nil
	}
	if strings.TrimSpace(received) != expected {
		return domain.NewError("EMAIL_TOPIC_FORBIDDEN", "Topic SNS email tidak diizinkan", 403)
	}
	return nil
}

func (s *AdminMailboxService) mailboxDomain() string {
	return strings.TrimPrefix(firstNonEmpty(s.emailCfg.MailboxDomain, "ppob.id"), "@")
}

func normalizeMailboxType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case domain.AdminMailboxTypeSystem, domain.AdminMailboxTypeShared, domain.AdminMailboxTypePersonal:
		return value
	default:
		return ""
	}
}

func normalizeMailboxAddress(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeThreadStatus(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case domain.AdminEmailThreadStatusBelumDibalas, domain.AdminEmailThreadStatusDibalas, domain.AdminEmailThreadStatusSelesai, domain.AdminEmailThreadStatusArsip:
		return value
	default:
		return ""
	}
}

func hasPermission(admin *domain.AdminUser, permission string) bool {
	if admin == nil {
		return false
	}
	return slices.Contains(admin.Permissions, permission)
}

func isExecutiveAdmin(admin *domain.AdminUser) bool {
	if admin == nil {
		return false
	}
	for _, role := range admin.Roles {
		switch strings.ToLower(strings.TrimSpace(role.ID)) {
		case "super_admin", "director", "commissioner":
			return true
		}
	}
	return false
}

func defaultRoleMailboxAccess(admin *domain.AdminUser, mailboxAddress string) bool {
	address := strings.ToLower(strings.TrimSpace(mailboxAddress))
	roleIDs := make([]string, 0, len(admin.Roles))
	for _, role := range admin.Roles {
		roleIDs = append(roleIDs, strings.ToLower(strings.TrimSpace(role.ID)))
	}
	switch address {
	case "cs@ppob.id":
		return slices.Contains(roleIDs, "customer_service")
	case "dpo@ppob.id", "legal@ppob.id":
		return slices.Contains(roleIDs, "compliance_kyc")
	case "partner@ppob.id", "partnership@ppob.id":
		return slices.Contains(roleIDs, "product_content") || slices.Contains(roleIDs, "customer_service")
	case "unmapped@ppob.id":
		return isExecutiveAdmin(admin)
	default:
		return false
	}
}

func parseSNSEnvelope(rawBody []byte) (*snsEnvelope, error) {
	var envelope snsEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return nil, err
	}
	return &envelope, nil
}

func parseInboundEmail(raw []byte, notification sesReceiptNotification) (*parsedInboundEmail, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	parsed := &parsedInboundEmail{
		MessageIDHeader: strings.TrimSpace(msg.Header.Get("Message-ID")),
		InReplyTo:       strings.TrimSpace(msg.Header.Get("In-Reply-To")),
		References:      parseReferenceHeaders(msg.Header),
		Subject:         firstNonEmpty(strings.TrimSpace(msg.Header.Get("Subject")), strings.TrimSpace(notification.Mail.CommonHeaders.Subject), "(Tanpa Subjek)"),
		ReceivedAt:      parseMailTimestamp(notification.Mail.Timestamp),
	}
	parsed.Normalized = normalizeEmailSubject(parsed.Subject)

	if fromList, err := mail.ParseAddressList(msg.Header.Get("From")); err == nil && len(fromList) > 0 {
		parsed.SenderName = strings.TrimSpace(fromList[0].Name)
		parsed.SenderAddress = strings.ToLower(strings.TrimSpace(fromList[0].Address))
	}
	if parsed.SenderAddress == "" {
		parsed.SenderAddress = strings.ToLower(strings.TrimSpace(notification.Mail.Source))
	}

	parsed.To = parseAddressHeaderList(msg.Header.Get("To"))
	if len(parsed.To) == 0 {
		parsed.To = sanitizeEmailList(notification.Mail.CommonHeaders.To)
	}
	parsed.Cc = parseAddressHeaderList(msg.Header.Get("Cc"))
	if len(parsed.Cc) == 0 {
		parsed.Cc = sanitizeEmailList(notification.Mail.CommonHeaders.Cc)
	}

	bodyBytes, err := io.ReadAll(msg.Body)
	if err != nil {
		return nil, err
	}
	if err := consumeEmailBody(msg.Header.Get("Content-Type"), bodyBytes, parsed); err != nil {
		return nil, err
	}
	if strings.TrimSpace(parsed.TextBody) == "" {
		parsed.TextBody = stripHTML(parsed.HTMLBody)
	}
	parsed.Preview = buildPreview(firstNonEmpty(parsed.TextBody, stripHTML(parsed.HTMLBody)))
	if parsed.MessageIDHeader == "" {
		parsed.MessageIDHeader = generateMessageHeader(firstNonEmpty(parsed.SenderAddress, "ppob.id"))
	}
	return parsed, nil
}

func consumeEmailBody(contentType string, body []byte, parsed *parsedInboundEmail) error {
	mediaType, params, _ := mime.ParseMediaType(contentType)
	switch {
	case strings.HasPrefix(mediaType, "multipart/"):
		boundary := params["boundary"]
		if boundary == "" {
			return nil
		}
		reader := multipart.NewReader(bytes.NewReader(body), boundary)
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			partBytes, err := io.ReadAll(part)
			part.Close()
			if err != nil {
				return err
			}
			partType := part.Header.Get("Content-Type")
			fileName := part.FileName()
			if fileName != "" || strings.HasPrefix(strings.ToLower(strings.TrimSpace(part.Header.Get("Content-Disposition"))), "attachment") {
				parsed.Attachments = append(parsed.Attachments, parsedInboundAttachment{
					FileName:    firstNonEmpty(fileName, "attachment.bin"),
					ContentType: firstNonEmpty(partType, "application/octet-stream"),
					Data:        partBytes,
				})
				continue
			}
			if err := consumeEmailBody(partType, partBytes, parsed); err != nil {
				return err
			}
		}
	case mediaType == "text/html":
		if parsed.HTMLBody == "" {
			parsed.HTMLBody = string(body)
		}
	case mediaType == "text/plain" || mediaType == "":
		if parsed.TextBody == "" {
			parsed.TextBody = string(body)
		}
	}
	return nil
}

func parseReferenceHeaders(header mail.Header) []string {
	values := make([]string, 0)
	for _, candidate := range []string{header.Get("References"), header.Get("In-Reply-To")} {
		for _, item := range strings.Fields(candidate) {
			item = strings.TrimSpace(item)
			if item != "" {
				values = append(values, item)
			}
		}
	}
	return uniqueStrings(values)
}

func parseAddressHeaderList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	addresses, err := mail.ParseAddressList(value)
	if err != nil {
		return sanitizeEmailList(strings.Split(value, ","))
	}
	result := make([]string, 0, len(addresses))
	for _, address := range addresses {
		if address != nil {
			result = append(result, strings.ToLower(strings.TrimSpace(address.Address)))
		}
	}
	return uniqueStrings(result)
}

func sanitizeEmailList(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if strings.Contains(value, "<") && strings.Contains(value, ">") {
			if parsed, err := mail.ParseAddress(value); err == nil {
				value = strings.ToLower(strings.TrimSpace(parsed.Address))
			}
		}
		result = append(result, value)
	}
	return uniqueStrings(result)
}

func collectThreadReferences(messages []map[string]interface{}) (string, []string) {
	headers := make([]string, 0, len(messages))
	for _, message := range messages {
		if header := stringValue(message["message_id_header"]); header != "" {
			headers = append(headers, header)
		}
	}
	if len(headers) == 0 {
		return "", nil
	}
	headers = uniqueStrings(headers)
	sort.Strings(headers)
	return headers[len(headers)-1], headers
}

func generateMessageHeader(address string) string {
	host := "ppob.id"
	if parts := strings.Split(strings.TrimSpace(address), "@"); len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
		host = parts[1]
	}
	return fmt.Sprintf("<%s@%s>", strings.ReplaceAll(uuid.New().String(), "-", ""), host)
}

func ensureReplySubject(subject string) string {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return "Re: Percakapan"
	}
	if strings.HasPrefix(strings.ToLower(subject), "re:") {
		return subject
	}
	return "Re: " + subject
}

func normalizeEmailSubject(subject string) string {
	subject = strings.TrimSpace(strings.ToLower(subject))
	for {
		switch {
		case strings.HasPrefix(subject, "re:"):
			subject = strings.TrimSpace(strings.TrimPrefix(subject, "re:"))
		case strings.HasPrefix(subject, "fw:"):
			subject = strings.TrimSpace(strings.TrimPrefix(subject, "fw:"))
		case strings.HasPrefix(subject, "fwd:"):
			subject = strings.TrimSpace(strings.TrimPrefix(subject, "fwd:"))
		default:
			return subject
		}
	}
}

func buildPreview(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\r", " "), "\n", " "))
	if len(value) > 220 {
		return value[:220]
	}
	return value
}

func stripHTML(value string) string {
	replacer := strings.NewReplacer("<br>", "\n", "<br/>", "\n", "<br />", "\n", "</p>", "\n\n", "</div>", "\n")
	value = html.UnescapeString(replacer.Replace(value))
	inTag := false
	builder := strings.Builder{}
	for _, r := range value {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				builder.WriteRune(r)
			}
		}
	}
	return strings.TrimSpace(builder.String())
}

func parseMailTimestamp(value string) time.Time {
	if ts := parseEventTimestamp(value); ts != nil {
		return *ts
	}
	return time.Now()
}

func parseEventTimestamp(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, time.RFC1123Z, time.RFC1123} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed
		}
	}
	return nil
}

func mapSESEventStatus(eventType string) string {
	switch strings.ToUpper(strings.TrimSpace(eventType)) {
	case "SEND":
		return "sent"
	case "DELIVERY":
		return "delivered"
	case "BOUNCE":
		return "bounced"
	case "COMPLAINT":
		return "complaint"
	default:
		return "failed"
	}
}

func extractDeliveryErrorMessage(payload map[string]interface{}, eventType string) *string {
	switch strings.ToUpper(strings.TrimSpace(eventType)) {
	case "BOUNCE":
		bounce := mapValue(payload["bounce"])
		message := firstNonEmpty(stringValue(bounce["bounceType"]), stringValue(bounce["bounceSubType"]))
		if message == "" {
			message = "Email mengalami bounce"
		}
		return stringPointer(message)
	case "COMPLAINT":
		return stringPointer("Email menerima complaint dari penerima")
	case "REJECT":
		return stringPointer("Email ditolak oleh SES")
	default:
		return nil
	}
}

func sanitizeFileName(value string) string {
	value = filepath.Base(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	if value == "." || value == "" {
		return "attachment.bin"
	}
	return value
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func stringPointer(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func stringPointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func mailboxOffset(page, perPage int) int {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	return (page - 1) * perPage
}

func toSMTPAttachments(inputs []EmailAttachmentInput) []internalsmtp.EmailAttachment {
	if len(inputs) == 0 {
		return nil
	}
	out := make([]internalsmtp.EmailAttachment, len(inputs))
	for i, a := range inputs {
		out[i] = internalsmtp.EmailAttachment{
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Data:        a.Data,
		}
	}
	return out
}

func (s *AdminMailboxService) buildSignature(ctx context.Context, adminID string) string {
	admin, err := s.repo.FindAdminByID(ctx, adminID)
	if err != nil || admin == nil {
		return ""
	}

	name := admin.DisplayName()

	// Use personal mailbox address (e.g. riko@ppob.id) instead of login email
	var personalEmail string
	_ = s.repo.DB().GetContext(ctx, &personalEmail, `SELECT address FROM admin_mailboxes WHERE type = 'personal' AND owner_admin_id = $1 LIMIT 1`, adminID)
	if personalEmail == "" {
		personalEmail = admin.Email
	}

	// Load position name
	var positionName string
	if admin.PositionID.Valid {
		_ = s.repo.DB().GetContext(ctx, &positionName, `SELECT name FROM admin_positions WHERE id = $1`, admin.PositionID.String)
	}

	var linkedinURL string
	if admin.LinkedinURL.Valid {
		linkedinURL = admin.LinkedinURL.String
	}

	sig := `<br/><br/><div style="border-top:1px solid #e2e8f0;padding-top:12px;margin-top:12px;">`
	sig += `<p style="font-size:13px;color:#475569;line-height:1.8;margin:0;">`
	sig += `Best Regards,<br/>`
	sig += `<strong style="color:#1e293b;">` + html.EscapeString(name) + `</strong><br/>`
	if positionName != "" {
		sig += html.EscapeString(positionName) + `<br/>`
	}
	sig += `Email: ` + html.EscapeString(personalEmail) + `<br/>`
	if linkedinURL != "" {
		sig += `LinkedIn: <a href="` + html.EscapeString(linkedinURL) + `" style="color:#2563eb;">` + html.EscapeString(linkedinURL) + `</a><br/>`
	}
	sig += `</p></div>`
	return sig
}
