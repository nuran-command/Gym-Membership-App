package grpc_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/ilnur/gym-membership-app/membership-service/internal/repository/postgres"
	"github.com/ilnur/gym-membership-app/membership-service/internal/usecase"
	"github.com/ilnur/gym-membership-app/membership-service/migrations"
	"github.com/ilnur/gym-membership-app/proto/asset"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/testcontainers/testcontainers-go"
	pgtc "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	googlegrpc "google.golang.org/grpc"
)

// --- Mocks for Integration Test ---

type MockAssetServiceClient struct {
	mock.Mock
}

func (m *MockAssetServiceClient) CheckAvailability(ctx context.Context, in *asset.CheckRequest, opts ...googlegrpc.CallOption) (*asset.AvailabilityResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*asset.AvailabilityResponse), args.Error(1)
}

func (m *MockAssetServiceClient) GetAsset(ctx context.Context, in *asset.GetAssetRequest, opts ...googlegrpc.CallOption) (*asset.Asset, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*asset.Asset), args.Error(1)
}

func (m *MockAssetServiceClient) ListAvailableAssets(ctx context.Context, in *asset.ListRequest, opts ...googlegrpc.CallOption) (*asset.AssetList, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*asset.AssetList), args.Error(1)
}

func (m *MockAssetServiceClient) UpdateAssetStatus(ctx context.Context, in *asset.UpdateStatusRequest, opts ...googlegrpc.CallOption) (*asset.Asset, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*asset.Asset), args.Error(1)
}

func (m *MockAssetServiceClient) GetHealthScore(ctx context.Context, in *asset.GetAssetRequest, opts ...googlegrpc.CallOption) (*asset.HealthResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*asset.HealthResponse), args.Error(1)
}

type MockMessagePublisher struct {
	mock.Mock
}

func (m *MockMessagePublisher) PublishBookingCreated(ctx context.Context, bookingID, userID, assetID string, startTime, endTime string) error {
	args := m.Called(ctx, bookingID, userID, assetID, startTime, endTime)
	return args.Error(0)
}

func (m *MockMessagePublisher) PublishBookingCancelled(ctx context.Context, bookingID string) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}

func (m *MockMessagePublisher) PublishBookingReturned(ctx context.Context, bookingID string) error {
	args := m.Called(ctx, bookingID)
	return args.Error(0)
}

// --- Migration helper for Integration Test ---

