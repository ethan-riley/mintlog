package pipeline

import (
	"encoding/json"

	"github.com/felipemonteiro/mintlog/pkg/logmodel"
)

// ParseJSON attempts to extract structured fields from the raw message.
// If the message is valid JSON, its fields are merged into the event's Fields map.
func ParseJSON(event *logmodel.LogEvent) {
	if event.Raw == "" {
		return
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(event.Raw), &parsed); err != nil {
		return
	}

	if event.Fields == nil {
		event.Fields = make(map[string]any)
	}

	for k, v := range parsed {
		switch k {
		case "message", "msg":
			if event.Message == "" {
				if s, ok := v.(string); ok {
					event.Message = s
				}
			}
		case "level", "severity":
			if event.Level == "" {
				if s, ok := v.(string); ok {
					event.Level = s
				}
			}
		case "service", "app":
			if event.Service == "" {
				if s, ok := v.(string); ok {
					event.Service = s
				}
			}
		case "host", "hostname":
			if event.Host == "" {
				if s, ok := v.(string); ok {
					event.Host = s
				}
			}
		case "trace_id":
			if event.TraceID == "" {
				if s, ok := v.(string); ok {
					event.TraceID = s
				}
			}
		case "span_id":
			if event.SpanID == "" {
				if s, ok := v.(string); ok {
					event.SpanID = s
				}
			}
		default:
			event.Fields[k] = v
		}
	}
}
