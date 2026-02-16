package queries

import (
	"context"

	"github.com/google/uuid"
)

const createTenant = `
INSERT INTO tenants (name, plan, retention_days)
VALUES ($1, $2, $3)
RETURNING id, name, plan, retention_days, created_at, updated_at
`

func (q *Queries) CreateTenant(ctx context.Context, name string, plan string, retentionDays int32) (Tenant, error) {
	row := q.db.QueryRow(ctx, createTenant, name, plan, retentionDays)
	var t Tenant
	err := row.Scan(&t.ID, &t.Name, &t.Plan, &t.RetentionDays, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

const getTenant = `
SELECT id, name, plan, retention_days, created_at, updated_at
FROM tenants WHERE id = $1
`

func (q *Queries) GetTenant(ctx context.Context, id uuid.UUID) (Tenant, error) {
	row := q.db.QueryRow(ctx, getTenant, id)
	var t Tenant
	err := row.Scan(&t.ID, &t.Name, &t.Plan, &t.RetentionDays, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

const getTenantByName = `
SELECT id, name, plan, retention_days, created_at, updated_at
FROM tenants WHERE name = $1
`

func (q *Queries) GetTenantByName(ctx context.Context, name string) (Tenant, error) {
	row := q.db.QueryRow(ctx, getTenantByName, name)
	var t Tenant
	err := row.Scan(&t.ID, &t.Name, &t.Plan, &t.RetentionDays, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

const listTenants = `
SELECT id, name, plan, retention_days, created_at, updated_at
FROM tenants ORDER BY created_at DESC
`

func (q *Queries) ListTenants(ctx context.Context) ([]Tenant, error) {
	rows, err := q.db.Query(ctx, listTenants)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Plan, &t.RetentionDays, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	if items == nil {
		items = []Tenant{}
	}
	return items, rows.Err()
}
