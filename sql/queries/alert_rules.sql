-- name: CreateAlertRule :one
INSERT INTO alert_rules (tenant_id, name, query, threshold, window_seconds, eval_interval)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAlertRule :one
SELECT * FROM alert_rules WHERE id = $1 AND tenant_id = $2;

-- name: ListAlertRules :many
SELECT * FROM alert_rules WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: UpdateAlertRule :one
UPDATE alert_rules
SET name = $3, query = $4, threshold = $5, window_seconds = $6, eval_interval = $7, is_active = $8, updated_at = now()
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: DeleteAlertRule :exec
DELETE FROM alert_rules WHERE id = $1 AND tenant_id = $2;

-- name: ListActiveAlertRules :many
SELECT * FROM alert_rules WHERE is_active = true;

-- name: UpsertAlertState :one
INSERT INTO alert_states (rule_id, tenant_id, state, last_value, last_eval_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (rule_id)
DO UPDATE SET state = $3, last_value = $4, last_eval_at = now(),
  fired_at = CASE WHEN $3 = 'firing' AND alert_states.state != 'firing' THEN now() ELSE alert_states.fired_at END,
  resolved_at = CASE WHEN $3 = 'ok' AND alert_states.state = 'firing' THEN now() ELSE alert_states.resolved_at END
RETURNING *;

-- name: GetAlertState :one
SELECT * FROM alert_states WHERE rule_id = $1;
