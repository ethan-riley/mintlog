package search

import "time"

// BuildSearchQuery constructs an OpenSearch DSL query from a SearchRequest.
func BuildSearchQuery(tenantID string, req *SearchRequest) map[string]any {
	must := []map[string]any{
		{"term": map[string]any{"tenant_id": tenantID}},
	}

	if req.Query != "" {
		must = append(must, map[string]any{
			"match": map[string]any{
				"message": map[string]any{
					"query":    req.Query,
					"operator": "and",
				},
			},
		})
	}

	if req.Level != "" {
		must = append(must, map[string]any{"term": map[string]any{"level": req.Level}})
	}
	if req.Service != "" {
		must = append(must, map[string]any{"term": map[string]any{"service": req.Service}})
	}
	if req.Host != "" {
		must = append(must, map[string]any{"term": map[string]any{"host": req.Host}})
	}
	if req.TraceID != "" {
		must = append(must, map[string]any{"term": map[string]any{"trace_id": req.TraceID}})
	}

	// Time range filter
	timeRange := map[string]any{}
	if !req.From.IsZero() {
		timeRange["gte"] = req.From.Format(time.RFC3339Nano)
	}
	if !req.To.IsZero() {
		timeRange["lte"] = req.To.Format(time.RFC3339Nano)
	}
	if len(timeRange) > 0 {
		must = append(must, map[string]any{
			"range": map[string]any{"timestamp": timeRange},
		})
	}

	size := req.Size
	if size <= 0 || size > 1000 {
		size = 50
	}

	sortOrder := "desc"
	if req.Sort == "asc" {
		sortOrder = "asc"
	}

	query := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": must,
			},
		},
		"size": size,
		"sort": []map[string]any{
			{"timestamp": map[string]any{"order": sortOrder}},
			{"id": map[string]any{"order": sortOrder}},
		},
	}

	if len(req.SearchAfter) > 0 {
		query["search_after"] = req.SearchAfter
	}

	return query
}

// BuildAggregateQuery constructs an OpenSearch aggregation DSL.
func BuildAggregateQuery(tenantID string, req *AggregateRequest) map[string]any {
	must := []map[string]any{
		{"term": map[string]any{"tenant_id": tenantID}},
	}

	if req.Query != "" {
		must = append(must, map[string]any{
			"match": map[string]any{"message": req.Query},
		})
	}
	if req.Level != "" {
		must = append(must, map[string]any{"term": map[string]any{"level": req.Level}})
	}
	if req.Service != "" {
		must = append(must, map[string]any{"term": map[string]any{"service": req.Service}})
	}

	timeRange := map[string]any{}
	if !req.From.IsZero() {
		timeRange["gte"] = req.From.Format(time.RFC3339Nano)
	}
	if !req.To.IsZero() {
		timeRange["lte"] = req.To.Format(time.RFC3339Nano)
	}
	if len(timeRange) > 0 {
		must = append(must, map[string]any{
			"range": map[string]any{"timestamp": timeRange},
		})
	}

	aggs := map[string]any{}

	if req.GroupBy != "" {
		aggs["group_by"] = map[string]any{
			"terms": map[string]any{
				"field": req.GroupBy,
				"size":  20,
			},
		}
	}

	if req.Interval != "" {
		aggs["over_time"] = map[string]any{
			"date_histogram": map[string]any{
				"field":          "timestamp",
				"fixed_interval": req.Interval,
			},
		}
	}

	// Default: count by level
	if len(aggs) == 0 {
		aggs["group_by"] = map[string]any{
			"terms": map[string]any{
				"field": "level",
				"size":  20,
			},
		}
	}

	return map[string]any{
		"query": map[string]any{
			"bool": map[string]any{"must": must},
		},
		"size":         0,
		"aggregations": aggs,
	}
}

// BuildTailQuery builds a query for the last N seconds of logs.
func BuildTailQuery(tenantID string, req *TailRequest, since time.Time) map[string]any {
	must := []map[string]any{
		{"term": map[string]any{"tenant_id": tenantID}},
		{"range": map[string]any{
			"timestamp": map[string]any{"gt": since.Format(time.RFC3339Nano)},
		}},
	}

	if req.Query != "" {
		must = append(must, map[string]any{
			"match": map[string]any{"message": req.Query},
		})
	}
	if req.Level != "" {
		must = append(must, map[string]any{"term": map[string]any{"level": req.Level}})
	}
	if req.Service != "" {
		must = append(must, map[string]any{"term": map[string]any{"service": req.Service}})
	}

	return map[string]any{
		"query": map[string]any{
			"bool": map[string]any{"must": must},
		},
		"size": 100,
		"sort": []map[string]any{
			{"timestamp": map[string]any{"order": "asc"}},
		},
	}
}
