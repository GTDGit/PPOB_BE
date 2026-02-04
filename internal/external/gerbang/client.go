package gerbang

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/GTDGit/PPOB_BE/pkg/httplog"
)

type Config struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	Timeout      time.Duration
}

type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new Gerbang API client
func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout:   config.Timeout,
			Transport: httplog.NewTransport(nil, nil),
		},
	}
}

// doRequest performs HTTP request with authentication headers
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.config.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication headers as per documentation
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.ClientSecret)
	req.Header.Set("X-Client-Id", c.config.ClientID)
	req.Header.Set("User-Agent", "PPOB.ID/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle 5xx errors BEFORE parsing JSON (server may return HTML error page)
	if resp.StatusCode >= 500 {
		return nil, &Error{
			Code:    resp.StatusCode,
			Message: fmt.Sprintf("server error: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		}
	}

	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		// Include status code and body snippet in error for debugging
		return nil, fmt.Errorf("failed to unmarshal response (status=%d): %w, body: %s",
			resp.StatusCode, err, truncateBody(respBody, 200))
	}

	// Check for API errors
	if !apiResp.Success {
		return nil, ParseError(&apiResp)
	}

	return &apiResp, nil
}

const maxRetries = 3

// doRequestWithRetry performs HTTP request with retry on transient errors
func (c *Client) doRequestWithRetry(ctx context.Context, method, path string, body interface{}) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<attempt) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		resp, err := c.doRequest(ctx, method, path, body)
		if err == nil {
			return resp, nil
		}

		// Only retry on retryable errors
		if c.isRetryableError(err) {
			lastErr = err
			continue
		}

		// Non-retryable error, return immediately
		return nil, err
	}

	return nil, fmt.Errorf("max retries (%d) exceeded: %w", maxRetries, lastErr)
}

// isRetryableError checks if error is transient and worth retrying
func (c *Client) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for Gerbang API error with 5xx status
	if gerr, ok := err.(*Error); ok {
		return gerr.Code >= 500
	}

	// Check for network/timeout errors
	errMsg := err.Error()
	return strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "no such host")
}

// truncateBody truncates body for error messages
func truncateBody(body []byte, maxLen int) string {
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen]) + "..."
}

// parseData parses the response data into target struct
func (c *Client) parseData(resp *Response, target interface{}) error {
	if resp.Data == nil {
		return fmt.Errorf("response data is nil")
	}

	jsonData, err := json.Marshal(resp.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}
