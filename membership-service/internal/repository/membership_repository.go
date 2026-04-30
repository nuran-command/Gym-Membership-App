package repository

import (
	"errors"
	"gym-membership/membership-service/internal/domain"
	"sync"
)

type inMemoryMembershipRepo struct {
	memberships map[string]*domain.Membership
	mu          sync.RWMutex
}

func NewInMemoryMembershipRepository() domain.MembershipRepository {
	return &inMemoryMembershipRepo{
		memberships: make(map[string]*domain.Membership),
	}
}

func (r *inMemoryMembershipRepo) Create(m *domain.Membership) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.memberships[m.ID] = m
	return nil
}

func (r *inMemoryMembershipRepo) GetByID(id string) (*domain.Membership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.memberships[id]
	if !ok {
		return nil, errors.New("membership not found")
	}
	return m, nil
}

func (r *inMemoryMembershipRepo) Cancel(id string) (*domain.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.memberships[id]
	if !ok {
		return nil, errors.New("membership not found")
	}
	m.Status = "cancelled"
	return m, nil
}

func (r *inMemoryMembershipRepo) ValidateAccess(userID, assetID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, m := range r.memberships {
		if m.UserID == userID && m.AssetID == assetID && m.Status == "active" {
			return true, nil
		}
	}
	return false, nil
}
