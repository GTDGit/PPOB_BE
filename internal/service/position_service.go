package service

import (
	"context"
	"strings"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

type PositionService struct {
	repo      *repository.PositionRepository
	adminRepo *repository.AdminRepository
}

func NewPositionService(repo *repository.PositionRepository, adminRepo *repository.AdminRepository) *PositionService {
	return &PositionService{repo: repo, adminRepo: adminRepo}
}

func (s *PositionService) ListPositions(ctx context.Context) ([]map[string]interface{}, error) {
	return s.repo.ListPositionsWithCount(ctx)
}

func (s *PositionService) CreatePosition(ctx context.Context, name string) (*domain.AdminPosition, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return nil, domain.ErrValidationFailed("Nama posisi minimal 2 karakter")
	}
	if len(name) > 150 {
		return nil, domain.ErrValidationFailed("Nama posisi maksimal 150 karakter")
	}
	pos, err := s.repo.CreatePosition(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return nil, domain.ErrValidationFailed("Nama posisi sudah digunakan")
		}
		return nil, err
	}
	return pos, nil
}

func (s *PositionService) UpdatePosition(ctx context.Context, id, name string) (*domain.AdminPosition, error) {
	name = strings.TrimSpace(name)
	if len(name) < 2 {
		return nil, domain.ErrValidationFailed("Nama posisi minimal 2 karakter")
	}
	if len(name) > 150 {
		return nil, domain.ErrValidationFailed("Nama posisi maksimal 150 karakter")
	}
	pos, err := s.repo.UpdatePosition(ctx, id, name)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return nil, domain.ErrValidationFailed("Nama posisi sudah digunakan")
		}
		return nil, err
	}
	if pos == nil {
		return nil, domain.NewError("POSITION_NOT_FOUND", "Posisi tidak ditemukan", 404)
	}
	return pos, nil
}

func (s *PositionService) DeletePosition(ctx context.Context, id string) error {
	err := s.repo.DeletePosition(ctx, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return domain.NewError("POSITION_NOT_FOUND", "Posisi tidak ditemukan", 404)
		}
		return err
	}
	return nil
}

func (s *PositionService) GetPositionAdmins(ctx context.Context, positionID string) ([]map[string]interface{}, error) {
	pos, err := s.repo.GetPositionByID(ctx, positionID)
	if err != nil {
		return nil, err
	}
	if pos == nil {
		return nil, domain.NewError("POSITION_NOT_FOUND", "Posisi tidak ditemukan", 404)
	}
	return s.repo.ListAdminsByPosition(ctx, positionID)
}

func (s *PositionService) AssignPosition(ctx context.Context, positionID, adminID string) error {
	pos, err := s.repo.GetPositionByID(ctx, positionID)
	if err != nil {
		return err
	}
	if pos == nil {
		return domain.NewError("POSITION_NOT_FOUND", "Posisi tidak ditemukan", 404)
	}
	admin, err := s.adminRepo.FindAdminByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin == nil {
		return domain.NewError("ADMIN_NOT_FOUND", "Admin tidak ditemukan", 404)
	}
	return s.repo.AssignPosition(ctx, adminID, positionID)
}

func (s *PositionService) RemoveAdminPosition(ctx context.Context, adminID string) error {
	admin, err := s.adminRepo.FindAdminByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin == nil {
		return domain.NewError("ADMIN_NOT_FOUND", "Admin tidak ditemukan", 404)
	}
	return s.repo.RemoveAdminPosition(ctx, adminID)
}
