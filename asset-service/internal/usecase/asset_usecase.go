package usecase

import (
	"context"
	"fmt"
	"gym-membership/asset-service/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type assetUsecase struct {
	repo  domain.AssetRepository
	redis *redis.Client
}

func NewAssetUsecase(repo domain.AssetRepository, redis *redis.Client) domain.AssetUsecase {
	return &assetUsecase{
		repo:  repo,
		redis: redis,
	}
}

func (u *assetUsecase) GetAsset(id string) (*domain.Asset, error) {
	return u.repo.GetByID(id)
}

func (u *assetUsecase) ListAvailableAssets(assetType string) ([]*domain.Asset, error) {
	return u.repo.ListByType(assetType)
}

func (u *assetUsecase) UpdateAssetStatus(id string, status string) (*domain.Asset, error) {
	asset, err := u.repo.UpdateStatus(id, status)
	if err != nil {
		return nil, err
	}

	// Phase 2: Invalidate cache on status update
	ctx := context.Background()
	cacheKey := fmt.Sprintf("availability:%s", id)
	u.redis.Del(ctx, cacheKey)

	return asset, nil
}

func (u *assetUsecase) CheckAvailability(id string, startTime, endTime string) (bool, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("availability:%s", id)

	// Phase 2: Cache check result for 10 seconds
	val, err := u.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		return val == "true", nil
	}

	available, err := u.repo.CheckAvailability(id, startTime, endTime)
	if err != nil {
		return false, err
	}

	// Set cache with 10s TTL
	u.redis.Set(ctx, cacheKey, fmt.Sprintf("%v", available), 10*time.Second)

	return available, nil
}
