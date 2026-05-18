package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	fmt.Println("Connected to NATS. Starting simulation...")

	assets := []string{"treadmill-1", "treadmill-2", "bench-press-1", "dumbbell-set-1", "rowing-machine-1"}
	users := []string{"user-A", "user-B", "user-C", "user-D"}

	for i := 1; i <= 20; i++ {
		bookingID := fmt.Sprintf("sim-booking-%d-%d", time.Now().Unix(), i)
		userID := users[rand.Intn(len(users))]
		assetID := assets[rand.Intn(len(assets))]
		duration := rand.Intn(120) + 15 // 15 to 135 minutes

		// 1. Publish booking.created
		createdPayload := map[string]interface{}{
			"user_id":    userID,
			"asset_id":   assetID,
			"booking_id": bookingID,
		}
		cb, _ := json.Marshal(createdPayload)
		nc.Publish("booking.created", cb)
		fmt.Printf("[Created] Booking: %s, User: %s, Asset: %s\n", bookingID, userID, assetID)

		// Wait briefly to simulate usage passing
		time.Sleep(100 * time.Millisecond)

		// 2. Publish booking.returned
		returnedPayload := map[string]interface{}{
			"user_id":          userID,
			"asset_id":         assetID,
			"booking_id":       bookingID,
			"duration_minutes": duration,
		}
		rb, _ := json.Marshal(returnedPayload)
		nc.Publish("booking.returned", rb)
		fmt.Printf("[Returned] Booking: %s, Duration: %d min\n", bookingID, duration)
		
		time.Sleep(400 * time.Millisecond)
	}

	fmt.Println("Simulation finished.")
}
