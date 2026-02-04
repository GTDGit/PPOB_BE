package job

import (
	"context"
	"log/slog"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
)

// TerritorySyncJob handles territory data synchronization from Gerbang API
type TerritorySyncJob struct {
	repo     repository.TerritoryRepository
	logger   *slog.Logger
	interval time.Duration
	enabled  bool
}

// NewTerritorySyncJob creates a new territory sync job
func NewTerritorySyncJob(
	repo repository.TerritoryRepository,
	logger *slog.Logger,
	interval time.Duration,
	enabled bool,
) *TerritorySyncJob {
	return &TerritorySyncJob{
		repo:     repo,
		logger:   logger,
		interval: interval,
		enabled:  enabled,
	}
}

// Start begins the sync job
func (j *TerritorySyncJob) Start(ctx context.Context) {
	if !j.enabled {
		j.logger.Info("territory sync job disabled")
		return
	}

	j.logger.Info("territory sync job started", "interval", j.interval.String())

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			j.logger.Info("territory sync job stopped")
			return
		case <-ticker.C:
			if err := j.RunOnce(ctx); err != nil {
				j.logger.Error("territory sync failed", slog.String("error", err.Error()))
			}
		}
	}
}

// RunOnce runs sync once
func (j *TerritorySyncJob) RunOnce(ctx context.Context) error {
	j.logger.Info("starting territory sync")
	startTime := time.Now()

	// Log sync start
	syncLog := &domain.TerritorySyncLog{
		SyncType:     domain.TerritorySyncTypeProvinces,
		TotalRecords: 0,
		Status:       domain.TerritorySyncStatusRunning,
		StartedAt:    startTime,
	}

	// TODO: Implement actual sync with Gerbang API
	// For now, just log that sync would run here

	// Log sync completion
	completedAt := time.Now()
	syncLog.CompletedAt = &completedAt
	syncLog.Status = domain.TerritorySyncStatusSuccess

	if err := j.repo.LogSync(ctx, syncLog); err != nil {
		j.logger.Warn("failed to log sync", slog.String("error", err.Error()))
	}

	duration := time.Since(startTime)
	j.logger.Info("territory sync completed",
		"duration", duration.String(),
	)

	return nil
}

// Stop gracefully stops the sync job
func (j *TerritorySyncJob) Stop() {
	j.logger.Info("stopping territory sync job")
}
