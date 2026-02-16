package search

import (
	"encoding/json"
	"time"
)

type SearchRequest struct {
	Query     string         `json:"query"`
	Level     string         `json:"level,omitempty"`
	Service   string         `json:"service,omitempty"`
	Host      string         `json:"host,omitempty"`
	TraceID   string         `json:"trace_id,omitempty"`
	From      time.Time      `json:"from,omitempty"`
	To        time.Time      `json:"to,omitempty"`
	Size      int            `json:"size,omitempty"`
	SearchAfter []any        `json:"search_after,omitempty"`
	Sort      string         `json:"sort,omitempty"` // "asc" or "desc"
}

type SearchResponse struct {
	Hits        []json.RawMessage `json:"hits"`
	Total       int               `json:"total"`
	SearchAfter []any             `json:"search_after,omitempty"`
}

type TailRequest struct {
	Query   string `json:"query,omitempty"`
	Level   string `json:"level,omitempty"`
	Service string `json:"service,omitempty"`
}

type AggregateRequest struct {
	Query    string `json:"query,omitempty"`
	Level    string `json:"level,omitempty"`
	Service  string `json:"service,omitempty"`
	From     time.Time `json:"from,omitempty"`
	To       time.Time `json:"to,omitempty"`
	GroupBy  string    `json:"group_by,omitempty"`  // "level", "service", "host"
	Interval string    `json:"interval,omitempty"`  // "1m", "5m", "1h", "1d"
}

type AggregateResponse struct {
	Buckets json.RawMessage `json:"buckets"`
}
