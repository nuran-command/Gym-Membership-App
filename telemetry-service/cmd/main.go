package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"gym-membership/telemetry-service/internal/delivery/subscriber"
	"gym-membership/telemetry-service/internal/repository"
	"gym-membership/telemetry-service/internal/usecase"
	// telemetry "gym-membership/telemetry-service/proto" // Uncomment after generation
)

func main() {
	// NATS connection
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()

	fmt.Printf("Connected to NATS at %s\n", natsURL)

	// Initialize layers
	repo := repository.NewInMemorySessionRepo()
	// Mock EmailSender for now
	uc := usecase.NewTelemetryUsecase(repo, nil)

	// Start NATS subscriber in a goroutine
	go func() {
		natsHandler := subscriber.NewNatsHandler(nc, uc)
		if err := natsHandler.Subscribe(); err != nil {
			log.Fatalf("Error subscribing to NATS: %v", err)
		}
		fmt.Println("NATS subscriber is running...")
	}()

	// Start gRPC server in a goroutine
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "50053"
		}
		lis, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := grpc.NewServer()
		// telemetry.RegisterTelemetryServiceServer(s, &grpcHandler{}) // Register after implementation

		fmt.Printf("gRPC server listening at %v\n", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	fmt.Println("Telemetry service is running (NATS + gRPC)...")

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down telemetry service...")
}
