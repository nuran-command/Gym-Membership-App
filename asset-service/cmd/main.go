package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	delivery "gym-membership/asset-service/internal/delivery/grpc"
	"gym-membership/asset-service/internal/repository"
	"gym-membership/asset-service/internal/usecase"
	pb "gym-membership/proto/asset"
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

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	repo := repository.NewPostgresAssetRepository(pool)
	uc := usecase.NewAssetUsecase(repo, rdb)
	handler := delivery.NewAssetHandler(uc)

	s := grpc.NewServer()
	pb.RegisterAssetServiceServer(s, handler)

	log.Printf("Asset Service (Phase 2) listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
