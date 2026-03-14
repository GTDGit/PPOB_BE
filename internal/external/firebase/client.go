package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/GTDGit/PPOB_BE/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const messagingScope = "https://www.googleapis.com/auth/firebase.messaging"

type Client struct {
	projectID   string
	tokenSource oauth2.TokenSource
	httpClient  *http.Client
	enabled     bool
}

type SendRequest struct {
	Token string
	Title string
	Body  string
	Data  map[string]string
}

type serviceAccount struct {
	ProjectID string `json:"project_id"`
}

type fcmRequest struct {
	Message messagePayload `json:"message"`
}

type messagePayload struct {
	Token        string            `json:"token"`
	Notification notificationBody  `json:"notification"`
	Data         map[string]string `json:"data,omitempty"`
}

type notificationBody struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type fcmErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient creates a Firebase Cloud Messaging client.
func NewClient(cfg config.FirebaseConfig) (*Client, error) {
	if !cfg.Enabled {
		return &Client{enabled: false}, nil
	}
	if cfg.ServiceAccountPath == "" {
		return nil, fmt.Errorf("firebase service account path is required")
	}

	raw, err := os.ReadFile(cfg.ServiceAccountPath)
	if err != nil {
		return nil, fmt.Errorf("read firebase service account: %w", err)
	}

	projectID := cfg.ProjectID
	if projectID == "" {
		var sa serviceAccount
		if err := json.Unmarshal(raw, &sa); err != nil {
			return nil, fmt.Errorf("parse firebase service account: %w", err)
		}
		projectID = sa.ProjectID
	}
	if projectID == "" {
		return nil, fmt.Errorf("firebase project id is required")
	}

	creds, err := google.CredentialsFromJSON(context.Background(), raw, messagingScope)
	if err != nil {
		return nil, fmt.Errorf("firebase credentials: %w", err)
	}

	return &Client{
		projectID:   projectID,
		tokenSource: creds.TokenSource,
		httpClient:  &http.Client{},
		enabled:     true,
	}, nil
}

// IsEnabled reports whether FCM is configured.
func (c *Client) IsEnabled() bool {
	return c != nil && c.enabled
}

// Send delivers a push notification to a single device token.
func (c *Client) Send(ctx context.Context, req SendRequest) error {
	if !c.IsEnabled() {
		return nil
	}
	if req.Token == "" {
		return fmt.Errorf("firebase token is required")
	}

	token, err := c.tokenSource.Token()
	if err != nil {
		return fmt.Errorf("firebase access token: %w", err)
	}

	body, err := json.Marshal(fcmRequest{
		Message: messagePayload{
			Token: req.Token,
			Notification: notificationBody{
				Title: req.Title,
				Body:  req.Body,
			},
			Data: req.Data,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal firebase payload: %w", err)
	}

	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", c.projectID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create firebase request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send firebase request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		var errResp fcmErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
			return fmt.Errorf("firebase send failed: %s", errResp.Error.Message)
		}
		return fmt.Errorf("firebase send failed: status %d", resp.StatusCode)
	}

	return nil
}
