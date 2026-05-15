package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) domain.UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	q := getQueryer(ctx, r.db)
	var user domain.User
	err := sqlx.GetContext(ctx, q, &user, "SELECT id, name, email, credits, created_at FROM users WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) UpdateCredits(ctx context.Context, userID string, amount int) error {
	q := getQueryer(ctx, r.db)
	res, err := q.ExecContext(ctx, "UPDATE users SET credits = credits + $1 WHERE id = $2", amount, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found for credits update")
	}
	return nil
}
