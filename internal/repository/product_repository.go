package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// ProductFilter for filtering products
type ProductFilter struct {
	Type     string // prepaid, postpaid
	Category string
	Brand    string
	Search   string
	IsActive *bool
	Page     int
	PerPage  int
}

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	// GTD Product Sync methods
	FindBySKU(ctx context.Context, skuCode string) (*domain.Product, error)
	Create(ctx context.Context, product *domain.Product) error
	Update(ctx context.Context, product *domain.Product) error
	BulkUpsert(ctx context.Context, products []*domain.Product) error

	// Query methods for user API
	FindAll(ctx context.Context, filter ProductFilter) ([]*domain.Product, error)
	FindByID(ctx context.Context, id string) (*domain.Product, error)
	FindByCategory(ctx context.Context, category string) ([]*domain.Product, error)
	FindByBrand(ctx context.Context, brand string) ([]*domain.Product, error)

	// Stats
	CountAll(ctx context.Context) (int, error)
	GetLastSyncTime(ctx context.Context) (*time.Time, error)

	// Static provider data (Operators, E-wallet, PDAM, Banks, TV)
	FindAllOperators(ctx context.Context) ([]*domain.Operator, error)
	FindOperatorByPrefix(ctx context.Context, prefix string) (*domain.Operator, error)
	FindAllEwalletProviders(ctx context.Context) ([]*domain.EwalletProvider, error)
	FindAllPDAMRegions(ctx context.Context) ([]*domain.PDAMRegion, error)
	FindAllBanks(ctx context.Context, filterType string) ([]*domain.Bank, error)
	FindBankByCode(ctx context.Context, code string) (*domain.Bank, error)
	UpsertBanks(ctx context.Context, banks []*domain.Bank) error
	FindAllTVProviders(ctx context.Context) ([]*domain.TVProvider, error)
}

// productRepository implements ProductRepository
type productRepository struct {
	db *sqlx.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sqlx.DB) ProductRepository {
	return &productRepository{db: db}
}

// FindAllOperators returns all mobile operators
func (r *productRepository) FindAllOperators(ctx context.Context) ([]*domain.Operator, error) {
	operators := []*domain.Operator{
		{
			ID:       "telkomsel",
			Name:     "Telkomsel",
			Prefixes: []string{"0811", "0812", "0813", "0821", "0822", "0823", "0852", "0853"},
			Icon:     "telkomsel",
			IconURL:  "https://cdn.ppob.id/operators/telkomsel.png",
			Status:   domain.StatusActive,
		},
		{
			ID:       "indosat",
			Name:     "Indosat Ooredoo",
			Prefixes: []string{"0814", "0815", "0816", "0855", "0856", "0857", "0858"},
			Icon:     "indosat",
			IconURL:  "https://cdn.ppob.id/operators/indosat.png",
			Status:   domain.StatusActive,
		},
		{
			ID:       "xl",
			Name:     "XL Axiata",
			Prefixes: []string{"0817", "0818", "0819", "0859", "0877", "0878"},
			Icon:     "xl",
			IconURL:  "https://cdn.ppob.id/operators/xl.png",
			Status:   domain.StatusActive,
		},
		{
			ID:       "axis",
			Name:     "Axis",
			Prefixes: []string{"0831", "0832", "0833", "0838"},
			Icon:     "axis",
			IconURL:  "https://cdn.ppob.id/operators/axis.png",
			Status:   domain.StatusActive,
		},
		{
			ID:       "three",
			Name:     "Tri",
			Prefixes: []string{"0895", "0896", "0897", "0898", "0899"},
			Icon:     "three",
			IconURL:  "https://cdn.ppob.id/operators/three.png",
			Status:   domain.StatusActive,
		},
		{
			ID:       "smartfren",
			Name:     "Smartfren",
			Prefixes: []string{"0881", "0882", "0883", "0884", "0885", "0886", "0887", "0888", "0889"},
			Icon:     "smartfren",
			IconURL:  "https://cdn.ppob.id/operators/smartfren.png",
			Status:   domain.StatusActive,
		},
		{
			ID:       "byu",
			Name:     "by.U",
			Prefixes: []string{"0851"},
			Icon:     "byu",
			IconURL:  "https://cdn.ppob.id/operators/byu.png",
			Status:   domain.StatusActive,
		},
	}
	return operators, nil
}

