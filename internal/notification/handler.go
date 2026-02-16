package notification

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

	var req ChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, apierror.BadRequest("invalid JSON: "+err.Error()))
		return
	}

	if req.Name == "" {
		apierror.Write(w, apierror.BadRequest("name is required"))
		return
	}
	if req.ChannelType == "" {
		req.ChannelType = "webhook"
	}
	if req.Config == nil {
		req.Config = json.RawMessage(`{}`)
	}

	ch, err := h.queries.CreateNotificationChannel(r.Context(), info.ID, req.Name, req.ChannelType, []byte(req.Config))
	if err != nil {
		apierror.Write(w, apierror.Internal("failed to create channel"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toChannelResponse(ch))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierror.Write(w, apierror.BadRequest("invalid channel ID"))
		return
	}

	ch, err := h.queries.GetNotificationChannel(r.Context(), id, info.ID)
	if err != nil {
		apierror.Write(w, apierror.NotFound("channel not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toChannelResponse(ch))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	channels, err := h.queries.ListNotificationChannels(r.Context(), info.ID)
	if err != nil {
		apierror.Write(w, apierror.Internal("failed to list channels"))
		return
	}

	resp := make([]ChannelResponse, len(channels))
	for i, ch := range channels {
		resp[i] = toChannelResponse(ch)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		apierror.Write(w, apierror.BadRequest("invalid channel ID"))
		return
	}

	if err := h.queries.DeleteNotificationChannel(r.Context(), id, info.ID); err != nil {
		apierror.Write(w, apierror.Internal("failed to delete channel"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toChannelResponse(c queries.NotificationChannel) ChannelResponse {
	return ChannelResponse{
		ID:          c.ID,
		TenantID:    c.TenantID,
		Name:        c.Name,
		ChannelType: c.ChannelType,
		Config:      c.Config,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
