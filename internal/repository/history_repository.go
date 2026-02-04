package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// HistoryRepository defines the interface for transaction history operations
type HistoryRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Transaction, error)
	FindByUserAndID(ctx context.Context, userID, id string) (*domain.Transaction, error)
	FindByUserID(ctx context.Context, userID string, filter TransactionFilter) ([]*domain.Transaction, int, error)
	Create(ctx context.Context, tx *domain.Transaction) error
	UpdateStatus(ctx context.Context, id, status string, receiptData *string) error
}

// TransactionFilter represents filter options for transaction list
type TransactionFilter struct {
	Type        string // prepaid, postpaid, transfer, all
	ServiceType string // pulsa, data, pln_prepaid, etc
	Status      string // success, pending, failed, cancelled, all
	StartDate   *time.Time
	EndDate     *time.Time
	Search      string
	Page        int
	PerPage     int
}

// historyRepository implements HistoryRepository
type historyRepository struct {
	db *sqlx.DB
}

// NewHistoryRepository creates a new history repository
func NewHistoryRepository(db *sqlx.DB) HistoryRepository {
	return &historyRepository{db: db}
}

// FindByID finds a transaction by ID
func (r *historyRepository) FindByID(ctx context.Context, id string) (*domain.Transaction, error) {
	// For now, use mock data
	transactions := r.getMockTransactions()
	for _, tx := range transactions {
		if tx.ID == id {
			return tx, nil
		}
	}
	return nil, nil
}

// FindByUserAndID finds a transaction by user ID and transaction ID (ownership validation)
func (r *historyRepository) FindByUserAndID(ctx context.Context, userID, id string) (*domain.Transaction, error) {
	transactions := r.getMockTransactions()
	for _, tx := range transactions {
		if tx.ID == id && tx.UserID == userID {
			return tx, nil
		}
	}
	return nil, nil
}

// FindByUserID finds transactions by user ID with filters and pagination
func (r *historyRepository) FindByUserID(ctx context.Context, userID string, filter TransactionFilter) ([]*domain.Transaction, int, error) {
	// For now, return mock transactions
	allTransactions := r.getMockTransactions()

	// Filter by user
	var userTransactions []*domain.Transaction
	for _, tx := range allTransactions {
		if tx.UserID == userID {
			userTransactions = append(userTransactions, tx)
		}
	}

	// Apply filters
	var filtered []*domain.Transaction
	for _, tx := range userTransactions {
		// Filter by type
		if filter.Type != "" && filter.Type != "all" && tx.Type != filter.Type {
			continue
		}

		// Filter by service type
		if filter.ServiceType != "" && tx.ServiceType != filter.ServiceType {
			continue
		}

		// Filter by status
		if filter.Status != "" && filter.Status != "all" && tx.Status != filter.Status {
			continue
		}

		// Filter by date range
		if filter.StartDate != nil && tx.CreatedAt.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && tx.CreatedAt.After(*filter.EndDate) {
			continue
		}

		// Filter by search (target)
		// In production, use proper SQL LIKE or full-text search

		filtered = append(filtered, tx)
	}

	total := len(filtered)

	// Apply pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 50 {
		perPage = 50
	}

	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(filtered) {
		return []*domain.Transaction{}, total, nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	paginated := filtered[start:end]

	return paginated, total, nil
}

// Create creates a new transaction
func (r *historyRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	// In production, insert into database
	// For now, just return success
	return nil
}

// UpdateStatus updates transaction status
func (r *historyRepository) UpdateStatus(ctx context.Context, id, status string, receiptData *string) error {
	// In production, update database
	// For now, just return success
	return nil
}

