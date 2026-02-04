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
	"github.com/GTDGit/PPOB_BE/pkg/redis"
)

// ProductSyncJob handles product synchronization from Gerbang API
type ProductSyncJob struct {
	gerbangClient *gerbang.Client
	productRepo   repository.ProductRepository
	redisClient   *redis.Client
	logger        *slog.Logger
	interval      time.Duration
}

// NewProductSyncJob creates a new product sync job
func NewProductSyncJob(
	gerbangClient *gerbang.Client,
	productRepo repository.ProductRepository,
	redisClient *redis.Client,
	logger *slog.Logger,
	interval time.Duration,
) *ProductSyncJob {
	return &ProductSyncJob{
		gerbangClient: gerbangClient,
		productRepo:   productRepo,
		redisClient:   redisClient,
		logger:        logger,
		interval:      interval,
	}
}

// Start begins the sync job (call in main.go)
func (j *ProductSyncJob) Start(ctx context.Context) {
	j.logger.Info("product sync job started", "interval", j.interval.String())

	// Run immediately on startup
	if err := j.RunOnce(ctx); err != nil {
		j.logger.Error("initial product sync failed", "error", err)
	}

	// Then run every interval
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			j.logger.Info("product sync job stopped")
			return
		case <-ticker.C:
			if err := j.RunOnce(ctx); err != nil {
				j.logger.Error("product sync failed", "error", err)
			}
		}
	}
}

// RunOnce runs sync once (for manual trigger or initial sync)
func (j *ProductSyncJob) RunOnce(ctx context.Context) error {
	j.logger.Info("starting product sync")
	startTime := time.Now()

	// Fetch all products from GTD (paginated)
	var allProducts []gerbang.Product
	page := 1
	limit := 100
	totalFetched := 0

	for {
		// Get products from Gerbang API
		resp, pagination, err := j.gerbangClient.GetProducts(ctx, "", "", "", "", page, limit)
		if err != nil {
			return fmt.Errorf("failed to fetch products from GTD (page %d): %w", page, err)
		}

		if len(resp) == 0 {
			break
		}

		allProducts = append(allProducts, resp...)
		totalFetched += len(resp)

		j.logger.Debug("fetched products page",
			"page", page,
			"count", len(resp),
			"total", totalFetched,
		)

		// Check if there are more pages
		if pagination == nil || page >= pagination.TotalPages {
			break
		}

		page++
	}

	j.logger.Info("fetched all products from GTD", "count", len(allProducts))

	if len(allProducts) == 0 {
		j.logger.Warn("no products fetched from GTD")
		return nil
	}

	// Convert to domain products
	var domainProducts []*domain.Product
	for _, p := range allProducts {
		// Parse GTD updated_at timestamp
		gtdUpdatedAt := time.Now()
		if p.UpdatedAt != "" {
			if parsed, err := time.Parse(time.RFC3339, p.UpdatedAt); err == nil {
				gtdUpdatedAt = parsed
			}
		}

		domainProducts = append(domainProducts, &domain.Product{
			ID:           uuid.New().String(),
			SKUCode:      p.SKUCode,
			Name:         p.Name,
			Category:     p.Category,
			Brand:        p.Brand,
			Type:         p.Type,
			Price:        p.Price,
			Admin:        p.Admin,
			Commission:   p.Commission,
			IsActive:     p.IsActive,
			Description:  p.Description,
			GTDUpdatedAt: gtdUpdatedAt,
		})
	}

	// Bulk upsert to database
	if err := j.productRepo.BulkUpsert(ctx, domainProducts); err != nil {
		return fmt.Errorf("failed to upsert products: %w", err)
	}

	// Invalidate all product caches
	keys, err := j.redisClient.Keys(ctx, "products:*").Result()
	if err == nil && len(keys) > 0 {
		if err := j.redisClient.Del(ctx, keys...).Err(); err != nil {
			j.logger.Warn("failed to invalidate product cache", "error", err)
		} else {
			j.logger.Debug("invalidated product caches", "count", len(keys))
		}
	}

	duration := time.Since(startTime)
	j.logger.Info("product sync completed",
		"products_synced", len(domainProducts),
		"duration", duration.String(),
	)

	return nil
}

// Stop gracefully stops the sync job
func (j *ProductSyncJob) Stop() {
	j.logger.Info("stopping product sync job")
}
