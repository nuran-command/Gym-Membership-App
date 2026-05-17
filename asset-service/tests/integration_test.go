//go:build integration
package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gym-membership/asset-service/internal/repository"
	"gym-membership/asset-service/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgres(ctx context.Context) (*postgres.PostgresContainer, *pgxpool.Pool, error) {
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("assets_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, nil, err
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, nil, err
	}

	return pgContainer, pool, nil
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Simple migration runner for tests
	migrationDir := "../migrations"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" && (f.Name()[len(f.Name())-7:] == ".up.sql" || f.Name() == "002_seed_assets.up.sql") {
			content, err := os.ReadFile(filepath.Join(migrationDir, f.Name()))
			if err != nil {
				return err
			}
			_, err = pool.Exec(ctx, string(content))
			if err != nil {
				return fmt.Errorf("error running migration %s: %v", f.Name(), err)
			}
		}
	}
	return nil
}

func TestAssetFlow_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	pgContainer, pool, err := setupPostgres(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	err = runMigrations(ctx, pool)
	assert.NoError(t, err)

	repo := repository.NewPostgresAssetRepository(pool)
	// We'll use nil for redis/nats in this integration test for simplicity,
	// focusing on the Postgres flow as requested.
	uc := usecase.NewAssetUsecase(repo, nil, nil)

	// 1. Create/Check asset from seed
	assets, err := uc.ListAvailableAssets("cardio")
	assert.NoError(t, err)
	assert.NotEmpty(t, assets)

	assetID := assets[0].ID

	// 2. Book asset (HandleBookingCreated)
	err = uc.HandleBookingCreated(assetID)
	assert.NoError(t, err)

	// Check status
	asset, err := uc.GetAsset(assetID)
	assert.NoError(t, err)
	assert.Equal(t, "in_use", asset.Status)

	// 3. Return asset (HandleBookingReturned) with 3 hours usage
	err = uc.HandleBookingReturned(assetID, 3.0)
	assert.NoError(t, err)

	// Check status and health
	updatedAsset, err := uc.GetAsset(assetID)
	assert.NoError(t, err)
	assert.Equal(t, "available", updatedAsset.Status)
	assert.Equal(t, asset.HealthScore-5, updatedAsset.HealthScore)

	log.Printf("Integration test passed for asset %s", assetID)
}