// FindOperatorByPrefix finds operator by phone prefix
func (r *productRepository) FindOperatorByPrefix(ctx context.Context, prefix string) (*domain.Operator, error) {
	operators, _ := r.FindAllOperators(ctx)

	// Check prefix (first 4 digits)
	if len(prefix) < 4 {
		return nil, nil
	}

	checkPrefix := prefix[:4]
	for _, op := range operators {
		for _, p := range op.Prefixes {
			if p == checkPrefix {
				return op, nil
			}
		}
	}
	return nil, nil
}

// FindAllEwalletProviders returns all e-wallet providers
func (r *productRepository) FindAllEwalletProviders(ctx context.Context) ([]*domain.EwalletProvider, error) {
	providers := []*domain.EwalletProvider{
		{
			ID:               "gopay",
			Name:             "GoPay",
			Icon:             "gopay",
			IconURL:          "https://cdn.ppob.id/ewallet/gopay.png",
			InputLabel:       "Nomor HP GoPay",
			InputPlaceholder: "08xxxxxxxxxx",
			InputType:        "phone",
			Status:           domain.StatusActive,
		},
		{
			ID:               "ovo",
			Name:             "OVO",
			Icon:             "ovo",
			IconURL:          "https://cdn.ppob.id/ewallet/ovo.png",
			InputLabel:       "Nomor HP OVO",
			InputPlaceholder: "08xxxxxxxxxx",
			InputType:        "phone",
			Status:           domain.StatusActive,
		},
		{
			ID:               "dana",
			Name:             "DANA",
			Icon:             "dana",
			IconURL:          "https://cdn.ppob.id/ewallet/dana.png",
			InputLabel:       "Nomor HP DANA",
			InputPlaceholder: "08xxxxxxxxxx",
			InputType:        "phone",
			Status:           domain.StatusActive,
		},
		{
			ID:               "shopeepay",
			Name:             "ShopeePay",
			Icon:             "shopeepay",
			IconURL:          "https://cdn.ppob.id/ewallet/shopeepay.png",
			InputLabel:       "Nomor HP ShopeePay",
			InputPlaceholder: "08xxxxxxxxxx",
			InputType:        "phone",
			Status:           domain.StatusActive,
		},
		{
			ID:               "linkaja",
			Name:             "LinkAja",
			Icon:             "linkaja",
			IconURL:          "https://cdn.ppob.id/ewallet/linkaja.png",
			InputLabel:       "Nomor HP LinkAja",
			InputPlaceholder: "08xxxxxxxxxx",
			InputType:        "phone",
			Status:           domain.StatusActive,
		},
	}
	return providers, nil
}

// FindAllPDAMRegions returns all PDAM regions
func (r *productRepository) FindAllPDAMRegions(ctx context.Context) ([]*domain.PDAMRegion, error) {
	regions := []*domain.PDAMRegion{
		{
			ID:       "pdam_jakarta",
			Name:     "PDAM DKI Jakarta",
			Province: "DKI Jakarta",
			Status:   domain.StatusActive,
		},
		{
			ID:       "pdam_bandung",
			Name:     "PDAM Tirta Wening Kota Bandung",
			Province: "Jawa Barat",
			Status:   domain.StatusActive,
		},
		{
			ID:       "pdam_surabaya",
			Name:     "PDAM Surya Sembada Surabaya",
			Province: "Jawa Timur",
			Status:   domain.StatusActive,
		},
		{
			ID:       "pdam_semarang",
			Name:     "PDAM Tirta Moedal Semarang",
			Province: "Jawa Tengah",
			Status:   domain.StatusActive,
		},
		{
			ID:       "pdam_medan",
			Name:     "PDAM Tirtanadi Medan",
			Province: "Sumatera Utara",
			Status:   domain.StatusActive,
		},
		{
			ID:       "pdam_palembang",
			Name:     "PDAM Tirta Musi Palembang",
			Province: "Sumatera Selatan",
			Status:   domain.StatusActive,
		},
		{
			ID:       "pdam_makassar",
			Name:     "PDAM Makassar",
			Province: "Sulawesi Selatan",
			Status:   domain.StatusActive,
		},
		{
			ID:       "pdam_denpasar",
			Name:     "PDAM Denpasar",
			Province: "Bali",
			Status:   domain.StatusActive,
		},
	}
	return regions, nil
}

