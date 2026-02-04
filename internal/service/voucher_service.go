package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// VoucherService handles voucher business logic
type VoucherService struct {
	voucherRepo repository.VoucherRepository
}

// NewVoucherService creates a new voucher service
func NewVoucherService(voucherRepo repository.VoucherRepository) *VoucherService {
	return &VoucherService{
		voucherRepo: voucherRepo,
	}
}

// List returns user's vouchers with optional status filter
func (s *VoucherService) List(ctx context.Context, userID, status string) (*domain.VoucherListResponse, error) {
	vouchers, err := s.voucherRepo.FindUserVouchers(ctx, userID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get vouchers: %w", err)
	}

	// Convert to response format
	details := make([]*domain.VoucherDetail, 0, len(vouchers))
	for _, v := range vouchers {
		detail := s.toVoucherDetail(v)
		details = append(details, detail)
	}

	return &domain.VoucherListResponse{
		Vouchers:      details,
		TotalVouchers: len(details),
	}, nil
}

// GetApplicable returns vouchers applicable for a transaction
func (s *VoucherService) GetApplicable(ctx context.Context, userID, serviceType string, amount int64) (*domain.ApplicableVouchersResponse, error) {
	// Get all user's vouchers
	allVouchers, err := s.voucherRepo.FindUserVouchers(ctx, userID, "active")
	if err != nil {
		return nil, fmt.Errorf("failed to get vouchers: %w", err)
	}

	// Check each voucher's applicability
	applicableVouchers := make([]*domain.ApplicableVoucher, 0, len(allVouchers))
	for _, v := range allVouchers {
		applicable, reason := s.checkApplicability(v, serviceType, amount)

		av := &domain.ApplicableVoucher{
			ID:            v.ID,
			Code:          v.Code,
			Name:          v.Name,
			Description:   v.Description,
			DiscountType:  v.DiscountType,
			DiscountValue: v.DiscountValue,
			Applicable:    applicable,
			ExpiresAt:     v.ExpiresAt.Format(time.RFC3339),
		}

		if applicable {
			// Calculate estimated discount
			discount := s.CalculateDiscount(v, amount)
			discountFormatted := formatCurrency(v.DiscountValue)
			estimatedFormatted := "-" + formatCurrency(discount)

			av.DiscountFormatted = &discountFormatted
			av.EstimatedDiscount = &discount
			av.EstimatedDiscountFormatted = &estimatedFormatted
		} else {
			av.NotApplicableReason = &reason
		}

		applicableVouchers = append(applicableVouchers, av)
	}

	return &domain.ApplicableVouchersResponse{
		Vouchers:                  applicableVouchers,
		MaxVouchersPerTransaction: 2, // TODO: Make configurable
	}, nil
}

// Validate validates a voucher code for a transaction
func (s *VoucherService) Validate(ctx context.Context, userID, code, serviceType string, amount int64) (*domain.ValidateVoucherResponse, error) {
	// Find voucher by code
	voucher, err := s.voucherRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to find voucher: %w", err)
	}

	if voucher == nil {
		reason := "INVALID_VOUCHER"
		message := "Kode voucher tidak valid"
		return &domain.ValidateVoucherResponse{
			Valid:   false,
			Reason:  &reason,
			Message: &message,
		}, nil
	}

	// Check if user has this voucher
	hasVoucher, err := s.voucherRepo.CheckUserHasVoucher(ctx, userID, voucher.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user voucher: %w", err)
	}
	if !hasVoucher {
		reason := "INVALID_VOUCHER"
		message := "Voucher tidak tersedia untuk Anda"
		return &domain.ValidateVoucherResponse{
			Valid:   false,
			Reason:  &reason,
			Message: &message,
		}, nil
	}

	// Check expiration
	if time.Now().After(voucher.ExpiresAt) {
		reason := "VOUCHER_EXPIRED"
		message := "Voucher sudah tidak berlaku"
		return &domain.ValidateVoucherResponse{
			Valid:   false,
			Reason:  &reason,
			Message: &message,
		}, nil
	}

	// Check usage limit
	if voucher.CurrentUsage >= voucher.MaxUsage {
		reason := "VOUCHER_USED"
		message := "Voucher sudah digunakan"
		return &domain.ValidateVoucherResponse{
			Valid:   false,
			Reason:  &reason,
			Message: &message,
		}, nil
	}

	// Check if active
	if !voucher.IsActive {
		reason := "VOUCHER_INACTIVE"
		message := "Voucher tidak aktif"
		return &domain.ValidateVoucherResponse{
			Valid:   false,
			Reason:  &reason,
			Message: &message,
		}, nil
	}

	// Check applicability
	applicable, reason := s.checkApplicability(voucher, serviceType, amount)
	if !applicable {
		message := reason
		reasonCode := "VOUCHER_NOT_APPLICABLE"
		if reason == "Minimum transaksi tidak terpenuhi" {
			reasonCode = "MIN_TRANSACTION_NOT_MET"
		}
		return &domain.ValidateVoucherResponse{
			Valid:   false,
			Reason:  &reasonCode,
			Message: &message,
		}, nil
	}

	// Valid - calculate discount
	discount := s.CalculateDiscount(voucher, amount)

	return &domain.ValidateVoucherResponse{
		Valid: true,
		Voucher: &domain.ValidVoucher{
			ID:                         voucher.ID,
			Code:                       voucher.Code,
			Name:                       voucher.Name,
			DiscountType:               voucher.DiscountType,
			DiscountValue:              voucher.DiscountValue,
			EstimatedDiscount:          discount,
			EstimatedDiscountFormatted: "-" + formatCurrency(discount),
			MaxDiscount:                voucher.MaxDiscount,
			MaxDiscountFormatted:       formatCurrency(voucher.MaxDiscount),
		},
	}, nil
}

