package domain

import "time"

type Asset struct {
	ID               string
	Name             string
	Type             string
	Status           string
	HealthScore      int
	Location         string
	CreatedAt        time.Time
	LastMaintainedAt time.Time
}

type AssetRepository interface {
	GetByID(id string) (*Asset, error)
	ListByType(assetType string) ([]*Asset, error)
	UpdateStatus(id string, status string) (*Asset, error)
	UpdateHealth(id string, healthDelta int) (*Asset, error)
	CheckAvailability(id string, startTime, endTime string) (bool, error)
}

type AssetUsecase interface {
	GetAsset(id string) (*Asset, error)
	ListAvailableAssets(assetType string) ([]*Asset, error)
	UpdateAssetStatus(id string, status string) (*Asset, error)
	CheckAvailability(id string, startTime, endTime string) (bool, error)
	
	// Phase 3: Event handlers
	HandleBookingCreated(assetID string) error
	HandleBookingReturned(assetID string, durationHours float64) error
	HandleBookingCancelled(assetID string) error
}
