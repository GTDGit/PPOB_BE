package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// VoucherRepository defines the interface for voucher data operations
type VoucherRepository interface {
	// Voucher operations
	FindByID(ctx context.Context, id string) (*domain.Voucher, error)
	FindByCode(ctx context.Context, code string) (*domain.Voucher, error)

	// User voucher operations
	FindUserVouchers(ctx context.Context, userID string, status string) ([]*domain.Voucher, error)
	FindApplicableVouchers(ctx context.Context, userID, serviceType string, amount int64) ([]*domain.Voucher, error)
	CheckUserHasVoucher(ctx context.Context, userID, voucherID string) (bool, error)
	UseVoucher(ctx context.Context, userID, voucherID string) error
}

// voucherRepository implements VoucherRepository
type voucherRepository struct {
	db *sqlx.DB
}

// NewVoucherRepository creates a new voucher repository
func NewVoucherRepository(db *sqlx.DB) VoucherRepository {
	return &voucherRepository{db: db}
}

// Column constants for explicit SELECT
const voucherColumns = `id, code, name, description, discount_type, discount_value, 
                        min_transaction, max_discount, applicable_services, 
                        max_usage, max_usage_per_user, current_usage, starts_at, 
                        expires_at, is_active, terms_url, created_at, updated_at`

