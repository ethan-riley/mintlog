package ingest

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/felipemonteiro/mintlog/internal/tenant"
	"github.com/felipemonteiro/mintlog/pkg/apierror"
	"github.com/felipemonteiro/mintlog/pkg/logmodel"
)

type Handler struct {
	publisher *LogPublisher
}

func NewHandler(publisher *LogPublisher) *Handler {
	return &Handler{publisher: publisher}
}

func (h *Handler) IngestLogs(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	var req logmodel.IngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, apierror.BadRequest("invalid JSON: "+err.Error()))
		return
	}

	if err := ValidateBatch(&req); err != nil {
		apierror.Write(w, apierror.BadRequest(err.Error()))
		return
	}

	var accepted, rejected int
	for i := range req.Events {
		if err := ValidateEvent(&req.Events[i]); err != nil {
			slog.Debug("rejected event", "index", i, "error", err)
			rejected++
			continue
		}

		event := toLogEvent(info.ID.String(), &req.Events[i])
		if err := h.publisher.Publish(info.ID.String(), event); err != nil {
			slog.Error("failed to publish event", "error", err)
			rejected++
			continue
		}
		accepted++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(logmodel.IngestResponse{
		Accepted: accepted,
		Rejected: rejected,
	})
}

func toLogEvent(tenantID string, e *logmodel.IngestEvent) *logmodel.LogEvent {
	ts := time.Now().UTC()
	if e.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, e.Timestamp); err == nil {
			ts = parsed.UTC()
		}
	}

	level := e.Level
	if level == "" {
		level = "info"
	}

	return &logmodel.LogEvent{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Timestamp: ts,
		Level:     level,
		Message:   e.Message,
		Service:   e.Service,
		Host:      e.Host,
		TraceID:   e.TraceID,
		SpanID:    e.SpanID,
		Fields:    e.Fields,
		Tags:      e.Tags,
	}
}
