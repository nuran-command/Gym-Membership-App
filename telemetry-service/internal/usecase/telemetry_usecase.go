package usecase

import (
	"context"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/domain"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/observability"
	"log/slog"
	"time"
)

type TelemetryUsecase struct {
	repo        domain.UsageSessionRepo
	emailSender domain.EmailSender
}

func NewTelemetryUsecase(repo domain.UsageSessionRepo, emailSender domain.EmailSender) *TelemetryUsecase {
	return &TelemetryUsecase{
		repo:        repo,
		emailSender: emailSender,
	}
}

func (u *TelemetryUsecase) HandleBookingCreated(ctx context.Context, userID, assetID, bookingID string) error {
	session := &domain.UsageSession{
		BookingID: bookingID,
		UserID:    userID,
		AssetID:   assetID,
		StartedAt: time.Now(),
		// EndedAt is implicitly nil
	}
	return u.repo.Create(ctx, session)
}

func (u *TelemetryUsecase) HandleBookingReturned(ctx context.Context, userID, assetID, bookingID string, duration int) error {
	session, err := u.repo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}

	now := time.Now()
	session.EndedAt = &now
	session.DurationMinutes = duration

	err = u.repo.Update(ctx, session)
	if err != nil {
		return err
	}

	// Send email if configured
	if u.emailSender != nil {
		// Replace with actual email lookup if needed later
		userEmail := "user_" + userID + "@example.com"
		
		slog.Info("Sending thank-you email",
			slog.String("booking_id", bookingID),
			slog.String("user_id", userID),
			slog.String("email", userEmail),
		)

		err = u.emailSender.SendThankYouEmail(ctx, userEmail, session)
		if err != nil {
			slog.Error("Failed to send thank-you email",
				slog.String("error", err.Error()),
				slog.String("booking_id", bookingID),
				slog.String("user_id", userID),
			)
			observability.EmailsSent.WithLabelValues("error").Inc()
			// Log error but do not fail the transaction since booking is returned
			return err
		}
		
		slog.Info("Successfully sent thank-you email",
			slog.String("booking_id", bookingID),
			slog.String("user_id", userID),
		)
		observability.EmailsSent.WithLabelValues("success").Inc()
		
		session.EmailSent = true
		_ = u.repo.Update(ctx, session)
	}

	return nil
}

func (u *TelemetryUsecase) GetUsageSession(ctx context.Context, bookingID string) (*domain.UsageSession, error) {
	return u.repo.GetByBookingID(ctx, bookingID)
}

func (u *TelemetryUsecase) ListUserSessions(ctx context.Context, userID string) ([]*domain.UsageSession, error) {
	return u.repo.ListByUserID(ctx, userID)
}

func (u *TelemetryUsecase) GetUsageStats(ctx context.Context, userID string) (int, int, error) {
	return u.repo.GetStatsByUserID(ctx, userID)
}

func (u *TelemetryUsecase) GetAssetUsageHistory(ctx context.Context, assetID string) ([]*domain.UsageSession, error) {
	return u.repo.ListByAssetID(ctx, assetID)
}

func (u *TelemetryUsecase) GetSystemUsageStats(ctx context.Context) ([]*domain.AssetUsageStats, error) {
	return u.repo.GetSystemUsageStats(ctx)
}

func (u *TelemetryUsecase) CreateUsageSession(ctx context.Context, bookingID, userID, assetID string) (*domain.UsageSession, error) {
	session := &domain.UsageSession{
		BookingID: bookingID,
		UserID:    userID,
		AssetID:   assetID,
		StartedAt: time.Now(),
	}
	err := u.repo.Create(ctx, session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (u *TelemetryUsecase) UpdateUsageSession(ctx context.Context, bookingID string, endedAt time.Time, duration int, emailSent bool) (*domain.UsageSession, error) {
	session, err := u.repo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	session.EndedAt = &endedAt
	session.DurationMinutes = duration
	session.EmailSent = emailSent

	err = u.repo.Update(ctx, session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (u *TelemetryUsecase) DeleteUsageSession(ctx context.Context, bookingID string) error {
	return u.repo.Delete(ctx, bookingID)
}

func (u *TelemetryUsecase) GetUserActiveSession(ctx context.Context, userID string) (*domain.UsageSession, error) {
	return u.repo.GetActiveByUserID(ctx, userID)
}

func (u *TelemetryUsecase) GetAssetActiveSession(ctx context.Context, assetID string) (*domain.UsageSession, error) {
	return u.repo.GetActiveByAssetID(ctx, assetID)
}
