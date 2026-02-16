package incident

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/felipemonteiro/mintlog/internal/storage/postgres/queries"
	"github.com/felipemonteiro/mintlog/internal/tenant"
	"github.com/felipemonteiro/mintlog/pkg/apierror"
)

type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, apierror.BadRequest("invalid JSON: "+err.Error()))
		return
	}

	if req.Title == "" {
		apierror.Write(w, apierror.BadRequest("title is required"))
		return
	}

	inc, err := h.service.Create(r.Context(), info.ID, req.Title, req.Severity, pgtype.UUID{})
	if err != nil {
		apierror.Write(w, apierror.Internal("failed to create incident"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toIncidentResponse(inc, nil))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierror.Write(w, apierror.BadRequest("invalid incident ID"))
		return
	}

	inc, err := h.service.Get(r.Context(), id, info.ID)
	if err != nil {
		apierror.Write(w, apierror.NotFound("incident not found"))
		return
	}

	timeline, _ := h.service.GetTimeline(r.Context(), id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toIncidentResponse(inc, timeline))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	status := r.URL.Query().Get("status")
	incidents, err := h.service.List(r.Context(), info.ID, status)
	if err != nil {
		apierror.Write(w, apierror.Internal("failed to list incidents"))
		return
	}

	resp := make([]IncidentResponse, len(incidents))
	for i, inc := range incidents {
		resp[i] = toIncidentResponse(inc, nil)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierror.Write(w, apierror.BadRequest("invalid incident ID"))
		return
	}

	var req PatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, apierror.BadRequest("invalid JSON: "+err.Error()))
		return
	}

	if req.Status == "" {
		apierror.Write(w, apierror.BadRequest("status is required"))
		return
	}

	inc, err := h.service.UpdateStatus(r.Context(), id, info.ID, req.Status)
	if err != nil {
		apierror.Write(w, apierror.BadRequest(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toIncidentResponse(inc, nil))
}

func (h *Handler) AddTimeline(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierror.Write(w, apierror.BadRequest("invalid incident ID"))
		return
	}

	// Verify incident belongs to tenant
	_, err = h.service.Get(r.Context(), id, info.ID)
	if err != nil {
		apierror.Write(w, apierror.NotFound("incident not found"))
		return
	}

	var req TimelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, apierror.BadRequest("invalid JSON: "+err.Error()))
		return
	}

	if req.EventType == "" {
		req.EventType = "comment"
	}

	entry, err := h.service.AddTimeline(r.Context(), id, req.EventType, req.Content)
	if err != nil {
		apierror.Write(w, apierror.Internal("failed to add timeline entry"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TimelineEntry{
		ID:        entry.ID,
		EventType: entry.EventType,
		Content:   entry.Content,
		CreatedAt: entry.CreatedAt,
	})
}

func toIncidentResponse(inc queries.Incident, timeline []queries.IncidentTimeline) IncidentResponse {
	resp := IncidentResponse{
		ID:          inc.ID,
		TenantID:    inc.TenantID,
		Title:       inc.Title,
		Status:      inc.Status,
		Severity:    inc.Severity,
		AlertRuleID: inc.AlertRuleID,
		CreatedAt:   inc.CreatedAt,
		UpdatedAt:   inc.UpdatedAt,
		ResolvedAt:  inc.ResolvedAt,
	}
	if timeline != nil {
		resp.Timeline = toTimelineEntries(timeline)
	}
	return resp
}
