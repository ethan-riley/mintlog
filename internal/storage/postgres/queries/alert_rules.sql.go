package queries

import (
	"context"

	"github.com/google/uuid"
)

const createAlertRule = `
INSERT INTO alert_rules (tenant_id, name, query, threshold, window_seconds, eval_interval)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, tenant_id, name, query, threshold, window_seconds, eval_interval, is_active, created_at, updated_at
`

func (q *Queries) CreateAlertRule(ctx context.Context, tenantID uuid.UUID, name string, query []byte, threshold, windowSeconds int32, evalInterval string) (AlertRule, error) {
	row := q.db.QueryRow(ctx, createAlertRule, tenantID, name, query, threshold, windowSeconds, evalInterval)
	var r AlertRule
	err := row.Scan(&r.ID, &r.TenantID, &r.Name, &r.Query, &r.Threshold, &r.WindowSeconds, &r.EvalInterval, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

const getAlertRule = `
SELECT id, tenant_id, name, query, threshold, window_seconds, eval_interval, is_active, created_at, updated_at
FROM alert_rules WHERE id = $1 AND tenant_id = $2
`

func (q *Queries) GetAlertRule(ctx context.Context, id, tenantID uuid.UUID) (AlertRule, error) {
	row := q.db.QueryRow(ctx, getAlertRule, id, tenantID)
	var r AlertRule
	err := row.Scan(&r.ID, &r.TenantID, &r.Name, &r.Query, &r.Threshold, &r.WindowSeconds, &r.EvalInterval, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

const listAlertRules = `
SELECT id, tenant_id, name, query, threshold, window_seconds, eval_interval, is_active, created_at, updated_at
FROM alert_rules WHERE tenant_id = $1 ORDER BY created_at DESC
`

func (q *Queries) ListAlertRules(ctx context.Context, tenantID uuid.UUID) ([]AlertRule, error) {
	rows, err := q.db.Query(ctx, listAlertRules, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AlertRule
	for rows.Next() {
		var r AlertRule
		if err := rows.Scan(&r.ID, &r.TenantID, &r.Name, &r.Query, &r.Threshold, &r.WindowSeconds, &r.EvalInterval, &r.IsActive, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	if items == nil {
		items = []AlertRule{}
	}
	return items, rows.Err()
}

const updateAlertRule = `
UPDATE alert_rules
SET name = $3, query = $4, threshold = $5, window_seconds = $6, eval_interval = $7, is_active = $8, updated_at = now()
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, name, query, threshold, window_seconds, eval_interval, is_active, created_at, updated_at
`

func (q *Queries) UpdateAlertRule(ctx context.Context, id, tenantID uuid.UUID, name string, query []byte, threshold, windowSeconds int32, evalInterval string, isActive bool) (AlertRule, error) {
	row := q.db.QueryRow(ctx, updateAlertRule, id, tenantID, name, query, threshold, windowSeconds, evalInterval, isActive)
	var r AlertRule
	err := row.Scan(&r.ID, &r.TenantID, &r.Name, &r.Query, &r.Threshold, &r.WindowSeconds, &r.EvalInterval, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

const deleteAlertRule = `DELETE FROM alert_rules WHERE id = $1 AND tenant_id = $2`

func (q *Queries) DeleteAlertRule(ctx context.Context, id, tenantID uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteAlertRule, id, tenantID)
	return err
}

const listActiveAlertRules = `
SELECT id, tenant_id, name, query, threshold, window_seconds, eval_interval, is_active, created_at, updated_at
FROM alert_rules WHERE is_active = true
`

func (q *Queries) ListActiveAlertRules(ctx context.Context) ([]AlertRule, error) {
	rows, err := q.db.Query(ctx, listActiveAlertRules)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AlertRule
	for rows.Next() {
		var r AlertRule
		if err := rows.Scan(&r.ID, &r.TenantID, &r.Name, &r.Query, &r.Threshold, &r.WindowSeconds, &r.EvalInterval, &r.IsActive, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	if items == nil {
		items = []AlertRule{}
	}
	return items, rows.Err()
}

const upsertAlertState = `
INSERT INTO alert_states (rule_id, tenant_id, state, last_value, last_eval_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (rule_id)
DO UPDATE SET state = $3, last_value = $4, last_eval_at = now(),
  fired_at = CASE WHEN $3 = 'firing' AND alert_states.state != 'firing' THEN now() ELSE alert_states.fired_at END,
  resolved_at = CASE WHEN $3 = 'ok' AND alert_states.state = 'firing' THEN now() ELSE alert_states.resolved_at END
RETURNING id, rule_id, tenant_id, state, last_value, last_eval_at, fired_at, resolved_at
`

func (q *Queries) UpsertAlertState(ctx context.Context, ruleID, tenantID uuid.UUID, state string, lastValue int32) (AlertState, error) {
	row := q.db.QueryRow(ctx, upsertAlertState, ruleID, tenantID, state, lastValue)
	var s AlertState
	err := row.Scan(&s.ID, &s.RuleID, &s.TenantID, &s.State, &s.LastValue, &s.LastEvalAt, &s.FiredAt, &s.ResolvedAt)
	return s, err
}

const getAlertState = `
SELECT id, rule_id, tenant_id, state, last_value, last_eval_at, fired_at, resolved_at
FROM alert_states WHERE rule_id = $1
`

func (q *Queries) GetAlertState(ctx context.Context, ruleID uuid.UUID) (AlertState, error) {
	row := q.db.QueryRow(ctx, getAlertState, ruleID)
	var s AlertState
	err := row.Scan(&s.ID, &s.RuleID, &s.TenantID, &s.State, &s.LastValue, &s.LastEvalAt, &s.FiredAt, &s.ResolvedAt)
	return s, err
}
