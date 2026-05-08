package postgres

import (
	"context"
	"errors"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
)

type creditRepo struct{}

func NewCreditRepo() domain.CreditRepo {
	return &creditRepo{}
}

func (r *creditRepo) CreateTransaction(ctx context.Context, tx *domain.CreditTransaction) error {
	return errors.New("not implemented")
}
