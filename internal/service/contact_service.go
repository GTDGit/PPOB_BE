package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// ContactService handles contact business logic
type ContactService struct {
	contactRepo repository.ContactRepository
	productRepo repository.ProductRepository
}

// NewContactService creates a new contact service
func NewContactService(
	contactRepo repository.ContactRepository,
	productRepo repository.ProductRepository,
) *ContactService {
	return &ContactService{
		contactRepo: contactRepo,
		productRepo: productRepo,
	}
}

// List returns user's contacts with optional filters
func (s *ContactService) List(ctx context.Context, userID string, filter repository.ContactFilter) (*domain.ContactListResponse, error) {
	// Set default limit
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	contacts, err := s.contactRepo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	// Convert to response format
	details := make([]*domain.ContactDetail, 0, len(contacts))
	for _, c := range contacts {
		detail := s.toContactDetail(c)
		details = append(details, detail)
	}

	return &domain.ContactListResponse{
		Contacts:      details,
		TotalContacts: len(details),
	}, nil
}

// Create creates a new contact
func (s *ContactService) Create(ctx context.Context, userID string, name, contactType, value string, metadata map[string]interface{}) (*domain.ContactResponse, error) {
	// Validate contact type
	if !s.isValidContactType(contactType) {
		return nil, domain.ErrValidationFailed("Invalid contact type")
	}

	// Validate value based on type
	if err := s.validateContactValue(contactType, value); err != nil {
		return nil, err
	}

	// Check for duplicate
	existing, err := s.contactRepo.FindByUserAndValue(ctx, userID, contactType, value)
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrValidationFailed("Contact already exists")
	}

	// Enhance metadata for phone type (detect operator)
	if contactType == domain.ContactTypePhone {
		operator, err := s.productRepo.FindOperatorByPrefix(ctx, value)
		if err == nil && operator != nil {
			if metadata == nil {
				metadata = make(map[string]interface{})
			}
			metadata["operator"] = operator.ID
			metadata["operatorName"] = operator.Name
		}
	}

	// Marshal metadata to JSON
	var metadataJSON *string
	if len(metadata) > 0 {
		bytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		str := string(bytes)
		metadataJSON = &str
	}

	// Create contact
	now := time.Now()
	contact := &domain.Contact{
		ID:         "cnt_" + uuid.New().String()[:8],
		UserID:     userID,
		Name:       name,
		Type:       contactType,
		Value:      value,
		Metadata:   metadataJSON,
		UsageCount: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.contactRepo.Create(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	// Return response
	return &domain.ContactResponse{
		Contact: s.toContactDetail(contact),
	}, nil
}

// Update updates a contact's name
func (s *ContactService) Update(ctx context.Context, userID, contactID, name string) (*domain.ContactResponse, error) {
	// Get contact with ownership validation
	contact, err := s.contactRepo.FindByUserAndID(ctx, userID, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}
	if contact == nil {
		return nil, domain.ErrValidationFailed("Contact not found")
	}

	// Update name
	contact.Name = name
	contact.UpdatedAt = time.Now()

	if err := s.contactRepo.Update(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	// Return response
	return &domain.ContactResponse{
		Contact: s.toContactDetail(contact),
	}, nil
}

// Delete deletes a contact
func (s *ContactService) Delete(ctx context.Context, userID, contactID string) (*domain.DeleteContactResponse, error) {
	// Get contact with ownership validation
	contact, err := s.contactRepo.FindByUserAndID(ctx, userID, contactID)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}
	if contact == nil {
		return nil, domain.ErrValidationFailed("Contact not found")
	}

	// Delete contact
	if err := s.contactRepo.Delete(ctx, contactID); err != nil {
		return nil, fmt.Errorf("failed to delete contact: %w", err)
	}

	return &domain.DeleteContactResponse{
		Deleted:   true,
		ContactID: contactID,
	}, nil
}

// Helper functions

func (s *ContactService) toContactDetail(c *domain.Contact) *domain.ContactDetail {
	detail := &domain.ContactDetail{
		ID:         c.ID,
		Name:       c.Name,
		Type:       c.Type,
		Value:      c.Value,
		UsageCount: c.UsageCount,
		CreatedAt:  c.CreatedAt.Format(time.RFC3339),
	}

	// Add last used at
	if c.LastUsedAt != nil {
		lastUsed := c.LastUsedAt.Format(time.RFC3339)
		detail.LastUsedAt = &lastUsed
	}

	// Add updated at
	updatedAt := c.UpdatedAt.Format(time.RFC3339)
	detail.UpdatedAt = &updatedAt

	// Parse metadata
	if c.Metadata != nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(*c.Metadata), &metadata); err == nil {
			// Add operator info for phone contacts
			if c.Type == domain.ContactTypePhone {
				if operatorID, ok := metadata["operator"].(string); ok {
					if operatorName, ok := metadata["operatorName"].(string); ok {
						detail.Operator = &domain.OperatorInfo{
							ID:   operatorID,
							Name: operatorName,
						}
					}
				}
			}

			// Add bank info for bank contacts
			if c.Type == domain.ContactTypeBank {
				if bankCode, ok := metadata["bankCode"].(string); ok {
					if bankName, ok := metadata["bankName"].(string); ok {
						detail.Bank = &domain.BankInfo{
							Code: bankCode,
							Name: bankName,
						}
					}
				}
				if accountName, ok := metadata["accountName"].(string); ok {
					detail.AccountName = &accountName
				}
			}

			// Add customer name for PLN/PDAM/BPJS
			if c.Type == domain.ContactTypePLN || c.Type == domain.ContactTypePDAM || c.Type == domain.ContactTypeBPJS {
				if customerName, ok := metadata["customerName"].(string); ok {
					detail.CustomerName = &customerName
				}
			}
		}
	}

	return detail
}

