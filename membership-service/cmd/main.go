package main

import (
	"log"
	"net"
	"os"

	"github.com/ilnur/gym-membership-app/proto/asset"
	"github.com/ilnur/gym-membership-app/proto/membership"

	"github.com/ilnur/gym-membership-app/membership-service/internal/delivery/grpc"
	"github.com/ilnur/gym-membership-app/membership-service/internal/repository/postgres"
	"github.com/ilnur/gym-membership-app/membership-service/internal/usecase"

	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	userRepo := postgres.NewUserRepo()
	bookingRepo := postgres.NewBookingRepo()
	creditRepo := postgres.NewCreditRepo()

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

	useCase := usecase.NewMembershipUseCase(userRepo, bookingRepo, creditRepo, assetClient)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := googlegrpc.NewServer()
	membership.RegisterMembershipServiceServer(s, grpc.NewMembershipHandler(useCase))

	log.Printf("Membership Service listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
