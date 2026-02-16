-- name: CreateIncident :one
INSERT INTO incidents (tenant_id, title, status, severity, alert_rule_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetIncident :one
SELECT * FROM incidents WHERE id = $1 AND tenant_id = $2;

-- name: ListIncidents :many
SELECT * FROM incidents WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: ListIncidentsByStatus :many
SELECT * FROM incidents WHERE tenant_id = $1 AND status = $2 ORDER BY created_at DESC;

-- name: UpdateIncidentStatus :one
UPDATE incidents
SET status = $3, updated_at = now(),
    resolved_at = CASE WHEN $3 = 'resolved' THEN now() ELSE resolved_at END
WHERE id = $1 AND tenant_id = $2
RETURNING *;

-- name: AddTimelineEntry :one
INSERT INTO incident_timeline (incident_id, event_type, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetIncidentTimeline :many
SELECT * FROM incident_timeline WHERE incident_id = $1 ORDER BY created_at ASC;
