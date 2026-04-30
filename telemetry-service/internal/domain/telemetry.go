package domain

type UsageStats struct {
	AssetID         string
	TotalAccesses   int32
	AverageDuration float32
}

type TelemetryRepository interface {
	RecordAccess(userID, assetID, timestamp string) error
	GetUsageStats(assetID, period string) (*UsageStats, error)
	LogEvent(source, level, message string) error
}

type TelemetryUsecase interface {
	RecordAccess(userID, assetID, timestamp string) error
	GetUsageStats(assetID, period string) (*UsageStats, error)
	Heartbeat(deviceID, status string) bool
	LogEvent(source, level, message string) error
}
