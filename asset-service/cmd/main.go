package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	delivery "github.com/ilnur/gym-membership-app/asset-service/internal/delivery/grpc"
	"github.com/ilnur/gym-membership-app/asset-service/internal/repository"
	"github.com/ilnur/gym-membership-app/asset-service/internal/usecase"
	pb "github.com/ilnur/gym-membership-app/proto/asset"
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
