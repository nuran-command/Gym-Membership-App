package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type membershipRepo struct {
	db *sqlx.DB
}

func NewMembershipRepo(db *sqlx.DB) domain.MembershipRepo {
	return &membershipRepo{db: db}
}

func (r *membershipRepo) Create(ctx context.Context, m *domain.Membership) error {
	q := getQueryer(ctx, r.db)
	_, err := q.ExecContext(ctx,
		"INSERT INTO memberships (id, user_id, type, status, expires_at) VALUES ($1, $2, $3, $4, $5)",
		m.ID, m.UserID, m.Type, m.Status, m.ExpiresAt,
	)
	return err
}

func (r *membershipRepo) GetByUserID(ctx context.Context, userID string) (*domain.Membership, error) {
	q := getQueryer(ctx, r.db)
	var m domain.Membership
	err := sqlx.GetContext(ctx, q, &m, "SELECT id, user_id, type, status, expires_at FROM memberships WHERE user_id = $1", userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("membership not found")
		}
		return nil, err
	}
	return &m, nil
}

func (r *membershipRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	q := getQueryer(ctx, r.db)
	res, err := q.ExecContext(ctx, "UPDATE memberships SET status = $1 WHERE id = $2", status, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("membership not found for status update")
	}
	return nil
}
