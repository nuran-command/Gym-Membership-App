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

	// For now, let's assume we have the user email or can get it. 
	// The requirement mentioned EmailSender implementation later.
	return nil
}
