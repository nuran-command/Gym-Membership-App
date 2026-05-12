package nats

import (
	"encoding/json"
	"gym-membership/asset-service/internal/domain"
	"log"

	"github.com/nats-io/nats.go"
)

type BookingEvent struct {
	AssetID       string  `json:"asset_id"`
	DurationHours float64 `json:"duration_hours"`
}

type AssetSubscriber struct {
	nc      *nats.Conn
	usecase domain.AssetUsecase
}

func NewAssetSubscriber(nc *nats.Conn, u domain.AssetUsecase) *AssetSubscriber {
	return &AssetSubscriber{
		nc:      nc,
		usecase: u,
	}
}

func (s *AssetSubscriber) Start() error {
	// Subscribe to booking.created
	_, err := s.nc.Subscribe("booking.created", func(m *nats.Msg) {
		var event BookingEvent
		if err := json.Unmarshal(m.Data, &event); err != nil {
			log.Printf("Error unmarshalling booking.created: %v", err)
			return
		}
		s.usecase.HandleBookingCreated(event.AssetID)
	})
	if err != nil {
		return err
	}

	// Subscribe to booking.returned
	_, err = s.nc.Subscribe("booking.returned", func(m *nats.Msg) {
		var event BookingEvent
		if err := json.Unmarshal(m.Data, &event); err != nil {
			log.Printf("Error unmarshalling booking.returned: %v", err)
			return
		}
		s.usecase.HandleBookingReturned(event.AssetID, event.DurationHours)
	})
	if err != nil {
		return err
	}

	// Subscribe to booking.cancelled
	_, err = s.nc.Subscribe("booking.cancelled", func(m *nats.Msg) {
		var event BookingEvent
		if err := json.Unmarshal(m.Data, &event); err != nil {
			log.Printf("Error unmarshalling booking.cancelled: %v", err)
			return
		}
		s.usecase.HandleBookingCancelled(event.AssetID)
	})
	if err != nil {
		return err
	}

	log.Printf("NATS Subscriber started (booking.*)")
	return nil
}
