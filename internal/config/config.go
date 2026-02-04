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
	OTP           OTPConfig
	WhatsApp      WhatsAppConfig
	Fazpass       FazpassConfig
	Brevo         BrevoConfig
	Gerbang       GerbangConfig
	S3            S3Config
	ProductSync   ProductSyncConfig
	BankCodeSync  BankCodeSyncConfig
	TerritorySync TerritorySyncConfig
}

type AppConfig struct {
	Name string
	Env  string
	Port int
	URL  string
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

type OTPConfig struct {
	Length         int
	TTL            time.Duration
	MaxAttempts    int
	MaxResend      int
	ResendCooldown time.Duration
}

type WhatsAppConfig struct {
	APIURL              string
	PhoneNumberID       string
	AccessToken         string
	OTPTemplateName     string
	OTPTemplateLanguage string
}

type FazpassConfig struct {
	APIURL      string
	MerchantKey string
	GatewayKey  string
}

type BrevoConfig struct {
	APIURL               string
	APIKey               string
	SenderEmail          string
	SenderName           string
	BaseURL              string // Application base URL for links
	TemplateVerification int64  // Email verification template ID
	TemplateNewLogin     int64  // New login alert template ID
	TemplatePINChanged   int64  // PIN changed alert template ID
	TemplateEmailChanged int64  // Email changed alert template ID
	TemplatePhoneChanged int64  // Phone changed alert template ID
}

type GerbangConfig struct {
	BaseURL        string
	ClientID       string
	ClientSecret   string
	CallbackSecret string
	Timeout        time.Duration
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
			Name: getEnv("APP_NAME", "ppob.id"),
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnvAsInt("APP_PORT", 8080),
			URL:  getEnv("APP_URL", "http://localhost:8080"),
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
		OTP: OTPConfig{
			Length:         getEnvAsInt("OTP_LENGTH", 4),                             // 4 digits (SMS standard)
			TTL:            time.Duration(getEnvAsInt("OTP_TTL", 300)) * time.Second, // 5 minutes (security best practice)
			MaxAttempts:    getEnvAsInt("OTP_MAX_ATTEMPTS", 5),
			MaxResend:      getEnvAsInt("OTP_MAX_RESEND", 3),
			ResendCooldown: time.Duration(getEnvAsInt("OTP_RESEND_COOLDOWN", 60)) * time.Second,
		},
		WhatsApp: WhatsAppConfig{
			APIURL:              getEnv("WA_API_URL", "https://graph.facebook.com/v18.0"),
			PhoneNumberID:       getEnv("WA_PHONE_NUMBER_ID", ""),
			AccessToken:         getEnv("WA_ACCESS_TOKEN", ""),
			OTPTemplateName:     getEnv("WA_OTP_TEMPLATE_NAME", "otp_verification"),
			OTPTemplateLanguage: getEnv("WA_OTP_TEMPLATE_LANGUAGE", "id"),
		},
		Fazpass: FazpassConfig{
			APIURL:      getEnv("FAZPASS_API_URL", "https://api.fazpass.com"),
			MerchantKey: getEnv("FAZPASS_MERCHANT_KEY", ""),
			GatewayKey:  getEnv("FAZPASS_GATEWAY_KEY", ""),
		},
		Brevo: BrevoConfig{
			APIURL:               getEnv("BREVO_API_URL", "https://api.brevo.com/v3"),
			APIKey:               getEnv("BREVO_API_KEY", ""),
			SenderEmail:          getEnv("BREVO_SENDER_EMAIL", "noreply@ppob.id"),
			SenderName:           getEnv("BREVO_SENDER_NAME", "PPOB.ID"),
			BaseURL:              getEnv("APP_URL", "http://localhost:8080"),
			TemplateVerification: int64(getEnvAsInt("BREVO_TEMPLATE_VERIFICATION", 1)),
			TemplateNewLogin:     int64(getEnvAsInt("BREVO_TEMPLATE_NEW_LOGIN", 2)),
			TemplatePINChanged:   int64(getEnvAsInt("BREVO_TEMPLATE_PIN_CHANGED", 3)),
			TemplateEmailChanged: int64(getEnvAsInt("BREVO_TEMPLATE_EMAIL_CHANGED", 4)),
			TemplatePhoneChanged: int64(getEnvAsInt("BREVO_TEMPLATE_PHONE_CHANGED", 5)),
		},
		Gerbang: GerbangConfig{
			BaseURL:        getEnv("GERBANG_BASE_URL", "https://api.gtd.co.id"),
			ClientID:       getEnv("GERBANG_CLIENT_ID", ""),
			ClientSecret:   getEnv("GERBANG_CLIENT_SECRET", ""),
			CallbackSecret: getEnv("GERBANG_CALLBACK_SECRET", ""),
			Timeout:        time.Duration(getEnvAsInt("GERBANG_TIMEOUT", 30)) * time.Second,
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
