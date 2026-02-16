-- name: CreateAPIKey :one
INSERT INTO api_keys (tenant_id, key_hash, key_prefix, name, scopes, rate_limit, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAPIKeyByHash :one
SELECT ak.*, t.name as tenant_name, t.plan as tenant_plan, t.retention_days
FROM api_keys ak
JOIN tenants t ON t.id = ak.tenant_id
WHERE ak.key_hash = $1 AND ak.is_active = true
AND (ak.expires_at IS NULL OR ak.expires_at > now());

-- name: ListAPIKeysByTenant :many
SELECT id, tenant_id, key_prefix, name, scopes, rate_limit, is_active, expires_at, created_at
FROM api_keys
WHERE tenant_id = $1
ORDER BY created_at DESC;

-- name: DeactivateAPIKey :exec
UPDATE api_keys SET is_active = false, updated_at = now() WHERE id = $1;
