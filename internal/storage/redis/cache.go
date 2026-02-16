package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewCache(rdb *redis.Client, ttl time.Duration) *Cache {
	return &Cache{rdb: rdb, ttl: ttl}
}

func (c *Cache) Get(ctx context.Context, key string, dest any) error {
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *Cache) Set(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}
	return c.rdb.Set(ctx, key, data, c.ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}
