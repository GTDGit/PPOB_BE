package httplog

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// Transport is an http.RoundTripper that logs requests and responses in development mode
type Transport struct {
	Transport http.RoundTripper
	Logger    *slog.Logger
}

// NewTransport creates a logging transport (only logs in development mode)
func NewTransport(base http.RoundTripper, logger *slog.Logger) http.RoundTripper {
	if os.Getenv("APP_ENV") != "development" {
		// In production, just return the base transport without logging
		if base == nil {
			return http.DefaultTransport
		}
		return base
	}

	if base == nil {
		base = http.DefaultTransport
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Transport{
		Transport: base,
		Logger:    logger,
	}
}

// RoundTrip implements http.RoundTripper
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Read and log request body
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	// Format request body for logging
	reqBodyStr := formatBody(reqBody)

	// Get client name from request host
	clientName := getClientName(req.URL.Host)

	t.Logger.Info("→ HTTP Request",
		"client", clientName,
		"method", req.Method,
		"url", req.URL.String(),
		"headers", sanitizeHeaders(req.Header),
		"body", reqBodyStr,
	)

	// Execute request
	resp, err := t.Transport.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		t.Logger.Error("✗ HTTP Error",
			"client", clientName,
			"method", req.Method,
			"url", req.URL.String(),
			"error", err.Error(),
			"duration_ms", duration.Milliseconds(),
		)
		return nil, err
	}

	// Read and log response body
	var respBody []byte
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	}

	// Format response body for logging
	respBodyStr := formatBody(respBody)

	logFn := t.Logger.Info
	if resp.StatusCode >= 400 {
		logFn = t.Logger.Warn
	}
	if resp.StatusCode >= 500 {
		logFn = t.Logger.Error
	}

	logFn("← HTTP Response",
		"client", clientName,
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode,
		"status_text", http.StatusText(resp.StatusCode),
		"body", respBodyStr,
		"duration_ms", duration.Milliseconds(),
	)

	return resp, nil
}

// formatBody formats body for logging (pretty print JSON, truncate if too long)
func formatBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	// Try to pretty print JSON
	var prettyJSON bytes.Buffer
	if json.Indent(&prettyJSON, body, "", "  ") == nil {
		return truncate(prettyJSON.String(), 2000)
	}

	return truncate(string(body), 2000)
}

// truncate truncates string to max length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "... (truncated)"
}

// sanitizeHeaders removes sensitive headers for logging
func sanitizeHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	sensitiveHeaders := []string{"authorization", "x-client-secret", "api-key", "x-api-key"}

	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		isSensitive := false
		for _, sensitive := range sensitiveHeaders {
			if lowerKey == sensitive {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			result[key] = "***REDACTED***"
		} else if len(values) > 0 {
			result[key] = values[0]
		}
	}

	return result
}

// getClientName extracts client name from host
func getClientName(host string) string {
	switch {
	case strings.Contains(host, "gtd.co.id"):
		return "Gerbang"
	case strings.Contains(host, "brevo.com") || strings.Contains(host, "sendinblue"):
		return "Brevo"
	case strings.Contains(host, "fazpass.com"):
		return "Fazpass"
	case strings.Contains(host, "facebook.com") || strings.Contains(host, "whatsapp"):
		return "WhatsApp"
	default:
		return host
	}
}
