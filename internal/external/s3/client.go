package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// Config represents S3 client configuration
type Config struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string // Optional for S3-compatible services
	PublicURL       string // Base URL for public access
}

// Client represents S3 client
type Client struct {
	s3Client  *s3.Client
	bucket    string
	publicURL string
}

// NewClient creates a new S3 client
func NewClient(cfg Config) (*Client, error) {
	// Configure AWS credentials
	creds := credentials.NewStaticCredentialsProvider(
		cfg.AccessKeyID,
		cfg.SecretAccessKey,
		"",
	)

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	return &Client{
		s3Client:  s3Client,
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
	}, nil
}

// UploadFile uploads a file to S3 and returns the public URL
func (c *Client) UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	// Open file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Read file content
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s/%s-%d%s", folder, uuid.New().String(), time.Now().Unix(), ext)

	// Upload to S3
	_, err = c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Return public URL
	publicURL := fmt.Sprintf("%s/%s", c.publicURL, filename)
	return publicURL, nil
}

// UploadBytes uploads raw bytes to S3 and returns the public URL
func (c *Client) UploadBytes(ctx context.Context, data []byte, folder, filename, contentType string) (string, error) {
	// Generate unique filename
	ext := filepath.Ext(filename)
	uniqueFilename := fmt.Sprintf("%s/%s-%d%s", folder, uuid.New().String(), time.Now().Unix(), ext)

	// Upload to S3
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(uniqueFilename),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Return public URL
	publicURL := fmt.Sprintf("%s/%s", c.publicURL, uniqueFilename)
	return publicURL, nil
}

// DeleteFile deletes a file from S3
func (c *Client) DeleteFile(ctx context.Context, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// GetPresignedURL generates a presigned URL for temporary access
func (c *Client) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	// TODO: Implement presigned URL generation
	// This requires aws-sdk-go-v2/service/s3/presign
	return "", fmt.Errorf("presigned URL not implemented")
}
