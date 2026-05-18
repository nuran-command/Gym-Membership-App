package domain

import "time"

type UsageSession struct {
	ID              int        `json:"id"`
	BookingID       string     `json:"booking_id"`
	UserID          string     `json:"user_id"`
	AssetID         string     `json:"asset_id"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DurationMinutes int        `json:"duration_minutes"`
	EmailSent       bool       `json:"email_sent"`
}

type AssetUsageStats struct {
	AssetID            string `json:"asset_id"`
	TotalSessions      int    `json:"total_sessions"`
	AvgDurationMinutes int    `json:"avg_duration_minutes"`
}
