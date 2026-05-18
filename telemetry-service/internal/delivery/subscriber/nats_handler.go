package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"gym-membership/telemetry-service/internal/observability"
	"gym-membership/telemetry-service/internal/usecase"
)

type NatsHandler struct {
	nc          *nats.Conn
	uc          *usecase.TelemetryUsecase
	createdSub  *nats.Subscription
	returnedSub *nats.Subscription
	stopChan    chan struct{}
}

func NewNatsHandler(nc *nats.Conn, uc *usecase.TelemetryUsecase) *NatsHandler {
	return &NatsHandler{
		nc:       nc,
		uc:       uc,
		stopChan: make(chan struct{}),
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
	subCreated, err := h.nc.Subscribe("booking.created", func(m *nats.Msg) {
		tr := otel.Tracer("telemetry-service")
		ctx, span := tr.Start(context.Background(), "NATS booking.created")
		defer span.End()

		var p BookingPayload
		if err := json.Unmarshal(m.Data, &p); err != nil {
			slog.Error("Error unmarshaling booking.created",
				slog.String("error", err.Error()),
				slog.String("subject", m.Subject),
			)
			span.RecordError(err)
			observability.EventsProcessed.WithLabelValues("booking.created", "error").Inc()
			return
		}

		// Structured log with booking_id and user_id
		slog.Info("Processing NATS event",
			slog.String("event_type", "booking.created"),
			slog.String("booking_id", p.BookingID),
			slog.String("user_id", p.UserID),
			slog.String("asset_id", p.AssetID),
		)

		span.SetAttributes(
			attribute.String("messaging.system", "nats"),
			attribute.String("messaging.destination", m.Subject),
			attribute.String("booking_id", p.BookingID),
			attribute.String("user_id", p.UserID),
			attribute.String("asset_id", p.AssetID),
		)

		if err := h.uc.HandleBookingCreated(ctx, p.UserID, p.AssetID, p.BookingID); err != nil {
			slog.Error("Error handling booking.created",
				slog.String("error", err.Error()),
				slog.String("booking_id", p.BookingID),
				slog.String("user_id", p.UserID),
			)
			span.RecordError(err)
			observability.EventsProcessed.WithLabelValues("booking.created", "error").Inc()
		} else {
			slog.Info("Successfully handled booking.created",
				slog.String("booking_id", p.BookingID),
				slog.String("user_id", p.UserID),
			)
			observability.EventsProcessed.WithLabelValues("booking.created", "success").Inc()
		}

		// Update lag gauge immediately
		if h.createdSub != nil {
			if msgs, _, err := h.createdSub.Pending(); err == nil {
				observability.NatsConsumerLag.WithLabelValues("booking.created").Set(float64(msgs))
			}
		}
	})
	if err != nil {
		return fmt.Errorf("error subscribing to booking.created: %w", err)
	}
	h.createdSub = subCreated

	// Subscribe to booking.returned
	subReturned, err := h.nc.Subscribe("booking.returned", func(m *nats.Msg) {
		tr := otel.Tracer("telemetry-service")
		ctx, span := tr.Start(context.Background(), "NATS booking.returned")
		defer span.End()

		var p BookingPayload
		if err := json.Unmarshal(m.Data, &p); err != nil {
			slog.Error("Error unmarshaling booking.returned",
				slog.String("error", err.Error()),
				slog.String("subject", m.Subject),
			)
			span.RecordError(err)
			observability.EventsProcessed.WithLabelValues("booking.returned", "error").Inc()
			return
		}

		// Structured log with booking_id and user_id
		slog.Info("Processing NATS event",
			slog.String("event_type", "booking.returned"),
			slog.String("booking_id", p.BookingID),
			slog.String("user_id", p.UserID),
			slog.String("asset_id", p.AssetID),
			slog.Int("duration_minutes", p.DurationMinutes),
		)

		span.SetAttributes(
			attribute.String("messaging.system", "nats"),
			attribute.String("messaging.destination", m.Subject),
			attribute.String("booking_id", p.BookingID),
			attribute.String("user_id", p.UserID),
			attribute.String("asset_id", p.AssetID),
			attribute.Int("duration_minutes", p.DurationMinutes),
		)

		if err := h.uc.HandleBookingReturned(ctx, p.UserID, p.AssetID, p.BookingID, p.DurationMinutes); err != nil {
			slog.Error("Error handling booking.returned",
				slog.String("error", err.Error()),
				slog.String("booking_id", p.BookingID),
				slog.String("user_id", p.UserID),
			)
			span.RecordError(err)
			observability.EventsProcessed.WithLabelValues("booking.returned", "error").Inc()
		} else {
			slog.Info("Successfully handled booking.returned",
				slog.String("booking_id", p.BookingID),
				slog.String("user_id", p.UserID),
			)
			observability.EventsProcessed.WithLabelValues("booking.returned", "success").Inc()
		}

		// Update lag gauge immediately
		if h.returnedSub != nil {
			if msgs, _, err := h.returnedSub.Pending(); err == nil {
				observability.NatsConsumerLag.WithLabelValues("booking.returned").Set(float64(msgs))
			}
		}
	})
	if err != nil {
		return fmt.Errorf("error subscribing to booking.returned: %w", err)
	}
	h.returnedSub = subReturned

	// Start periodic lag monitoring loop in a background goroutine
	go h.monitorLag()

	return nil
}

func (h *NatsHandler) monitorLag() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopChan:
			return
		case <-ticker.C:
			if h.createdSub != nil && h.createdSub.IsValid() {
				if msgs, _, err := h.createdSub.Pending(); err == nil {
					observability.NatsConsumerLag.WithLabelValues("booking.created").Set(float64(msgs))
				}
			}
			if h.returnedSub != nil && h.returnedSub.IsValid() {
				if msgs, _, err := h.returnedSub.Pending(); err == nil {
					observability.NatsConsumerLag.WithLabelValues("booking.returned").Set(float64(msgs))
				}
			}
		}
	}
}

func (h *NatsHandler) Close() {
	close(h.stopChan)
	if h.createdSub != nil {
		_ = h.createdSub.Unsubscribe()
	}
	if h.returnedSub != nil {
		_ = h.returnedSub.Unsubscribe()
	}
}
