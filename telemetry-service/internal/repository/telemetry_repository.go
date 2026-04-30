package repository

import (
	"gym-membership/telemetry-service/internal/domain"
	"log"
)

type inMemoryTelemetryRepo struct {
}

func NewInMemoryTelemetryRepository() domain.TelemetryRepository {
	return &inMemoryTelemetryRepo{}
}

func (r *inMemoryTelemetryRepo) RecordAccess(userID, assetID, timestamp string) error {
	log.Printf("Access recorded: User %s, Asset %s at %s", userID, assetID, timestamp)
	return nil
}

func (r *inMemoryTelemetryRepo) GetUsageStats(assetID, period string) (*domain.UsageStats, error) {
	return &domain.UsageStats{
		AssetID:         assetID,
		TotalAccesses:   10,
		AverageDuration: 45.5,
	}, nil
}

func (r *inMemoryTelemetryRepo) LogEvent(source, level, message string) error {
	log.Printf("[%s] %s: %s", level, source, message)
	return nil
}
