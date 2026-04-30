package usecase

import (
	"gym-membership/asset-service/internal/domain"
)

type assetUsecase struct {
	repo domain.AssetRepository
}

func NewAssetUsecase(repo domain.AssetRepository) domain.AssetUsecase {
	return &assetUsecase{
		repo: repo,
	}
}

func (u *assetUsecase) GetAsset(id string) (*domain.Asset, error) {
	return u.repo.GetByID(id)
}

func (u *assetUsecase) ListAvailableAssets(assetType string) ([]*domain.Asset, error) {
	return u.repo.ListByType(assetType)
}

func (u *assetUsecase) UpdateAssetStatus(id string, status string) (*domain.Asset, error) {
	return u.repo.UpdateStatus(id, status)
}

func (u *assetUsecase) CheckAvailability(id string, startTime, endTime string) (bool, error) {
	return u.repo.CheckAvailability(id, startTime, endTime)
}
