package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createIncident = `
INSERT INTO incidents (tenant_id, title, status, severity, alert_rule_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, tenant_id, title, status, severity, alert_rule_id, created_at, updated_at, resolved_at
`

func (q *Queries) CreateIncident(ctx context.Context, tenantID uuid.UUID, title, status, severity string, alertRuleID pgtype.UUID) (Incident, error) {
	row := q.db.QueryRow(ctx, createIncident, tenantID, title, status, severity, alertRuleID)
	var i Incident
	err := row.Scan(&i.ID, &i.TenantID, &i.Title, &i.Status, &i.Severity, &i.AlertRuleID, &i.CreatedAt, &i.UpdatedAt, &i.ResolvedAt)
	return i, err
}

const getIncident = `
SELECT id, tenant_id, title, status, severity, alert_rule_id, created_at, updated_at, resolved_at
FROM incidents WHERE id = $1 AND tenant_id = $2
`

func (q *Queries) GetIncident(ctx context.Context, id, tenantID uuid.UUID) (Incident, error) {
	row := q.db.QueryRow(ctx, getIncident, id, tenantID)
	var i Incident
	err := row.Scan(&i.ID, &i.TenantID, &i.Title, &i.Status, &i.Severity, &i.AlertRuleID, &i.CreatedAt, &i.UpdatedAt, &i.ResolvedAt)
	return i, err
}

const listIncidents = `
SELECT id, tenant_id, title, status, severity, alert_rule_id, created_at, updated_at, resolved_at
FROM incidents WHERE tenant_id = $1
ORDER BY created_at DESC
`

func (q *Queries) ListIncidents(ctx context.Context, tenantID uuid.UUID) ([]Incident, error) {
	rows, err := q.db.Query(ctx, listIncidents, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Incident
	for rows.Next() {
		var i Incident
		if err := rows.Scan(&i.ID, &i.TenantID, &i.Title, &i.Status, &i.Severity, &i.AlertRuleID, &i.CreatedAt, &i.UpdatedAt, &i.ResolvedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if items == nil {
		items = []Incident{}
	}
	return items, rows.Err()
}

const listIncidentsByStatus = `
SELECT id, tenant_id, title, status, severity, alert_rule_id, created_at, updated_at, resolved_at
FROM incidents WHERE tenant_id = $1 AND status = $2
ORDER BY created_at DESC
`

func (q *Queries) ListIncidentsByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]Incident, error) {
	rows, err := q.db.Query(ctx, listIncidentsByStatus, tenantID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Incident
	for rows.Next() {
		var i Incident
		if err := rows.Scan(&i.ID, &i.TenantID, &i.Title, &i.Status, &i.Severity, &i.AlertRuleID, &i.CreatedAt, &i.UpdatedAt, &i.ResolvedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if items == nil {
		items = []Incident{}
	}
	return items, rows.Err()
}

const updateIncidentStatus = `
UPDATE incidents
SET status = $3, updated_at = now(),
    resolved_at = CASE WHEN $3 = 'resolved' THEN now() ELSE resolved_at END
WHERE id = $1 AND tenant_id = $2
RETURNING id, tenant_id, title, status, severity, alert_rule_id, created_at, updated_at, resolved_at
`

func (q *Queries) UpdateIncidentStatus(ctx context.Context, id, tenantID uuid.UUID, status string) (Incident, error) {
	row := q.db.QueryRow(ctx, updateIncidentStatus, id, tenantID, status)
	var i Incident
	err := row.Scan(&i.ID, &i.TenantID, &i.Title, &i.Status, &i.Severity, &i.AlertRuleID, &i.CreatedAt, &i.UpdatedAt, &i.ResolvedAt)
	return i, err
}

const addTimelineEntry = `
INSERT INTO incident_timeline (incident_id, event_type, content)
VALUES ($1, $2, $3)
RETURNING id, incident_id, event_type, content, created_at
`

func (q *Queries) AddTimelineEntry(ctx context.Context, incidentID uuid.UUID, eventType, content string) (IncidentTimeline, error) {
	row := q.db.QueryRow(ctx, addTimelineEntry, incidentID, eventType, content)
	var t IncidentTimeline
	err := row.Scan(&t.ID, &t.IncidentID, &t.EventType, &t.Content, &t.CreatedAt)
	return t, err
}

const getIncidentTimeline = `
SELECT id, incident_id, event_type, content, created_at
FROM incident_timeline WHERE incident_id = $1 ORDER BY created_at ASC
`

func (q *Queries) GetIncidentTimeline(ctx context.Context, incidentID uuid.UUID) ([]IncidentTimeline, error) {
	rows, err := q.db.Query(ctx, getIncidentTimeline, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []IncidentTimeline
	for rows.Next() {
		var t IncidentTimeline
		if err := rows.Scan(&t.ID, &t.IncidentID, &t.EventType, &t.Content, &t.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	if items == nil {
		items = []IncidentTimeline{}
	}
	return items, rows.Err()
}