// getMockTransactions returns mock transaction data
func (r *historyRepository) getMockTransactions() []*domain.Transaction {
	now := time.Now()
	completedAt1 := now.Add(-5 * time.Second)
	completedAt2 := now.Add(-10 * time.Second)

	serialNum1 := "SN123456789012"
	serialNum2 := "REF20251214170005"
	serialNum3 := "REF20251214170010"
	token := "1234-5678-9012-3456-7890"

	return []*domain.Transaction{
		// Transfer - Processing
		{
			ID:           "trx_tf_abc123",
			UserID:       "mock_user_id",
			Type:         domain.TransactionTypeTransfer,
			ServiceType:  "transfer_bank",
			Target:       "1234567890",
			ProductName:  "Transfer Bank BCA",
			Amount:       10000,
			AdminFee:     6500,
			Discount:     0,
			TotalPayment: 16500,
			Status:       domain.TransactionStatusProcessing,
			CreatedAt:    now.Add(-5 * time.Minute),
			UpdatedAt:    now.Add(-5 * time.Minute),
		},
		// Pulsa - Cancelled
		{
			ID:           "trx_pulsa_xyz789",
			UserID:       "mock_user_id",
			Type:         domain.TransactionTypePrepaid,
			ServiceType:  "pulsa",
			Target:       "081234567890",
			ProductName:  "Pulsa 10.000",
			Amount:       10000,
			AdminFee:     0,
			Discount:     0,
			TotalPayment: 10000,
			Status:       domain.TransactionStatusCancelled,
			CreatedAt:    now.Add(-10 * time.Minute),
			UpdatedAt:    now.Add(-10 * time.Minute),
		},
		// Paket Data - Success
		{
			ID:           "trx_data_def456",
			UserID:       "mock_user_id",
			Type:         domain.TransactionTypePrepaid,
			ServiceType:  "data",
			Target:       "081234567890",
			ProductName:  "Paket Data 5GB",
			Amount:       5000,
			AdminFee:     1250,
			Discount:     0,
			TotalPayment: 6250,
			Status:       domain.TransactionStatusSuccess,
			SerialNumber: &serialNum1,
			CompletedAt:  &completedAt1,
			CreatedAt:    now.Add(-15 * time.Minute),
			UpdatedAt:    now.Add(-15 * time.Minute),
		},
		// Transfer - Success
		{
			ID:           "trx_tf_ghi789",
			UserID:       "mock_user_id",
			Type:         domain.TransactionTypeTransfer,
			ServiceType:  "transfer_bank",
			Target:       "9876543210",
			ProductName:  "Transfer Bank BCA",
			Amount:       5000,
			AdminFee:     1250,
			Discount:     0,
			TotalPayment: 6250,
			Status:       domain.TransactionStatusSuccess,
			SerialNumber: &serialNum2,
			CompletedAt:  &completedAt2,
			CreatedAt:    now.Add(-20 * time.Minute),
			UpdatedAt:    now.Add(-20 * time.Minute),
		},
		// PLN Token - Success
		{
			ID:           "trx_pln_jkl012",
			UserID:       "mock_user_id",
			Type:         domain.TransactionTypePrepaid,
			ServiceType:  "pln_prepaid",
			Target:       "123456789012",
			ProductName:  "Token PLN 50.000",
			Amount:       50000,
			AdminFee:     2500,
			Discount:     0,
			TotalPayment: 52500,
			Status:       domain.TransactionStatusSuccess,
			SerialNumber: &serialNum3,
			Token:        &token,
			CompletedAt:  &completedAt1,
			CreatedAt:    now.Add(-30 * time.Minute),
			UpdatedAt:    now.Add(-30 * time.Minute),
		},
		// PDAM - Success
		{
			ID:           "trx_pdam_mno345",
			UserID:       "mock_user_id",
			Type:         domain.TransactionTypePostpaid,
			ServiceType:  "pdam",
			Target:       "987654321",
			ProductName:  "Tagihan PDAM",
			Amount:       75000,
			AdminFee:     2500,
			Discount:     0,
			TotalPayment: 77500,
			Status:       domain.TransactionStatusSuccess,
			SerialNumber: &serialNum1,
			CompletedAt:  &completedAt2,
			CreatedAt:    now.Add(-45 * time.Minute),
			UpdatedAt:    now.Add(-45 * time.Minute),
		},
	}
}
