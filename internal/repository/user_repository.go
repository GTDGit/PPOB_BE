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

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)
	FindByMIC(ctx context.Context, mic string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	UpdatePIN(ctx context.Context, userID, pinHash string) error
	UpdatePhone(ctx context.Context, userID, phone string) error
	UpdateLastLogin(ctx context.Context, userID string) error
	VerifyEmail(ctx context.Context, userID, email string) error
	GenerateMIC(ctx context.Context) (string, error)
	GenerateReferralCode(ctx context.Context) (string, error)
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

// Column constants for explicit SELECT
const userColumns = `id, mic, phone, full_name, email, gender, tier, avatar_url, 
                     kyc_status, business_type, source, referred_by, referral_code, 
                     used_referral_code, pin_hash, is_active, is_locked, locked_until, 
                     phone_verified_at, created_at, updated_at`

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	query := fmt.Sprintf(`SELECT %s FROM users WHERE id = $1`, userColumns)
	err := r.db.GetContext(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var user domain.User
	query := fmt.Sprintf(`SELECT %s FROM users WHERE phone = $1`, userColumns)
	err := r.db.GetContext(ctx, &user, query, phone)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByMIC(ctx context.Context, mic string) (*domain.User, error) {
	var user domain.User
	query := fmt.Sprintf(`SELECT %s FROM users WHERE mic = $1`, userColumns)
	err := r.db.GetContext(ctx, &user, query, mic)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, mic, phone, full_name, email, gender, tier, avatar_url,
			kyc_status, business_type, source, referred_by, referral_code,
			used_referral_code, pin_hash, is_active, is_locked, phone_verified_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.MIC, user.Phone, user.FullName, user.Email, user.Gender,
		user.Tier, user.AvatarURL, user.KYCStatus, user.BusinessType, user.Source,
		user.ReferredBy, user.ReferralCode, user.UsedReferralCode, user.PINHash,
		user.IsActive, user.IsLocked, user.PhoneVerifiedAt, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users SET
			full_name = $2, email = $3, gender = $4, tier = $5, avatar_url = $6,
			kyc_status = $7, business_type = $8, source = $9, referred_by = $10,
			is_active = $11, is_locked = $12, locked_until = $13, updated_at = $14
		WHERE id = $1
	`
	user.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.FullName, user.Email, user.Gender, user.Tier, user.AvatarURL,
		user.KYCStatus, user.BusinessType, user.Source, user.ReferredBy,
		user.IsActive, user.IsLocked, user.LockedUntil, user.UpdatedAt,
	)
	return err
}

func (r *userRepository) UpdatePIN(ctx context.Context, userID, pinHash string) error {
	query := `UPDATE users SET pin_hash = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, pinHash, time.Now())
	return err
}

func (r *userRepository) UpdatePhone(ctx context.Context, userID, phone string) error {
	query := `UPDATE users SET phone = $2, phone_verified_at = $3, updated_at = $3 WHERE id = $1`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, phone, now)
	return err
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `UPDATE users SET updated_at = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, time.Now())
	return err
}

func (r *userRepository) VerifyEmail(ctx context.Context, userID, email string) error {
	query := `UPDATE users SET email = $2, email_verified_at = $3, updated_at = $3 WHERE id = $1`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, email, now)
	return err
}

func (r *userRepository) GenerateMIC(ctx context.Context) (string, error) {
	// Generate unique MIC (Merchant Identification Code)
	// Format: PID + 5 digits
	for i := 0; i < 10; i++ {
		mic := "PID" + generateRandomDigits(5)
		exists, err := r.micExists(ctx, mic)
		if err != nil {
			return "", err
		}
		if !exists {
			return mic, nil
		}
	}
	// If still collision after 10 tries, use UUID-based approach
	return "PID" + uuid.New().String()[:5], nil
}

func (r *userRepository) GenerateReferralCode(ctx context.Context) (string, error) {
	// Generate unique referral code
	// Format: 8 alphanumeric characters
	for i := 0; i < 10; i++ {
		code := generateRandomAlphanumeric(8)
		exists, err := r.referralCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return generateRandomAlphanumeric(10), nil
}

func (r *userRepository) micExists(ctx context.Context, mic string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE mic = $1`
	err := r.db.GetContext(ctx, &count, query, mic)
	return count > 0, err
}

func (r *userRepository) referralCodeExists(ctx context.Context, code string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE referral_code = $1`
	err := r.db.GetContext(ctx, &count, query, code)
	return count > 0, err
}

// Helper functions for random generation
func generateRandomDigits(n int) string {
	const digits = "0123456789"
	return generateRandom(n, digits)
}

func generateRandomAlphanumeric(n int) string {
	const alphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return generateRandom(n, alphanumeric)
}

func generateRandom(n int, charset string) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond)
	}
	return string(result)
}
