package usecase

import (
	"gym-membership/telemetry-service/internal/domain"
)

type telemetryUsecase struct {
	repo domain.TelemetryRepository
}

func NewTelemetryUsecase(repo domain.TelemetryRepository) domain.TelemetryUsecase {
	return &telemetryUsecase{
		repo: repo,
	}
}

func (u *telemetryUsecase) RecordAccess(userID, assetID, timestamp string) error {
	return u.repo.RecordAccess(userID, assetID, timestamp)
}

func (u *telemetryUsecase) GetUsageStats(assetID, period string) (*domain.UsageStats, error) {
	return u.repo.GetUsageStats(assetID, period)
}

func (u *telemetryUsecase) Heartbeat(deviceID, status string) bool {
	// For now just acknowledge
	return true
}

func (u *telemetryUsecase) LogEvent(source, level, message string) error {
	return u.repo.LogEvent(source, level, message)
}
