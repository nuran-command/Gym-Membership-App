package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/ilnur/gym-membership-app/proto/asset"
)

type membershipUseCase struct {
	userRepo    domain.UserRepo
	bookingRepo domain.BookingRepo
	creditRepo  domain.CreditRepo
	assetClient asset.AssetServiceClient
}

func NewMembershipUseCase(
	userRepo domain.UserRepo,
	bookingRepo domain.BookingRepo,
	creditRepo domain.CreditRepo,
	assetClient asset.AssetServiceClient,
) domain.MembershipUseCase {
	return &membershipUseCase{
		userRepo:    userRepo,
		bookingRepo: bookingRepo,
		creditRepo:  creditRepo,
		assetClient: assetClient,
	}
}

func (u *membershipUseCase) CreateBooking(ctx context.Context, userID, assetID string, startTime, endTime string) (*domain.Booking, error) {
	resp, err := u.assetClient.CheckAvailability(ctx, &asset.CheckRequest{
		Id:        assetID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check asset availability: %w", err)
	}

	if !resp.Available {
		return nil, errors.New("asset is not available for the selected time")
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.Credits < 10 {
		return nil, errors.New("insufficient credits")
	}

	err = u.userRepo.UpdateCredits(ctx, userID, -10)
	if err != nil {
		return nil, fmt.Errorf("failed to deduct credits: %w", err)
	}

	booking := &domain.Booking{
		ID:      uuid.New().String(),
		UserID:  userID,
		AssetID: assetID,
		Status:  "Confirmed",
	}

	err = u.bookingRepo.Create(ctx, booking)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	return booking, nil
}

func (u *membershipUseCase) CancelBooking(ctx context.Context, bookingID string) error {
	return u.bookingRepo.UpdateStatus(ctx, bookingID, "Cancelled")
}

func (u *membershipUseCase) GetUserCredits(ctx context.Context, userID string) (int, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.Credits, nil
}

func (u *membershipUseCase) DeductCredits(ctx context.Context, userID string, amount int) error {
	return u.userRepo.UpdateCredits(ctx, userID, -amount)
}

func (u *membershipUseCase) AddCredits(ctx context.Context, userID string, amount int) error {
	return u.userRepo.UpdateCredits(ctx, userID, amount)
}

func (u *membershipUseCase) GetUserBookings(ctx context.Context, userID string) ([]*domain.Booking, error) {
	return u.bookingRepo.GetByUserID(ctx, userID)
}
