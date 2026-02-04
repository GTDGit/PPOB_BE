package database

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Migrator handles database migrations
type Migrator struct {
	db     *sqlx.DB
	logger *slog.Logger
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sqlx.DB, logger *slog.Logger) *Migrator {
	return &Migrator{
		db:     db,
		logger: logger,
	}
}

// RunMigrations runs all pending migrations from the specified directory
func (m *Migrator) RunMigrations(migrationsDir string) error {
	// Create migrations tracking table if not exists
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migrations by filename (they should be numbered like 001_xxx.sql)
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Mark existing tables' migrations as applied (handles pre-migrator databases)
	if err := m.markExistingMigrationsAsApplied(migrationFiles); err != nil {
		return fmt.Errorf("failed to mark existing migrations: %w", err)
	}

	// Get already applied migrations
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	pendingCount := 0
	for _, filename := range migrationFiles {
		if applied[filename] {
			continue
		}
		pendingCount++

		// Read migration file
		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// Execute migration
		m.logger.Info("running migration", "file", filename)
		if err := m.executeMigration(filename, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		m.logger.Info("migration completed", "file", filename)
	}

	if pendingCount == 0 {
		m.logger.Info("no pending migrations")
	} else {
		m.logger.Info("migrations applied", "count", pendingCount)
	}

	return nil
}

// createMigrationsTable creates the schema_migrations table to track applied migrations
func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) UNIQUE NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := m.db.Exec(query)
	return err
}

// markExistingMigrationsAsApplied checks if tables already exist and marks their migrations as applied
// This handles the case where the database was set up before the migration system was introduced
func (m *Migrator) markExistingMigrationsAsApplied(migrationFiles []string) error {
	// Map of migration files to their primary table names
	migrationTables := map[string]string{
		"001_create_users_table.sql":                      "users",
		"002_create_devices_table.sql":                    "devices",
		"003_create_sessions_table.sql":                   "sessions",
		"004_create_otp_sessions_table.sql":               "otp_sessions",
		"005_create_balances_table.sql":                   "balances",
		"006_create_user_settings_table.sql":              "user_settings",
		"007_create_services_tables.sql":                  "services",
		"008_create_banners_table.sql":                    "banners",
		"009_create_operators_table.sql":                  "operators",
		"010_create_products_table.sql":                   "products",
		"011_create_providers_tables.sql":                 "providers",
		"012_create_transactions_table.sql":               "transactions",
		"013_create_vouchers_tables.sql":                  "vouchers",
		"014_create_contacts_table.sql":                   "contacts",
		"015_create_deposits_table.sql":                   "deposits",
		"016_create_qris_tables.sql":                      "qris_transactions",
		"017_create_notifications_tables.sql":             "notifications",
		"018_create_referrals_table.sql":                  "referrals",
		"019_create_audit_logs_table.sql":                 "audit_logs",
		"020_create_email_templates_table.sql":            "email_templates",
		"021_create_prepaid_postpaid_transfer_tables.sql": "prepaid_transactions",
		"022_recreate_products_for_gtd.sql":               "product_categories",
		"023_add_phone_verified_at_to_users.sql":          "",
		"024_add_columns_to_deposits.sql":                 "",
		"025_add_columns_to_notifications.sql":            "",
		"026_create_balance_history_table.sql":            "balance_history",
		"027_add_composite_indexes.sql":                   "",
		"028_create_banks_table.sql":                      "banks",
		"029_update_transfer_tables.sql":                  "",
		"030_create_territory_tables.sql":                 "territories",
		"031_create_kyc_tables.sql":                       "kyc_documents",
	}

	for _, filename := range migrationFiles {
		tableName, exists := migrationTables[filename]
		if !exists {
			continue
		}

		// For migrations that add columns (empty tableName), check if already recorded
		if tableName == "" {
			continue
		}

		// Check if the table already exists
		var tableExists bool
		err := m.db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' AND table_name = $1
			)
		`, tableName).Scan(&tableExists)
		if err != nil {
			return fmt.Errorf("failed to check table existence for %s: %w", tableName, err)
		}

		if tableExists {
			// Mark this migration as already applied
			_, err := m.db.Exec(`
				INSERT INTO schema_migrations (filename) VALUES ($1)
				ON CONFLICT (filename) DO NOTHING
			`, filename)
			if err != nil {
				return fmt.Errorf("failed to mark migration %s as applied: %w", filename, err)
			}
			m.logger.Info("marked existing migration as applied", "file", filename, "table", tableName)
		}
	}

	return nil
}

// getAppliedMigrations returns a map of already applied migrations
func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	applied := make(map[string]bool)

	rows, err := m.db.Query("SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}

	return applied, rows.Err()
}

// executeMigration executes a single migration within a transaction
func (m *Migrator) executeMigration(filename, content string) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(content); err != nil {
		return fmt.Errorf("SQL error: %w", err)
	}

	// Record migration as applied
	if _, err := tx.Exec("INSERT INTO schema_migrations (filename) VALUES ($1)", filename); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}
