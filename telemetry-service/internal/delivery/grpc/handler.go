package grpc

import (
	"context"
	"gym-membership/telemetry-service/internal/domain"
	pb "gym-membership/proto/telemetry"
)

type TelemetryHandler struct {
	pb.UnimplementedTelemetryServiceServer
	usecase domain.TelemetryUsecase
}

func NewTelemetryHandler(u domain.TelemetryUsecase) *TelemetryHandler {
	return &TelemetryHandler{
		usecase: u,
	}
}

func (h *TelemetryHandler) RecordAccess(ctx context.Context, req *pb.RecordAccessRequest) (*pb.TelemetryResponse, error) {
	err := h.usecase.RecordAccess(req.UserId, req.AssetId, req.Timestamp)
	if err != nil {
		return nil, err
	}
	return &pb.TelemetryResponse{Success: true}, nil
}

func (h *TelemetryHandler) GetUsageStats(ctx context.Context, req *pb.GetUsageRequest) (*pb.UsageStats, error) {
	stats, err := h.usecase.GetUsageStats(req.AssetId, req.Period)
	if err != nil {
		return nil, err
	}
	return &pb.UsageStats{
		AssetId:         stats.AssetID,
		TotalAccesses:   stats.TotalAccesses,
		AverageDuration: stats.AverageDuration,
	}, nil
}

func (h *TelemetryHandler) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	ok := h.usecase.Heartbeat(req.DeviceId, req.Status)
	return &pb.HeartbeatResponse{Acknowledged: ok}, nil
}

func (h *TelemetryHandler) LogEvent(ctx context.Context, req *pb.LogEventRequest) (*pb.TelemetryResponse, error) {
	err := h.usecase.LogEvent(req.Source, req.Level, req.Message)
	if err != nil {
		return nil, err
	}
	return &pb.TelemetryResponse{Success: true}, nil
}
