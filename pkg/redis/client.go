package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/config"
	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with helper methods
type Client struct {
	*redis.Client
}

// NewClient creates a new Redis client
func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Ping to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Client{rdb}, nil
}

// SetJSON stores a value as JSON
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return c.Set(ctx, key, data, expiration).Err()
}

// GetJSON retrieves and unmarshals a JSON value
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// OTP Session Keys
func OTPSessionKey(phone, sessionID string) string {
	return fmt.Sprintf("otp:%s:%s", phone, sessionID)
}

func OTPRateLimitKey(phone string) string {
	return fmt.Sprintf("otp:rate:%s", phone)
}

func OTPResendCountKey(phone string) string {
	return fmt.Sprintf("otp:resend:%s", phone)
}

// Temp Token Keys
func TempTokenKey(token string) string {
	return fmt.Sprintf("temp_token:%s", token)
}

// PIN Attempts Keys
func PINAttemptsKey(userID string) string {
	return fmt.Sprintf("pin:attempts:%s", userID)
}

func PINLockKey(userID string) string {
	return fmt.Sprintf("pin:lock:%s", userID)
}

// Session Keys
func SessionKey(accessToken string) string {
	return fmt.Sprintf("session:%s", accessToken)
}

// Rate Limit Keys
func RateLimitKey(ip, endpoint string) string {
	return fmt.Sprintf("rate:%s:%s", ip, endpoint)
}

// Email Verification Key
func EmailVerificationKey(token string) string {
	return fmt.Sprintf("email_verify:%s", token)
}

// Product cache keys
func ProductListKey(category string) string {
	return fmt.Sprintf("products:list:%s", category)
}

func ProductByIDKey(id string) string {
	return fmt.Sprintf("products:id:%s", id)
}

func ProductBySKUKey(skuCode string) string {
	return fmt.Sprintf("products:sku:%s", skuCode)
}

func ProductsLastSyncKey() string {
	return "products:last_sync"
}
