package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// NotificationService handles notification business logic
type NotificationService struct {
	notificationRepo repository.NotificationRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(notificationRepo repository.NotificationRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

// List returns notification list with pagination
func (s *NotificationService) List(ctx context.Context, userID string, filter repository.NotificationFilter) (*domain.NotificationListResponse, error) {
	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}
	if filter.PerPage > 50 {
		filter.PerPage = 50
	}

	// Get notifications
	notifications, total, err := s.notificationRepo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	// Convert to summary
	summaries := make([]*domain.NotificationSummary, 0, len(notifications))
	for _, notif := range notifications {
		summary := s.toNotificationSummary(notif)
		summaries = append(summaries, summary)
	}

	// Build pagination
	totalPages := int(math.Ceil(float64(total) / float64(filter.PerPage)))
	pagination := &domain.Pagination{
		CurrentPage:  filter.Page,
		TotalPages:   totalPages,
		TotalItems:   total,
		ItemsPerPage: filter.PerPage,
		HasNextPage:  filter.Page < totalPages,
		HasPrevPage:  filter.Page > 1,
	}

	response := &domain.NotificationListResponse{
		Notifications: summaries,
		Pagination:    pagination,
	}

	// Add empty state if no notifications
	if len(summaries) == 0 {
		emptyImageURL := "https://cdn.ppob.id/empty-states/no-notifications.png"
		response.EmptyState = &domain.EmptyState{
			Title:    "Tidak Ada Notifikasi Terbaru",
			Message:  "Transaksi yang masih diproses, perlu dibayar atau membutuhkan konfirmasi operator akan muncul disini",
			ImageURL: &emptyImageURL,
		}
	}

	return response, nil
}

// GetDetail returns detailed notification information and marks as read
func (s *NotificationService) GetDetail(ctx context.Context, userID, notificationID string) (*domain.NotificationDetailResponse, error) {
	// Get notification with ownership validation
	notif, err := s.notificationRepo.FindByUserAndID(ctx, userID, notificationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	if notif == nil {
		return nil, domain.ErrValidationFailed("Notification not found")
	}

	// Auto mark as read if not already read
	if !notif.IsRead {
		if err := s.notificationRepo.MarkAsRead(ctx, notificationID); err != nil {
			// Log error but don't fail the request
			slog.Error("failed to mark notification as read",
				slog.String("notification_id", notificationID),
				slog.String("error", err.Error()),
			)
		}
		// Update local state
		now := time.Now()
		notif.IsRead = true
		notif.ReadAt = &now
	}

	// Build detailed response
	detail := s.toNotificationDetail(notif)

	return &domain.NotificationDetailResponse{
		Notification: detail,
	}, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID string) (*domain.MarkAsReadResponse, error) {
	// Validate ownership
	notif, err := s.notificationRepo.FindByUserAndID(ctx, userID, notificationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	if notif == nil {
		return nil, domain.ErrValidationFailed("Notification not found")
	}

	// Mark as read
	if err := s.notificationRepo.MarkAsRead(ctx, notificationID); err != nil {
		return nil, fmt.Errorf("failed to mark as read: %w", err)
	}

	now := time.Now()
	return &domain.MarkAsReadResponse{
		NotificationID: notificationID,
		IsRead:         true,
		ReadAt:         now.Format(time.RFC3339),
	}, nil
}

// MarkAllAsRead marks all notifications as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID string, category *string) (*domain.MarkAllAsReadResponse, error) {
	// Mark all as read and get count
	count, err := s.notificationRepo.MarkAllAsRead(ctx, userID, category)
	if err != nil {
		return nil, fmt.Errorf("failed to mark all as read: %w", err)
	}

	message := fmt.Sprintf("%d notifikasi ditandai sudah dibaca", count)
	if count == 0 {
		message = "Tidak ada notifikasi yang perlu ditandai"
	}

	return &domain.MarkAllAsReadResponse{
		MarkedCount: count,
		Message:     message,
	}, nil
}

// GetUnreadCount returns unread notification count by category
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (*domain.UnreadCountResponse, error) {
	// Get total unread count
	total, err := s.notificationRepo.CountUnread(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count unread: %w", err)
	}

	// Get unread count by category
	byCategory, err := s.notificationRepo.CountUnreadByCategory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count by category: %w", err)
	}

	return &domain.UnreadCountResponse{
		UnreadCount: total,
		ByCategory:  byCategory,
	}, nil
}

// Delete deletes a notification
func (s *NotificationService) Delete(ctx context.Context, userID, notificationID string) (*domain.DeleteNotificationResponse, error) {
	// Validate ownership
	notif, err := s.notificationRepo.FindByUserAndID(ctx, userID, notificationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	if notif == nil {
		return nil, domain.ErrValidationFailed("Notification not found")
	}

	// Delete
	if err := s.notificationRepo.Delete(ctx, notificationID); err != nil {
		return nil, fmt.Errorf("failed to delete notification: %w", err)
	}

	return &domain.DeleteNotificationResponse{
		Deleted:        true,
		NotificationID: notificationID,
	}, nil
}

// Helper functions

func (s *NotificationService) toNotificationSummary(notif *domain.Notification) *domain.NotificationSummary {
	shortBody := truncateBody(notif.Body, 50)
	formattedDate := formatNotificationDate(notif.CreatedAt)

	summary := &domain.NotificationSummary{
		ID:                 notif.ID,
		Category:           notif.Category,
		Title:              notif.Title,
		Body:               notif.Body,
		ShortBody:          shortBody,
		ImageURL:           notif.ImageURL,
		IsRead:             notif.IsRead,
		CreatedAt:          notif.CreatedAt.Format(time.RFC3339),
		CreatedAtFormatted: formattedDate,
	}

	// Add action if present
	if notif.ActionType != nil && notif.ActionValue != nil {
		summary.Action = &domain.NotificationAction{
			Type:  *notif.ActionType,
			Value: *notif.ActionValue,
		}
	}

	return summary
}

func (s *NotificationService) toNotificationDetail(notif *domain.Notification) *domain.NotificationDetail {
	formattedDate := formatNotificationDate(notif.CreatedAt)

	detail := &domain.NotificationDetail{
		ID:                 notif.ID,
		Category:           notif.Category,
		Title:              notif.Title,
		Body:               notif.Body,
		ImageURL:           notif.ImageURL,
		IsRead:             notif.IsRead,
		CreatedAt:          notif.CreatedAt.Format(time.RFC3339),
		CreatedAtFormatted: formattedDate,
	}

	// Add action if present
	if notif.ActionType != nil && notif.ActionValue != nil {
		buttonText := getButtonText(notif.Category)
		detail.Action = &domain.NotificationAction{
			Type:       *notif.ActionType,
			Value:      *notif.ActionValue,
			ButtonText: &buttonText,
		}
	}

	// Parse metadata if present
	if notif.Metadata != nil {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(*notif.Metadata), &metadata); err == nil {
			detail.Metadata = metadata
		}
	}

	return detail
}

func truncateBody(body string, maxLen int) string {
	if len(body) <= maxLen {
		return body
	}
	return body[:maxLen] + "..."
}

func formatNotificationDate(t time.Time) string {
	// Format: "Senin, 11 Desember 2025, 8:31"
	// For simplicity, using English format
	// In production, use proper Indonesian localization
	return t.Format("Monday, 2 January 2006, 15:04")
}

func getButtonText(category string) string {
	buttonTexts := map[string]string{
		domain.NotificationCategorySecurity:    "Lihat Aktivitas Login",
		domain.NotificationCategoryTransaction: "Lihat Detail Transaksi",
		domain.NotificationCategoryDeposit:     "Lihat Detail Deposit",
		domain.NotificationCategoryPromo:       "Lihat Promo",
		domain.NotificationCategoryInfo:        "Lihat Detail",
		domain.NotificationCategoryQRIS:        "Lihat Detail Pembayaran",
	}
	if text, ok := buttonTexts[category]; ok {
		return text
	}
	return "Lihat Detail"
}
