package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App           AppConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	JWT           JWTConfig
	Admin         AdminConfig
	OTP           OTPConfig
	WhatsApp      WhatsAppConfig
	Fazpass       FazpassConfig
	Email         EmailConfig
	Brevo         BrevoConfig
	Firebase      FirebaseConfig
	Gerbang       GerbangConfig
	S3            S3Config
	Fallback      FallbackConfig
	ProductSync   ProductSyncConfig
	BankCodeSync  BankCodeSyncConfig
	TerritorySync TerritorySyncConfig
}

type AppConfig struct {
	Name        string
	Env         string
	Port        int
	URL         string
	FrontendURL string // Frontend app URL for email links
}

type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxConnections  int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret       string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	TempTokenTTL time.Duration
}

type AdminConfig struct {
	FrontendURL       string
	InviteBaseURL     string
	CORSOrigin        string
	JWTSecret         string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
	InviteTTL         time.Duration
	TOTPIssuer        string
	BootstrapSecret   string
	BootstrapEmail    string
	BootstrapPhone    string
	BootstrapFullName string
	BootstrapRoleID   string
}

type OTPConfig struct {
	Length         int
	TTL            time.Duration
	MaxAttempts    int
	MaxResend      int
	ResendCooldown time.Duration
	TestEnabled    bool
	TestPhone      string
	TestOTP        string
}

type WhatsAppConfig struct {
	APIURL              string
	PhoneNumberID       string
	AccessToken         string
	OTPTemplateName     string
	OTPTemplateLanguage string
	OTPButtonSubType    string
	OTPButtonIndex      string
	WebhookVerifyToken  string
	AppSecret           string
}

type FazpassConfig struct {
	APIURL      string
	MerchantKey string
	GatewayKey  string
}

type EmailConfig struct {
	Provider         string
	DefaultFromEmail string
	DefaultFromName  string
	ReplyToEmail     string
	MailboxDomain    string
	SES              SESConfig
}

type SESConfig struct {
	Region                         string
	AccessKeyID                    string
	SecretAccessKey                string
	ConfigurationSetTransactional  string
	ConfigurationSetOperations     string
	MailFromDomain                 string
	InboundBucket                  string
	InboundTopicARN                string
	DeliveryTopicARN               string
}

type BrevoConfig struct {
	APIURL               string
	APIKey               string
	SenderEmail          string
	SenderName           string
	BaseURL              string // Backend API base URL
	FrontendURL          string // Frontend app URL for email links
	TemplateVerification int64  // Email verification template ID
	TemplateNewLogin     int64  // New login alert template ID
	TemplatePINChanged   int64  // PIN changed alert template ID
	TemplateEmailChanged int64  // Email changed alert template ID
	TemplatePhoneChanged int64  // Phone changed alert template ID
}

type FirebaseConfig struct {
	Enabled            bool
	ProjectID          string
	ServiceAccountPath string
}

type GerbangConfig struct {
	BaseURL        string
	ClientID       string
	ClientSecret   string
	CallbackSecret string
	Timeout        time.Duration
}

type FallbackConfig struct {
	KYCEnabled     bool
	PaymentEnabled bool
	PPOBEnabled    bool
}

type ProductSyncConfig struct {
	Interval      time.Duration // Sync interval (default: 15 minutes)
	EnableOnStart bool          // Run sync immediately on startup
	Enabled       bool          // Enable/disable sync job
}

type BankCodeSyncConfig struct {
	Interval      time.Duration // Sync interval (default: 72 hours / 3 days)
	EnableOnStart bool          // Run sync immediately on startup
	Enabled       bool          // Enable/disable sync job
}

// S3Config untuk menyimpan file KYC (KTP, face photos)
// Note: Liveness pakai Gerbang API, bukan S3 langsung
type S3Config struct {
	Bucket    string
	Region    string // ap-southeast-3 (Jakarta) for data residency
	AccessKey string
	SecretKey string
	BaseURL   string // untuk generate signed URLs
}

type TerritorySyncConfig struct {
	Interval      time.Duration // Sync interval (default: 30 days)
	EnableOnStart bool          // Run sync immediately on startup
	Enabled       bool          // Enable/disable sync job
}

