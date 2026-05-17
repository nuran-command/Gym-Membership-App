package domain

import "context"

type UsageSessionRepo interface {
	Create(ctx context.Context, session *UsageSession) error
	Update(ctx context.Context, session *UsageSession) error
	GetByBookingID(ctx context.Context, bookingID string) (*UsageSession, error)
	ListByUserID(ctx context.Context, userID string) ([]*UsageSession, error)
	ListByAssetID(ctx context.Context, assetID string) ([]*UsageSession, error)
	GetStatsByUserID(ctx context.Context, userID string) (int, int, error)
}

type EmailSender interface {
	SendThankYouEmail(ctx context.Context, email string, session *UsageSession) error
}
