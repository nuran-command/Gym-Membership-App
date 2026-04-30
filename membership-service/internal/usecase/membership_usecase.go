package usecase

import (
	"context"
	"errors"
	"fmt"
	"gym-membership/membership-service/internal/domain"
	assetpb "gym-membership/proto/asset" // Importing asset proto for the client
	"time"
)

type membershipUsecase struct {
	repo        domain.MembershipRepository
	assetClient assetpb.AssetServiceClient
}

func NewMembershipUsecase(repo domain.MembershipRepository, assetClient assetpb.AssetServiceClient) domain.MembershipUsecase {
	return &membershipUsecase{
		repo:        repo,
		assetClient: assetClient,
	}
}

func (u *membershipUsecase) CreateMembership(userID, assetID, start, end string) (*domain.Membership, error) {
	// Service B calls Service A по gRPC при создании бронирования
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := u.assetClient.CheckAvailability(ctx, &assetpb.CheckRequest{
		Id:        assetID,
		StartTime: start,
		EndTime:   end,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to check asset availability: %v", err)
	}

	if !resp.Available {
		return nil, errors.New("asset is not available for the selected time")
	}

	m := &domain.Membership{
		ID:        fmt.Sprintf("m-%d", time.Now().Unix()),
		UserID:    userID,
		AssetID:   assetID,
		Status:    "active",
		StartDate: start,
		EndDate:   end,
	}

	err = u.repo.Create(m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (u *membershipUsecase) GetMembership(id string) (*domain.Membership, error) {
	return u.repo.GetByID(id)
}

func (u *membershipUsecase) CancelMembership(id string) (*domain.Membership, error) {
	return u.repo.Cancel(id)
}

func (u *membershipUsecase) ValidateAccess(userID, assetID string) (bool, error) {
	return u.repo.ValidateAccess(userID, assetID)
}
