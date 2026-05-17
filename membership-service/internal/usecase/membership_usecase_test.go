package usecase

import (
	"context"
	"testing"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/ilnur/gym-membership-app/proto/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	googlegrpc "google.golang.org/grpc"
)

// --- Mocks ---

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepo) UpdateCredits(ctx context.Context, userID string, amount int) error {
	args := m.Called(ctx, userID, amount)
	return args.Error(0)
}

type MockBookingRepo struct {
	mock.Mock
}

func (m *MockBookingRepo) Create(ctx context.Context, booking *domain.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockBookingRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Booking, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Booking), args.Error(1)
}

func (m *MockBookingRepo) UpdateStatus(ctx context.Context, bookingID string, status string) error {
	args := m.Called(ctx, bookingID, status)
	return args.Error(0)
}

func (m *MockBookingRepo) GetByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	args := m.Called(ctx, bookingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

type MockCreditRepo struct {
	mock.Mock
}

func (m *MockCreditRepo) CreateTransaction(ctx context.Context, tx *domain.CreditTransaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

type MockTxManager struct {
	mock.Mock
}

func (m *MockTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type MockMessagePublisher struct {
	mock.Mock
}

func (m *MockMessagePublisher) PublishBookingCreated(ctx context.Context, bookingID, userID, assetID string, startTime, endTime string) error {
	args := m.Called(ctx, bookingID, userID, assetID, startTime, endTime)
	return args.Error(0)
}

func (m *MockMessagePublisher) PublishBookingCancelled(ctx context.Context, bookingID string) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}

func (m *MockMessagePublisher) PublishBookingReturned(ctx context.Context, bookingID string) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}

type MockAssetClient struct {
	mock.Mock
}

func (m *MockAssetClient) GetAsset(ctx context.Context, in *asset.GetAssetRequest, opts ...googlegrpc.CallOption) (*asset.Asset, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*asset.Asset), args.Error(1)
}

func (m *MockAssetClient) ListAvailableAssets(ctx context.Context, in *asset.ListRequest, opts ...googlegrpc.CallOption) (*asset.AssetList, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*asset.AssetList), args.Error(1)
}

func (m *MockAssetClient) UpdateAssetStatus(ctx context.Context, in *asset.UpdateStatusRequest, opts ...googlegrpc.CallOption) (*asset.Asset, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*asset.Asset), args.Error(1)
}

func (m *MockAssetClient) CheckAvailability(ctx context.Context, in *asset.CheckRequest, opts ...googlegrpc.CallOption) (*asset.AvailabilityResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.AvailabilityResponse), args.Error(1)
}

func (m *MockAssetClient) GetHealthScore(ctx context.Context, in *asset.GetAssetRequest, opts ...googlegrpc.CallOption) (*asset.HealthResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.HealthResponse), args.Error(1)
}

// --- Test Cases ---

func TestCreateBooking_InsufficientCredits(t *testing.T) {
	ctx := context.Background()

	userRepo := new(MockUserRepo)
	bookingRepo := new(MockBookingRepo)
	creditRepo := new(MockCreditRepo)
	txManager := new(MockTxManager)
	publisher := new(MockMessagePublisher)
	assetClient := new(MockAssetClient)

	useCase := NewMembershipUseCase(userRepo, bookingRepo, creditRepo, txManager, publisher, assetClient)

	userID := "user-123"
	assetID := "asset-456"
	startTime := "2026-05-18T10:00:00Z"
	endTime := "2026-05-18T11:00:00Z"

	// Mock Asset check: Available
	assetClient.On("CheckAvailability", ctx, &asset.CheckRequest{
		Id:        assetID,
		StartTime: startTime,
		EndTime:   endTime,
	}).Return(&asset.AvailabilityResponse{Available: true}, nil)

	// Mock User: has only 5 credits
	userRepo.On("GetByID", ctx, userID).Return(&domain.User{
		ID:      userID,
		Credits: 5,
	}, nil)

	booking, err := useCase.CreateBooking(ctx, userID, assetID, startTime, endTime)

	assert.Error(t, err)
	assert.Nil(t, booking)
	assert.Equal(t, "insufficient credits", err.Error())

	assetClient.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestCreateBooking_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := new(MockUserRepo)
	bookingRepo := new(MockBookingRepo)
	creditRepo := new(MockCreditRepo)
	txManager := new(MockTxManager)
	publisher := new(MockMessagePublisher)
	assetClient := new(MockAssetClient)

	useCase := NewMembershipUseCase(userRepo, bookingRepo, creditRepo, txManager, publisher, assetClient)

	userID := "user-123"
	assetID := "asset-456"
	startTime := "2026-05-18T10:00:00Z"
	endTime := "2026-05-18T11:00:00Z"

	// Mock Asset: Available
	assetClient.On("CheckAvailability", ctx, &asset.CheckRequest{
		Id:        assetID,
		StartTime: startTime,
		EndTime:   endTime,
	}).Return(&asset.AvailabilityResponse{Available: true}, nil)

	// Mock User: has 20 credits
	userRepo.On("GetByID", ctx, userID).Return(&domain.User{
		ID:      userID,
		Credits: 20,
	}, nil)

	// Transactional mocks
	userRepo.On("UpdateCredits", ctx, userID, -10).Return(nil)
	creditRepo.On("CreateTransaction", ctx, mock.Anything).Return(nil)
	bookingRepo.On("Create", ctx, mock.Anything).Return(nil)

	// NATS Publisher
	publisher.On("PublishBookingCreated", ctx, mock.Anything, userID, assetID, startTime, endTime).Return(nil)

	booking, err := useCase.CreateBooking(ctx, userID, assetID, startTime, endTime)

	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, userID, booking.UserID)
	assert.Equal(t, assetID, booking.AssetID)
	assert.Equal(t, "Confirmed", booking.Status)

	assetClient.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	creditRepo.AssertExpectations(t)
	bookingRepo.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestDeductCredits_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := new(MockUserRepo)
	bookingRepo := new(MockBookingRepo)
	creditRepo := new(MockCreditRepo)
	txManager := new(MockTxManager)
	publisher := new(MockMessagePublisher)
	assetClient := new(MockAssetClient)

	useCase := NewMembershipUseCase(userRepo, bookingRepo, creditRepo, txManager, publisher, assetClient)

	userID := "user-123"
	amount := 15

	// Mock repo and tx updates
	userRepo.On("UpdateCredits", ctx, userID, -amount).Return(nil)
	creditRepo.On("CreateTransaction", ctx, mock.Anything).Return(nil)

	err := useCase.DeductCredits(ctx, userID, amount)

	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
	creditRepo.AssertExpectations(t)
}
