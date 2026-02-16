package notification

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ChannelRequest struct {
	Name        string          `json:"name"`
	ChannelType string          `json:"channel_type"`
	Config      json.RawMessage `json:"config"`
}

type ChannelResponse struct {
	ID          uuid.UUID       `json:"id"`
	TenantID    uuid.UUID       `json:"tenant_id"`
	Name        string          `json:"name"`
	ChannelType string          `json:"channel_type"`
	Config      json.RawMessage `json:"config"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type WebhookConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Secret  string            `json:"secret,omitempty"`
}
