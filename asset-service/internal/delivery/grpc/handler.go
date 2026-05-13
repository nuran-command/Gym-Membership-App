package grpc

import (
	"context"
	"gym-membership/asset-service/internal/domain"
	"gym-membership/asset-service/internal/observability"
	pb "gym-membership/proto/asset"

	"go.opentelemetry.io/otel"
)

type AssetHandler struct {
	pb.UnimplementedAssetServiceServer
	usecase domain.AssetUsecase
}

func NewAssetHandler(u domain.AssetUsecase) *AssetHandler {
	return &AssetHandler{
		usecase: u,
	}
}

func (h *AssetHandler) GetAsset(ctx context.Context, req *pb.GetAssetRequest) (*pb.Asset, error) {
	observability.GrpcRequestsTotal.WithLabelValues("GetAsset", "started").Inc()
	asset, err := h.usecase.GetAsset(req.Id)
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("GetAsset", "error").Inc()
		return nil, err
	}
	observability.GrpcRequestsTotal.WithLabelValues("GetAsset", "success").Inc()
	return &pb.Asset{
		Id:          asset.ID,
		Name:        asset.Name,
		Type:        asset.Type,
		Status:      asset.Status,
		HealthScore: int32(asset.HealthScore),
	}, nil
}

func (h *AssetHandler) ListAvailableAssets(ctx context.Context, req *pb.ListRequest) (*pb.AssetList, error) {
	observability.GrpcRequestsTotal.WithLabelValues("ListAvailableAssets", "started").Inc()
	assets, err := h.usecase.ListAvailableAssets(req.Type)
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("ListAvailableAssets", "error").Inc()
		return nil, err
	}
	var pbAssets []*pb.Asset
	for _, a := range assets {
		pbAssets = append(pbAssets, &pb.Asset{
			Id:          a.ID,
			Name:        a.Name,
			Type:        a.Type,
			Status:      a.Status,
			HealthScore: int32(a.HealthScore),
		})
	}
	observability.GrpcRequestsTotal.WithLabelValues("ListAvailableAssets", "success").Inc()
	return &pb.AssetList{Assets: pbAssets}, nil
}

func (h *AssetHandler) UpdateAssetStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.Asset, error) {
	observability.GrpcRequestsTotal.WithLabelValues("UpdateAssetStatus", "started").Inc()
	asset, err := h.usecase.UpdateAssetStatus(req.Id, req.Status)
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("UpdateAssetStatus", "error").Inc()
		return nil, err
	}
	observability.GrpcRequestsTotal.WithLabelValues("UpdateAssetStatus", "success").Inc()
	return &pb.Asset{
		Id:          asset.ID,
		Name:        asset.Name,
		Type:        asset.Type,
		Status:      asset.Status,
		HealthScore: int32(asset.HealthScore),
	}, nil
}

func (h *AssetHandler) CheckAvailability(ctx context.Context, req *pb.CheckRequest) (*pb.AvailabilityResponse, error) {
	observability.GrpcRequestsTotal.WithLabelValues("CheckAvailability", "started").Inc()
	
	// Phase 6: OpenTelemetry span
	tracer := otel.Tracer("asset-service")
	ctx, span := tracer.Start(ctx, "CheckAvailability")
	defer span.End()

	available, err := h.usecase.CheckAvailability(req.Id, req.StartTime, req.EndTime)
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("CheckAvailability", "error").Inc()
		return nil, err
	}
	observability.GrpcRequestsTotal.WithLabelValues("CheckAvailability", "success").Inc()
	return &pb.AvailabilityResponse{Available: available}, nil
}

func (h *AssetHandler) GetHealthScore(ctx context.Context, req *pb.GetAssetRequest) (*pb.HealthResponse, error) {
	observability.GrpcRequestsTotal.WithLabelValues("GetHealthScore", "started").Inc()
	asset, err := h.usecase.GetAsset(req.Id)
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("GetHealthScore", "error").Inc()
		return nil, err
	}
	observability.GrpcRequestsTotal.WithLabelValues("GetHealthScore", "success").Inc()
	return &pb.HealthResponse{
		HealthScore: int32(asset.HealthScore),
	}, nil
}
