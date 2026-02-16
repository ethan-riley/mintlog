-- name: CreateTenant :one
INSERT INTO tenants (name, plan, retention_days)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTenant :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantByName :one
SELECT * FROM tenants WHERE name = $1;

-- name: ListTenants :many
SELECT * FROM tenants ORDER BY created_at DESC;
