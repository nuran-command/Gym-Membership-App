package grpc

import (
	"gym-membership/telemetry-service/internal/usecase"
)

type TelemetryHandler struct {
	// telemetry.UnimplementedTelemetryServiceServer // Uncomment after generation
	uc *usecase.TelemetryUsecase
}

func NewTelemetryHandler(uc *usecase.TelemetryUsecase) *TelemetryHandler {
	return &TelemetryHandler{
		uc: uc,
	}
}

// Add gRPC methods here after proto generation
// Example:
// func (h *TelemetryHandler) GetUsageSession(ctx context.Context, req *telemetry.GetUsageSessionRequest) (*telemetry.UsageSessionResponse, error) {
//     ...
// }
