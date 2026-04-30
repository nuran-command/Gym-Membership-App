package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	delivery "gym-membership/telemetry-service/internal/delivery/grpc"
	"gym-membership/telemetry-service/internal/repository"
	"gym-membership/telemetry-service/internal/usecase"
	pb "gym-membership/proto/telemetry"
)

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	repo := repository.NewInMemoryTelemetryRepository()
	uc := usecase.NewTelemetryUsecase(repo)
	handler := delivery.NewTelemetryHandler(uc)

	s := grpc.NewServer()
	pb.RegisterTelemetryServiceServer(s, handler)

	log.Printf("Telemetry Service listening on :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
