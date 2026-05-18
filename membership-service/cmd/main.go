package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ilnur/gym-membership-app/proto/asset"
	"github.com/ilnur/gym-membership-app/proto/membership"

	"github.com/ilnur/gym-membership-app/membership-service/internal/delivery/grpc"
	"github.com/ilnur/gym-membership-app/membership-service/internal/repository/nats"
	"github.com/ilnur/gym-membership-app/membership-service/internal/repository/postgres"
	"github.com/ilnur/gym-membership-app/membership-service/internal/repository/redis"
	"github.com/ilnur/gym-membership-app/membership-service/internal/usecase"
	"github.com/ilnur/gym-membership-app/membership-service/migrations"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	natsio "github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	redisio "github.com/redis/go-redis/v9"

	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func runMigrations(db *sqlx.DB) error {
	slog.Info("Running database migrations...")

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INT PRIMARY KEY)`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrationFiles := []struct {
		version int
		name    string
	}{
		{1, "001_create_users.up.sql"},
		{2, "002_create_memberships.up.sql"},
		{3, "003_create_bookings.up.sql"},
		{4, "004_create_credit_transactions.up.sql"},
	}

	for _, m := range migrationFiles {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", m.version).Scan(&exists)
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		slog.Info("Executing migration", "file", m.name)
		content, err := migrations.FS.ReadFile(m.name)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", m.name, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", m.name, err)
		}

		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", m.version)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
		slog.Info("Migration completed successfully", "version", m.version)
	}

	return nil
}

func initTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "jaeger:4317"
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otlpEndpoint),
	)
	if err != nil {
		slog.Warn("Failed to initialize OTLP exporter, running with standard tracer provider", "error", err)
		return sdktrace.NewTracerProvider(), nil
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("membership-service"),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

func main() {
	// Set structured logging default handler to slog (JSON logs)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting Membership Service...")

	// 1. Initialize OpenTelemetry Tracing
	tp, err := initTracer()
	if err == nil {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				slog.Error("Error shutting down Tracer Provider", "error", err)
			}
		}()
	}

	// 2. Connect to Postgres DB
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/membership?sslmode=disable"
	}

	var db *sqlx.DB
	for i := 0; i < 10; i++ {
		db, err = sqlx.Connect("postgres", dbURL)
		if err == nil {
			break
		}
		slog.Warn("Waiting for Postgres database connection...", "attempt", i+1, "error", err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}

	// 3. Connect to Redis Cache
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}
	rClient := redisio.NewClient(&redisio.Options{
		Addr: redisURL,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rClient.Ping(ctx).Err(); err != nil {
		slog.Warn("Could not ping Redis server, caching will operate on best effort", "error", err)
	}

	// 4. Connect to NATS Message Queue
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	var nc *natsio.Conn
	for i := 0; i < 10; i++ {
		nc, err = natsio.Connect(natsURL)
		if err == nil {
			break
		}
		slog.Warn("Waiting for NATS server connection...", "attempt", i+1, "error", err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// 5. Connect to Service A (Asset Service) via gRPC
	assetAddr := os.Getenv("ASSET_SERVICE_ADDR")
	if assetAddr == "" {
		assetAddr = "asset-service:50051"
	}
	assetConn, err := googlegrpc.Dial(assetAddr, googlegrpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to asset service: %v", err)
	}
	defer assetConn.Close()
	assetClient := asset.NewAssetServiceClient(assetConn)

	// 6. Setup Clean Architecture layers
	baseUserRepo := postgres.NewUserRepo(db)
	cachedUserRepo := redis.NewCachedUserRepo(baseUserRepo, rClient)

	bookingRepo := postgres.NewBookingRepo(db)
	creditRepo := postgres.NewCreditRepo(db)
	membershipRepo := postgres.NewMembershipRepo(db)
	txManager := postgres.NewTxManager(db)
	natsPub := nats.NewNatsPublisher(nc)

	useCase := usecase.NewMembershipUseCase(cachedUserRepo, bookingRepo, creditRepo, membershipRepo, txManager, natsPub, assetClient)

	// 7. Start Prometheus HTTP server
	go func() {
		slog.Info("Starting Prometheus metrics server on :2112/metrics")
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			slog.Error("Failed to start Prometheus metrics server", "error", err)
		}
	}()

	// 8. Start gRPC server
	port := os.Getenv("PORT")
	if port == "" {
		port = "50052"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := googlegrpc.NewServer(
		googlegrpc.UnaryInterceptor(grpc.UnaryServerInterceptor()),
	)
	membership.RegisterMembershipServiceServer(s, grpc.NewMembershipHandler(useCase))

	slog.Info("Membership Service is listening", "port", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
