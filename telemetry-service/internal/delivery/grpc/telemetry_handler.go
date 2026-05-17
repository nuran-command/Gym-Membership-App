package grpc

import (
	"context"
	"gym-membership/telemetry-service/internal/usecase"
	telemetry "gym-membership/telemetry-service/proto"
	"time"
)

type TelemetryHandler struct {
	telemetry.UnimplementedTelemetryServiceServer
	uc *usecase.TelemetryUsecase
}

func NewTelemetryHandler(uc *usecase.TelemetryUsecase) *TelemetryHandler {
	return &TelemetryHandler{
		uc: uc,
	}
}

func (h *TelemetryHandler) GetUsageSession(ctx context.Context, req *telemetry.GetUsageSessionRequest) (*telemetry.UsageSessionResponse, error) {
	session, err := h.uc.GetUsageSession(ctx, req.BookingId)
	if err != nil {
		return nil, err
	}

	var endedAt string
	if session.EndedAt != nil {
		endedAt = session.EndedAt.Format(time.RFC3339)
	}

	return &telemetry.UsageSessionResponse{
		BookingId:       session.BookingID,
		UserId:          session.UserID,
		AssetId:         session.AssetID,
		StartedAt:       session.StartedAt.Format(time.RFC3339),
		EndedAt:         endedAt,
		DurationMinutes: int32(session.DurationMinutes),
	}, nil
}

func (h *TelemetryHandler) ListUserSessions(ctx context.Context, req *telemetry.ListUserSessionsRequest) (*telemetry.ListUserSessionsResponse, error) {
	sessions, err := h.uc.ListUserSessions(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	var responseSessions []*telemetry.UsageSessionResponse
	for _, session := range sessions {
		var endedAt string
		if session.EndedAt != nil {
			endedAt = session.EndedAt.Format(time.RFC3339)
		}
		responseSessions = append(responseSessions, &telemetry.UsageSessionResponse{
			BookingId:       session.BookingID,
			UserId:          session.UserID,
			AssetId:         session.AssetID,
			StartedAt:       session.StartedAt.Format(time.RFC3339),
			EndedAt:         endedAt,
			DurationMinutes: int32(session.DurationMinutes),
		})
	}

	return &telemetry.ListUserSessionsResponse{
		Sessions: responseSessions,
	}, nil
}

func (h *TelemetryHandler) GetUsageStats(ctx context.Context, req *telemetry.GetUsageStatsRequest) (*telemetry.UsageStatsResponse, error) {
	totalSessions, totalDuration, err := h.uc.GetUsageStats(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &telemetry.UsageStatsResponse{
		TotalSessions:        int32(totalSessions),
		TotalDurationMinutes: int32(totalDuration),
	}, nil
}

func (h *TelemetryHandler) GetAssetUsageHistory(ctx context.Context, req *telemetry.GetAssetUsageHistoryRequest) (*telemetry.GetAssetUsageHistoryResponse, error) {
	sessions, err := h.uc.GetAssetUsageHistory(ctx, req.AssetId)
	if err != nil {
		return nil, err
	}

	var responseSessions []*telemetry.UsageSessionResponse
	for _, session := range sessions {
		var endedAt string
		if session.EndedAt != nil {
			endedAt = session.EndedAt.Format(time.RFC3339)
		}
		responseSessions = append(responseSessions, &telemetry.UsageSessionResponse{
			BookingId:       session.BookingID,
			UserId:          session.UserID,
			AssetId:         session.AssetID,
			StartedAt:       session.StartedAt.Format(time.RFC3339),
			EndedAt:         endedAt,
			DurationMinutes: int32(session.DurationMinutes),
		})
	}

	return &telemetry.GetAssetUsageHistoryResponse{
		Sessions: responseSessions,
	}, nil
}

func (h *TelemetryHandler) GetSystemUsageStats(ctx context.Context, req *telemetry.GetSystemUsageStatsRequest) (*telemetry.SystemUsageStatsResponse, error) {
	stats, err := h.uc.GetSystemUsageStats(ctx)
	if err != nil {
		return nil, err
	}

	var responseStats []*telemetry.AssetUsageStats
	for _, stat := range stats {
		responseStats = append(responseStats, &telemetry.AssetUsageStats{
			AssetId:            stat.AssetID,
			TotalSessions:      int32(stat.TotalSessions),
			AvgDurationMinutes: int32(stat.AvgDurationMinutes),
		})
	}

	return &telemetry.SystemUsageStatsResponse{
		AssetStats: responseStats,
	}, nil
}
