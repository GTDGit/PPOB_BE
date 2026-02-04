package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// ContactRepository defines the interface for contact data operations
type ContactRepository interface {
	// CRUD operations
	FindByID(ctx context.Context, id string) (*domain.Contact, error)
	FindByUserID(ctx context.Context, userID string, filter ContactFilter) ([]*domain.Contact, error)
	FindByUserAndID(ctx context.Context, userID, contactID string) (*domain.Contact, error)
	FindByUserAndValue(ctx context.Context, userID, contactType, value string) (*domain.Contact, error)
	Create(ctx context.Context, contact *domain.Contact) error
	Update(ctx context.Context, contact *domain.Contact) error
	Delete(ctx context.Context, id string) error
	IncrementUsage(ctx context.Context, id string) error
}

// ContactFilter represents filter options for listing contacts
type ContactFilter struct {
	Type   string
	Search string
	Limit  int
}

// contactRepository implements ContactRepository
type contactRepository struct {
	db *sqlx.DB
}

// NewContactRepository creates a new contact repository
func NewContactRepository(db *sqlx.DB) ContactRepository {
	return &contactRepository{db: db}
}

// FindByID finds a contact by ID
func (r *contactRepository) FindByID(ctx context.Context, id string) (*domain.Contact, error) {
	// For now, use mock data
	// In production, query database
	contacts := r.getMockContacts()
	for _, c := range contacts {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}

// FindByUserID finds contacts by user ID with optional filters
func (r *contactRepository) FindByUserID(ctx context.Context, userID string, filter ContactFilter) ([]*domain.Contact, error) {
	// For now, return mock contacts
	// In production, query database with filters
	contacts := r.getMockContacts()

	// Apply filters
	result := []*domain.Contact{}
	for _, c := range contacts {
		if c.UserID != userID {
			continue
		}

		// Filter by type
		if filter.Type != "" && filter.Type != "all" && c.Type != filter.Type {
			continue
		}

		// Filter by search (name or value)
		if filter.Search != "" {
			// Simple contains check
			// In production, use proper SQL LIKE or full-text search
			continue // Skip search for mock
		}

		result = append(result, c)

		// Apply limit
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}

	return result, nil
}

// FindByUserAndID finds a contact by user ID and contact ID (ownership validation)
func (r *contactRepository) FindByUserAndID(ctx context.Context, userID, contactID string) (*domain.Contact, error) {
	contacts := r.getMockContacts()
	for _, c := range contacts {
		if c.ID == contactID && c.UserID == userID {
			return c, nil
		}
	}
	return nil, nil
}

// FindByUserAndValue finds a contact by user ID, type, and value (for duplicate check)
func (r *contactRepository) FindByUserAndValue(ctx context.Context, userID, contactType, value string) (*domain.Contact, error) {
	contacts := r.getMockContacts()
	for _, c := range contacts {
		if c.UserID == userID && c.Type == contactType && c.Value == value {
			return c, nil
		}
	}
	return nil, nil
}

// Create creates a new contact
func (r *contactRepository) Create(ctx context.Context, contact *domain.Contact) error {
	// In production, insert into database
	// For now, just return success
	return nil
}

// Update updates a contact
func (r *contactRepository) Update(ctx context.Context, contact *domain.Contact) error {
	// In production, update database
	// For now, just return success
	return nil
}

// Delete deletes a contact
func (r *contactRepository) Delete(ctx context.Context, id string) error {
	// In production, delete from database
	// For now, just return success
	return nil
}

// IncrementUsage increments usage count and updates last used time
func (r *contactRepository) IncrementUsage(ctx context.Context, id string) error {
	// In production, update database
	// For now, just return success
	return nil
}

// getMockContacts returns mock contacts for testing
func (r *contactRepository) getMockContacts() []*domain.Contact {
	now := time.Now()
	lastUsed1 := now.Add(-24 * time.Hour)
	lastUsed2 := now.Add(-72 * time.Hour)
	lastUsed3 := now.Add(-168 * time.Hour)
	lastUsed4 := now.Add(-240 * time.Hour)

	// Mock metadata
	phoneMetadata, _ := json.Marshal(map[string]string{
		"operator":     "telkomsel",
		"operatorName": "Telkomsel",
	})

	plnMetadata, _ := json.Marshal(map[string]string{
		"customerName": "BUDI SANTOSO",
	})

	bankMetadata, _ := json.Marshal(map[string]string{
		"bankCode":    "014",
		"bankName":    "BCA",
		"accountName": "BUDI SANTOSO",
	})

	phoneMetadataStr := string(phoneMetadata)
	plnMetadataStr := string(plnMetadata)
	bankMetadataStr := string(bankMetadata)

	return []*domain.Contact{
		{
			ID:         "cnt_abc123",
			UserID:     "mock_user_id", // Will be replaced with actual user ID
			Name:       "Ayah",
			Type:       domain.ContactTypePhone,
			Value:      "081234567890", // 12 digits - valid format
			Metadata:   &phoneMetadataStr,
			LastUsedAt: &lastUsed1,
			UsageCount: 5,
			CreatedAt:  now.AddDate(0, -1, 0),
			UpdatedAt:  now,
		},
		{
			ID:         "cnt_xyz789",
			UserID:     "mock_user_id",
			Name:       "Ibu",
			Type:       domain.ContactTypePhone,
			Value:      "081298765432", // 12 digits - valid format
			Metadata:   &phoneMetadataStr,
			LastUsedAt: &lastUsed2,
			UsageCount: 3,
			CreatedAt:  now.AddDate(0, -1, 0),
			UpdatedAt:  now,
		},
		{
			ID:         "cnt_pln123",
			UserID:     "mock_user_id",
			Name:       "Rumah",
			Type:       domain.ContactTypePLN,
			Value:      "123456789012", // 12 digits - valid PLN meter format
			Metadata:   &plnMetadataStr,
			LastUsedAt: &lastUsed3,
			UsageCount: 2,
			CreatedAt:  now.AddDate(0, 0, -15),
			UpdatedAt:  now,
		},
		{
			ID:         "cnt_bank123",
			UserID:     "mock_user_id",
			Name:       "Budi BCA",
			Type:       domain.ContactTypeBank,
			Value:      "1234567890", // 10 digits - valid bank account format
			Metadata:   &bankMetadataStr,
			LastUsedAt: &lastUsed4,
			UsageCount: 1,
			CreatedAt:  now.AddDate(0, 0, -10),
			UpdatedAt:  now,
		},
	}
}
