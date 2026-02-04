package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

type UserSettingsRepository interface {
	FindByUserID(ctx context.Context, userID string) (*domain.UserSettings, error)
	Create(ctx context.Context, settings *domain.UserSettings) error
	Update(ctx context.Context, settings *domain.UserSettings) error
}

type userSettingsRepository struct {
	db *sqlx.DB
}

func NewUserSettingsRepository(db *sqlx.DB) UserSettingsRepository {
	return &userSettingsRepository{db: db}
}

// Column constants for explicit SELECT
const settingsColumns = `id, user_id, notification_enabled, transaction_alert, 
                         promo_notification, email_notification, auto_save_contact, 
                         show_profit_on_receipt, language, currency, theme, 
                         created_at, updated_at`

func (r *userSettingsRepository) FindByUserID(ctx context.Context, userID string) (*domain.UserSettings, error) {
	var settings domain.UserSettings
	query := fmt.Sprintf(`SELECT %s FROM user_settings WHERE user_id = $1`, settingsColumns)
	err := r.db.GetContext(ctx, &settings, query, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *userSettingsRepository) Create(ctx context.Context, settings *domain.UserSettings) error {
	if settings.ID == "" {
		settings.ID = "ust_" + uuid.New().String()[:8]
	}
	now := time.Now()
	settings.CreatedAt = now
	settings.UpdatedAt = now

	// Set defaults
	if settings.Language == "" {
		settings.Language = domain.LanguageID
	}
	if settings.Currency == "" {
		settings.Currency = "IDR"
	}
	if settings.Theme == "" {
		settings.Theme = domain.ThemeLight
	}

	query := `
		INSERT INTO user_settings (
			id, user_id, pin_required_for_transaction, pin_required_min_amount,
			biometric_enabled, default_selling_price_markup, auto_save_contact,
			show_profit_on_receipt, language, currency, theme,
			show_phone_on_qris, show_name_on_qris, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		settings.ID, settings.UserID, settings.PINRequiredForTransaction,
		settings.PINRequiredMinAmount, settings.BiometricEnabled,
		settings.DefaultSellingPriceMarkup, settings.AutoSaveContact,
		settings.ShowProfitOnReceipt, settings.Language, settings.Currency,
		settings.Theme, settings.ShowPhoneOnQRIS, settings.ShowNameOnQRIS,
		settings.CreatedAt, settings.UpdatedAt,
	)
	return err
}

func (r *userSettingsRepository) Update(ctx context.Context, settings *domain.UserSettings) error {
	settings.UpdatedAt = time.Now()
	query := `
		UPDATE user_settings SET
			pin_required_for_transaction = $2, pin_required_min_amount = $3,
			biometric_enabled = $4, default_selling_price_markup = $5,
			auto_save_contact = $6, show_profit_on_receipt = $7,
			language = $8, currency = $9, theme = $10,
			show_phone_on_qris = $11, show_name_on_qris = $12, updated_at = $13
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		settings.ID, settings.PINRequiredForTransaction, settings.PINRequiredMinAmount,
		settings.BiometricEnabled, settings.DefaultSellingPriceMarkup,
		settings.AutoSaveContact, settings.ShowProfitOnReceipt,
		settings.Language, settings.Currency, settings.Theme,
		settings.ShowPhoneOnQRIS, settings.ShowNameOnQRIS, settings.UpdatedAt,
	)
	return err
}

// CreateDefaultSettings creates default settings for a new user
func CreateDefaultSettings(userID string) *domain.UserSettings {
	return &domain.UserSettings{
		UserID:                    userID,
		PINRequiredForTransaction: true,
		PINRequiredMinAmount:      0,
		BiometricEnabled:          false,
		DefaultSellingPriceMarkup: 0,
		AutoSaveContact:           true,
		ShowProfitOnReceipt:       true,
		Language:                  domain.LanguageID,
		Currency:                  "IDR",
		Theme:                     domain.ThemeLight,
		ShowPhoneOnQRIS:           false,
		ShowNameOnQRIS:            true,
	}
}
