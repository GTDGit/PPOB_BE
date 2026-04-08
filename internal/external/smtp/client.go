package smtp

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/GTDGit/PPOB_BE/internal/config"
	"github.com/google/uuid"
)

type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type Client struct {
	cfg config.SMTPConfig
}

type SendMessageInput struct {
	EnvelopeFrom     string // SMTP MAIL FROM (envelope sender); if empty, falls back to FromAddress
	FromAddress      string
	FromName         string
	ToAddresses      []string
	CcAddresses      []string
	BccAddresses     []string
	ReplyToAddresses []string
	Subject          string
	HTMLBody         string
	TextBody         string
	Headers          map[string]string
	Attachments      []EmailAttachment
}

func NewClient(cfg config.SMTPConfig) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) IsEnabled() bool {
	return c != nil && strings.TrimSpace(c.cfg.Host) != "" && c.cfg.Port > 0
}

func (c *Client) Send(ctx context.Context, input SendMessageInput) (string, error) {
	if !c.IsEnabled() {
		return "", fmt.Errorf("smtp client is not enabled")
	}
	if len(input.ToAddresses) == 0 {
		return "", fmt.Errorf("at least one recipient is required")
	}

	from := strings.TrimSpace(input.FromAddress)
	if from == "" {
		return "", fmt.Errorf("from address is required")
	}

	// EnvelopeFrom is the SMTP MAIL FROM (envelope sender).
	// Some providers (e.g. Alibaba DirectMail) require this to match the
	// authenticated account, while the From header can differ.
	envelopeFrom := strings.TrimSpace(input.EnvelopeFrom)
	if envelopeFrom == "" {
		envelopeFrom = from
	}

	messageID := fmt.Sprintf("<%s@ppob.id>", uuid.New().String())

	msg := buildMessage(input, messageID)

	allRecipients := make([]string, 0, len(input.ToAddresses)+len(input.CcAddresses)+len(input.BccAddresses))
	allRecipients = append(allRecipients, input.ToAddresses...)
	allRecipients = append(allRecipients, input.CcAddresses...)
	allRecipients = append(allRecipients, input.BccAddresses...)

	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)
	auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)

	if c.cfg.UseTLS {
		if err := sendWithTLS(addr, auth, envelopeFrom, allRecipients, msg, c.cfg.Host); err != nil {
			return "", fmt.Errorf("smtp send (TLS): %w", err)
		}
	} else {
		if err := sendWithSTARTTLS(addr, auth, envelopeFrom, allRecipients, msg, c.cfg.Host); err != nil {
			return "", fmt.Errorf("smtp send (STARTTLS): %w", err)
		}
	}

	return messageID, nil
}

func sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	tlsConfig := &tls.Config{ServerName: host}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("new client: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("rcpt to %s: %w", addr, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		w.Close()
		return fmt.Errorf("write: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}

	return client.Quit()
}

func sendWithSTARTTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, host string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("new client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: host}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("rcpt to %s: %w", addr, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		w.Close()
		return fmt.Errorf("write: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}

	return client.Quit()
}

func buildMessage(input SendMessageInput, messageID string) []byte {
	var buf strings.Builder

	fromAddr := input.FromAddress
	if strings.TrimSpace(input.FromName) != "" {
		fromAddr = (&mail.Address{Name: input.FromName, Address: input.FromAddress}).String()
	}

	header := textproto.MIMEHeader{}
	header.Set("From", fromAddr)
	header.Set("To", strings.Join(input.ToAddresses, ", "))
	if len(input.CcAddresses) > 0 {
		header.Set("Cc", strings.Join(input.CcAddresses, ", "))
	}
	if len(input.ReplyToAddresses) > 0 {
		header.Set("Reply-To", strings.Join(input.ReplyToAddresses, ", "))
	}
	header.Set("Subject", input.Subject)
	header.Set("Message-ID", messageID)
	header.Set("MIME-Version", "1.0")

	for key, value := range input.Headers {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		k := strings.ToLower(key)
		if k == "from" || k == "to" || k == "cc" || k == "subject" || k == "message-id" || k == "mime-version" {
			continue
		}
		header.Set(key, value)
	}

	for key, values := range header {
		for _, v := range values {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, v))
		}
	}

	hasHTML := strings.TrimSpace(input.HTMLBody) != ""
	hasText := strings.TrimSpace(input.TextBody) != ""
	hasAttachments := len(input.Attachments) > 0

	if hasAttachments {
		// multipart/mixed: body part (alternative or single) + attachment parts
		mixedBoundary := fmt.Sprintf("mixed_%s", uuid.New().String()[:8])
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", mixedBoundary))
		buf.WriteString("\r\n")

		mixedWriter := multipart.NewWriter(&buf)
		mixedWriter.SetBoundary(mixedBoundary)

		// Write body part
		if hasHTML && hasText {
			altBoundary := fmt.Sprintf("alt_%s", uuid.New().String()[:8])
			bodyPart, _ := mixedWriter.CreatePart(textproto.MIMEHeader{
				"Content-Type": {fmt.Sprintf("multipart/alternative; boundary=\"%s\"", altBoundary)},
			})
			altWriter := multipart.NewWriter(bodyPart)
			altWriter.SetBoundary(altBoundary)

			textPart, _ := altWriter.CreatePart(textproto.MIMEHeader{
				"Content-Type":              {"text/plain; charset=UTF-8"},
				"Content-Transfer-Encoding": {"quoted-printable"},
			})
			textPart.Write([]byte(input.TextBody))

			htmlPart, _ := altWriter.CreatePart(textproto.MIMEHeader{
				"Content-Type":              {"text/html; charset=UTF-8"},
				"Content-Transfer-Encoding": {"quoted-printable"},
			})
			htmlPart.Write([]byte(input.HTMLBody))
			altWriter.Close()
		} else if hasHTML {
			bodyPart, _ := mixedWriter.CreatePart(textproto.MIMEHeader{
				"Content-Type":              {"text/html; charset=UTF-8"},
				"Content-Transfer-Encoding": {"quoted-printable"},
			})
			bodyPart.Write([]byte(input.HTMLBody))
		} else {
			bodyPart, _ := mixedWriter.CreatePart(textproto.MIMEHeader{
				"Content-Type":              {"text/plain; charset=UTF-8"},
				"Content-Transfer-Encoding": {"quoted-printable"},
			})
			bodyPart.Write([]byte(input.TextBody))
		}

		// Write attachment parts
		for _, att := range input.Attachments {
			ct := att.ContentType
			if ct == "" {
				ct = "application/octet-stream"
			}
			attPart, _ := mixedWriter.CreatePart(textproto.MIMEHeader{
				"Content-Type":              {fmt.Sprintf("%s; name=\"%s\"", ct, att.Filename)},
				"Content-Transfer-Encoding": {"base64"},
				"Content-Disposition":       {fmt.Sprintf("attachment; filename=\"%s\"", att.Filename)},
			})
			encoded := base64.StdEncoding.EncodeToString(att.Data)
			// Write base64 in 76-char lines
			for i := 0; i < len(encoded); i += 76 {
				end := i + 76
				if end > len(encoded) {
					end = len(encoded)
				}
				attPart.Write([]byte(encoded[i:end] + "\r\n"))
			}
		}
		mixedWriter.Close()
	} else if hasHTML && hasText {
		boundary := fmt.Sprintf("boundary_%s", uuid.New().String()[:8])
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		buf.WriteString("\r\n")

		writer := multipart.NewWriter(&buf)
		writer.SetBoundary(boundary)

		textPart, _ := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {"text/plain; charset=UTF-8"},
			"Content-Transfer-Encoding": {"quoted-printable"},
		})
		textPart.Write([]byte(input.TextBody))

		htmlPart, _ := writer.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {"text/html; charset=UTF-8"},
			"Content-Transfer-Encoding": {"quoted-printable"},
		})
		htmlPart.Write([]byte(input.HTMLBody))

		writer.Close()
	} else if hasHTML {
		buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(input.HTMLBody)
	} else {
		buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
		buf.WriteString("\r\n")
		buf.WriteString(input.TextBody)
	}

	return []byte(buf.String())
}