// FindByID finds a voucher by ID
func (r *voucherRepository) FindByID(ctx context.Context, id string) (*domain.Voucher, error) {
	var voucher domain.Voucher
	query := fmt.Sprintf(`SELECT %s FROM vouchers WHERE id = $1`, voucherColumns)
	err := r.db.GetContext(ctx, &voucher, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &voucher, err
}

// FindByCode finds a voucher by code
func (r *voucherRepository) FindByCode(ctx context.Context, code string) (*domain.Voucher, error) {
	var voucher domain.Voucher
	query := fmt.Sprintf(`SELECT %s FROM vouchers WHERE code = $1 AND is_active = true`, voucherColumns)
	err := r.db.GetContext(ctx, &voucher, query, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &voucher, err
}

// FindUserVouchers finds user's vouchers with optional status filter
func (r *voucherRepository) FindUserVouchers(ctx context.Context, userID string, status string) ([]*domain.Voucher, error) {
	// For now, return mock vouchers
	// In production, this would join user_vouchers table
	vouchers := r.getMockVouchers()

	// Filter by status
	if status != "" && status != "all" {
		filtered := []*domain.Voucher{}
		for _, v := range vouchers {
			vStatus := r.getVoucherStatus(v)
			if vStatus == status {
				filtered = append(filtered, v)
			}
		}
		return filtered, nil
	}

	return vouchers, nil
}

// FindApplicableVouchers finds vouchers applicable for a transaction
func (r *voucherRepository) FindApplicableVouchers(ctx context.Context, userID, serviceType string, amount int64) ([]*domain.Voucher, error) {
	// Get all user's active vouchers
	vouchers, err := r.FindUserVouchers(ctx, userID, "active")
	if err != nil {
		return nil, err
	}

	result := []*domain.Voucher{}
	for _, v := range vouchers {
		// Parse applicable services
		services := []string{}
		if v.ApplicableServices != "" {
			json.Unmarshal([]byte(v.ApplicableServices), &services)
		}

		// Check if applicable
		if r.isApplicable(v, services, serviceType, amount) {
			result = append(result, v)
		}
	}

	return result, nil
}

// CheckUserHasVoucher checks if user has a specific voucher
func (r *voucherRepository) CheckUserHasVoucher(ctx context.Context, userID, voucherID string) (bool, error) {
	// For now, assume user has all mock vouchers
	// In production, query user_vouchers table
	voucher, err := r.FindByID(ctx, voucherID)
	if err != nil {
		return false, err
	}
	return voucher != nil, nil
}

// UseVoucher marks a voucher as used
func (r *voucherRepository) UseVoucher(ctx context.Context, userID, voucherID string) error {
	// In production, update user_vouchers table
	// For now, just return success
	return nil
}

// Helper functions

func (r *voucherRepository) getVoucherStatus(v *domain.Voucher) string {
	now := time.Now()
	if now.After(v.ExpiresAt) {
		return domain.VoucherStatusExpired
	}
	if v.CurrentUsage >= v.MaxUsage {
		return domain.VoucherStatusUsed
	}
	if !v.IsActive {
		return "inactive"
	}
	return domain.VoucherStatusActive
}

func (r *voucherRepository) isApplicable(v *domain.Voucher, services []string, serviceType string, amount int64) bool {
	// Check service type
	applicable := false
	for _, s := range services {
		if s == "all" || s == serviceType {
			applicable = true
			break
		}
	}
	if !applicable {
		return false
	}

	// Check minimum transaction
	if amount < v.MinTransaction {
		return false
	}

	return true
}

// getMockVouchers returns mock vouchers for testing
func (r *voucherRepository) getMockVouchers() []*domain.Voucher {
	now := time.Now()

	// Mock applicable services
	pulsaDataPLN, _ := json.Marshal([]string{"pulsa", "data", "pln_prepaid"})
	allServices, _ := json.Marshal([]string{"all"})
	pdamOnly, _ := json.Marshal([]string{"pdam"})

	return []*domain.Voucher{
		{
			ID:                 "vch_abc123",
			Code:               "DISKON10",
			Name:               "Discount Rp12.000",
			Description:        "Khusus untuk pengguna pertama",
			DiscountType:       domain.DiscountFixed,
			DiscountValue:      12000,
			MinTransaction:     10000,
			MaxDiscount:        12000,
			ApplicableServices: string(pulsaDataPLN),
			MaxUsage:           1,
			MaxUsagePerUser:    1,
			CurrentUsage:       0,
			StartsAt:           nil,
			ExpiresAt:          now.AddDate(0, 1, 0), // 1 month from now
			IsActive:           true,
			TermsURL:           "https://ppob.id/voucher/diskon10/terms",
			CreatedAt:          now,
			UpdatedAt:          now,
		},
		{
			ID:                 "vch_xyz789",
			Code:               "CASHBACK5",
			Name:               "Cashback 5%",
			Description:        "Cashback 5% max Rp10.000",
			DiscountType:       domain.DiscountPercentage,
			DiscountValue:      5,
			MinTransaction:     50000,
			MaxDiscount:        10000,
			ApplicableServices: string(allServices),
			MaxUsage:           3,
			MaxUsagePerUser:    1,
			CurrentUsage:       0,
			StartsAt:           nil,
			ExpiresAt:          now.AddDate(0, 2, 0), // 2 months from now
			IsActive:           true,
			TermsURL:           "https://ppob.id/voucher/cashback5/terms",
			CreatedAt:          now,
			UpdatedAt:          now,
		},
		{
			ID:                 "vch_pdam50",
			Code:               "PDAM50",
			Name:               "Discount PDAM",
			Description:        "Khusus pembayaran PDAM",
			DiscountType:       domain.DiscountFixed,
			DiscountValue:      5000,
			MinTransaction:     50000,
			MaxDiscount:        5000,
			ApplicableServices: string(pdamOnly),
			MaxUsage:           5,
			MaxUsagePerUser:    2,
			CurrentUsage:       0,
			StartsAt:           nil,
			ExpiresAt:          now.AddDate(0, 1, 0),
			IsActive:           true,
			TermsURL:           "https://ppob.id/voucher/pdam50/terms",
			CreatedAt:          now,
			UpdatedAt:          now,
		},
		{
			ID:                 "vch_newuser",
			Code:               "NEWUSER2025",
			Name:               "Discount New User",
			Description:        "Promo untuk pengguna baru",
			DiscountType:       domain.DiscountPercentage,
			DiscountValue:      10,
			MinTransaction:     20000,
			MaxDiscount:        10000,
			ApplicableServices: string(allServices),
			MaxUsage:           1,
			MaxUsagePerUser:    1,
			CurrentUsage:       0,
			StartsAt:           nil,
			ExpiresAt:          now.AddDate(0, 3, 0), // 3 months from now
			IsActive:           true,
			TermsURL:           "https://ppob.id/voucher/newuser/terms",
			CreatedAt:          now,
			UpdatedAt:          now,
		},
	}
}
