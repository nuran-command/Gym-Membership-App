package postgres

import (
	"context"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type creditRepo struct {
	db *sqlx.DB
}

func NewCreditRepo(db *sqlx.DB) domain.CreditRepo {
	return &creditRepo{db: db}
}

func (r *creditRepo) CreateTransaction(ctx context.Context, tx *domain.CreditTransaction) error {
	q := getQueryer(ctx, r.db)
	_, err := q.ExecContext(ctx,
		"INSERT INTO credit_transactions (id, user_id, amount, type, reason, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		tx.ID, tx.UserID, tx.Amount, tx.Type, tx.Reason, tx.CreatedAt,
	)
	return err
}

func (r *creditRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.CreditTransaction, error) {
	q := getQueryer(ctx, r.db)
	var transactions []*domain.CreditTransaction
	err := sqlx.SelectContext(ctx, q, &transactions,
		"SELECT id, user_id, amount, type, reason, created_at FROM credit_transactions WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}
