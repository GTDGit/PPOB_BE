package job

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/external/gerbang"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// BankSyncJob handles bank synchronization from Gerbang API
type BankSyncJob struct {
	gerbangClient *gerbang.Client
	productRepo   repository.ProductRepository
	logger        *slog.Logger
	interval      time.Duration
	enableOnStart bool
}

// NewBankSyncJob creates a new bank sync job
func NewBankSyncJob(
	gerbangClient *gerbang.Client,
	productRepo repository.ProductRepository,
	logger *slog.Logger,
	interval time.Duration,
	enableOnStart bool,
) *BankSyncJob {
	return &BankSyncJob{
		gerbangClient: gerbangClient,
		productRepo:   productRepo,
		logger:        logger,
		interval:      interval,
		enableOnStart: enableOnStart,
	}
}

// Start begins the sync job (call in main.go)
func (j *BankSyncJob) Start(ctx context.Context) {
	j.logger.Info("bank sync job started", "interval", j.interval.String())

	// Run immediately on startup if enabled
	if j.enableOnStart {
		if err := j.RunOnce(ctx); err != nil {
			j.logger.Error("initial bank sync failed", "error", err)
		}
	}

	// Then run every interval (3 days = 72 hours)
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			j.logger.Info("bank sync job stopped")
			return
		case <-ticker.C:
			if err := j.RunOnce(ctx); err != nil {
				j.logger.Error("bank sync failed", "error", err)
			}
		}
	}
}

// RunOnce runs sync once (for manual trigger or initial sync)
func (j *BankSyncJob) RunOnce(ctx context.Context) error {
	j.logger.Info("starting bank sync")
	startTime := time.Now()

	// Fetch bank codes from Gerbang API
	bankCodeItems, err := j.gerbangClient.GetBankCodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch bank codes from Gerbang API: %w", err)
	}

	if len(bankCodeItems) == 0 {
		j.logger.Warn("no bank codes returned from Gerbang API")
		return nil
	}

	// Convert Gerbang bank code items to domain.Bank
	banks := make([]*domain.Bank, len(bankCodeItems))
	for i, item := range bankCodeItems {
		banks[i] = &domain.Bank{
			ID:                   uuid.New().String(),
			Code:                 item.Code,
			Name:                 item.Name,
			ShortName:            item.ShortName,
			SwiftCode:            item.SwiftCode,
			Icon:                 "", // Will be populated later if needed
			IconURL:              "", // Will be populated later if needed
			TransferFee:          6500, // Default transfer fee
			TransferFeeFormatted: "Rp6.500",
			IsPopular:            false, // Can be updated manually or via additional logic
			Status:               domain.StatusActive,
		}
	}

	// Upsert banks to database via ProductRepository
	if err := j.productRepo.UpsertBanks(ctx, banks); err != nil {
		return fmt.Errorf("failed to upsert banks: %w", err)
	}

	duration := time.Since(startTime)
	j.logger.Info("bank sync completed",
		slog.Int("count", len(banks)),
		slog.Duration("duration", duration))

	return nil
}
