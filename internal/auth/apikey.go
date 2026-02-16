package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/felipemonteiro/mintlog/internal/storage/postgres/queries"
	rediscache "github.com/felipemonteiro/mintlog/internal/storage/redis"
	"github.com/felipemonteiro/mintlog/internal/tenant"
	goredis "github.com/redis/go-redis/v9"
)

type KeyResolver struct {
	queries *queries.Queries
	cache   *rediscache.Cache
}

func NewKeyResolver(q *queries.Queries, cache *rediscache.Cache) *KeyResolver {
	return &KeyResolver{queries: q, cache: cache}
}

func HashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func KeyPrefix(raw string) string {
	if len(raw) < 8 {
		return raw
	}
	return raw[:8]
}

func (kr *KeyResolver) Resolve(ctx context.Context, rawKey string) (*tenant.Info, error) {
	hash := HashKey(rawKey)
	cacheKey := fmt.Sprintf("apikey:%s", hash)

	var info tenant.Info
	err := kr.cache.Get(ctx, cacheKey, &info)
	if err == nil {
		return &info, nil
	}
	if err != goredis.Nil {
		// Log but don't fail; fall through to Postgres
	}

	row, err := kr.queries.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("api key not found or expired")
	}

	info = tenant.Info{
		ID:            row.TenantID,
		Name:          row.TenantName,
		Plan:          row.TenantPlan,
		RetentionDays: row.RetentionDays,
		Scopes:        row.Scopes,
		RateLimit:     row.RateLimit,
	}

	_ = kr.cache.Set(ctx, cacheKey, &info)

	return &info, nil
}
