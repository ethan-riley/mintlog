package search

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	osstore "github.com/felipemonteiro/mintlog/internal/storage/opensearch"
	"github.com/felipemonteiro/mintlog/internal/tenant"
	"github.com/felipemonteiro/mintlog/pkg/apierror"
)

type Handler struct {
	searcher *osstore.Searcher
}

func NewHandler(searcher *osstore.Searcher) *Handler {
	return &Handler{searcher: searcher}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, apierror.BadRequest("invalid JSON: "+err.Error()))
		return
	}

	query := BuildSearchQuery(info.ID.String(), &req)
	indices := []string{fmt.Sprintf("mintlog-%s-*", info.ID.String())}

	result, err := h.searcher.Search(r.Context(), indices, query)
	if err != nil {
		slog.Error("search failed", "error", err)
		apierror.Write(w, apierror.Internal("search failed"))
		return
	}

	resp := SearchResponse{
		Hits:  result.Hits,
		Total: result.Total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Tail(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	var req TailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, apierror.BadRequest("invalid JSON: "+err.Error()))
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		apierror.Write(w, apierror.Internal("streaming not supported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	since := time.Now().UTC().Add(-10 * time.Second)
	indices := []string{fmt.Sprintf("mintlog-%s-*", info.ID.String())}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			query := BuildTailQuery(info.ID.String(), &req, since)
			result, err := h.searcher.Search(r.Context(), indices, query)
			if err != nil {
				slog.Error("tail search failed", "error", err)
				continue
			}
			for _, hit := range result.Hits {
				fmt.Fprintf(w, "data: %s\n\n", hit)
				flusher.Flush()
			}
			since = time.Now().UTC()
		}
	}
}

func (h *Handler) Aggregate(w http.ResponseWriter, r *http.Request) {
	info := tenant.FromContext(r.Context())
	if info == nil {
		apierror.Write(w, apierror.Unauthorized("not authenticated"))
		return
	}

	var req AggregateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.Write(w, apierror.BadRequest("invalid JSON: "+err.Error()))
		return
	}

	query := BuildAggregateQuery(info.ID.String(), &req)
	indices := []string{fmt.Sprintf("mintlog-%s-*", info.ID.String())}

	aggs, err := h.searcher.Aggregate(r.Context(), indices, query)
	if err != nil {
		slog.Error("aggregate failed", "error", err)
		apierror.Write(w, apierror.Internal("aggregation failed"))
		return
	}

	resp := AggregateResponse{Buckets: aggs}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
