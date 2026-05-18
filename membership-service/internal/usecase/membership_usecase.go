package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/ilnur/gym-membership-app/proto/asset"
)

type membershipUseCase struct {
	userRepo       domain.UserRepo
	bookingRepo    domain.BookingRepo
	creditRepo     domain.CreditRepo
	membershipRepo domain.MembershipRepo
	txManager      domain.TxManager
	publisher      domain.MessagePublisher
	assetClient    asset.AssetServiceClient
}

func NewMembershipUseCase(
	userRepo domain.UserRepo,
	bookingRepo domain.BookingRepo,
	creditRepo domain.CreditRepo,
	membershipRepo domain.MembershipRepo,
	txManager domain.TxManager,
	publisher domain.MessagePublisher,
	assetClient asset.AssetServiceClient,
) domain.MembershipUseCase {
	return &membershipUseCase{
		userRepo:       userRepo,
		bookingRepo:    bookingRepo,
		creditRepo:     creditRepo,
		membershipRepo: membershipRepo,
		txManager:      txManager,
		publisher:      publisher,
		assetClient:    assetClient,
	}
}

func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02T15:04:05", s)
}

func (u *membershipUseCase) CreateBooking(ctx context.Context, userID, assetID string, startTime, endTime string) (*domain.Booking, error) {
	slog.Info("CreateBooking request received", "user_id", userID, "asset_id", assetID, "start_time", startTime, "end_time", endTime)

	// 1. Call Service A (Asset Service) via gRPC to check availability
	resp, err := u.assetClient.CheckAvailability(ctx, &asset.CheckRequest{
		Id:        assetID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		slog.Error("Failed to check asset availability from Service A", "error", err)
		return nil, fmt.Errorf("failed to check asset availability: %w", err)
	}

	if !resp.Available {
		slog.Warn("Asset is not available for booking", "asset_id", assetID)
		return nil, errors.New("asset is not available for the selected time")
	}

	// 2. Parse times
	start, err := parseTime(startTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format: %w", err)
	}
	end, err := parseTime(endTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format: %w", err)
	}

	// 3. Get user and verify credits before transaction starts
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		slog.Error("Failed to fetch user", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.Credits < 10 {
		slog.Warn("Insufficient credits for booking", "user_id", userID, "credits", user.Credits)
		return nil, errors.New("insufficient credits")
	}

	booking := &domain.Booking{
		ID:        uuid.New().String(),
		UserID:    userID,
		AssetID:   assetID,
		StartTime: start,
		EndTime:   end,
		Status:    "Confirmed",
	}

	// 4. Atomic transaction: Deduct credits + Create transaction log + Create booking
	err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Deduct 10 credits
		err := u.userRepo.UpdateCredits(txCtx, userID, -10)
		if err != nil {
			return fmt.Errorf("failed to deduct credits: %w", err)
		}

		// Log credit transaction
		txLog := &domain.CreditTransaction{
			ID:        uuid.New().String(),
			UserID:    userID,
			Amount:    -10,
			Type:      "DEDUCTION",
			Reason:    fmt.Sprintf("Booking Confirmed: %s", booking.ID),
			CreatedAt: time.Now(),
		}
		err = u.creditRepo.CreateTransaction(txCtx, txLog)
		if err != nil {
			return fmt.Errorf("failed to log credit transaction: %w", err)
		}

		// Save booking
		err = u.bookingRepo.Create(txCtx, booking)
		if err != nil {
			return fmt.Errorf("failed to save booking: %w", err)
		}

		return nil
	})

	if err != nil {
		slog.Error("CreateBooking transaction aborted", "error", err)
		return nil, err
	}

	slog.Info("CreateBooking transaction successful", "booking_id", booking.ID)

	// 5. Publish event to NATS after successful commit
	if err := u.publisher.PublishBookingCreated(ctx, booking.ID, userID, assetID, startTime, endTime); err != nil {
		slog.Error("Failed to publish booking.created event to NATS", "error", err)
	} else {
		slog.Info("Published booking.created event to NATS", "booking_id", booking.ID)
	}

	return booking, nil
}

func (u *membershipUseCase) CancelBooking(ctx context.Context, bookingID string) error {
	slog.Info("CancelBooking request received", "booking_id", bookingID)

	// 1. Fetch booking details
	booking, err := u.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		slog.Error("Booking not found", "booking_id", bookingID, "error", err)
		return err
	}

	if booking.Status != "Confirmed" {
		slog.Warn("Booking cannot be cancelled because its status is not Confirmed", "booking_id", bookingID, "status", booking.Status)
		return errors.New("booking is not in Confirmed state")
	}

	// 2. Atomic transaction: Update booking status + Refund credits + Log credit transaction
	err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Update booking status
		err := u.bookingRepo.UpdateStatus(txCtx, bookingID, "Cancelled")
		if err != nil {
			return err
		}

		// Refund 10 credits to the user
		err = u.userRepo.UpdateCredits(txCtx, booking.UserID, 10)
		if err != nil {
			return err
		}

		// Log refund transaction
		txLog := &domain.CreditTransaction{
			ID:        uuid.New().String(),
			UserID:    booking.UserID,
			Amount:    10,
			Type:      "REFUND",
			Reason:    fmt.Sprintf("Booking Cancelled: %s", bookingID),
			CreatedAt: time.Now(),
		}
		err = u.creditRepo.CreateTransaction(txCtx, txLog)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		slog.Error("CancelBooking transaction aborted", "error", err)
		return err
	}

	slog.Info("CancelBooking transaction successful", "booking_id", bookingID)

	// 3. Publish event to NATS
	if err := u.publisher.PublishBookingCancelled(ctx, bookingID, booking.AssetID); err != nil {
		slog.Error("Failed to publish booking.cancelled event to NATS", "error", err)
	} else {
		slog.Info("Published booking.cancelled event to NATS", "booking_id", bookingID)
	}

	return nil
}

