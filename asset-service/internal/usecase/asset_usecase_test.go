package usecase

import (
	"gym-membership/asset-service/internal/domain"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAssetRepository struct {
	mock.Mock
}

func (m *MockAssetRepository) GetByID(id string) (*domain.Asset, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Asset), args.Error(1)
}

func (m *MockAssetRepository) ListByType(assetType string) ([]*domain.Asset, error) {
	args := m.Called(assetType)
	return args.Get(0).([]*domain.Asset), args.Error(1)
}

func (m *MockAssetRepository) UpdateStatus(id string, status string) (*domain.Asset, error) {
	args := m.Called(id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Asset), args.Error(1)
}

func (m *MockAssetRepository) UpdateHealth(id string, healthDelta int) (*domain.Asset, error) {
	args := m.Called(id, healthDelta)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Asset), args.Error(1)
}

func (m *MockAssetRepository) CheckAvailability(id string, startTime, endTime string) (bool, error) {
	args := m.Called(id, startTime, endTime)
	return args.Bool(0), args.Error(1)
}

func TestCheckAvailability_AssetInUse(t *testing.T) {
	mockRepo := new(MockAssetRepository)
	uc := &assetUsecase{repo: mockRepo}

	mockRepo.On("CheckAvailability", "asset-1", "", "").Return(false, nil)

	available, err := uc.CheckAvailability("asset-1", "", "")
	assert.NoError(t, err)
	assert.False(t, available)
	mockRepo.AssertExpectations(t)
}

func TestUpdateHealthScore_TriggersMaintenance(t *testing.T) {
	mockRepo := new(MockAssetRepository)
	uc := &assetUsecase{repo: mockRepo}

	asset := &domain.Asset{ID: "asset-1", HealthScore: 25}
	mockRepo.On("UpdateHealth", "asset-1", -5).Return(asset, nil)
	mockRepo.On("UpdateStatus", "asset-1", "maintenance").Return(asset, nil)

	err := uc.HandleBookingReturned("asset-1", 3.0)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCheckAvailability_RaceCondition(t *testing.T) {
	mockRepo := new(MockAssetRepository)
	uc := &assetUsecase{repo: mockRepo}

	mockRepo.On("CheckAvailability", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = uc.CheckAvailability("asset-1", "", "")
		}()
	}
	wg.Wait()
}
