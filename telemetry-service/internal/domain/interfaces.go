package domain

import "context"

type UsageSessionRepo interface {
	Create(ctx context.Context, session *UsageSession) error
	Update(ctx context.Context, session *UsageSession) error
	GetByBookingID(ctx context.Context, bookingID string) (*UsageSession, error)
}

type EmailSender interface {
	SendThankYouEmail(ctx context.Context, email string, session *UsageSession) error
}