func runMigrationsOnDb(db *sqlx.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INT PRIMARY KEY)`)
	if err != nil {
		return err
	}

	migrationFiles := []struct {
		version int
		name    string
	}{
		{1, "001_create_users.up.sql"},
		{2, "002_create_memberships.up.sql"},
		{3, "003_create_bookings.up.sql"},
		{4, "004_create_credit_transactions.up.sql"},
	}

	for _, m := range migrationFiles {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", m.version).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		content, err := migrations.FS.ReadFile(m.name)
		if err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", m.version)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		_ = tx.Commit()
	}
	return nil
}

func TestCreateBooking_Integration(t *testing.T) {
	// Skip the test gracefully if Docker is not available in the local execution environment
	if _, dockerSet := os.LookupEnv("DOCKER_HOST"); !dockerSet {
		// Quick heuristic: try to ping standard named pipe / socket
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{Image: "alpine"},
		})
		if err != nil {
			t.Skip("Skipping Integration Test because Docker daemon is not running or accessible")
			return
		}
	}

	ctx := context.Background()

	// 1. Setup Postgres Test Container
	pgContainer, err := pgtc.Run(ctx,
		"postgres:15-alpine",
		pgtc.WithDatabase("testdb"),
		pgtc.WithUsername("postgres"),
		pgtc.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		_ = pgContainer.Terminate(ctx)
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// 2. Connect to test DB
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}
	defer db.Close()

	// 3. Run migrations on container DB
	if err := runMigrationsOnDb(db); err != nil {
		t.Fatalf("failed to run migrations on test db: %v", err)
	}

	// 4. Setup repositories and usecase
	userRepo := postgres.NewUserRepo(db)
	bookingRepo := postgres.NewBookingRepo(db)
	creditRepo := postgres.NewCreditRepo(db)
	membershipRepo := postgres.NewMembershipRepo(db)
	txManager := postgres.NewTxManager(db)

	mockAsset := new(MockAssetServiceClient)
	mockPub := new(MockMessagePublisher)

	useCase := usecase.NewMembershipUseCase(userRepo, bookingRepo, creditRepo, membershipRepo, txManager, mockPub, mockAsset)

	t.Run("CreateBooking_InsufficientCredits_RollsBackTransaction", func(t *testing.T) {
		userID := "user-poor"
		assetID := "asset-poor"
		startTime := "2026-05-18T10:00:00Z"
		endTime := "2026-05-18T11:00:00Z"

		// Seed user with 5 credits
		_, err := db.Exec(`INSERT INTO users (id, name, email, credits) VALUES ($1, 'Poor User', 'poor@gym.com', 5)`, userID)
		assert.NoError(t, err)

		// Mock asset availability
		mockAsset.On("CheckAvailability", mock.Anything, &asset.CheckRequest{
			Id:        assetID,
			StartTime: startTime,
			EndTime:   endTime,
		}).Return(&asset.AvailabilityResponse{Available: true}, nil).Once()

		// Call CreateBooking (should abort due to 5 < 10 credits)
		booking, err := useCase.CreateBooking(ctx, userID, assetID, startTime, endTime)

		assert.Error(t, err)
		assert.Nil(t, booking)
		assert.Equal(t, "insufficient credits", err.Error())

		// Verify no booking was stored in database (rollback works)
		var bookingCount int
		err = db.Get(&bookingCount, "SELECT COUNT(*) FROM bookings WHERE user_id = $1", userID)
		assert.NoError(t, err)
		assert.Equal(t, 0, bookingCount)

		// Verify user credits are untouched
		var credits int
		err = db.Get(&credits, "SELECT credits FROM users WHERE id = $1", userID)
		assert.NoError(t, err)
		assert.Equal(t, 5, credits)

		mockAsset.AssertExpectations(t)
	})

	t.Run("CreateBooking_Success_CommitsTransaction", func(t *testing.T) {
		userID := "user-rich"
		assetID := "asset-rich"
		startTime := "2026-05-18T12:00:00Z"
		endTime := "2026-05-18T13:00:00Z"

		// Seed user with 20 credits
		_, err := db.Exec(`INSERT INTO users (id, name, email, credits) VALUES ($1, 'Rich User', 'rich@gym.com', 20)`, userID)
		assert.NoError(t, err)

		// Mock asset availability
		mockAsset.On("CheckAvailability", mock.Anything, &asset.CheckRequest{
			Id:        assetID,
			StartTime: startTime,
			EndTime:   endTime,
		}).Return(&asset.AvailabilityResponse{Available: true}, nil).Once()

		// Mock publisher
		mockPub.On("PublishBookingCreated", mock.Anything, mock.Anything, userID, assetID, startTime, endTime).Return(nil).Once()

		// Call CreateBooking
		booking, err := useCase.CreateBooking(ctx, userID, assetID, startTime, endTime)

		assert.NoError(t, err)
		assert.NotNil(t, booking)

		// Verify booking is present in database (commit worked)
		var storedBooking domain.Booking
		err = db.Get(&storedBooking, "SELECT id, user_id, asset_id, status FROM bookings WHERE user_id = $1", userID)
		assert.NoError(t, err)
		assert.Equal(t, booking.ID, storedBooking.ID)
		assert.Equal(t, "Confirmed", storedBooking.Status)

		// Verify user credits are deducted to 10
		var credits int
		err = db.Get(&credits, "SELECT credits FROM users WHERE id = $1", userID)
		assert.NoError(t, err)
		assert.Equal(t, 10, credits)

		// Verify transaction record was stored
		var txCount int
		err = db.Get(&txCount, "SELECT COUNT(*) FROM credit_transactions WHERE user_id = $1", userID)
		assert.NoError(t, err)
		assert.Equal(t, 1, txCount)

		mockAsset.AssertExpectations(t)
		mockPub.AssertExpectations(t)
	})
}
