package cache

import (
	"fmt"
	"time"
)

type RateLimiter struct {
	redis *RedisCache
}

func NewRateLimiter(redis *RedisCache) *RateLimiter {
	return &RateLimiter{redis: redis}
}

// CheckLimit verifies if request is within rate limit
func (r *RateLimiter) CheckLimit(userID string, limit int, window time.Duration) (bool, int, error) {
	key := fmt.Sprintf("ratelimit:%s", userID)

	// Increment counter
	count, err := r.redis.client.Incr(r.redis.ctx, key).Result()
	if err != nil {
		return false, 0, err
	}

	// Set expiry on first request
	if count == 1 {
		r.redis.client.Expire(r.redis.ctx, key, window)
	}

	// Check if limit exceeded
	if count > int64(limit) {
		return false, int(count), fmt.Errorf("rate limit exceeded: %d/%d requests", count, limit)
	}

	return true, int(count), nil
}

// GetRemaining returns remaining requests
func (r *RateLimiter) GetRemaining(userID string, limit int) (int, error) {
	key := fmt.Sprintf("ratelimit:%s", userID)

	count, err := r.redis.client.Get(r.redis.ctx, key).Int64()
	if err != nil {
		return limit, nil // No requests yet
	}

	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// Reset clears rate limit for user
func (r *RateLimiter) Reset(userID string) error {
	key := fmt.Sprintf("ratelimit:%s", userID)
	return r.redis.client.Del(r.redis.ctx, key).Err()
}
