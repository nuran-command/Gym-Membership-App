package grpc

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	googlegrpc "google.golang.org/grpc"
)

var (
	BookingOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "booking_operations_total",
			Help: "Total number of booking operations (create, cancel, return)",
		},
		[]string{"operation", "status"},
	)

	CreditOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "credit_operations_total",
			Help: "Total number of credit operations (deduct, add)",
		},
		[]string{"operation", "status"},
	)

	GrpcLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Latency of gRPC requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status"},
	)
)

func UnaryServerInterceptor() googlegrpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *googlegrpc.UnaryServerInfo, handler googlegrpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		tr := otel.Tracer("membership-service")
		ctx, span := tr.Start(ctx, info.FullMethod)
		defer span.End()

		resp, err := handler(ctx, req)

		duration := time.Since(startTime).Seconds()
		status := "success"
		if err != nil {
			status = "failure"
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "OK")
		}

		GrpcLatency.WithLabelValues(info.FullMethod, status).Observe(duration)

		switch info.FullMethod {
		case "/membership.MembershipService/CreateBooking":
			BookingOperationsTotal.WithLabelValues("create", status).Inc()
		case "/membership.MembershipService/CancelBooking":
			BookingOperationsTotal.WithLabelValues("cancel", status).Inc()
		case "/membership.MembershipService/ReturnBooking":
			BookingOperationsTotal.WithLabelValues("return", status).Inc()
		case "/membership.MembershipService/DeductCredits":
			CreditOperationsTotal.WithLabelValues("deduct", status).Inc()
		case "/membership.MembershipService/AddCredits":
			CreditOperationsTotal.WithLabelValues("add", status).Inc()
		}

		return resp, err
	}
}
