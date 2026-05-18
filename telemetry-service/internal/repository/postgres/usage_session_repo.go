package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/domain"

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

func (r *UsageSessionRepo) ListByUserID(ctx context.Context, userID string) ([]*domain.UsageSession, error) {
	query := `
		SELECT id, booking_id, user_id, asset_id, started_at, ended_at, duration_minutes, email_sent
		FROM usage_sessions
		WHERE user_id = $1
		ORDER BY started_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.UsageSession
	for rows.Next() {
		session := &domain.UsageSession{}
		var endedAt sql.NullTime
		if err := rows.Scan(
			&session.ID,
			&session.BookingID,
			&session.UserID,
			&session.AssetID,
			&session.StartedAt,
			&endedAt,
			&session.DurationMinutes,
			&session.EmailSent,
		); err != nil {
			return nil, err
		}
		if endedAt.Valid {
			session.EndedAt = &endedAt.Time
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (r *UsageSessionRepo) ListByAssetID(ctx context.Context, assetID string) ([]*domain.UsageSession, error) {
	query := `
		SELECT id, booking_id, user_id, asset_id, started_at, ended_at, duration_minutes, email_sent
		FROM usage_sessions
		WHERE asset_id = $1
		ORDER BY started_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.UsageSession
	for rows.Next() {
		session := &domain.UsageSession{}
		var endedAt sql.NullTime
		if err := rows.Scan(
			&session.ID,
			&session.BookingID,
			&session.UserID,
			&session.AssetID,
			&session.StartedAt,
			&endedAt,
			&session.DurationMinutes,
			&session.EmailSent,
		); err != nil {
			return nil, err
		}
		if endedAt.Valid {
			session.EndedAt = &endedAt.Time
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (r *UsageSessionRepo) GetStatsByUserID(ctx context.Context, userID string) (int, int, error) {
	query := `
		SELECT COUNT(id), COALESCE(SUM(duration_minutes), 0)
		FROM usage_sessions
		WHERE user_id = $1 AND ended_at IS NOT NULL
	`
	var totalSessions, totalDuration int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&totalSessions, &totalDuration)
	if err != nil {
		return 0, 0, err
	}
	return totalSessions, totalDuration, nil
}

func (r *UsageSessionRepo) GetSystemUsageStats(ctx context.Context) ([]*domain.AssetUsageStats, error) {
	query := `
		SELECT asset_id, COUNT(id) as total_sessions, COALESCE(AVG(duration_minutes), 0) as avg_duration
		FROM usage_sessions
		WHERE ended_at IS NOT NULL
		GROUP BY asset_id
		ORDER BY total_sessions DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*domain.AssetUsageStats
	for rows.Next() {
		stat := &domain.AssetUsageStats{}
		var avg float64
		if err := rows.Scan(&stat.AssetID, &stat.TotalSessions, &avg); err != nil {
			return nil, err
		}
		stat.AvgDurationMinutes = int(avg)
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

func (r *UsageSessionRepo) Delete(ctx context.Context, bookingID string) error {
	query := `DELETE FROM usage_sessions WHERE booking_id = $1`
	_, err := r.db.ExecContext(ctx, query, bookingID)
	return err
}

func (r *UsageSessionRepo) GetActiveByUserID(ctx context.Context, userID string) (*domain.UsageSession, error) {
	query := `
		SELECT id, booking_id, user_id, asset_id, started_at, ended_at, duration_minutes, email_sent
		FROM usage_sessions
		WHERE user_id = $1 AND ended_at IS NULL
		LIMIT 1
	`
	session := &domain.UsageSession{}
	var endedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
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
			return nil, errors.New("active session not found")
		}
		return nil, err
	}
	return session, nil
}

func (r *UsageSessionRepo) GetActiveByAssetID(ctx context.Context, assetID string) (*domain.UsageSession, error) {
	query := `
		SELECT id, booking_id, user_id, asset_id, started_at, ended_at, duration_minutes, email_sent
		FROM usage_sessions
		WHERE asset_id = $1 AND ended_at IS NULL
		LIMIT 1
	`
	session := &domain.UsageSession{}
	var endedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, assetID).Scan(
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
			return nil, errors.New("active session not found")
		}
		return nil, err
	}
	return session, nil
}