func (u *membershipUseCase) ReturnBooking(ctx context.Context, bookingID string) error {
	slog.Info("ReturnBooking request received", "booking_id", bookingID)

	// 1. Fetch booking
	booking, err := u.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		slog.Error("Booking not found", "booking_id", bookingID, "error", err)
		return err
	}

	if booking.Status != "Confirmed" {
		slog.Warn("Booking cannot be returned because its status is not Confirmed", "booking_id", bookingID, "status", booking.Status)
		return errors.New("booking is not in Confirmed state")
	}

	// 2. Atomic transaction: Update booking status to Returned
	err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		return u.bookingRepo.UpdateStatus(txCtx, bookingID, "Returned")
	})

	if err != nil {
		slog.Error("ReturnBooking transaction aborted", "error", err)
		return err
	}

	slog.Info("ReturnBooking transaction successful", "booking_id", bookingID)

	durationHours := time.Since(booking.StartTime).Hours()
	if durationHours < 0 {
		durationHours = 0
	}

	// 3. Publish event to NATS
	if err := u.publisher.PublishBookingReturned(ctx, bookingID, booking.AssetID, durationHours); err != nil {
		slog.Error("Failed to publish booking.returned event to NATS", "error", err)
	} else {
		slog.Info("Published booking.returned event to NATS", "booking_id", bookingID)
	}

	return nil
}

func (u *membershipUseCase) GetUserCredits(ctx context.Context, userID string) (int, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.Credits, nil
}

func (u *membershipUseCase) DeductCredits(ctx context.Context, userID string, amount int) error {
	slog.Info("DeductCredits request received", "user_id", userID, "amount", amount)

	return u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		err := u.userRepo.UpdateCredits(txCtx, userID, -amount)
		if err != nil {
			return err
		}

		txLog := &domain.CreditTransaction{
			ID:        uuid.New().String(),
			UserID:    userID,
			Amount:    -amount,
			Type:      "DEDUCTION",
			Reason:    "Manual DeductCredits RPC",
			CreatedAt: time.Now(),
		}
		return u.creditRepo.CreateTransaction(txCtx, txLog)
	})
}

func (u *membershipUseCase) AddCredits(ctx context.Context, userID string, amount int) error {
	slog.Info("AddCredits request received", "user_id", userID, "amount", amount)

	return u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		err := u.userRepo.UpdateCredits(txCtx, userID, amount)
		if err != nil {
			return err
		}

		txLog := &domain.CreditTransaction{
			ID:        uuid.New().String(),
			UserID:    userID,
			Amount:    amount,
			Type:      "ADDITION",
			Reason:    "Manual AddCredits RPC",
			CreatedAt: time.Now(),
		}
		return u.creditRepo.CreateTransaction(txCtx, txLog)
	})
}

func (u *membershipUseCase) GetUserBookings(ctx context.Context, userID string) ([]*domain.Booking, error) {
	return u.bookingRepo.GetByUserID(ctx, userID)
}

func (u *membershipUseCase) CreateUser(ctx context.Context, name, email string, startingCredits int) (*domain.User, error) {
	slog.Info("CreateUser request received", "name", name, "email", email, "starting_credits", startingCredits)

	user := &domain.User{
		ID:        uuid.New().String(),
		Name:      name,
		Email:     email,
		Credits:   startingCredits,
		CreatedAt: time.Now(),
	}

	err := u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		err := u.userRepo.Create(txCtx, user)
		if err != nil {
			return err
		}

		// Auto-create a standard active membership for the user
		membership := &domain.Membership{
			ID:        uuid.New().String(),
			UserID:    user.ID,
			Type:      "Standard",
			Status:    "Active",
			ExpiresAt: time.Now().AddDate(1, 0, 0), // 1 year expiration
		}
		err = u.membershipRepo.Create(txCtx, membership)
		if err != nil {
			return err
		}

		// Log welcome credits if > 0
		if startingCredits > 0 {
			txLog := &domain.CreditTransaction{
				ID:        uuid.New().String(),
				UserID:    user.ID,
				Amount:    startingCredits,
				Type:      "ADDITION",
				Reason:    "Welcome Credits",
				CreatedAt: time.Now(),
			}
			err = u.creditRepo.CreateTransaction(txCtx, txLog)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("CreateUser transaction aborted", "error", err)
		return nil, err
	}

	slog.Info("CreateUser successful", "user_id", user.ID)
	return user, nil
}

func (u *membershipUseCase) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return u.userRepo.GetByID(ctx, userID)
}

func (u *membershipUseCase) UpdateUser(ctx context.Context, userID, name, email string) (*domain.User, error) {
	slog.Info("UpdateUser request received", "user_id", userID, "name", name, "email", email)

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Name = name
	user.Email = email

	err = u.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	slog.Info("UpdateUser successful", "user_id", userID)
	return user, nil
}

func (u *membershipUseCase) GetUserMembership(ctx context.Context, userID string) (*domain.Membership, error) {
	return u.membershipRepo.GetByUserID(ctx, userID)
}

func (u *membershipUseCase) GetCreditTransactions(ctx context.Context, userID string) ([]*domain.CreditTransaction, error) {
	return u.creditRepo.GetByUserID(ctx, userID)
}