// FindAllBanks returns all banks from database with optional filtering
func (r *productRepository) FindAllBanks(ctx context.Context, filterType string) ([]*domain.Bank, error) {
	var query string

	// Build query based on filter type
	switch filterType {
	case "popular":
		query = `
			SELECT id, code, name, short_name, swift_code, icon, icon_url,
			       transfer_fee, transfer_fee_formatted, is_popular, status,
			       created_at, updated_at
			FROM banks
			WHERE is_popular = true AND status = 'active'
			ORDER BY name ASC
		`
	case "transfer_supported":
		// All active banks support transfer
		query = `
			SELECT id, code, name, short_name, swift_code, icon, icon_url,
			       transfer_fee, transfer_fee_formatted, is_popular, status,
			       created_at, updated_at
			FROM banks
			WHERE status = 'active'
			ORDER BY name ASC
		`
	default:
		// Return all active banks
		query = `
			SELECT id, code, name, short_name, swift_code, icon, icon_url,
			       transfer_fee, transfer_fee_formatted, is_popular, status,
			       created_at, updated_at
			FROM banks
			WHERE status = 'active'
			ORDER BY name ASC
		`
	}

	var banks []*domain.Bank
	if err := r.db.SelectContext(ctx, &banks, query); err != nil {
		// If table doesn't exist or is empty, return empty array (not error)
		// This allows graceful degradation before first sync
		return []*domain.Bank{}, nil
	}

	return banks, nil
}

// FindBankByCode finds bank by code from database
func (r *productRepository) FindBankByCode(ctx context.Context, code string) (*domain.Bank, error) {
	query := `
		SELECT id, code, name, short_name, swift_code, icon, icon_url,
		       transfer_fee, transfer_fee_formatted, is_popular, status,
		       created_at, updated_at
		FROM banks
		WHERE code = $1 AND status = 'active'
		LIMIT 1
	`

	var bank domain.Bank
	if err := r.db.GetContext(ctx, &bank, query, code); err != nil {
		// Return nil if not found (graceful degradation)
		return nil, nil
	}

	return &bank, nil
}

