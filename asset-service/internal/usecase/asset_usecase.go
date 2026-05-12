package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"gym-membership/asset-service/internal/domain"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

type assetUsecase struct {
	repo  domain.AssetRepository
	redis *redis.Client
	nats  *nats.Conn
}

func NewAssetUsecase(repo domain.AssetRepository, redis *redis.Client, nc *nats.Conn) domain.AssetUsecase {
	return &assetUsecase{
		repo:  repo,
		redis: redis,
		nats:  nc,
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
	if u.redis != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("availability:%s", id)
		u.redis.Del(ctx, cacheKey)
	}

	return asset, nil
}

func (u *assetUsecase) CheckAvailability(id string, startTime, endTime string) (bool, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("availability:%s", id)

	// Phase 2: Cache check result for 10 seconds
	if u.redis != nil {
		val, err := u.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			return val == "true", nil
		}
	}

	available, err := u.repo.CheckAvailability(id, startTime, endTime)
	if err != nil {
		return false, err
	}

	// Set cache with 10s TTL
	if u.redis != nil {
		u.redis.Set(ctx, cacheKey, fmt.Sprintf("%v", available), 10*time.Second)
	}

	return available, nil
}

// Phase 3: Event handlers

func (u *assetUsecase) HandleBookingCreated(assetID string) error {
	log.Printf("[NATS] Setting asset %s to in_use", assetID)
	_, err := u.UpdateAssetStatus(assetID, "in_use")
	return err
}

func (u *assetUsecase) HandleBookingReturned(assetID string, durationHours float64) error {
	log.Printf("[NATS] Asset %s returned after %.2f hours", assetID, durationHours)

	// Health logic: -5 for every 2+ hours
	healthDelta := 0
	if durationHours >= 2.0 {
		healthDelta = -5
	}

	asset, err := u.repo.UpdateHealth(assetID, healthDelta)
	if err != nil {
		return err
	}

	newStatus := "available"
	if asset.HealthScore < 30 {
		log.Printf("[NATS] Asset %s health dropped to %d, setting to maintenance", assetID, asset.HealthScore)
		newStatus = "maintenance"

		// Publish maintenance event
		if u.nats != nil {
			event := map[string]interface{}{
				"asset_id":     assetID,
				"health_score": asset.HealthScore,
				"timestamp":    time.Now().Format(time.RFC3339),
			}
			data, _ := json.Marshal(event)
			u.nats.Publish("asset.needs_maintenance", data)
		}
	}

	_, err = u.UpdateAssetStatus(assetID, newStatus)
	return err
}

func (u *assetUsecase) HandleBookingCancelled(assetID string) error {
	log.Printf("[NATS] Booking cancelled for asset %s, returning to available", assetID)
	_, err := u.UpdateAssetStatus(assetID, "available")
	return err
}
