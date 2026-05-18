package repository

import (
	"errors"
	"fmt"
	"gym-membership/asset-service/internal/domain"
	"sync"
)

type inMemoryAssetRepo struct {
	assets map[string]*domain.Asset
	mu     sync.RWMutex
}

func NewInMemoryAssetRepository() domain.AssetRepository {
	return &inMemoryAssetRepo{
		assets: make(map[string]*domain.Asset),
	}
}

func (r *inMemoryAssetRepo) GetByID(id string) (*domain.Asset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	asset, ok := r.assets[id]
	if !ok {
		return nil, errors.New("asset not found")
	}
	return asset, nil
}

func (r *inMemoryAssetRepo) ListByType(assetType string) ([]*domain.Asset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Asset
	for _, a := range r.assets {
		if (a.Type == assetType || assetType == "") && a.Status == "available" {
			result = append(result, a)
		}
	}
	return result, nil
}

func (r *inMemoryAssetRepo) UpdateStatus(id string, status string) (*domain.Asset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	asset, ok := r.assets[id]
	if !ok {
		return nil, errors.New("asset not found")
	}
	asset.Status = status
	return asset, nil
}

func (r *inMemoryAssetRepo) CheckAvailability(id string, startTime, endTime string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	asset, ok := r.assets[id]
	if !ok {
		return false, errors.New("asset not found")
	}
	return asset.Status == "available", nil
}

func (r *inMemoryAssetRepo) UpdateHealth(id string, healthDelta int) (*domain.Asset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	asset, ok := r.assets[id]
	if !ok {
		return nil, errors.New("asset not found")
	}
	asset.HealthScore += healthDelta
	if asset.HealthScore < 0 {
		asset.HealthScore = 0
	}
	if asset.HealthScore > 100 {
		asset.HealthScore = 100
	}
	return asset, nil
}

func (r *inMemoryAssetRepo) Create(asset *domain.Asset) (*domain.Asset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if asset.ID == "" {
		asset.ID = fmt.Sprintf("in-memory-uuid-%d", len(r.assets)+1)
	}
	r.assets[asset.ID] = asset
	return asset, nil
}

func (r *inMemoryAssetRepo) Update(asset *domain.Asset) (*domain.Asset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.assets[asset.ID]
	if !ok {
		return nil, errors.New("asset not found")
	}
	existing.Name = asset.Name
	existing.Type = asset.Type
	existing.Location = asset.Location
	return existing, nil
}

func (r *inMemoryAssetRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.assets[id]
	if !ok {
		return errors.New("asset not found")
	}
	delete(r.assets, id)
	return nil
}

func (r *inMemoryAssetRepo) ListAll(assetType string) ([]*domain.Asset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Asset
	for _, a := range r.assets {
		if a.Type == assetType || assetType == "" {
			result = append(result, a)
		}
	}
	return result, nil
}
