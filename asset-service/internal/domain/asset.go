package domain

type Asset struct {
	ID     string
	Name   string
	Type   string
	Status string
}

type AssetRepository interface {
	GetByID(id string) (*Asset, error)
	ListByType(assetType string) ([]*Asset, error)
	UpdateStatus(id string, status string) (*Asset, error)
	CheckAvailability(id string, startTime, endTime string) (bool, error)
}

type AssetUsecase interface {
	GetAsset(id string) (*Asset, error)
	ListAvailableAssets(assetType string) ([]*Asset, error)
	UpdateAssetStatus(id string, status string) (*Asset, error)
	CheckAvailability(id string, startTime, endTime string) (bool, error)
}