func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "ppob.id"),
			Env:         getEnv("APP_ENV", "development"),
			Port:        getEnvAsInt("APP_PORT", 8080),
			URL:         getEnv("APP_URL", "http://localhost:8080"),
			FrontendURL: getEnv("FRONTEND_URL", "https://ppob.id"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			Name:            getEnv("DB_NAME", "ppob_db"),
			User:            getEnv("DB_USER", "ppob_user"),
			Password:        getEnvRequired("DB_PASSWORD"), // SECURITY: No default - must be set!
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxConnections:  getEnvAsInt("DB_MAX_CONNECTIONS", 100),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 10),
			ConnMaxLifetime: time.Hour,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:       getEnvRequired("JWT_SECRET"),                                         // SECURITY: No default - must be set!
			AccessTTL:    time.Duration(getEnvAsInt("JWT_ACCESS_TTL", 3600)) * time.Second,     // 1 hour (fintech UX)
			RefreshTTL:   time.Duration(getEnvAsInt("JWT_REFRESH_TTL", 7776000)) * time.Second, // 90 days (fintech UX)
			TempTokenTTL: time.Duration(getEnvAsInt("TEMP_TOKEN_TTL", 900)) * time.Second,
		},
		Admin: AdminConfig{
			FrontendURL:       getEnv("ADMIN_FRONTEND_URL", "https://console.ppob.id"),
			InviteBaseURL:     getEnv("ADMIN_INVITE_BASE_URL", "https://console.ppob.id/activate"),
			CORSOrigin:        getEnv("ADMIN_CORS_ORIGIN", "https://console.ppob.id"),
			JWTSecret:         getEnv("ADMIN_JWT_SECRET", getEnvRequired("JWT_SECRET")),
			AccessTTL:         time.Duration(getEnvAsInt("ADMIN_JWT_ACCESS_TTL", 3600)) * time.Second,
			RefreshTTL:        time.Duration(getEnvAsInt("ADMIN_JWT_REFRESH_TTL", 2592000)) * time.Second,
			InviteTTL:         time.Duration(getEnvAsInt("ADMIN_INVITE_TTL_HOURS", 72)) * time.Hour,
			TOTPIssuer:        getEnv("ADMIN_TOTP_ISSUER", "PPOB.ID Admin"),
			BootstrapSecret:   getEnv("ADMIN_BOOTSTRAP_SECRET", ""),
			BootstrapEmail:    getEnv("ADMIN_BOOTSTRAP_EMAIL", ""),
			BootstrapPhone:    getEnv("ADMIN_BOOTSTRAP_PHONE", ""),
			BootstrapFullName: getEnv("ADMIN_BOOTSTRAP_FULL_NAME", ""),
			BootstrapRoleID:   getEnv("ADMIN_BOOTSTRAP_ROLE_ID", "super_admin"),
		},
		OTP: OTPConfig{
			Length:         getEnvAsInt("OTP_LENGTH", 4),                             // 4 digits (SMS standard)
			TTL:            time.Duration(getEnvAsInt("OTP_TTL", 300)) * time.Second, // 5 minutes (security best practice)
			MaxAttempts:    getEnvAsInt("OTP_MAX_ATTEMPTS", 5),
			MaxResend:      getEnvAsInt("OTP_MAX_RESEND", 3),
			ResendCooldown: time.Duration(getEnvAsInt("OTP_RESEND_COOLDOWN", 60)) * time.Second,
			TestEnabled:    getEnv("OTP_TEST_ENABLED", "false") == "true",
			TestPhone:      getEnv("OTP_TEST_PHONE", ""),
			TestOTP:        getEnv("OTP_TEST_CODE", ""),
		},
		WhatsApp: WhatsAppConfig{
			APIURL:              getEnv("WA_API_URL", "https://graph.facebook.com/v24.0"),
			PhoneNumberID:       getEnv("WA_PHONE_NUMBER_ID", ""),
			AccessToken:         getEnv("WA_ACCESS_TOKEN", ""),
			OTPTemplateName:     getEnv("WA_OTP_TEMPLATE_NAME", "otp_verification"),
			OTPTemplateLanguage: getEnv("WA_OTP_TEMPLATE_LANGUAGE", "id"),
			OTPButtonSubType:    getEnv("WA_OTP_BUTTON_SUBTYPE", ""),
			OTPButtonIndex:      getEnv("WA_OTP_BUTTON_INDEX", "0"),
			WebhookVerifyToken:  getEnv("WA_WEBHOOK_VERIFY_TOKEN", ""),
			AppSecret:           getEnv("WA_APP_SECRET", ""),
		},
		Fazpass: FazpassConfig{
			APIURL:      getEnv("FAZPASS_API_URL", "https://api.fazpass.com"),
			MerchantKey: getEnv("FAZPASS_MERCHANT_KEY", ""),
			GatewayKey:  getEnv("FAZPASS_GATEWAY_KEY", ""),
		},
		Email: EmailConfig{
			Provider:         getEnv("EMAIL_PROVIDER", "brevo"),
			DefaultFromEmail: getEnv("EMAIL_FROM_DEFAULT", "noreply@ppob.id"),
			DefaultFromName:  getEnv("EMAIL_FROM_NAME", "PPOB.ID"),
			ReplyToEmail:     getEnv("EMAIL_REPLY_TO", "cs@ppob.id"),
			MailboxDomain:    getEnv("EMAIL_MAILBOX_DOMAIN", "ppob.id"),
			SES: SESConfig{
				Region:                        getEnv("SES_REGION", "ap-southeast-3"),
				AccessKeyID:                   getEnv("SES_ACCESS_KEY_ID", getEnv("S3_ACCESS_KEY", "")),
				SecretAccessKey:               getEnv("SES_SECRET_ACCESS_KEY", getEnv("S3_SECRET_KEY", "")),
				ConfigurationSetTransactional: getEnv("SES_CONFIGURATION_SET_TRANSACTIONAL", "ppob-transactional"),
				ConfigurationSetOperations:    getEnv("SES_CONFIGURATION_SET_OPERATIONS", "ppob-operations"),
				MailFromDomain:                getEnv("SES_MAIL_FROM_DOMAIN", "bounce.ppob.id"),
				InboundBucket:                 getEnv("SES_INBOUND_BUCKET", getEnv("S3_BUCKET", "")),
				InboundTopicARN:               getEnv("SES_INBOUND_TOPIC_ARN", ""),
				DeliveryTopicARN:              getEnv("SES_DELIVERY_TOPIC_ARN", ""),
			},
		},
		Brevo: BrevoConfig{
			APIURL:               getEnv("BREVO_API_URL", "https://api.brevo.com/v3"),
			APIKey:               getEnv("BREVO_API_KEY", ""),
			SenderEmail:          getEnv("BREVO_SENDER_EMAIL", "noreply@ppob.id"),
			SenderName:           getEnv("BREVO_SENDER_NAME", "PPOB.ID"),
			BaseURL:              getEnv("APP_URL", "http://localhost:8080"),
			FrontendURL:          getEnv("FRONTEND_URL", "https://ppob.id"),
			TemplateVerification: int64(getEnvAsInt("BREVO_TEMPLATE_VERIFICATION", 1)),
			TemplateNewLogin:     int64(getEnvAsInt("BREVO_TEMPLATE_NEW_LOGIN", 2)),
			TemplatePINChanged:   int64(getEnvAsInt("BREVO_TEMPLATE_PIN_CHANGED", 3)),
			TemplateEmailChanged: int64(getEnvAsInt("BREVO_TEMPLATE_EMAIL_CHANGED", 4)),
			TemplatePhoneChanged: int64(getEnvAsInt("BREVO_TEMPLATE_PHONE_CHANGED", 5)),
		},
		Firebase: FirebaseConfig{
			Enabled:            getEnv("FIREBASE_ENABLED", "false") == "true",
			ProjectID:          getEnv("FIREBASE_PROJECT_ID", ""),
			ServiceAccountPath: getEnv("FIREBASE_SERVICE_ACCOUNT_PATH", ""),
		},
		Gerbang: GerbangConfig{
			BaseURL:        getEnv("GERBANG_BASE_URL", "https://api.gtd.co.id"),
			ClientID:       getEnv("GERBANG_CLIENT_ID", ""),
			ClientSecret:   getEnv("GERBANG_CLIENT_SECRET", ""),
			CallbackSecret: getEnv("GERBANG_CALLBACK_SECRET", ""),
			Timeout:        time.Duration(getEnvAsInt("GERBANG_TIMEOUT", 30)) * time.Second,
		},
		Fallback: FallbackConfig{
			KYCEnabled:     getEnv("DUMMY_KYC_FALLBACK_ENABLED", "true") == "true",
			PaymentEnabled: getEnv("DUMMY_PAYMENT_FALLBACK_ENABLED", "true") == "true",
			PPOBEnabled:    getEnv("DUMMY_PPOB_FALLBACK_ENABLED", "true") == "true",
		},
		ProductSync: ProductSyncConfig{
			Interval:      time.Duration(getEnvAsInt("PRODUCT_SYNC_INTERVAL", 15)) * time.Minute,
			EnableOnStart: getEnv("PRODUCT_SYNC_ON_START", "true") == "true",
			Enabled:       getEnv("PRODUCT_SYNC_ENABLED", "true") == "true",
		},
		BankCodeSync: BankCodeSyncConfig{
			Interval:      time.Duration(getEnvAsInt("BANK_CODE_SYNC_INTERVAL", 4320)) * time.Minute, // 72 hours = 3 days
			EnableOnStart: getEnv("BANK_CODE_SYNC_ON_START", "true") == "true",
			Enabled:       getEnv("BANK_CODE_SYNC_ENABLED", "true") == "true",
		},
		S3: S3Config{
			Bucket:    getEnv("S3_BUCKET", "ppob-id-kyc"),
			Region:    getEnv("S3_REGION", "ap-southeast-3"), // Jakarta for data residency
			AccessKey: getEnvRequired("S3_ACCESS_KEY"),
			SecretKey: getEnvRequired("S3_SECRET_KEY"),
			BaseURL:   getEnv("S3_BASE_URL", ""),
		},
		TerritorySync: TerritorySyncConfig{
			Interval:      time.Duration(getEnvAsInt("TERRITORY_SYNC_INTERVAL_DAYS", 30)) * 24 * time.Hour,
			EnableOnStart: getEnv("TERRITORY_SYNC_ON_START", "false") == "true",
			Enabled:       getEnv("TERRITORY_SYNC_ENABLED", "false") == "true",
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvRequired returns env var value or fatally exits if not set
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("FATAL: Environment variable %s is required but not set", key)
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
