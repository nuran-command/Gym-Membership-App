package observability

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.uber.org/zap"
)

var (
	// Prometheus Metrics
	AssetsByStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "assets_by_status_total",
		Help: "The total number of assets by status",
	}, []string{"status"})

	AvgHealthScore = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "assets_avg_health_score",
		Help: "The average health score of all assets",
	})

	GrpcRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "The total number of gRPC requests",
	}, []string{"method", "status"})

	// Zap Logger
	Logger *zap.Logger
)

func InitLogger() {
	var err error
	Logger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
}

func InitTracer() (*trace.TracerProvider, error) {
	// For demo purposes, we export to stdout. 
	// In production, this would be an OTLP exporter to Jaeger/Collector.
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("asset-service"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}
