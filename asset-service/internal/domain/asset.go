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
	
	// New methods
	Create(asset *Asset) (*Asset, error)
	Update(asset *Asset) (*Asset, error)
	Delete(id string) error
	ListAll(assetType string) ([]*Asset, error)
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

	// New usecase methods
	CreateAsset(name, assetType, status, location string, health int) (*Asset, error)
	UpdateAsset(id, name, assetType, location string) (*Asset, error)
	DeleteAsset(id string) error
	ReportDamage(id string, damageAmount int) (*Asset, error)
	ResolveMaintenance(id string) (*Asset, error)
	ListAllAssets(assetType string) ([]*Asset, error)
	BatchCreateAssets(assets []*Asset) ([]*Asset, error)
}
