package domain

import "context"

type UsageSessionRepo interface {
	Create(ctx context.Context, session *UsageSession) error
	Update(ctx context.Context, session *UsageSession) error
	Delete(ctx context.Context, bookingID string) error
	GetByBookingID(ctx context.Context, bookingID string) (*UsageSession, error)
	ListByUserID(ctx context.Context, userID string) ([]*UsageSession, error)
	ListByAssetID(ctx context.Context, assetID string) ([]*UsageSession, error)
	GetStatsByUserID(ctx context.Context, userID string) (int, int, error)
	GetSystemUsageStats(ctx context.Context) ([]*AssetUsageStats, error)
	GetActiveByUserID(ctx context.Context, userID string) (*UsageSession, error)
	GetActiveByAssetID(ctx context.Context, assetID string) (*UsageSession, error)
}

type EmailSender interface {
	SendThankYouEmail(ctx context.Context, email string, session *UsageSession) error
}
