package grpc

import (
	"context"
	"gym-membership/membership-service/internal/domain"
	pb "gym-membership/proto/membership"
)

type MembershipHandler struct {
	pb.UnimplementedMembershipServiceServer
	usecase domain.MembershipUsecase
}

func NewMembershipHandler(u domain.MembershipUsecase) *MembershipHandler {
	return &MembershipHandler{
		usecase: u,
	}
}

func (h *MembershipHandler) CreateMembership(ctx context.Context, req *pb.CreateMembershipRequest) (*pb.Membership, error) {
	m, err := h.usecase.CreateMembership(req.UserId, req.AssetId, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	return &pb.Membership{
		Id:        m.ID,
		UserId:    m.UserID,
		AssetId:   m.AssetID,
		Status:    m.Status,
		StartDate: m.StartDate,
		EndDate:   m.EndDate,
	}, nil
}

func (h *MembershipHandler) GetMembership(ctx context.Context, req *pb.GetMembershipRequest) (*pb.Membership, error) {
	m, err := h.usecase.GetMembership(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.Membership{
		Id:        m.ID,
		UserId:    m.UserID,
		AssetId:   m.AssetID,
		Status:    m.Status,
		StartDate: m.StartDate,
		EndDate:   m.EndDate,
	}, nil
}

func (h *MembershipHandler) CancelMembership(ctx context.Context, req *pb.CancelMembershipRequest) (*pb.Membership, error) {
	m, err := h.usecase.CancelMembership(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.Membership{
		Id:        m.ID,
		UserId:    m.UserID,
		AssetId:   m.AssetID,
		Status:    m.Status,
		StartDate: m.StartDate,
		EndDate:   m.EndDate,
	}, nil
}

func (h *MembershipHandler) ValidateAccess(ctx context.Context, req *pb.ValidateAccessRequest) (*pb.ValidateAccessResponse, error) {
	ok, err := h.usecase.ValidateAccess(req.UserId, req.AssetId)
	if err != nil {
		return nil, err
	}
	return &pb.ValidateAccessResponse{HasAccess: ok}, nil
}
