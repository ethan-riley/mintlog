package alerting

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AlertRuleRequest struct {
	Name          string          `json:"name"`
	Query         json.RawMessage `json:"query"`
	Threshold     int32           `json:"threshold"`
	WindowSeconds int32           `json:"window_seconds"`
	EvalInterval  string          `json:"eval_interval,omitempty"`
	IsActive      *bool           `json:"is_active,omitempty"`
}

type AlertRuleResponse struct {
	ID            uuid.UUID       `json:"id"`
	TenantID      uuid.UUID       `json:"tenant_id"`
	Name          string          `json:"name"`
	Query         json.RawMessage `json:"query"`
	Threshold     int32           `json:"threshold"`
	WindowSeconds int32           `json:"window_seconds"`
	EvalInterval  string          `json:"eval_interval"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type AlertEvent struct {
	RuleID    uuid.UUID `json:"rule_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	RuleName  string    `json:"rule_name"`
	State     string    `json:"state"` // "firing" or "resolved"
	Value     int32     `json:"value"`
	Threshold int32     `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
}
