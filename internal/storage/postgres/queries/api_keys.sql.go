package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createAPIKey = `
INSERT INTO api_keys (tenant_id, key_hash, key_prefix, name, scopes, rate_limit, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, tenant_id, key_hash, key_prefix, name, scopes, rate_limit, is_active, expires_at, created_at, updated_at
`

func (q *Queries) CreateAPIKey(ctx context.Context, tenantID uuid.UUID, keyHash, keyPrefix, name string, scopes []string, rateLimit int32, expiresAt pgtype.Timestamptz) (ApiKey, error) {
	row := q.db.QueryRow(ctx, createAPIKey, tenantID, keyHash, keyPrefix, name, scopes, rateLimit, expiresAt)
	var k ApiKey
	err := row.Scan(&k.ID, &k.TenantID, &k.KeyHash, &k.KeyPrefix, &k.Name, &k.Scopes, &k.RateLimit, &k.IsActive, &k.ExpiresAt, &k.CreatedAt, &k.UpdatedAt)
	return k, err
}

const getAPIKeyByHash = `
SELECT ak.id, ak.tenant_id, ak.key_hash, ak.key_prefix, ak.name, ak.scopes, ak.rate_limit, ak.is_active, ak.expires_at, ak.created_at, ak.updated_at,
       t.name as tenant_name, t.plan as tenant_plan, t.retention_days
FROM api_keys ak
JOIN tenants t ON t.id = ak.tenant_id
WHERE ak.key_hash = $1 AND ak.is_active = true
AND (ak.expires_at IS NULL OR ak.expires_at > now())
`

func (q *Queries) GetAPIKeyByHash(ctx context.Context, keyHash string) (ApiKeyWithTenant, error) {
	row := q.db.QueryRow(ctx, getAPIKeyByHash, keyHash)
	var k ApiKeyWithTenant
	err := row.Scan(&k.ID, &k.TenantID, &k.KeyHash, &k.KeyPrefix, &k.Name, &k.Scopes, &k.RateLimit, &k.IsActive, &k.ExpiresAt, &k.CreatedAt, &k.UpdatedAt, &k.TenantName, &k.TenantPlan, &k.RetentionDays)
	return k, err
}

const listAPIKeysByTenant = `
SELECT id, tenant_id, key_prefix, name, scopes, rate_limit, is_active, expires_at, created_at
FROM api_keys
WHERE tenant_id = $1
ORDER BY created_at DESC
`

type ListAPIKeysRow struct {
	ID        uuid.UUID          `json:"id"`
	TenantID  uuid.UUID          `json:"tenant_id"`
	KeyPrefix string             `json:"key_prefix"`
	Name      string             `json:"name"`
	Scopes    []string           `json:"scopes"`
	RateLimit int32              `json:"rate_limit"`
	IsActive  bool               `json:"is_active"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

func (q *Queries) ListAPIKeysByTenant(ctx context.Context, tenantID uuid.UUID) ([]ListAPIKeysRow, error) {
	rows, err := q.db.Query(ctx, listAPIKeysByTenant, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListAPIKeysRow
	for rows.Next() {
		var k ListAPIKeysRow
		if err := rows.Scan(&k.ID, &k.TenantID, &k.KeyPrefix, &k.Name, &k.Scopes, &k.RateLimit, &k.IsActive, &k.ExpiresAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, k)
	}
	if items == nil {
		items = []ListAPIKeysRow{}
	}
	return items, rows.Err()
}

const deactivateAPIKey = `UPDATE api_keys SET is_active = false, updated_at = now() WHERE id = $1`

func (q *Queries) DeactivateAPIKey(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deactivateAPIKey, id)
	return err
}
