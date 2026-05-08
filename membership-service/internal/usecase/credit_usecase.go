package usecase

import (
	"context"

	"github.com/your-username/gym-membership-app/membership-service/internal/domain"
)

type creditUseCase struct {
	userRepo   domain.UserRepo
	creditRepo domain.CreditRepo
}

func NewCreditUseCase(userRepo domain.UserRepo, creditRepo domain.CreditRepo) *creditUseCase {
	return &creditUseCase{
		userRepo:   userRepo,
		creditRepo: creditRepo,
	}
}

func (u *creditUseCase) GetUserCredits(ctx context.Context, userID string) (int, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.Credits, nil
}

func (u *creditUseCase) DeductCredits(ctx context.Context, userID string, amount int) error {
	return u.userRepo.UpdateCredits(ctx, userID, -amount)
}

func (u *creditUseCase) AddCredits(ctx context.Context, userID string, amount int) error {
	return u.userRepo.UpdateCredits(ctx, userID, amount)
}