// CalculateDiscount calculates the discount amount for a voucher
func (s *VoucherService) CalculateDiscount(voucher *domain.Voucher, amount int64) int64 {
	var discount int64

	if voucher.DiscountType == domain.DiscountFixed {
		discount = voucher.DiscountValue
	} else if voucher.DiscountType == domain.DiscountPercentage {
		discount = (amount * voucher.DiscountValue) / 100
		// Apply max discount limit
		if discount > voucher.MaxDiscount {
			discount = voucher.MaxDiscount
		}
	}

	// Discount cannot exceed transaction amount
	if discount > amount {
		discount = amount
	}

	return discount
}

// Helper functions

func (s *VoucherService) toVoucherDetail(v *domain.Voucher) *domain.VoucherDetail {
	// Parse applicable services
	services := []string{}
	if v.ApplicableServices != "" {
		if err := json.Unmarshal([]byte(v.ApplicableServices), &services); err != nil {
			slog.Error("failed to unmarshal applicable services",
				slog.String("voucher_id", v.ID),
				slog.String("error", err.Error()),
			)
			// Continue with empty services
		}
	}

	// Determine status
	status := domain.VoucherStatusActive
	now := time.Now()
	if now.After(v.ExpiresAt) {
		status = domain.VoucherStatusExpired
	} else if v.CurrentUsage >= v.MaxUsage {
		status = domain.VoucherStatusUsed
	} else if !v.IsActive {
		status = "inactive"
	}

	// Format discount
	discountFormatted := formatCurrency(v.DiscountValue)
	if v.DiscountType == domain.DiscountPercentage {
		discountFormatted = fmt.Sprintf("%d%%", v.DiscountValue)
	}

	return &domain.VoucherDetail{
		ID:                      v.ID,
		Code:                    v.Code,
		Name:                    v.Name,
		Description:             v.Description,
		DiscountType:            v.DiscountType,
		DiscountValue:           v.DiscountValue,
		DiscountFormatted:       discountFormatted,
		MinTransaction:          v.MinTransaction,
		MinTransactionFormatted: formatCurrency(v.MinTransaction),
		MaxDiscount:             v.MaxDiscount,
		MaxDiscountFormatted:    formatCurrency(v.MaxDiscount),
		ApplicableServices:      services,
		ExpiresAt:               v.ExpiresAt.Format(time.RFC3339),
		Status:                  status,
		TermsURL:                v.TermsURL,
	}
}

func (s *VoucherService) checkApplicability(voucher *domain.Voucher, serviceType string, amount int64) (bool, string) {
	// Check expiration
	if time.Now().After(voucher.ExpiresAt) {
		return false, "Voucher sudah expired"
	}

	// Check usage limit
	if voucher.CurrentUsage >= voucher.MaxUsage {
		return false, "Voucher sudah digunakan"
	}

	// Check if active
	if !voucher.IsActive {
		return false, "Voucher tidak aktif"
	}

	// Parse applicable services
	services := []string{}
	if voucher.ApplicableServices != "" {
		if err := json.Unmarshal([]byte(voucher.ApplicableServices), &services); err != nil {
			slog.Error("failed to unmarshal applicable services for validation",
				slog.String("voucher_id", voucher.ID),
				slog.String("error", err.Error()),
			)
			// Continue with empty services (will fail service check)
		}
	}

	// Check service type
	serviceApplicable := false
	for _, s := range services {
		if s == "all" || s == serviceType {
			serviceApplicable = true
			break
		}
	}
	if !serviceApplicable {
		return false, fmt.Sprintf("Voucher hanya berlaku untuk %v", services)
	}

	// Check minimum transaction
	if amount < voucher.MinTransaction {
		return false, "Minimum transaksi tidak terpenuhi"
	}

	return true, ""
}