func (s *ContactService) isValidContactType(contactType string) bool {
	validTypes := []string{
		domain.ContactTypePhone,
		domain.ContactTypePLN,
		domain.ContactTypePDAM,
		domain.ContactTypeBPJS,
		domain.ContactTypeTelkom,
		domain.ContactTypeBank,
	}
	for _, t := range validTypes {
		if t == contactType {
			return true
		}
	}
	return false
}

func (s *ContactService) validateContactValue(contactType, value string) error {
	if value == "" {
		return domain.ErrValidationFailed("Value is required")
	}

	switch contactType {
	case domain.ContactTypePhone:
		// Phone number: 10-13 digits, starts with 08
		if len(value) < 10 || len(value) > 13 {
			return domain.ErrValidationFailed("Invalid phone number length")
		}
		if value[:2] != "08" {
			return domain.ErrValidationFailed("Phone number must start with 08")
		}
		for _, c := range value {
			if c < '0' || c > '9' {
				return domain.ErrValidationFailed("Phone number must contain only digits")
			}
		}
	case domain.ContactTypePLN:
		// PLN meter: 11-12 digits
		if len(value) < 11 || len(value) > 12 {
			return domain.ErrValidationFailed("Invalid PLN meter number length")
		}
		for _, c := range value {
			if c < '0' || c > '9' {
				return domain.ErrValidationFailed("PLN meter must contain only digits")
			}
		}
	case domain.ContactTypeBPJS:
		// BPJS: 13 digits
		if len(value) != 13 {
			return domain.ErrValidationFailed("BPJS number must be 13 digits")
		}
		for _, c := range value {
			if c < '0' || c > '9' {
				return domain.ErrValidationFailed("BPJS number must contain only digits")
			}
		}
	case domain.ContactTypeBank:
		// Bank account: at least 6 digits
		if len(value) < 6 {
			return domain.ErrValidationFailed("Bank account number too short")
		}
	case domain.ContactTypePDAM, domain.ContactTypeTelkom:
		// General validation: at least 3 chars
		if len(value) < 3 {
			return domain.ErrValidationFailed("Value too short")
		}
	}

	return nil
}
