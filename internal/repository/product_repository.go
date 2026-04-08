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

// FindAllOperators returns all mobile operators from database
func (r *productRepository) FindAllOperators(ctx context.Context) ([]*domain.Operator, error) {
	query := `
		SELECT id, name, prefixes, icon, icon_url, status, sort_order
		FROM operators
		WHERE status = 'active'
		ORDER BY sort_order ASC
	`

	var operators []*domain.Operator
	if err := r.db.SelectContext(ctx, &operators, query); err != nil {
		return []*domain.Operator{}, nil
	}

	// Parse JSON prefixes from DB text to []string
	for _, op := range operators {
		if op.PrefixesDB != "" {
			json.Unmarshal([]byte(op.PrefixesDB), &op.Prefixes)
		}
		if op.Prefixes == nil {
			op.Prefixes = []string{}
		}
	}

	return operators, nil
}

// FindOperatorByPrefix finds operator by phone prefix
func (r *productRepository) FindOperatorByPrefix(ctx context.Context, prefix string) (*domain.Operator, error) {
	if len(prefix) < 4 {
		return nil, nil
	}

	checkPrefix := prefix[:4]

	// Query all active operators and check prefix in Go
	// (prefixes is stored as JSON text, checked after parsing)
	operators, err := r.FindAllOperators(ctx)
	if err != nil {
		return nil, err
	}

	for _, op := range operators {
		for _, p := range op.Prefixes {
			if p == checkPrefix {
				return op, nil
			}
		}
	}
	return nil, nil
}

// FindAllEwalletProviders returns all e-wallet providers from database
func (r *productRepository) FindAllEwalletProviders(ctx context.Context) ([]*domain.EwalletProvider, error) {
	query := `
		SELECT id, name, icon, icon_url, input_label, input_placeholder, input_type, status, sort_order
		FROM ewallet_providers
		WHERE status = 'active'
		ORDER BY sort_order ASC
	`

	var providers []*domain.EwalletProvider
	if err := r.db.SelectContext(ctx, &providers, query); err != nil {
		return []*domain.EwalletProvider{}, nil
	}

	return providers, nil
}

