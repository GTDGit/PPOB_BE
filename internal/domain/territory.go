package domain

import "time"

// Province represents Indonesian province
type Province struct {
	Code      string    `json:"code" db:"code"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

// City represents Indonesian city/regency
type City struct {
	Code         string    `json:"code" db:"code"`
	ProvinceCode string    `json:"provinceCode" db:"province_code"`
	Name         string    `json:"name" db:"name"`
	CreatedAt    time.Time `json:"-" db:"created_at"`
	UpdatedAt    time.Time `json:"-" db:"updated_at"`
}

// District represents Indonesian district
type District struct {
	Code      string    `json:"code" db:"code"`
	CityCode  string    `json:"cityCode" db:"city_code"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

// SubDistrict represents Indonesian sub-district/village
type SubDistrict struct {
	Code         string    `json:"code" db:"code"`
	DistrictCode string    `json:"districtCode" db:"district_code"`
	Name         string    `json:"name" db:"name"`
	CreatedAt    time.Time `json:"-" db:"created_at"`
	UpdatedAt    time.Time `json:"-" db:"updated_at"`
}

// PostalCode represents postal code data
type PostalCode struct {
	ID              int       `json:"-" db:"id"`
	SubDistrictCode string    `json:"subDistrictCode" db:"sub_district_code"`
	PostalCode      string    `json:"postalCode" db:"postal_code"`
	CreatedAt       time.Time `json:"-" db:"created_at"`
}

// PostalCodeSearchResult represents postal code search result with full hierarchy
type PostalCodeSearchResult struct {
	PostalCode  string   `json:"postalCode"`
	SubDistrict Location `json:"subDistrict"`
	District    Location `json:"district"`
	City        Location `json:"city"`
	Province    Location `json:"province"`
}

// Location represents a generic location (code + name)
type Location struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// TerritorySyncLog represents territory sync metadata
type TerritorySyncLog struct {
	ID           int        `db:"id"`
	SyncType     string     `db:"sync_type"`
	TotalRecords int        `db:"total_records"`
	Status       string     `db:"status"`
	ErrorMessage *string    `db:"error_message"`
	StartedAt    time.Time  `db:"started_at"`
	CompletedAt  *time.Time `db:"completed_at"`
	CreatedAt    time.Time  `db:"created_at"`
}

// Territory sync status constants
const (
	TerritorySyncStatusRunning = "running"
	TerritorySyncStatusSuccess = "success"
	TerritorySyncStatusFailed  = "failed"
)

// Territory sync types
const (
	TerritorySyncTypeProvinces    = "provinces"
	TerritorySyncTypeCities       = "cities"
	TerritorySyncTypeDistricts    = "districts"
	TerritorySyncTypeSubDistricts = "sub_districts"
	TerritorySyncTypePostalCodes  = "postal_codes"
)
