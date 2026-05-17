package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type bookingRepo struct {
	db *sqlx.DB
}

func NewBookingRepo(db *sqlx.DB) domain.BookingRepo {
	return &bookingRepo{db: db}
}

func (r *bookingRepo) Create(ctx context.Context, booking *domain.Booking) error {
	q := getQueryer(ctx, r.db)
	_, err := q.ExecContext(ctx,
		"INSERT INTO bookings (id, user_id, asset_id, start_time, end_time, status) VALUES ($1, $2, $3, $4, $5, $6)",
		booking.ID, booking.UserID, booking.AssetID, booking.StartTime, booking.EndTime, booking.Status,
	)
	return err
}

func (r *bookingRepo) GetByUserID(ctx context.Context, userID string) ([]*domain.Booking, error) {
	q := getQueryer(ctx, r.db)
	var bookings []*domain.Booking
	err := sqlx.SelectContext(ctx, q, &bookings,
		"SELECT id, user_id, asset_id, start_time, end_time, status FROM bookings WHERE user_id = $1 ORDER BY start_time DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *bookingRepo) UpdateStatus(ctx context.Context, bookingID string, status string) error {
	q := getQueryer(ctx, r.db)
	res, err := q.ExecContext(ctx, "UPDATE bookings SET status = $1 WHERE id = $2", status, bookingID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("booking not found for status update")
	}
	return nil
}

func (r *bookingRepo) GetByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	q := getQueryer(ctx, r.db)
	var booking domain.Booking
	err := sqlx.GetContext(ctx, q, &booking,
		"SELECT id, user_id, asset_id, start_time, end_time, status FROM bookings WHERE id = $1",
		bookingID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("booking not found")
		}
		return nil, err
	}
	return &booking, nil
}
