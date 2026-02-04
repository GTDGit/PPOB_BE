package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// TerritoryRepository defines the interface for territory data operations
type TerritoryRepository interface {
	// Province
	GetAllProvinces(ctx context.Context) ([]*domain.Province, error)
	GetProvinceByCode(ctx context.Context, code string) (*domain.Province, error)
	UpsertProvinces(ctx context.Context, provinces []*domain.Province) error

	// City
	GetCitiesByProvince(ctx context.Context, provinceCode string) ([]*domain.City, error)
	GetCityByCode(ctx context.Context, code string) (*domain.City, error)
	UpsertCities(ctx context.Context, cities []*domain.City) error

	// District
	GetDistrictsByCity(ctx context.Context, cityCode string) ([]*domain.District, error)
	GetDistrictByCode(ctx context.Context, code string) (*domain.District, error)
	UpsertDistricts(ctx context.Context, districts []*domain.District) error

	// SubDistrict
	GetSubDistrictsByDistrict(ctx context.Context, districtCode string) ([]*domain.SubDistrict, error)
	GetSubDistrictByCode(ctx context.Context, code string) (*domain.SubDistrict, error)
	UpsertSubDistricts(ctx context.Context, subDistricts []*domain.SubDistrict) error

	// PostalCode
	GetPostalCodesBySubDistrict(ctx context.Context, subDistrictCode string) ([]*domain.PostalCode, error)
	SearchByPostalCode(ctx context.Context, postalCode string) ([]*domain.PostalCodeSearchResult, error)
	UpsertPostalCodes(ctx context.Context, postalCodes []*domain.PostalCode) error

	// Sync
	LogSync(ctx context.Context, log *domain.TerritorySyncLog) error
}

// territoryRepository implements TerritoryRepository
type territoryRepository struct {
	db *sqlx.DB
}

// NewTerritoryRepository creates a new territory repository
func NewTerritoryRepository(db *sqlx.DB) TerritoryRepository {
	return &territoryRepository{db: db}
}

// ========== Province Methods ==========

func (r *territoryRepository) GetAllProvinces(ctx context.Context) ([]*domain.Province, error) {
	query := `SELECT code, name, created_at, updated_at FROM provinces ORDER BY name`
	var provinces []*domain.Province
	err := r.db.SelectContext(ctx, &provinces, query)
	return provinces, err
}

