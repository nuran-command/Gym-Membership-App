package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/nats-io/nats.go"
)

type natsPublisher struct {
	nc *nats.Conn
}

func NewNatsPublisher(nc *nats.Conn) domain.MessagePublisher {
	return &natsPublisher{nc: nc}
}

type BookingCreatedPayload struct {
	UserID    string `json:"user_id"`
	AssetID   string `json:"asset_id"`
	BookingID string `json:"booking_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type BookingCancelledPayload struct {
	BookingID string `json:"booking_id"`
}

type BookingReturnedPayload struct {
	BookingID string `json:"booking_id"`
}

func (p *natsPublisher) PublishBookingCreated(ctx context.Context, bookingID, userID, assetID string, startTime, endTime string) error {
	payload := BookingCreatedPayload{
		UserID:    userID,
		AssetID:   assetID,
		BookingID: bookingID,
		StartTime: startTime,
		EndTime:   endTime,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal booking created payload: %w", err)
	}
	return p.nc.Publish("booking.created", data)
}

func (p *natsPublisher) PublishBookingCancelled(ctx context.Context, bookingID string) error {
	payload := BookingCancelledPayload{
		BookingID: bookingID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal booking cancelled payload: %w", err)
	}
	return p.nc.Publish("booking.cancelled", data)
}

func (p *natsPublisher) PublishBookingReturned(ctx context.Context, bookingID string) error {
	payload := BookingReturnedPayload{
		BookingID: bookingID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal booking returned payload: %w", err)
	}
	return p.nc.Publish("booking.returned", data)
}