// FindAllPDAMRegions returns all PDAM regions from database
func (r *productRepository) FindAllPDAMRegions(ctx context.Context) ([]*domain.PDAMRegion, error) {
	query := `
		SELECT id, name, province, status, sort_order
		FROM pdam_regions
		WHERE status = 'active'
		ORDER BY sort_order ASC
	`

	var regions []*domain.PDAMRegion
	if err := r.db.SelectContext(ctx, &regions, query); err != nil {
		return []*domain.PDAMRegion{}, nil
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

// FindAllTVProviders returns all TV cable providers from database
func (r *productRepository) FindAllTVProviders(ctx context.Context) ([]*domain.TVProvider, error) {
	query := `
		SELECT id, name, icon, icon_url, input_label, status, sort_order
		FROM tv_providers
		WHERE status = 'active'
		ORDER BY sort_order ASC
	`

	var providers []*domain.TVProvider
	if err := r.db.SelectContext(ctx, &providers, query); err != nil {
		return []*domain.TVProvider{}, nil
	}

	return providers, nil
}

// ========== GTD Product Sync Methods ==========

const productColumns = `id, sku_code, name, category, brand, type, price, admin,
	commission, is_active, description, gtd_updated_at, created_at, updated_at`

// FindBySKU finds product by SKU code
func (r *productRepository) FindBySKU(ctx context.Context, skuCode string) (*domain.Product, error) {
	query := `SELECT ` + productColumns + ` FROM products WHERE sku_code = $1 LIMIT 1`

	var product domain.Product
	if err := r.db.GetContext(ctx, &product, query, skuCode); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

// Create creates a new product
func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (
			id, sku_code, name, category, brand, type, price, admin,
			commission, is_active, description, gtd_updated_at, created_at, updated_at
		) VALUES (
			:id, :sku_code, :name, :category, :brand, :type, :price, :admin,
			:commission, :is_active, :description, :gtd_updated_at, NOW(), NOW()
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, product)
	return err
}

// Update updates an existing product
func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET
			id = :id,
			name = :name,
			category = :category,
			brand = :brand,
			type = :type,
			price = :price,
			admin = :admin,
			commission = :commission,
			is_active = :is_active,
			description = :description,
			gtd_updated_at = :gtd_updated_at,
			updated_at = NOW()
		WHERE sku_code = :sku_code
	`
	_, err := r.db.NamedExecContext(ctx, query, product)
	return err
}

// BulkUpsert upserts multiple products efficiently
func (r *productRepository) BulkUpsert(ctx context.Context, products []*domain.Product) error {
	if len(products) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin product upsert transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO products (
			id, sku_code, name, category, brand, type, price, admin,
			commission, is_active, description, gtd_updated_at, created_at, updated_at
		) VALUES (
			:id, :sku_code, :name, :category, :brand, :type, :price, :admin,
			:commission, :is_active, :description, :gtd_updated_at, NOW(), NOW()
		)
		ON CONFLICT (sku_code) DO UPDATE SET
			id = EXCLUDED.id,
			name = EXCLUDED.name,
			category = EXCLUDED.category,
			brand = EXCLUDED.brand,
			type = EXCLUDED.type,
			price = EXCLUDED.price,
			admin = EXCLUDED.admin,
			commission = EXCLUDED.commission,
			is_active = EXCLUDED.is_active,
			description = EXCLUDED.description,
			gtd_updated_at = EXCLUDED.gtd_updated_at,
			updated_at = NOW()
	`

	for _, product := range products {
		if _, err := tx.NamedExecContext(ctx, query, product); err != nil {
			return fmt.Errorf("failed to upsert product %s: %w", product.SKUCode, err)
		}
	}

	return tx.Commit()
}

// FindAll returns products with filters
func (r *productRepository) FindAll(ctx context.Context, filter ProductFilter) ([]*domain.Product, error) {
	query := `SELECT ` + productColumns + ` FROM products WHERE 1=1`
	args := make([]interface{}, 0, 6)
	argIdx := 1

	if filter.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, filter.Type)
		argIdx++
	}
	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, filter.Category)
		argIdx++
	}
	if filter.Brand != "" {
		query += fmt.Sprintf(" AND brand = $%d", argIdx)
		args = append(args, filter.Brand)
		argIdx++
	}
	if filter.Search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR sku_code ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}
	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argIdx)
		args = append(args, *filter.IsActive)
		argIdx++
	}

	query += ` ORDER BY category ASC, brand ASC, price ASC, name ASC`

	if filter.Page > 0 && filter.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, filter.PerPage, (filter.Page-1)*filter.PerPage)
	}

	var products []*domain.Product
	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return []*domain.Product{}, nil
		}
		return nil, err
	}

	return products, nil
}

// FindByID finds product by ID
func (r *productRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `SELECT ` + productColumns + ` FROM products WHERE id = $1 LIMIT 1`

	var product domain.Product
	if err := r.db.GetContext(ctx, &product, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &product, nil
}

// FindByCategory finds products by category
func (r *productRepository) FindByCategory(ctx context.Context, category string) ([]*domain.Product, error) {
	query := `SELECT ` + productColumns + ` FROM products WHERE category = $1 AND is_active = true ORDER BY price ASC, name ASC`

	var products []*domain.Product
	if err := r.db.SelectContext(ctx, &products, query, category); err != nil {
		if err == sql.ErrNoRows {
			return []*domain.Product{}, nil
		}
		return nil, err
	}

	return products, nil
}

// FindByBrand finds products by brand
func (r *productRepository) FindByBrand(ctx context.Context, brand string) ([]*domain.Product, error) {
	query := `SELECT ` + productColumns + ` FROM products WHERE brand = $1 AND is_active = true ORDER BY price ASC, name ASC`

	var products []*domain.Product
	if err := r.db.SelectContext(ctx, &products, query, brand); err != nil {
		if err == sql.ErrNoRows {
			return []*domain.Product{}, nil
		}
		return nil, err
	}

	return products, nil
}

// CountAll returns total products count
func (r *productRepository) CountAll(ctx context.Context) (int, error) {
	var count int
	if err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM products`); err != nil {
		return 0, err
	}
	return count, nil
}

// GetLastSyncTime returns last sync timestamp
func (r *productRepository) GetLastSyncTime(ctx context.Context) (*time.Time, error) {
	var lastSync sql.NullTime
	query := `SELECT MAX(COALESCE(gtd_updated_at, updated_at)) FROM products`

	if err := r.db.GetContext(ctx, &lastSync, query); err != nil {
		return nil, err
	}
	if !lastSync.Valid {
		return nil, nil
	}

	return &lastSync.Time, nil
}

// ========== Static Provider Data Methods ==========
