package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	delivery_grpc "gym-membership/asset-service/internal/delivery/grpc"
	delivery_nats "gym-membership/asset-service/internal/delivery/nats"
	"gym-membership/asset-service/internal/observability"
	"gym-membership/asset-service/internal/repository"
	"gym-membership/asset-service/internal/usecase"
	pb "gym-membership/proto/asset"
	"net/http"
)

func main() {
	// Initialize Postgres
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/assets?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Initialize Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Initialize NATS (Phase 3)
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Unable to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Phase 6: Initialize Observability
	observability.InitLogger()
	tp, err := observability.InitTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer tp.Shutdown(context.Background())

	// Start Prometheus Metrics Server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Prometheus metrics server starting on :9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Fatalf("failed to start metrics server: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	repo := repository.NewPostgresAssetRepository(pool)
	uc := usecase.NewAssetUsecase(repo, rdb, nc)
	handler := delivery_grpc.NewAssetHandler(uc)

	// Start NATS Subscriber (Phase 3)
	subscriber := delivery_nats.NewAssetSubscriber(nc, uc)
	if err := subscriber.Start(); err != nil {
		log.Fatalf("failed to start NATS subscriber: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAssetServiceServer(s, handler)

	log.Printf("Asset Service (Phase 3) listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
