package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	assetpb "gym-membership/proto/asset"
	pb "gym-membership/proto/membership"

	delivery "gym-membership/membership-service/internal/delivery/grpc"
	"gym-membership/membership-service/internal/repository"
	"gym-membership/membership-service/internal/usecase"
)

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	assetServiceAddr := os.Getenv("ASSET_SERVICE_ADDR")
	if assetServiceAddr == "" {
		assetServiceAddr = "localhost:50051"
	}

	conn, err := grpc.Dial(assetServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to asset service: %v", err)
	}
	defer conn.Close()

	assetClient := assetpb.NewAssetServiceClient(conn)

	repo := repository.NewInMemoryMembershipRepository()
	uc := usecase.NewMembershipUsecase(repo, assetClient)
	handler := delivery.NewMembershipHandler(uc)

	s := grpc.NewServer()
	pb.RegisterMembershipServiceServer(s, handler)

	log.Printf("Membership Service listening on :50052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
