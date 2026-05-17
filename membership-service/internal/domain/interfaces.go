package domain

import "context"

type UserRepo interface {
	GetByID(ctx context.Context, id string) (*User, error)
	UpdateCredits(ctx context.Context, userID string, amount int) error
}

type BookingRepo interface {
	Create(ctx context.Context, booking *Booking) error
	GetByUserID(ctx context.Context, userID string) ([]*Booking, error)
	UpdateStatus(ctx context.Context, bookingID string, status string) error
	GetByID(ctx context.Context, bookingID string) (*Booking, error)
}

type CreditRepo interface {
	CreateTransaction(ctx context.Context, tx *CreditTransaction) error
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type MessagePublisher interface {
	PublishBookingCreated(ctx context.Context, bookingID, userID, assetID string, startTime, endTime string) error
	PublishBookingCancelled(ctx context.Context, bookingID string) error
	PublishBookingReturned(ctx context.Context, bookingID string) error
}

type MembershipUseCase interface {
	CreateBooking(ctx context.Context, userID, assetID string, startTime, endTime string) (*Booking, error)
	CancelBooking(ctx context.Context, bookingID string) error
	ReturnBooking(ctx context.Context, bookingID string) error
	GetUserCredits(ctx context.Context, userID string) (int, error)
	DeductCredits(ctx context.Context, userID string, amount int) error
	AddCredits(ctx context.Context, userID string, amount int) error
	GetUserBookings(ctx context.Context, userID string) ([]*Booking, error)
}
