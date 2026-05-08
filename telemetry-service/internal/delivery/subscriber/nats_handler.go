package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"gym-membership/telemetry-service/internal/usecase"
)

type NatsHandler struct {
	nc *nats.Conn
	uc *usecase.TelemetryUsecase
}

func NewNatsHandler(nc *nats.Conn, uc *usecase.TelemetryUsecase) *NatsHandler {
	return &NatsHandler{
		nc: nc,
		uc: uc,
	}
}

type BookingPayload struct {
	UserID          string `json:"user_id"`
	AssetID         string `json:"asset_id"`
	BookingID       string `json:"booking_id"`
	DurationMinutes int    `json:"duration_minutes"`
}

func (h *NatsHandler) Subscribe() error {
	// Subscribe to booking.created
	_, err := h.nc.Subscribe("booking.created", func(m *nats.Msg) {
		var p BookingPayload
		if err := json.Unmarshal(m.Data, &p); err != nil {
			log.Printf("Error unmarshaling booking.created: %v", err)
			return
		}

		if err := h.uc.HandleBookingCreated(context.Background(), p.UserID, p.AssetID, p.BookingID); err != nil {
			log.Printf("Error handling booking.created: %v", err)
		} else {
			log.Printf("Successfully handled booking.created for booking %s", p.BookingID)
		}
	})
	if err != nil {
		return fmt.Errorf("error subscribing to booking.created: %w", err)
	}

	// Subscribe to booking.returned
	_, err = h.nc.Subscribe("booking.returned", func(m *nats.Msg) {
		var p BookingPayload
		if err := json.Unmarshal(m.Data, &p); err != nil {
			log.Printf("Error unmarshaling booking.returned: %v", err)
			return
		}

		if err := h.uc.HandleBookingReturned(context.Background(), p.UserID, p.AssetID, p.BookingID, p.DurationMinutes); err != nil {
			log.Printf("Error handling booking.returned: %v", err)
		} else {
			log.Printf("Successfully handled booking.returned for booking %s", p.BookingID)
		}
	})
	if err != nil {
		return fmt.Errorf("error subscribing to booking.returned: %w", err)
	}

	return nil
}
