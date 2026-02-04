package service

import (
	"context"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// TerritoryService handles territory business logic
type TerritoryService struct {
	repo repository.TerritoryRepository
}

// NewTerritoryService creates a new territory service
func NewTerritoryService(repo repository.TerritoryRepository) *TerritoryService {
	return &TerritoryService{repo: repo}
}

// GetProvinces returns all provinces
func (s *TerritoryService) GetProvinces(ctx context.Context) ([]*domain.Province, error) {
	return s.repo.GetAllProvinces(ctx)
}

// GetCitiesByProvince returns cities/regencies for a province
func (s *TerritoryService) GetCitiesByProvince(ctx context.Context, provinceCode string) ([]*domain.City, error) {
	// Validate province exists
	province, err := s.repo.GetProvinceByCode(ctx, provinceCode)
	if err != nil {
		return nil, err
	}
	if province == nil {
		return nil, domain.ErrNotFound("Provinsi")
	}

	return s.repo.GetCitiesByProvince(ctx, provinceCode)
}

// GetDistrictsByCity returns districts for a city
func (s *TerritoryService) GetDistrictsByCity(ctx context.Context, cityCode string) ([]*domain.District, error) {
	// Validate city exists
	city, err := s.repo.GetCityByCode(ctx, cityCode)
	if err != nil {
		return nil, err
	}
	if city == nil {
		return nil, domain.ErrNotFound("Kota/Kabupaten")
	}

	return s.repo.GetDistrictsByCity(ctx, cityCode)
}

// GetSubDistrictsByDistrict returns sub-districts for a district
func (s *TerritoryService) GetSubDistrictsByDistrict(ctx context.Context, districtCode string) ([]*domain.SubDistrict, error) {
	// Validate district exists
	district, err := s.repo.GetDistrictByCode(ctx, districtCode)
	if err != nil {
		return nil, err
	}
	if district == nil {
		return nil, domain.ErrNotFound("Kecamatan")
	}

	return s.repo.GetSubDistrictsByDistrict(ctx, districtCode)
}

// GetPostalCodesBySubDistrict returns postal codes for a sub-district
func (s *TerritoryService) GetPostalCodesBySubDistrict(ctx context.Context, subDistrictCode string) ([]*domain.PostalCode, error) {
	// Validate sub-district exists
	subDistrict, err := s.repo.GetSubDistrictByCode(ctx, subDistrictCode)
	if err != nil {
		return nil, err
	}
	if subDistrict == nil {
		return nil, domain.ErrNotFound("Kelurahan")
	}

	return s.repo.GetPostalCodesBySubDistrict(ctx, subDistrictCode)
}

// SearchByPostalCode searches location by postal code
func (s *TerritoryService) SearchByPostalCode(ctx context.Context, postalCode string) ([]*domain.PostalCodeSearchResult, error) {
	if postalCode == "" {
		return nil, domain.ErrValidationFailed("Kode pos wajib diisi")
	}

	if len(postalCode) != 5 {
		return nil, domain.ErrValidationFailed("Kode pos harus 5 digit")
	}

	results, err := s.repo.SearchByPostalCode(ctx, postalCode)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, domain.ErrNotFound("Kode pos")
	}

	return results, nil
}
