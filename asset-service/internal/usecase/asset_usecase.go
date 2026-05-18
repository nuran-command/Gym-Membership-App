package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"gym-membership/asset-service/internal/domain"
	"gym-membership/asset-service/internal/observability"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

	// Phase 6: Structured logging
	if observability.Logger != nil {
		observability.Logger.Info("Asset status updated",
			zap.String("asset_id", id),
			zap.String("new_status", status),
			zap.Int("health_score", asset.HealthScore),
		)
	}

	// Phase 6: Update Prometheus metrics
	u.updateStatusMetrics()

	// Phase 2: Invalidate cache on status update
	if u.redis != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("availability:%s", id)
		u.redis.Del(ctx, cacheKey)
	}

	return asset, nil
}

func (u *assetUsecase) updateStatusMetrics() {
	assets, err := u.repo.ListAll("")
	if err != nil {
		return
	}

	counts := make(map[string]int)
	totalHealth := 0
	for _, a := range assets {
		counts[a.Status]++
		totalHealth += a.HealthScore
	}

	for status, count := range counts {
		observability.AssetsByStatus.WithLabelValues(status).Set(float64(count))
	}
	if len(assets) > 0 {
		observability.AvgHealthScore.Set(float64(totalHealth) / float64(len(assets)))
	}
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

// New Usecase Methods

func (u *assetUsecase) CreateAsset(name, assetType, status, location string, health int) (*domain.Asset, error) {
	if status == "" {
		status = "available"
	}
	if health <= 0 {
		health = 100
	}
	asset := &domain.Asset{
		Name:        name,
		Type:        assetType,
		Status:      status,
		HealthScore: health,
		Location:    location,
	}
	newAsset, err := u.repo.Create(asset)
	if err != nil {
		return nil, err
	}

	if observability.Logger != nil {
		observability.Logger.Info("Asset created",
			zap.String("asset_id", newAsset.ID),
			zap.String("name", name),
			zap.String("type", assetType),
		)
	}

	u.updateStatusMetrics()
	return newAsset, nil
}

func (u *assetUsecase) UpdateAsset(id, name, assetType, location string) (*domain.Asset, error) {
	asset := &domain.Asset{
		ID:       id,
		Name:     name,
		Type:     assetType,
		Location: location,
	}
	updatedAsset, err := u.repo.Update(asset)
	if err != nil {
		return nil, err
	}

	if observability.Logger != nil {
		observability.Logger.Info("Asset updated",
			zap.String("asset_id", id),
			zap.String("name", name),
			zap.String("type", assetType),
		)
	}

	// Invalidate cache
	if u.redis != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("availability:%s", id)
		u.redis.Del(ctx, cacheKey)
	}

	return updatedAsset, nil
}

func (u *assetUsecase) DeleteAsset(id string) error {
	err := u.repo.Delete(id)
	if err != nil {
		return err
	}

	if observability.Logger != nil {
		observability.Logger.Info("Asset deleted",
			zap.String("asset_id", id),
		)
	}

	u.updateStatusMetrics()

	// Invalidate cache
	if u.redis != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("availability:%s", id)
		u.redis.Del(ctx, cacheKey)
	}

	return nil
}

func (u *assetUsecase) ReportDamage(id string, damageAmount int) (*domain.Asset, error) {
	asset, err := u.repo.UpdateHealth(id, -damageAmount)
	if err != nil {
		return nil, err
	}

	if observability.Logger != nil {
		observability.Logger.Info("Asset damage reported",
			zap.String("asset_id", id),
			zap.Int("damage_amount", damageAmount),
			zap.Int("new_health", asset.HealthScore),
		)
	}

	// If health < 30, trigger maintenance
	if asset.HealthScore < 30 && asset.Status != "maintenance" {
		asset, err = u.UpdateAssetStatus(id, "maintenance")
		if err != nil {
			return nil, err
		}

		// Publish maintenance event
		if u.nats != nil {
			event := map[string]interface{}{
				"asset_id":     id,
				"health_score": asset.HealthScore,
				"timestamp":    time.Now().Format(time.RFC3339),
			}
			data, _ := json.Marshal(event)
			u.nats.Publish("asset.needs_maintenance", data)
		}
	}

	return asset, nil
}

func (u *assetUsecase) ResolveMaintenance(id string) (*domain.Asset, error) {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	delta := 100 - existing.HealthScore
	asset, err := u.repo.UpdateHealth(id, delta)
	if err != nil {
		return nil, err
	}

	asset, err = u.UpdateAssetStatus(id, "available")
	if err != nil {
		return nil, err
	}

	if observability.Logger != nil {
		observability.Logger.Info("Asset maintenance resolved",
			zap.String("asset_id", id),
		)
	}

	return asset, nil
}

func (u *assetUsecase) ListAllAssets(assetType string) ([]*domain.Asset, error) {
	return u.repo.ListAll(assetType)
}

func (u *assetUsecase) BatchCreateAssets(assets []*domain.Asset) ([]*domain.Asset, error) {
	var created []*domain.Asset
	for _, a := range assets {
		if a.Status == "" {
			a.Status = "available"
		}
		if a.HealthScore <= 0 {
			a.HealthScore = 100
		}
		newAsset, err := u.repo.Create(a)
		if err != nil {
			return nil, err
		}
		created = append(created, newAsset)
	}

	if observability.Logger != nil {
		observability.Logger.Info("Batch assets created",
			zap.Int("count", len(created)),
		)
	}

	u.updateStatusMetrics()
	return created, nil
}
