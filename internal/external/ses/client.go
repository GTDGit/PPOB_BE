package ses

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/GTDGit/PPOB_BE/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type Client struct {
	cfg    config.SESConfig
	client *sesv2.Client
}

type SendMessageInput struct {
	FromAddress          string
	FromName             string
	ToAddresses          []string
	CcAddresses          []string
	BccAddresses         []string
	ReplyToAddresses     []string
	Subject              string
	HTMLBody             string
	TextBody             string
	Headers              map[string]string
	ConfigurationSetName string
	Tags                 map[string]string
}

func NewClient(cfg config.SESConfig) (*Client, error) {
	creds := credentials.NewStaticCredentialsProvider(
		cfg.AccessKeyID,
		cfg.SecretAccessKey,
		"",
	)

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	return &Client{
		cfg:    cfg,
		client: sesv2.NewFromConfig(awsCfg),
	}, nil
}

func (c *Client) IsEnabled() bool {
	return c != nil && c.client != nil && strings.TrimSpace(c.cfg.Region) != ""
}

func (c *Client) Send(ctx context.Context, input SendMessageInput) (string, error) {
	if !c.IsEnabled() {
		return "", fmt.Errorf("ses client is not enabled")
	}
	if len(input.ToAddresses) == 0 {
		return "", fmt.Errorf("at least one recipient is required")
	}

	from := strings.TrimSpace(input.FromAddress)
	if from == "" {
		return "", fmt.Errorf("from address is required")
	}
	if strings.TrimSpace(input.FromName) != "" {
		from = (&mail.Address{
			Name:    input.FromName,
			Address: input.FromAddress,
		}).String()
	}

	headers := make([]types.MessageHeader, 0, len(input.Headers))
	for key, value := range input.Headers {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		headers = append(headers, types.MessageHeader{
			Name:  aws.String(key),
			Value: aws.String(value),
		})
	}

	emailTags := make([]types.MessageTag, 0, len(input.Tags))
	for key, value := range input.Tags {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		emailTags = append(emailTags, types.MessageTag{
			Name:  aws.String(key),
			Value: aws.String(value),
		})
	}

	out, err := c.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress:   aws.String(from),
		ReplyToAddresses:   input.ReplyToAddresses,
		ConfigurationSetName: aws.String(strings.TrimSpace(input.ConfigurationSetName)),
		Destination: &types.Destination{
			ToAddresses:  input.ToAddresses,
			CcAddresses:  input.CcAddresses,
			BccAddresses: input.BccAddresses,
		},
		EmailTags: emailTags,
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    aws.String(input.Subject),
					Charset: aws.String("UTF-8"),
				},
				Body: &types.Body{
					Html: optionalContent(input.HTMLBody),
					Text: optionalContent(input.TextBody),
				},
				Headers: headers,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return aws.ToString(out.MessageId), nil
}

func optionalContent(value string) *types.Content {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &types.Content{
		Data:    aws.String(value),
		Charset: aws.String("UTF-8"),
	}
}
