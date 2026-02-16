package alerting

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/felipemonteiro/mintlog/internal/storage/postgres/queries"
	"github.com/felipemonteiro/mintlog/internal/tenant"
	"github.com/felipemonteiro/mintlog/pkg/apierror"
)

type Handler struct {
	queries *queries.Queries
}

func NewHandler(q *queries.Queries) *Handler {
	return &Handler{queries: q}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	var req AlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, apierror.BadRequest("invalid JSON: "+err.Error()))
		return
	}

	if req.Name == "" {
		apierror.Write(w, apierror.BadRequest("name is required"))
		return
	}
	if req.Threshold <= 0 {
		req.Threshold = 1
	}
	if req.WindowSeconds <= 0 {
		req.WindowSeconds = 300
	}
	if req.EvalInterval == "" {
		req.EvalInterval = "1m"
	}
	if req.Query == nil {
		req.Query = json.RawMessage(`{}`)
	}

	rule, err := h.queries.CreateAlertRule(r.Context(), info.ID, req.Name, []byte(req.Query), req.Threshold, req.WindowSeconds, req.EvalInterval)
	if err != nil {
		apierror.Write(w, apierror.Internal("failed to create alert rule"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toAlertRuleResponse(rule))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierror.Write(w, apierror.BadRequest("invalid rule ID"))
		return
	}

	rule, err := h.queries.GetAlertRule(r.Context(), id, info.ID)
	if err != nil {
		apierror.Write(w, apierror.NotFound("alert rule not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toAlertRuleResponse(rule))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	rules, err := h.queries.ListAlertRules(r.Context(), info.ID)
	if err != nil {
		apierror.Write(w, apierror.Internal("failed to list alert rules"))
		return
	}

	resp := make([]AlertRuleResponse, len(rules))
	for i, rule := range rules {
		resp[i] = toAlertRuleResponse(rule)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierror.Write(w, apierror.BadRequest("invalid rule ID"))
		return
	}

	var req AlertRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, apierror.BadRequest("invalid JSON: "+err.Error()))
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.EvalInterval == "" {
		req.EvalInterval = "1m"
	}
	if req.Query == nil {
		req.Query = json.RawMessage(`{}`)
	}

	rule, err := h.queries.UpdateAlertRule(r.Context(), id, info.ID, req.Name, []byte(req.Query), req.Threshold, req.WindowSeconds, req.EvalInterval, isActive)
	if err != nil {
		apierror.Write(w, apierror.NotFound("alert rule not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toAlertRuleResponse(rule))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierror.Write(w, apierror.BadRequest("invalid rule ID"))
		return
	}

	if err := h.queries.DeleteAlertRule(r.Context(), id, info.ID); err != nil {
		apierror.Write(w, apierror.Internal("failed to delete alert rule"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toAlertRuleResponse(r queries.AlertRule) AlertRuleResponse {
	return AlertRuleResponse{
		ID:            r.ID,
		TenantID:      r.TenantID,
		Name:          r.Name,
		Query:         r.Query,
		Threshold:     r.Threshold,
		WindowSeconds: r.WindowSeconds,
		EvalInterval:  r.EvalInterval,
		IsActive:      r.IsActive,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
