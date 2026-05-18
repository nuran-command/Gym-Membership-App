package grpc

import (
	"context"
	"gym-membership/asset-service/internal/domain"
	"gym-membership/asset-service/internal/observability"
	pb "gym-membership/proto/asset"
	"time"

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

func toPbAsset(a *domain.Asset) *pb.Asset {
	if a == nil {
		return nil
	}
	return &pb.Asset{
		Id:               a.ID,
		Name:             a.Name,
		Type:             a.Type,
		Status:           a.Status,
		HealthScore:      int32(a.HealthScore),
		Location:         a.Location,
		CreatedAt:        a.CreatedAt.Format(time.RFC3339),
		LastMaintainedAt: a.LastMaintainedAt.Format(time.RFC3339),
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
	return toPbAsset(asset), nil
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
		pbAssets = append(pbAssets, toPbAsset(a))
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
	return toPbAsset(asset), nil
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

func (h *AssetHandler) CreateAsset(ctx context.Context, req *pb.CreateAssetRequest) (*pb.Asset, error) {
	observability.GrpcRequestsTotal.WithLabelValues("CreateAsset", "started").Inc()
	asset, err := h.usecase.CreateAsset(req.Name, req.Type, req.Status, req.Location, int(req.HealthScore))
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("CreateAsset", "error").Inc()
		return nil, err
	}
	observability.GrpcRequestsTotal.WithLabelValues("CreateAsset", "success").Inc()
	return toPbAsset(asset), nil
}

func (h *AssetHandler) UpdateAsset(ctx context.Context, req *pb.UpdateAssetRequest) (*pb.Asset, error) {
	observability.GrpcRequestsTotal.WithLabelValues("UpdateAsset", "started").Inc()
	asset, err := h.usecase.UpdateAsset(req.Id, req.Name, req.Type, req.Location)
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("UpdateAsset", "error").Inc()
		return nil, err
	}
	observability.GrpcRequestsTotal.WithLabelValues("UpdateAsset", "success").Inc()
	return toPbAsset(asset), nil
}

func (h *AssetHandler) DeleteAsset(ctx context.Context, req *pb.DeleteAssetRequest) (*pb.DeleteAssetResponse, error) {
	observability.GrpcRequestsTotal.WithLabelValues("DeleteAsset", "started").Inc()
	err := h.usecase.DeleteAsset(req.Id)
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("DeleteAsset", "error").Inc()
		return nil, err
	}
	observability.GrpcRequestsTotal.WithLabelValues("DeleteAsset", "success").Inc()
	return &pb.DeleteAssetResponse{Id: req.Id, Success: true}, nil
}

func (h *AssetHandler) ReportDamage(ctx context.Context, req *pb.ReportDamageRequest) (*pb.Asset, error) {
	observability.GrpcRequestsTotal.WithLabelValues("ReportDamage", "started").Inc()
	asset, err := h.usecase.ReportDamage(req.Id, int(req.DamageAmount))
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("ReportDamage", "error").Inc()
		return nil, err
	}
	observability.GrpcRequestsTotal.WithLabelValues("ReportDamage", "success").Inc()
	return toPbAsset(asset), nil
}

func (h *AssetHandler) ResolveMaintenance(ctx context.Context, req *pb.GetAssetRequest) (*pb.Asset, error) {
	observability.GrpcRequestsTotal.WithLabelValues("ResolveMaintenance", "started").Inc()
	asset, err := h.usecase.ResolveMaintenance(req.Id)
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("ResolveMaintenance", "error").Inc()
		return nil, err
	}
	observability.GrpcRequestsTotal.WithLabelValues("ResolveMaintenance", "success").Inc()
	return toPbAsset(asset), nil
}

func (h *AssetHandler) ListAllAssets(ctx context.Context, req *pb.ListAllRequest) (*pb.AssetList, error) {
	observability.GrpcRequestsTotal.WithLabelValues("ListAllAssets", "started").Inc()
	assets, err := h.usecase.ListAllAssets(req.Type)
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("ListAllAssets", "error").Inc()
		return nil, err
	}
	var pbAssets []*pb.Asset
	for _, a := range assets {
		pbAssets = append(pbAssets, toPbAsset(a))
	}
	observability.GrpcRequestsTotal.WithLabelValues("ListAllAssets", "success").Inc()
	return &pb.AssetList{Assets: pbAssets}, nil
}

func (h *AssetHandler) BatchCreateAssets(ctx context.Context, req *pb.BatchCreateRequest) (*pb.BatchCreateResponse, error) {
	observability.GrpcRequestsTotal.WithLabelValues("BatchCreateAssets", "started").Inc()
	var domainAssets []*domain.Asset
	for _, a := range req.Assets {
		domainAssets = append(domainAssets, &domain.Asset{
			Name:        a.Name,
			Type:        a.Type,
			Status:      a.Status,
			HealthScore: int(a.HealthScore),
			Location:    a.Location,
		})
	}
	createdAssets, err := h.usecase.BatchCreateAssets(domainAssets)
	if err != nil {
		observability.GrpcRequestsTotal.WithLabelValues("BatchCreateAssets", "error").Inc()
		return nil, err
	}
	var pbAssets []*pb.Asset
	for _, a := range createdAssets {
		pbAssets = append(pbAssets, toPbAsset(a))
	}
	observability.GrpcRequestsTotal.WithLabelValues("BatchCreateAssets", "success").Inc()
	return &pb.BatchCreateResponse{Assets: pbAssets}, nil
}