// UpsertBanks inserts or updates multiple banks in a transaction
func (r *productRepository) UpsertBanks(ctx context.Context, banks []*domain.Bank) error {
	if len(banks) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO banks (
			id, code, name, short_name, swift_code, icon, icon_url,
			transfer_fee, transfer_fee_formatted, is_popular, status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (code)
		DO UPDATE SET
			name = EXCLUDED.name,
			short_name = EXCLUDED.short_name,
			swift_code = EXCLUDED.swift_code,
			icon = EXCLUDED.icon,
			icon_url = EXCLUDED.icon_url,
			transfer_fee = EXCLUDED.transfer_fee,
			transfer_fee_formatted = EXCLUDED.transfer_fee_formatted,
			is_popular = EXCLUDED.is_popular,
			status = EXCLUDED.status,
			updated_at = NOW()
	`

	stmt, err := tx.PreparexContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, bank := range banks {
		_, err := stmt.ExecContext(ctx,
			bank.ID,
			bank.Code,
			bank.Name,
			bank.ShortName,
			bank.SwiftCode,
			bank.Icon,
			bank.IconURL,
			bank.TransferFee,
			bank.TransferFeeFormatted,
			bank.IsPopular,
			bank.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert bank %s: %w", bank.Code, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindAllTVProviders returns all TV cable providers
func (r *productRepository) FindAllTVProviders(ctx context.Context) ([]*domain.TVProvider, error) {
	providers := []*domain.TVProvider{
		{
			ID:         "indovision",
			Name:       "Indovision",
			Icon:       "indovision",
			IconURL:    "https://cdn.ppob.id/tv/indovision.png",
			InputLabel: "Nomor Pelanggan",
			Status:     domain.StatusActive,
		},
		{
			ID:         "transvision",
			Name:       "Transvision",
			Icon:       "transvision",
			IconURL:    "https://cdn.ppob.id/tv/transvision.png",
			InputLabel: "Nomor Pelanggan",
			Status:     domain.StatusActive,
		},
		{
			ID:         "topas",
			Name:       "Topas TV",
			Icon:       "topas",
			IconURL:    "https://cdn.ppob.id/tv/topas.png",
			InputLabel: "Nomor Pelanggan",
			Status:     domain.StatusActive,
		},
		{
			ID:         "firstmedia",
			Name:       "First Media",
			Icon:       "firstmedia",
			IconURL:    "https://cdn.ppob.id/tv/firstmedia.png",
			InputLabel: "Nomor Pelanggan",
			Status:     domain.StatusActive,
		},
		{
			ID:         "k_vision",
			Name:       "K-Vision",
			Icon:       "kvision",
			IconURL:    "https://cdn.ppob.id/tv/kvision.png",
			InputLabel: "Nomor Pelanggan",
			Status:     domain.StatusActive,
		},
		{
			ID:         "nex_parabola",
			Name:       "Nex Parabola",
			Icon:       "nexparabola",
			IconURL:    "https://cdn.ppob.id/tv/nexparabola.png",
			InputLabel: "Nomor Pelanggan",
			Status:     domain.StatusActive,
		},
	}
	return providers, nil
}

// ========== GTD Product Sync Methods ==========

// Mock storage for synced products (in-memory for now)
var mockProducts = []*domain.Product{}

// FindBySKU finds product by SKU code
func (r *productRepository) FindBySKU(ctx context.Context, skuCode string) (*domain.Product, error) {
	// Mock implementation - will use DB query in production
	for _, product := range mockProducts {
		if product.SKUCode == skuCode {
			return product, nil
		}
	}
	return nil, nil
}

// Create creates a new product
func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	// Mock implementation
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	mockProducts = append(mockProducts, product)
	return nil
}

// Update updates an existing product
func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	// Mock implementation
	for i, p := range mockProducts {
		if p.SKUCode == product.SKUCode {
			product.UpdatedAt = time.Now()
			mockProducts[i] = product
			return nil
		}
	}
	return fmt.Errorf("product not found")
}

// BulkUpsert upserts multiple products efficiently
func (r *productRepository) BulkUpsert(ctx context.Context, products []*domain.Product) error {
	// Mock implementation - will use INSERT ON CONFLICT in production
	for _, product := range products {
		existing, _ := r.FindBySKU(ctx, product.SKUCode)
		if existing != nil {
			// Update existing
			r.Update(ctx, product)
		} else {
			// Create new
			r.Create(ctx, product)
		}
	}
	return nil
}

// FindAll returns products with filters
func (r *productRepository) FindAll(ctx context.Context, filter ProductFilter) ([]*domain.Product, error) {
	// Mock implementation
	result := []*domain.Product{}

	for _, product := range mockProducts {
		// Apply filters
		if filter.Type != "" && product.Type != filter.Type {
			continue
		}
		if filter.Category != "" && product.Category != filter.Category {
			continue
		}
		if filter.Brand != "" && product.Brand != filter.Brand {
			continue
		}
		if filter.IsActive != nil && product.IsActive != *filter.IsActive {
			continue
		}

		result = append(result, product)
	}

	// Apply pagination
	if filter.Page > 0 && filter.PerPage > 0 {
		start := (filter.Page - 1) * filter.PerPage
		end := start + filter.PerPage

		if start >= len(result) {
			return []*domain.Product{}, nil
		}
		if end > len(result) {
			end = len(result)
		}

		result = result[start:end]
	}

	return result, nil
}

// FindByID finds product by ID
func (r *productRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	for _, product := range mockProducts {
		if product.ID == id {
			return product, nil
		}
	}
	return nil, nil
}

// FindByCategory finds products by category
func (r *productRepository) FindByCategory(ctx context.Context, category string) ([]*domain.Product, error) {
	result := []*domain.Product{}
	for _, product := range mockProducts {
		if product.Category == category && product.IsActive {
			result = append(result, product)
		}
	}
	return result, nil
}

// FindByBrand finds products by brand
func (r *productRepository) FindByBrand(ctx context.Context, brand string) ([]*domain.Product, error) {
	result := []*domain.Product{}
	for _, product := range mockProducts {
		if product.Brand == brand && product.IsActive {
			result = append(result, product)
		}
	}
	return result, nil
}

// CountAll returns total products count
func (r *productRepository) CountAll(ctx context.Context) (int, error) {
	return len(mockProducts), nil
}

// GetLastSyncTime returns last sync timestamp
func (r *productRepository) GetLastSyncTime(ctx context.Context) (*time.Time, error) {
	if len(mockProducts) == 0 {
		return nil, nil
	}

	// Return latest UpdatedAt
	var latest time.Time
	for _, p := range mockProducts {
		if p.UpdatedAt.After(latest) {
			latest = p.UpdatedAt
		}
	}

	return &latest, nil
}

// ========== Static Provider Data Methods ==========
