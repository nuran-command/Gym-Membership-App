package postgres

import (
	"context"
	"errors"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
)

type bookingRepo struct{}

func NewBookingRepo() domain.BookingRepo {
	return &bookingRepo{}
}

func (r *bookingRepo) Create(ctx context.Context, booking *domain.Booking) error {
	return errors.New("not implemented")
}

func (r *bookingRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Booking, error) {
	return nil, errors.New("not implemented")
}

func (r *bookingRepo) UpdateStatus(ctx context.Context, bookingID string, status string) error {
	return errors.New("not implemented")
}

func (r *bookingRepo) GetByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	return nil, errors.New("not implemented")
}
