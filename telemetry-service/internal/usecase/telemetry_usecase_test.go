package usecase

import (
	"context"
	"errors"
	"gym-membership/telemetry-service/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUsageSessionRepo struct {
	mock.Mock
}

func (m *MockUsageSessionRepo) Create(ctx context.Context, session *domain.UsageSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockUsageSessionRepo) Update(ctx context.Context, session *domain.UsageSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockUsageSessionRepo) GetByBookingID(ctx context.Context, bookingID string) (*domain.UsageSession, error) {
	args := m.Called(ctx, bookingID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UsageSession), args.Error(1)
}

func (m *MockUsageSessionRepo) ListByUserID(ctx context.Context, userID string) ([]*domain.UsageSession, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.UsageSession), args.Error(1)
}

func (m *MockUsageSessionRepo) ListByAssetID(ctx context.Context, assetID string) ([]*domain.UsageSession, error) {
	args := m.Called(ctx, assetID)
	return args.Get(0).([]*domain.UsageSession), args.Error(1)
}

func (m *MockUsageSessionRepo) GetStatsByUserID(ctx context.Context, userID string) (int, int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockUsageSessionRepo) GetSystemUsageStats(ctx context.Context) ([]*domain.AssetUsageStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.AssetUsageStats), args.Error(1)
}

type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendThankYouEmail(ctx context.Context, email string, session *domain.UsageSession) error {
	args := m.Called(ctx, email, session)
	return args.Error(0)
}

func TestHandleBookingReturned_SendsEmail(t *testing.T) {
	mockRepo := new(MockUsageSessionRepo)
	mockEmailSender := new(MockEmailSender)
	uc := NewTelemetryUsecase(mockRepo, mockEmailSender)

	ctx := context.Background()
	bookingID := "booking-123"
	userID := "user-123"
	assetID := "asset-123"
	duration := 45

	session := &domain.UsageSession{
		ID:        1,
		BookingID: bookingID,
		UserID:    userID,
		AssetID:   assetID,
		StartedAt: time.Now().Add(-45 * time.Minute),
	}

	mockRepo.On("GetByBookingID", ctx, bookingID).Return(session, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.UsageSession")).Return(nil).Times(2)
	mockEmailSender.On("SendThankYouEmail", ctx, "user_user-123@example.com", mock.AnythingOfType("*domain.UsageSession")).Return(nil)

	err := uc.HandleBookingReturned(ctx, userID, assetID, bookingID, duration)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEmailSender.AssertExpectations(t)
}

func TestHandleBookingReturned_SavesSession(t *testing.T) {
	mockRepo := new(MockUsageSessionRepo)
	uc := NewTelemetryUsecase(mockRepo, nil) // no email sender

	ctx := context.Background()
	bookingID := "booking-123"
	userID := "user-123"
	assetID := "asset-123"
	duration := 45

	session := &domain.UsageSession{
		ID:        1,
		BookingID: bookingID,
		UserID:    userID,
		AssetID:   assetID,
		StartedAt: time.Now().Add(-45 * time.Minute),
	}

	mockRepo.On("GetByBookingID", ctx, bookingID).Return(session, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(s *domain.UsageSession) bool {
		return s.DurationMinutes == duration && s.EndedAt != nil && !s.EmailSent
	})).Return(nil).Once()

	err := uc.HandleBookingReturned(ctx, userID, assetID, bookingID, duration)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestHandleBookingReturned_EmailError_LeavesEmailSentFalse(t *testing.T) {
	mockRepo := new(MockUsageSessionRepo)
	mockEmailSender := new(MockEmailSender)
	uc := NewTelemetryUsecase(mockRepo, mockEmailSender)

	ctx := context.Background()
	bookingID := "booking-123"
	userID := "user-123"
	assetID := "asset-123"
	duration := 45

	session := &domain.UsageSession{
		ID:        1,
		BookingID: bookingID,
		UserID:    userID,
		AssetID:   assetID,
		StartedAt: time.Now().Add(-45 * time.Minute),
	}

	mockRepo.On("GetByBookingID", ctx, bookingID).Return(session, nil)
	
	// First update for ended_at and duration
	mockRepo.On("Update", ctx, mock.MatchedBy(func(s *domain.UsageSession) bool {
		return s.EmailSent == false
	})).Return(nil).Once()
	
	// Email sending fails
	mockEmailSender.On("SendThankYouEmail", ctx, "user_user-123@example.com", mock.AnythingOfType("*domain.UsageSession")).Return(errors.New("smtp down"))

	err := uc.HandleBookingReturned(ctx, userID, assetID, bookingID, duration)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
	mockEmailSender.AssertExpectations(t)
	assert.False(t, session.EmailSent)
}
