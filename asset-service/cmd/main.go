package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	delivery "gym-membership/asset-service/internal/delivery/grpc"
	"gym-membership/asset-service/internal/repository"
	"gym-membership/asset-service/internal/usecase"
	pb "gym-membership/proto/asset"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	repo := repository.NewInMemoryAssetRepository()
	uc := usecase.NewAssetUsecase(repo)
	handler := delivery.NewAssetHandler(uc)

	s := grpc.NewServer()
	pb.RegisterAssetServiceServer(s, handler)

	log.Printf("Asset Service listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
