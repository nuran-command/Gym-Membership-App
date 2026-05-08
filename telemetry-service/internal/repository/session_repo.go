package repository

import (
	"context"
	"fmt"
	"gym-membership/telemetry-service/internal/domain"
	"sync"
)

type InMemorySessionRepo struct {
	mu       sync.RWMutex
	sessions map[string]*domain.UsageSession
}

func NewInMemorySessionRepo() *InMemorySessionRepo {
	return &InMemorySessionRepo{
		sessions: make(map[string]*domain.UsageSession),
	}
}

func (r *InMemorySessionRepo) Create(ctx context.Context, session *domain.UsageSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.BookingID] = session
	return nil
}

func (r *InMemorySessionRepo) Update(ctx context.Context, session *domain.UsageSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.BookingID] = session
	return nil
}

func (r *InMemorySessionRepo) GetByBookingID(ctx context.Context, bookingID string) (*domain.UsageSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[bookingID]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}
