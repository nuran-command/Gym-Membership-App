package postgres

import (
	"context"
	"database/sql"
	"errors"
	"gym-membership/telemetry-service/internal/domain"

	_ "github.com/lib/pq"
)

type UsageSessionRepo struct {
	db *sql.DB
}

func NewUsageSessionRepo(db *sql.DB) *UsageSessionRepo {
	return &UsageSessionRepo{db: db}
}

func (r *UsageSessionRepo) Create(ctx context.Context, session *domain.UsageSession) error {
	query := `
		INSERT INTO usage_sessions (booking_id, user_id, asset_id, started_at, ended_at, duration_minutes, email_sent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		session.BookingID,
		session.UserID,
		session.AssetID,
		session.StartedAt,
		session.EndedAt,
		session.DurationMinutes,
		session.EmailSent,
	).Scan(&session.ID)
	return err
}

func (r *UsageSessionRepo) Update(ctx context.Context, session *domain.UsageSession) error {
	query := `
		UPDATE usage_sessions
		SET ended_at = $1, duration_minutes = $2, email_sent = $3
		WHERE booking_id = $4
	`
	_, err := r.db.ExecContext(ctx, query,
		session.EndedAt,
		session.DurationMinutes,
		session.EmailSent,
		session.BookingID,
	)
	return err
}

func (r *UsageSessionRepo) GetByBookingID(ctx context.Context, bookingID string) (*domain.UsageSession, error) {
	query := `
		SELECT id, booking_id, user_id, asset_id, started_at, ended_at, duration_minutes, email_sent
		FROM usage_sessions
		WHERE booking_id = $1
	`
	session := &domain.UsageSession{}
	var endedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, bookingID).Scan(
		&session.ID,
		&session.BookingID,
		&session.UserID,
		&session.AssetID,
		&session.StartedAt,
		&endedAt,
		&session.DurationMinutes,
		&session.EmailSent,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}
	return session, nil
}
