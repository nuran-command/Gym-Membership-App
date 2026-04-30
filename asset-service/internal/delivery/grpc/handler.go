package grpc

import (
	"context"
	"gym-membership/asset-service/internal/domain"
	pb "gym-membership/proto/asset"
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
	asset, err := h.usecase.GetAsset(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.Asset{
		Id:     asset.ID,
		Name:   asset.Name,
		Type:   asset.Type,
		Status: asset.Status,
	}, nil
}

func (h *AssetHandler) ListAvailableAssets(ctx context.Context, req *pb.ListRequest) (*pb.AssetList, error) {
	assets, err := h.usecase.ListAvailableAssets(req.Type)
	if err != nil {
		return nil, err
	}
	var pbAssets []*pb.Asset
	for _, a := range assets {
		pbAssets = append(pbAssets, &pb.Asset{
			Id:     a.ID,
			Name:   a.Name,
			Type:   a.Type,
			Status: a.Status,
		})
	}
	return &pb.AssetList{Assets: pbAssets}, nil
}

func (h *AssetHandler) UpdateAssetStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.Asset, error) {
	asset, err := h.usecase.UpdateAssetStatus(req.Id, req.Status)
	if err != nil {
		return nil, err
	}
	return &pb.Asset{
		Id:     asset.ID,
		Name:   asset.Name,
		Type:   asset.Type,
		Status: asset.Status,
	}, nil
}

func (h *AssetHandler) CheckAvailability(ctx context.Context, req *pb.CheckRequest) (*pb.AvailabilityResponse, error) {
	available, err := h.usecase.CheckAvailability(req.Id, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	return &pb.AvailabilityResponse{Available: available}, nil
}
