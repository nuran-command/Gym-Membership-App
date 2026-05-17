package grpc

import (
	"context"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/ilnur/gym-membership-app/proto/membership"
)

type MembershipHandler struct {
	membership.UnimplementedMembershipServiceServer
	useCase domain.MembershipUseCase
}

func NewMembershipHandler(useCase domain.MembershipUseCase) *MembershipHandler {
	return &MembershipHandler{
		useCase: useCase,
	}
}

func (h *MembershipHandler) CreateBooking(ctx context.Context, req *membership.CreateBookingRequest) (*membership.Booking, error) {
	booking, err := h.useCase.CreateBooking(ctx, req.UserId, req.AssetId, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}

	return &membership.Booking{
		Id:        booking.ID,
		UserId:    booking.UserID,
		AssetId:   booking.AssetID,
		StartTime: booking.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:   booking.EndTime.Format("2006-01-02 15:04:05"),
		Status:    booking.Status,
	}, nil
}

func (h *MembershipHandler) CancelBooking(ctx context.Context, req *membership.CancelBookingRequest) (*membership.Booking, error) {
	err := h.useCase.CancelBooking(ctx, req.BookingId)
	if err != nil {
		return nil, err
	}
	return &membership.Booking{Id: req.BookingId, Status: "Cancelled"}, nil
}

func (h *MembershipHandler) ReturnBooking(ctx context.Context, req *membership.ReturnBookingRequest) (*membership.Booking, error) {
	err := h.useCase.ReturnBooking(ctx, req.BookingId)
	if err != nil {
		return nil, err
	}
	return &membership.Booking{Id: req.BookingId, Status: "Returned"}, nil
}

func (h *MembershipHandler) GetUserCredits(ctx context.Context, req *membership.GetUserCreditsRequest) (*membership.CreditsResponse, error) {
	balance, err := h.useCase.GetUserCredits(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &membership.CreditsResponse{UserId: req.UserId, Balance: int32(balance)}, nil
}

func (h *MembershipHandler) DeductCredits(ctx context.Context, req *membership.DeductCreditsRequest) (*membership.CreditsResponse, error) {
	err := h.useCase.DeductCredits(ctx, req.UserId, int(req.Amount))
	if err != nil {
		return nil, err
	}
	balance, _ := h.useCase.GetUserCredits(ctx, req.UserId)
	return &membership.CreditsResponse{UserId: req.UserId, Balance: int32(balance)}, nil
}

func (h *MembershipHandler) AddCredits(ctx context.Context, req *membership.AddCreditsRequest) (*membership.CreditsResponse, error) {
	err := h.useCase.AddCredits(ctx, req.UserId, int(req.Amount))
	if err != nil {
		return nil, err
	}
	balance, _ := h.useCase.GetUserCredits(ctx, req.UserId)
	return &membership.CreditsResponse{UserId: req.UserId, Balance: int32(balance)}, nil
}

func (h *MembershipHandler) GetUserBookings(ctx context.Context, req *membership.GetUserBookingsRequest) (*membership.BookingList, error) {
	bookings, err := h.useCase.GetUserBookings(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	var pbBookings []*membership.Booking
	for _, b := range bookings {
		pbBookings = append(pbBookings, &membership.Booking{
			Id:        b.ID,
			UserId:    b.UserID,
			AssetId:   b.AssetID,
			StartTime: b.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:   b.EndTime.Format("2006-01-02 15:04:05"),
			Status:    b.Status,
		})
	}

	return &membership.BookingList{Bookings: pbBookings}, nil
}
