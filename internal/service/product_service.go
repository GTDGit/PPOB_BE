package service

import (
	"context"
	"fmt"
	"time"

	"github.com/GTDGit/PPOB_BE/internal/domain"
	"github.com/GTDGit/PPOB_BE/internal/repository"
	"github.com/GTDGit/PPOB_BE/pkg/redis"
)

const productCacheTTL = 5 * time.Minute

// ProductService handles product data business logic
type ProductService struct {
	productRepo repository.ProductRepository
	redisClient *redis.Client
}

// NewProductService creates a new product service
func NewProductService(productRepo repository.ProductRepository, redisClient *redis.Client) *ProductService {
	return &ProductService{
		productRepo: productRepo,
		redisClient: redisClient,
	}
}

// GetOperators returns all mobile operators
func (s *ProductService) GetOperators(ctx context.Context) ([]*domain.Operator, error) {
	// 1. Try cache first
	cacheKey := redis.ProductListKey("operators")
	var operators []*domain.Operator
	if err := s.redisClient.GetJSON(ctx, cacheKey, &operators); err == nil && len(operators) > 0 {
		return operators, nil // Cache hit
	}

	// 2. Fetch from DB
	operators, err := s.productRepo.FindAllOperators(ctx)
	if err != nil {
		return nil, err
	}

	// 3. Cache with TTL
	if len(operators) > 0 {
		s.redisClient.SetJSON(ctx, cacheKey, operators, productCacheTTL)
	}

	return operators, nil
}

// GetEwalletProviders returns all e-wallet providers
func (s *ProductService) GetEwalletProviders(ctx context.Context) ([]*domain.EwalletProvider, error) {
	// 1. Try cache first
	cacheKey := redis.ProductListKey("ewallet")
	var providers []*domain.EwalletProvider
	if err := s.redisClient.GetJSON(ctx, cacheKey, &providers); err == nil && len(providers) > 0 {
		return providers, nil // Cache hit
	}

	// 2. Fetch from DB
	providers, err := s.productRepo.FindAllEwalletProviders(ctx)
	if err != nil {
		return nil, err
	}

	// 3. Cache with TTL
	if len(providers) > 0 {
		s.redisClient.SetJSON(ctx, cacheKey, providers, productCacheTTL)
	}

	return providers, nil
}

// GetPDAMRegions returns all PDAM regions
func (s *ProductService) GetPDAMRegions(ctx context.Context) ([]*domain.PDAMRegion, error) {
	// 1. Try cache first
	cacheKey := redis.ProductListKey("pdam")
	var regions []*domain.PDAMRegion
	if err := s.redisClient.GetJSON(ctx, cacheKey, &regions); err == nil && len(regions) > 0 {
		return regions, nil // Cache hit
	}

	// 2. Fetch from DB
	regions, err := s.productRepo.FindAllPDAMRegions(ctx)
	if err != nil {
		return nil, err
	}

	// 3. Cache with TTL
	if len(regions) > 0 {
		s.redisClient.SetJSON(ctx, cacheKey, regions, productCacheTTL)
	}

	return regions, nil
}

// GetBanks returns all banks with optional filtering
func (s *ProductService) GetBanks(ctx context.Context, filterType string) ([]*domain.Bank, error) {
	// 1. Try cache first
	cacheKey := redis.ProductListKey(fmt.Sprintf("banks:%s", filterType))
	var banks []*domain.Bank
	if err := s.redisClient.GetJSON(ctx, cacheKey, &banks); err == nil && len(banks) > 0 {
		return banks, nil // Cache hit
	}

	// 2. Fetch from DB
	banks, err := s.productRepo.FindAllBanks(ctx, filterType)
	if err != nil {
		return nil, err
	}

	// 3. Cache with TTL
	if len(banks) > 0 {
		s.redisClient.SetJSON(ctx, cacheKey, banks, productCacheTTL)
	}

	return banks, nil
}

// GetTVProviders returns all TV cable providers
func (s *ProductService) GetTVProviders(ctx context.Context) ([]*domain.TVProvider, error) {
	// 1. Try cache first
	cacheKey := redis.ProductListKey("tv")
	var providers []*domain.TVProvider
	if err := s.redisClient.GetJSON(ctx, cacheKey, &providers); err == nil && len(providers) > 0 {
		return providers, nil // Cache hit
	}

	// 2. Fetch from DB
	providers, err := s.productRepo.FindAllTVProviders(ctx)
	if err != nil {
		return nil, err
	}

	// 3. Cache with TTL
	if len(providers) > 0 {
		s.redisClient.SetJSON(ctx, cacheKey, providers, productCacheTTL)
	}

	return providers, nil
}
