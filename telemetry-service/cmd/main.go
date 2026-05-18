package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"log/slog"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/delivery/email"
	telemetryGrpc "github.com/ilnur/gym-membership-app/telemetry-service/internal/delivery/grpc"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/delivery/subscriber"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/repository/postgres"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/usecase"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/domain"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/observability"
	telemetry "github.com/ilnur/gym-membership-app/telemetry-service/proto"
)

func main() {
	// Initialize structured JSON logging globally as default
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	slog.Info("Starting telemetry service...")

	// Initialize OpenTelemetry Tracer
	_, cleanup, err := observability.InitTracer(context.Background(), "telemetry-service")
	if err != nil {
		slog.Error("Failed to initialize OpenTelemetry tracer", slog.String("error", err.Error()))
	} else {
		defer func() {
			if err := cleanup(context.Background()); err != nil {
				slog.Error("Failed to shutdown OpenTelemetry tracer", slog.String("error", err.Error()))
			}
		}()
		slog.Info("OpenTelemetry tracer initialized successfully")
	}

	// Start Prometheus HTTP metrics server in a goroutine
	go func() {
		metricsPort := os.Getenv("METRICS_PORT")
		if metricsPort == "" {
			metricsPort = "2112"
		}
		slog.Info("Starting Prometheus metrics HTTP server", slog.String("port", metricsPort))
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			slog.Error("Failed to serve Prometheus metrics", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// NATS connection
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		slog.Error("Error connecting to NATS", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer nc.Close()

	slog.Info("Connected to NATS", slog.String("url", natsURL))

	// Initialize DB
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/telemetry?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("Failed to open DB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("Failed to ping DB", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize Email Sender
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	var emailSender domain.EmailSender
	if smtpHost != "" && smtpUser != "" {
		emailSender = email.NewSMTPEmailSender(smtpHost, smtpPort, smtpUser, smtpPass)
	}

	// Initialize layers
	repo := postgres.NewUsageSessionRepo(db)
	uc := usecase.NewTelemetryUsecase(repo, emailSender)

	var natsHandler *subscriber.NatsHandler

	// Start NATS subscriber in a goroutine
	go func() {
		natsHandler = subscriber.NewNatsHandler(nc, uc)
		if err := natsHandler.Subscribe(); err != nil {
			slog.Error("Error subscribing to NATS", slog.String("error", err.Error()))
			os.Exit(1)
		}
		slog.Info("NATS subscriber is running...")
	}()
	defer func() {
		if natsHandler != nil {
			natsHandler.Close()
		}
	}()

	// Start gRPC server in a goroutine
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "50053"
		}
		lis, err := net.Listen("tcp", ":"+port)
		if err != nil {
			slog.Error("failed to listen", slog.String("error", err.Error()))
			os.Exit(1)
		}

		s := grpc.NewServer()
		grpcHandler := telemetryGrpc.NewTelemetryHandler(uc)
		telemetry.RegisterTelemetryServiceServer(s, grpcHandler)

		slog.Info("gRPC server listening", slog.String("address", lis.Addr().String()))

		if err := s.Serve(lis); err != nil {
			slog.Error("failed to serve", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	slog.Info("Telemetry service is running (NATS + gRPC + Metrics)...")

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down telemetry service...")
}
