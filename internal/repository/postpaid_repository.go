package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// PostpaidRepository handles postpaid data access
type PostpaidRepository interface {
	CreateInquiry(ctx context.Context, inquiry *domain.PostpaidInquiry) error
	FindInquiryByID(ctx context.Context, id string) (*domain.PostpaidInquiry, error)
	FindInquiryByUserAndID(ctx context.Context, userID, id string) (*domain.PostpaidInquiry, error)

	CreateTransaction(ctx context.Context, tx *domain.PostpaidTransaction) error
	FindTransactionByID(ctx context.Context, id string) (*domain.PostpaidTransaction, error)
	FindTransactionByInquiryID(ctx context.Context, inquiryID string) (*domain.PostpaidTransaction, error)
	UpdateTransactionStatus(ctx context.Context, id, status string) error

	// Transaction methods for atomic operations
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

type postpaidRepository struct {
	db *sqlx.DB
}

// NewPostpaidRepository creates a new postpaid repository
func NewPostpaidRepository(db *sqlx.DB) PostpaidRepository {
	return &postpaidRepository{db: db}
}

// CreateInquiry creates a new inquiry
func (r *postpaidRepository) CreateInquiry(ctx context.Context, inquiry *domain.PostpaidInquiry) error {
	// Mock implementation - store in memory for now
	mockPostpaidInquiries = append(mockPostpaidInquiries, inquiry)
	return nil
}

// FindInquiryByID finds inquiry by ID
func (r *postpaidRepository) FindInquiryByID(ctx context.Context, id string) (*domain.PostpaidInquiry, error) {
	for _, inquiry := range mockPostpaidInquiries {
		if inquiry.ID == id {
			return inquiry, nil
		}
	}
	return nil, nil
}

// FindInquiryByUserAndID finds inquiry by user and ID (ownership check)
func (r *postpaidRepository) FindInquiryByUserAndID(ctx context.Context, userID, id string) (*domain.PostpaidInquiry, error) {
	for _, inquiry := range mockPostpaidInquiries {
		if inquiry.ID == id && inquiry.UserID == userID {
			return inquiry, nil
		}
	}
	return nil, nil
}

// CreateTransaction creates a new transaction
func (r *postpaidRepository) CreateTransaction(ctx context.Context, tx *domain.PostpaidTransaction) error {
	// Mock implementation
	mockPostpaidTransactions = append(mockPostpaidTransactions, tx)
	return nil
}

// FindTransactionByID finds transaction by ID
func (r *postpaidRepository) FindTransactionByID(ctx context.Context, id string) (*domain.PostpaidTransaction, error) {
	for _, tx := range mockPostpaidTransactions {
		if tx.ID == id {
			return tx, nil
		}
	}
	return nil, nil
}

// FindTransactionByInquiryID finds transaction by inquiry ID (idempotency check)
func (r *postpaidRepository) FindTransactionByInquiryID(ctx context.Context, inquiryID string) (*domain.PostpaidTransaction, error) {
	for _, tx := range mockPostpaidTransactions {
		if tx.InquiryID == inquiryID {
			return tx, nil
		}
	}
	return nil, nil
}

// UpdateTransactionStatus updates transaction status
func (r *postpaidRepository) UpdateTransactionStatus(ctx context.Context, id, status string) error {
	for _, tx := range mockPostpaidTransactions {
		if tx.ID == id {
			tx.Status = status
			tx.UpdatedAt = time.Now()
			return nil
		}
	}
	return nil
}

// BeginTx begins a database transaction
func (r *postpaidRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

// Mock data storage
var mockPostpaidInquiries = []*domain.PostpaidInquiry{
	{
		ID:           "inq_post_mock_1",
		UserID:       "user_mock_1",
		ServiceType:  domain.ServicePLNPostpaid,
		Target:       "123456789012",
		CustomerID:   "123456789012",
		CustomerName: "BUDI SANTOSO",
		Period:       "Januari 2025",
		BillAmount:   350000,
		AdminFee:     2500,
		Penalty:      0,
		TotalPayment: 352500,
		HasBill:      true,
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		CreatedAt:    time.Now(),
	},
}

var mockPostpaidTransactions = []*domain.PostpaidTransaction{}
