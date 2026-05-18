package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/delivery/subscriber"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/repository/postgres"
	"github.com/ilnur/gym-membership-app/telemetry-service/internal/usecase"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	natsTC "github.com/testcontainers/testcontainers-go/modules/nats"
	postgresTC "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNATSToPostgresIntegration(t *testing.T) {
	ctx := context.Background()

	// 1. Start PostgreSQL container
	pgContainer, err := postgresTC.Run(ctx,
		"postgres:15-alpine",
		postgresTC.WithDatabase("testdb"),
		postgresTC.WithUsername("testuser"),
		postgresTC.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	defer db.Close()

	// Run migrations (simulate up.sql)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS usage_sessions (
			id SERIAL PRIMARY KEY,
			booking_id VARCHAR(255) NOT NULL UNIQUE,
			user_id VARCHAR(255) NOT NULL,
			asset_id VARCHAR(255) NOT NULL,
			started_at TIMESTAMP NOT NULL,
			ended_at TIMESTAMP,
			duration_minutes INT DEFAULT 0,
			email_sent BOOLEAN DEFAULT FALSE
		);
	`)
	require.NoError(t, err)

	// 2. Start NATS container
	natsContainer, err := natsTC.Run(ctx, "nats:latest")
	require.NoError(t, err)
	defer natsContainer.Terminate(ctx)

	natsURL, err := natsContainer.ConnectionString(ctx)
	require.NoError(t, err)

	nc, err := nats.Connect(natsURL)
	require.NoError(t, err)
	defer nc.Close()

	// 3. Initialize App layers
	repo := postgres.NewUsageSessionRepo(db)
	uc := usecase.NewTelemetryUsecase(repo, nil) // Mock email sender
	natsHandler := subscriber.NewNatsHandler(nc, uc)
	err = natsHandler.Subscribe()
	require.NoError(t, err)

	// 4. Test execution
	bookingID := "test-booking-xyz"
	userID := "test-user-1"
	assetID := "test-asset-1"

	// Publish booking.created
	createdPayload := map[string]interface{}{
		"user_id":    userID,
		"asset_id":   assetID,
		"booking_id": bookingID,
	}
	createdBytes, _ := json.Marshal(createdPayload)
	err = nc.Publish("booking.created", createdBytes)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond) // Give subscriber time to process

	// Verify it was created
	session, err := repo.GetByBookingID(ctx, bookingID)
	require.NoError(t, err)
	assert.Equal(t, userID, session.UserID)
	assert.Nil(t, session.EndedAt)

	// Publish booking.returned
	returnedPayload := map[string]interface{}{
		"user_id":          userID,
		"asset_id":         assetID,
		"booking_id":       bookingID,
		"duration_minutes": 120,
	}
	returnedBytes, _ := json.Marshal(returnedPayload)
	err = nc.Publish("booking.returned", returnedBytes)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	// Verify it was updated
	updatedSession, err := repo.GetByBookingID(ctx, bookingID)
	require.NoError(t, err)
	assert.NotNil(t, updatedSession.EndedAt)
	assert.Equal(t, 120, updatedSession.DurationMinutes)
}
