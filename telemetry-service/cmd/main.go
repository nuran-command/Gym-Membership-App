package main

import (
	"log"
	"net"

	"github.com/ilnur/gym-membership-app/proto/telemetry"

	"google.golang.org/grpc"
)

type telemetryServer struct {
	telemetry.UnimplementedTelemetryServiceServer
}

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	telemetry.RegisterTelemetryServiceServer(s, &telemetryServer{})
	log.Printf("Telemetry Service listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
