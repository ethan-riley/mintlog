package incident

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	StatusTriggered    = "triggered"
	StatusAcknowledged = "acknowledged"
	StatusResolved     = "resolved"
)

type CreateRequest struct {
	Title    string `json:"title"`
	Severity string `json:"severity,omitempty"`
}

type PatchRequest struct {
	Status string `json:"status"` // "acknowledged" or "resolved"
}

type TimelineRequest struct {
	EventType string `json:"event_type"` // "comment", "status_change", etc.
	Content   string `json:"content"`
}

type IncidentResponse struct {
	ID          uuid.UUID          `json:"id"`
	TenantID    uuid.UUID          `json:"tenant_id"`
	Title       string             `json:"title"`
	Status      string             `json:"status"`
	Severity    string             `json:"severity"`
	AlertRuleID pgtype.UUID        `json:"alert_rule_id,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	ResolvedAt  pgtype.Timestamptz `json:"resolved_at,omitempty"`
	Timeline    []TimelineEntry    `json:"timeline,omitempty"`
}

type TimelineEntry struct {
	ID        uuid.UUID `json:"id"`
	EventType string    `json:"event_type"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
