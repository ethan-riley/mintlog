package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb    *redis.Client
	window time.Duration
}

func NewRateLimiter(rdb *redis.Client, window time.Duration) *RateLimiter {
	return &RateLimiter{rdb: rdb, window: window}
}

// Allow checks if the tenant is within their rate limit using a sliding window.
// Returns (allowed bool, remaining int, error).
func (rl *RateLimiter) Allow(ctx context.Context, tenantID string, limit int) (bool, int, error) {
	key := fmt.Sprintf("ratelimit:%s", tenantID)
	now := time.Now()
	windowStart := now.Add(-rl.window)

	pipe := rl.rdb.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixMilli()))
	countCmd := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixMilli()), Member: now.UnixNano()})
	pipe.Expire(ctx, key, rl.window+time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, fmt.Errorf("rate limit pipeline: %w", err)
	}

	count := int(countCmd.Val())
	remaining := limit - count - 1
	if remaining < 0 {
		remaining = 0
	}

	return count < int(limit), remaining, nil
}
