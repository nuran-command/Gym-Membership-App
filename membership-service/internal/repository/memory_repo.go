package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/your-username/gym-membership-app/membership-service/internal/domain"
)

type MemoryRepo struct {
	users    map[string]*domain.User
	bookings map[string]*domain.Booking
	mu       sync.RWMutex
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		users:    make(map[string]*domain.User),
		bookings: make(map[string]*domain.Booking),
	}
}

func (r *MemoryRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *MemoryRepo) UpdateCredits(ctx context.Context, userID string, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	user.Credits += amount
	return nil
}

func (r *MemoryRepo) Create(ctx context.Context, booking *domain.Booking) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bookings[booking.ID] = booking
	return nil
}

func (r *MemoryRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Booking
	for _, b := range r.bookings {
		if b.UserID == userID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (r *MemoryRepo) UpdateStatus(ctx context.Context, bookingID string, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	booking, ok := r.bookings[bookingID]
	if !ok {
		return errors.New("booking not found")
	}
	booking.Status = status
	return nil
}

func (r *MemoryRepo) GetByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	booking, ok := r.bookings[bookingID]
	if !ok {
		return nil, errors.New("booking not found")
	}
	return booking, nil
}

func (r *MemoryRepo) CreateTransaction(ctx context.Context, tx *domain.CreditTransaction) error {
	return nil
}