func (r *territoryRepository) GetProvinceByCode(ctx context.Context, code string) (*domain.Province, error) {
	query := `SELECT code, name, created_at, updated_at FROM provinces WHERE code = $1`
	var province domain.Province
	err := r.db.GetContext(ctx, &province, query, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &province, err
}

func (r *territoryRepository) UpsertProvinces(ctx context.Context, provinces []*domain.Province) error {
	if len(provinces) == 0 {
		return nil
	}

	query := `
		INSERT INTO provinces (code, name, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = CURRENT_TIMESTAMP
	`

	for _, p := range provinces {
		_, err := r.db.ExecContext(ctx, query, p.Code, p.Name)
		if err != nil {
			return fmt.Errorf("failed to upsert province %s: %w", p.Code, err)
		}
	}

	return nil
}

// ========== City Methods ==========

func (r *territoryRepository) GetCitiesByProvince(ctx context.Context, provinceCode string) ([]*domain.City, error) {
	query := `SELECT code, province_code, name, created_at, updated_at FROM cities WHERE province_code = $1 ORDER BY name`
	var cities []*domain.City
	err := r.db.SelectContext(ctx, &cities, query, provinceCode)
	return cities, err
}

func (r *territoryRepository) GetCityByCode(ctx context.Context, code string) (*domain.City, error) {
	query := `SELECT code, province_code, name, created_at, updated_at FROM cities WHERE code = $1`
	var city domain.City
	err := r.db.GetContext(ctx, &city, query, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &city, err
}

func (r *territoryRepository) UpsertCities(ctx context.Context, cities []*domain.City) error {
	if len(cities) == 0 {
		return nil
	}

	query := `
		INSERT INTO cities (code, province_code, name, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (code) DO UPDATE SET
			province_code = EXCLUDED.province_code,
			name = EXCLUDED.name,
			updated_at = CURRENT_TIMESTAMP
	`

	for _, c := range cities {
		_, err := r.db.ExecContext(ctx, query, c.Code, c.ProvinceCode, c.Name)
		if err != nil {
			return fmt.Errorf("failed to upsert city %s: %w", c.Code, err)
		}
	}

	return nil
}

// ========== District Methods ==========

func (r *territoryRepository) GetDistrictsByCity(ctx context.Context, cityCode string) ([]*domain.District, error) {
	query := `SELECT code, city_code, name, created_at, updated_at FROM districts WHERE city_code = $1 ORDER BY name`
	var districts []*domain.District
	err := r.db.SelectContext(ctx, &districts, query, cityCode)
	return districts, err
}

func (r *territoryRepository) GetDistrictByCode(ctx context.Context, code string) (*domain.District, error) {
	query := `SELECT code, city_code, name, created_at, updated_at FROM districts WHERE code = $1`
	var district domain.District
	err := r.db.GetContext(ctx, &district, query, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &district, err
}

func (r *territoryRepository) UpsertDistricts(ctx context.Context, districts []*domain.District) error {
	if len(districts) == 0 {
		return nil
	}

	query := `
		INSERT INTO districts (code, city_code, name, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (code) DO UPDATE SET
			city_code = EXCLUDED.city_code,
			name = EXCLUDED.name,
			updated_at = CURRENT_TIMESTAMP
	`

	for _, d := range districts {
		_, err := r.db.ExecContext(ctx, query, d.Code, d.CityCode, d.Name)
		if err != nil {
			return fmt.Errorf("failed to upsert district %s: %w", d.Code, err)
		}
	}

	return nil
}

// ========== SubDistrict Methods ==========

func (r *territoryRepository) GetSubDistrictsByDistrict(ctx context.Context, districtCode string) ([]*domain.SubDistrict, error) {
	query := `SELECT code, district_code, name, created_at, updated_at FROM sub_districts WHERE district_code = $1 ORDER BY name`
	var subDistricts []*domain.SubDistrict
	err := r.db.SelectContext(ctx, &subDistricts, query, districtCode)
	return subDistricts, err
}

func (r *territoryRepository) GetSubDistrictByCode(ctx context.Context, code string) (*domain.SubDistrict, error) {
	query := `SELECT code, district_code, name, created_at, updated_at FROM sub_districts WHERE code = $1`
	var subDistrict domain.SubDistrict
	err := r.db.GetContext(ctx, &subDistrict, query, code)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &subDistrict, err
}

func (r *territoryRepository) UpsertSubDistricts(ctx context.Context, subDistricts []*domain.SubDistrict) error {
	if len(subDistricts) == 0 {
		return nil
	}

	query := `
		INSERT INTO sub_districts (code, district_code, name, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (code) DO UPDATE SET
			district_code = EXCLUDED.district_code,
			name = EXCLUDED.name,
			updated_at = CURRENT_TIMESTAMP
	`

	for _, sd := range subDistricts {
		_, err := r.db.ExecContext(ctx, query, sd.Code, sd.DistrictCode, sd.Name)
		if err != nil {
			return fmt.Errorf("failed to upsert sub_district %s: %w", sd.Code, err)
		}
	}

	return nil
}

// ========== PostalCode Methods ==========

func (r *territoryRepository) GetPostalCodesBySubDistrict(ctx context.Context, subDistrictCode string) ([]*domain.PostalCode, error) {
	query := `SELECT id, sub_district_code, postal_code, created_at FROM postal_codes WHERE sub_district_code = $1`
	var postalCodes []*domain.PostalCode
	err := r.db.SelectContext(ctx, &postalCodes, query, subDistrictCode)
	return postalCodes, err
}

func (r *territoryRepository) SearchByPostalCode(ctx context.Context, postalCode string) ([]*domain.PostalCodeSearchResult, error) {
	query := `
		SELECT 
			pc.postal_code,
			sd.code as sub_district_code, sd.name as sub_district_name,
			d.code as district_code, d.name as district_name,
			c.code as city_code, c.name as city_name,
			p.code as province_code, p.name as province_name
		FROM postal_codes pc
		JOIN sub_districts sd ON pc.sub_district_code = sd.code
		JOIN districts d ON sd.district_code = d.code
		JOIN cities c ON d.city_code = c.code
		JOIN provinces p ON c.province_code = p.code
		WHERE pc.postal_code = $1
		ORDER BY sd.name
	`

	rows, err := r.db.QueryContext(ctx, query, postalCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.PostalCodeSearchResult
	for rows.Next() {
		var result domain.PostalCodeSearchResult
		var subDistrictCode, subDistrictName string
		var districtCode, districtName string
		var cityCode, cityName string
		var provinceCode, provinceName string

		err := rows.Scan(
			&result.PostalCode,
			&subDistrictCode, &subDistrictName,
			&districtCode, &districtName,
			&cityCode, &cityName,
			&provinceCode, &provinceName,
		)
		if err != nil {
			return nil, err
		}

		result.SubDistrict = domain.Location{Code: subDistrictCode, Name: subDistrictName}
		result.District = domain.Location{Code: districtCode, Name: districtName}
		result.City = domain.Location{Code: cityCode, Name: cityName}
		result.Province = domain.Location{Code: provinceCode, Name: provinceName}

		results = append(results, &result)
	}

	return results, nil
}

func (r *territoryRepository) UpsertPostalCodes(ctx context.Context, postalCodes []*domain.PostalCode) error {
	if len(postalCodes) == 0 {
		return nil
	}

	query := `
		INSERT INTO postal_codes (sub_district_code, postal_code, created_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (sub_district_code, postal_code) DO NOTHING
	`

	for _, pc := range postalCodes {
		_, err := r.db.ExecContext(ctx, query, pc.SubDistrictCode, pc.PostalCode)
		if err != nil {
			return fmt.Errorf("failed to upsert postal_code %s: %w", pc.PostalCode, err)
		}
	}

	return nil
}

// ========== Sync Log Methods ==========

func (r *territoryRepository) LogSync(ctx context.Context, log *domain.TerritorySyncLog) error {
	query := `
		INSERT INTO territory_sync_log (sync_type, total_records, status, error_message, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		log.SyncType,
		log.TotalRecords,
		log.Status,
		log.ErrorMessage,
		log.StartedAt,
		log.CompletedAt,
	).Scan(&log.ID)
}
