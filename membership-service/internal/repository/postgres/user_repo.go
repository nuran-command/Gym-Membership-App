package postgres

import (
	"context"
	"errors"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
)

type userRepo struct{}

func NewUserRepo() domain.UserRepo {
	return &userRepo{}
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return nil, errors.New("not implemented")
}

func (r *userRepo) UpdateCredits(ctx context.Context, userID string, amount int) error {
	return errors.New("not implemented")
}
