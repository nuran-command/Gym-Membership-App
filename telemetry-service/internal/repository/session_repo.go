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

func (r *InMemorySessionRepo) Delete(ctx context.Context, bookingID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, bookingID)
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

func (r *InMemorySessionRepo) ListByUserID(ctx context.Context, userID string) ([]*domain.UsageSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var sessions []*domain.UsageSession
	for _, s := range r.sessions {
		if s.UserID == userID {
			sessions = append(sessions, s)
		}
	}
	return sessions, nil
}

func (r *InMemorySessionRepo) ListByAssetID(ctx context.Context, assetID string) ([]*domain.UsageSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var sessions []*domain.UsageSession
	for _, s := range r.sessions {
		if s.AssetID == assetID {
			sessions = append(sessions, s)
		}
	}
	return sessions, nil
}

func (r *InMemorySessionRepo) GetStatsByUserID(ctx context.Context, userID string) (int, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count, duration int
	for _, s := range r.sessions {
		if s.UserID == userID && s.EndedAt != nil {
			count++
			duration += s.DurationMinutes
		}
	}
	return count, duration, nil
}

func (r *InMemorySessionRepo) GetSystemUsageStats(ctx context.Context) ([]*domain.AssetUsageStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Quick dummy implementation for compilation
	return nil, nil
}

func (r *InMemorySessionRepo) GetActiveByUserID(ctx context.Context, userID string) (*domain.UsageSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.sessions {
		if s.UserID == userID && s.EndedAt == nil {
			return s, nil
		}
	}
	return nil, fmt.Errorf("active session not found")
}

func (r *InMemorySessionRepo) GetActiveByAssetID(ctx context.Context, assetID string) (*domain.UsageSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.sessions {
		if s.AssetID == assetID && s.EndedAt == nil {
			return s, nil
		}
	}
	return nil, fmt.Errorf("active session not found")
}
