package usecase

import (
	"context"
	"gym-membership/telemetry-service/internal/domain"
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
		err = u.emailSender.SendThankYouEmail(ctx, userEmail, session)
		if err != nil {
			// Log error but do not fail the transaction since booking is returned
			return err
		}
		
		session.EmailSent = true
		_ = u.repo.Update(ctx, session)
	}

	return nil
}
