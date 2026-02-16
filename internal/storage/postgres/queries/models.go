package queries

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Tenant struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Plan          string    `json:"plan"`
	RetentionDays int32     `json:"retention_days"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ApiKey struct {
	ID        uuid.UUID          `json:"id"`
	TenantID  uuid.UUID          `json:"tenant_id"`
	KeyHash   string             `json:"key_hash"`
	KeyPrefix string             `json:"key_prefix"`
	Name      string             `json:"name"`
	Scopes    []string           `json:"scopes"`
	RateLimit int32              `json:"rate_limit"`
	IsActive  bool               `json:"is_active"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type ApiKeyWithTenant struct {
	ID            uuid.UUID          `json:"id"`
	TenantID      uuid.UUID          `json:"tenant_id"`
	KeyHash       string             `json:"key_hash"`
	KeyPrefix     string             `json:"key_prefix"`
	Name          string             `json:"name"`
	Scopes        []string           `json:"scopes"`
	RateLimit     int32              `json:"rate_limit"`
	IsActive      bool               `json:"is_active"`
	ExpiresAt     pgtype.Timestamptz `json:"expires_at"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	TenantName    string             `json:"tenant_name"`
	TenantPlan    string             `json:"tenant_plan"`
	RetentionDays int32              `json:"retention_days"`
}

type AlertRule struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	Name           string    `json:"name"`
	Query          []byte    `json:"query"`
	Threshold      int32     `json:"threshold"`
	WindowSeconds  int32     `json:"window_seconds"`
	EvalInterval   string    `json:"eval_interval"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type AlertState struct {
	ID          uuid.UUID          `json:"id"`
	RuleID      uuid.UUID          `json:"rule_id"`
	TenantID    uuid.UUID          `json:"tenant_id"`
	State       string             `json:"state"`
	LastValue   int32              `json:"last_value"`
	LastEvalAt  pgtype.Timestamptz `json:"last_eval_at"`
	FiredAt     pgtype.Timestamptz `json:"fired_at"`
	ResolvedAt  pgtype.Timestamptz `json:"resolved_at"`
}

type Incident struct {
	ID          uuid.UUID          `json:"id"`
	TenantID    uuid.UUID          `json:"tenant_id"`
	Title       string             `json:"title"`
	Status      string             `json:"status"`
	Severity    string             `json:"severity"`
	AlertRuleID pgtype.UUID        `json:"alert_rule_id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	ResolvedAt  pgtype.Timestamptz `json:"resolved_at"`
}

type IncidentTimeline struct {
	ID         uuid.UUID `json:"id"`
	IncidentID uuid.UUID `json:"incident_id"`
	EventType  string    `json:"event_type"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type NotificationChannel struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	ChannelType string    `json:"channel_type"`
	Config      []byte    `json:"config"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
