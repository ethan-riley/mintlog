package logmodel

import "time"

// LogEvent is the canonical log event structure used throughout the system.
type LogEvent struct {
	ID        string            `json:"id"`
	TenantID  string            `json:"tenant_id"`
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	Service   string            `json:"service"`
	Host      string            `json:"host,omitempty"`
	TraceID   string            `json:"trace_id,omitempty"`
	SpanID    string            `json:"span_id,omitempty"`
	Fields    map[string]any    `json:"fields,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
	Raw       string            `json:"raw,omitempty"`
}

// IngestRequest is the payload for POST /v1/ingest/logs.
type IngestRequest struct {
	Events []IngestEvent `json:"events"`
}

// IngestEvent is a single event in an ingest request (before normalization).
type IngestEvent struct {
	Timestamp string         `json:"timestamp,omitempty"`
	Level     string         `json:"level,omitempty"`
	Message   string         `json:"message"`
	Service   string         `json:"service"`
	Host      string         `json:"host,omitempty"`
	TraceID   string         `json:"trace_id,omitempty"`
	SpanID    string         `json:"span_id,omitempty"`
	Fields    map[string]any `json:"fields,omitempty"`
	Tags      []string       `json:"tags,omitempty"`
}

// IngestResponse is returned from POST /v1/ingest/logs.
type IngestResponse struct {
	Accepted int `json:"accepted"`
	Rejected int `json:"rejected"`
}
