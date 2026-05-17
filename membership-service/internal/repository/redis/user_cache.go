package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ilnur/gym-membership-app/membership-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type cachedUserRepo struct {
	postgresRepo domain.UserRepo
	redisClient  *redis.Client
	ttl          time.Duration
}

func NewCachedUserRepo(postgresRepo domain.UserRepo, redisClient *redis.Client) domain.UserRepo {
	return &cachedUserRepo{
		postgresRepo: postgresRepo,
		redisClient:  redisClient,
		ttl:          30 * time.Second,
	}
}

func (r *cachedUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	key := "user_credits:" + id

	// Try getting from Redis cache
	val, err := r.redisClient.Get(ctx, key).Result()
	if err == nil {
		var user domain.User
		if err := json.Unmarshal([]byte(val), &user); err == nil {
			return &user, nil
		}
	}

	// Fetch from Postgres if not cached or cache corrupted
	user, err := r.postgresRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Write to Redis cache
	if data, err := json.Marshal(user); err == nil {
		_ = r.redisClient.Set(ctx, key, data, r.ttl).Err()
	}

	return user, nil
}

func (r *cachedUserRepo) UpdateCredits(ctx context.Context, userID string, amount int) error {
	// Execute database update
	err := r.postgresRepo.UpdateCredits(ctx, userID, amount)
	if err != nil {
		return err
	}

	// Invalidate Redis cache
	key := "user_credits:" + userID
	_ = r.redisClient.Del(ctx, key).Err()

	return nil
}
