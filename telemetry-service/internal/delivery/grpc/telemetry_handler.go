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

func (h *TelemetryHandler) CreateUsageSession(ctx context.Context, req *telemetry.CreateUsageSessionRequest) (*telemetry.UsageSessionResponse, error) {
	session, err := h.uc.CreateUsageSession(ctx, req.BookingId, req.UserId, req.AssetId)
	if err != nil {
		return nil, err
	}

	return &telemetry.UsageSessionResponse{
		BookingId:       session.BookingID,
		UserId:          session.UserID,
		AssetId:         session.AssetID,
		StartedAt:       session.StartedAt.Format(time.RFC3339),
		DurationMinutes: 0,
	}, nil
}

func (h *TelemetryHandler) UpdateUsageSession(ctx context.Context, req *telemetry.UpdateUsageSessionRequest) (*telemetry.UsageSessionResponse, error) {
	endedAt, err := time.Parse(time.RFC3339, req.EndedAt)
	if err != nil {
		endedAt = time.Now()
	}

	session, err := h.uc.UpdateUsageSession(ctx, req.BookingId, endedAt, int(req.DurationMinutes), req.EmailSent)
	if err != nil {
		return nil, err
	}

	var endedAtStr string
	if session.EndedAt != nil {
		endedAtStr = session.EndedAt.Format(time.RFC3339)
	}

	return &telemetry.UsageSessionResponse{
		BookingId:       session.BookingID,
		UserId:          session.UserID,
		AssetId:         session.AssetID,
		StartedAt:       session.StartedAt.Format(time.RFC3339),
		EndedAt:         endedAtStr,
		DurationMinutes: int32(session.DurationMinutes),
	}, nil
}

func (h *TelemetryHandler) DeleteUsageSession(ctx context.Context, req *telemetry.DeleteUsageSessionRequest) (*telemetry.DeleteUsageSessionResponse, error) {
	err := h.uc.DeleteUsageSession(ctx, req.BookingId)
	if err != nil {
		return &telemetry.DeleteUsageSessionResponse{Success: false}, err
	}
	return &telemetry.DeleteUsageSessionResponse{Success: true}, nil
}

func (h *TelemetryHandler) GetUserActiveSession(ctx context.Context, req *telemetry.GetUserActiveSessionRequest) (*telemetry.UsageSessionResponse, error) {
	session, err := h.uc.GetUserActiveSession(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	var endedAtStr string
	if session.EndedAt != nil {
		endedAtStr = session.EndedAt.Format(time.RFC3339)
	}

	return &telemetry.UsageSessionResponse{
		BookingId:       session.BookingID,
		UserId:          session.UserID,
		AssetId:         session.AssetID,
		StartedAt:       session.StartedAt.Format(time.RFC3339),
		EndedAt:         endedAtStr,
		DurationMinutes: int32(session.DurationMinutes),
	}, nil
}

func (h *TelemetryHandler) GetAssetActiveSession(ctx context.Context, req *telemetry.GetAssetActiveSessionRequest) (*telemetry.UsageSessionResponse, error) {
	session, err := h.uc.GetAssetActiveSession(ctx, req.AssetId)
	if err != nil {
		return nil, err
	}

	var endedAtStr string
	if session.EndedAt != nil {
		endedAtStr = session.EndedAt.Format(time.RFC3339)
	}

	return &telemetry.UsageSessionResponse{
		BookingId:       session.BookingID,
		UserId:          session.UserID,
		AssetId:         session.AssetID,
		StartedAt:       session.StartedAt.Format(time.RFC3339),
		EndedAt:         endedAtStr,
		DurationMinutes: int32(session.DurationMinutes),
	}, nil
}

func (h *TelemetryHandler) Heartbeat(ctx context.Context, req *telemetry.HeartbeatRequest) (*telemetry.HeartbeatResponse, error) {
	return &telemetry.HeartbeatResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

func (h *TelemetryHandler) LogEvent(ctx context.Context, req *telemetry.LogEventRequest) (*telemetry.LogEventResponse, error) {
	// Dummy log event recorder
	return &telemetry.LogEventResponse{
		Success: true,
	}, nil
}
