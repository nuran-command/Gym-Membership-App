package observability

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var (
	// EventsProcessed tracks the total number of processed NATS events.
	EventsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telemetry_events_processed_total",
			Help: "Total number of NATS events processed by telemetry-service.",
		},
		[]string{"event_type", "status"}, // event_type e.g. booking.created, booking.returned; status e.g. success, error
	)

	// EmailsSent tracks the total number of emails sent.
	EmailsSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "telemetry_emails_sent_total",
			Help: "Total number of thank-you emails sent by telemetry-service.",
		},
		[]string{"status"}, // status e.g. success, error
	)

	// NatsConsumerLag tracks the NATS subscription consumer lag (client-side pending messages).
	NatsConsumerLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "telemetry_nats_consumer_lag",
			Help: "Number of pending messages in the client subscription queue.",
		},
		[]string{"subject"},
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(EventsProcessed)
	prometheus.MustRegister(EmailsSent)
	prometheus.MustRegister(NatsConsumerLag)
}

// InitTracer initializes OpenTelemetry tracer provider with a stdout exporter.
func InitTracer(ctx context.Context, serviceName string) (*sdktrace.TracerProvider, func(context.Context) error, error) {
	// For production, this could write to Jaeger, Zipkin, or OTLP collector.
	// For high-quality observability here, we use a stdouttrace exporter writing to a trace log file or standard out/err.
	var w io.Writer = os.Stdout
	
	// Optional: write to a file or stdout based on env
	if logFile := os.Getenv("OTEL_LOG_FILE"); logFile != "" {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			w = f
		}
	}

	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(w),
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

	cleanup := func(cleanupCtx context.Context) error {
		return tp.Shutdown(cleanupCtx)
	}

	return tp, cleanup, nil
}
