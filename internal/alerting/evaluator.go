package alerting

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/felipemonteiro/mintlog/internal/bus"
	osstore "github.com/felipemonteiro/mintlog/internal/storage/opensearch"
	"github.com/felipemonteiro/mintlog/internal/storage/postgres/queries"
)

type Evaluator struct {
	queries  *queries.Queries
	searcher *osstore.Searcher
	pub      *bus.Publisher
	cron     *cron.Cron
}

func NewEvaluator(q *queries.Queries, searcher *osstore.Searcher, pub *bus.Publisher) *Evaluator {
	return &Evaluator{
		queries:  q,
		searcher: searcher,
		pub:      pub,
		cron:     cron.New(cron.WithSeconds()),
	}
}

func (e *Evaluator) Start() error {
	// Run evaluation every 30 seconds
	_, err := e.cron.AddFunc("*/30 * * * * *", e.evaluateAll)
	if err != nil {
		return fmt.Errorf("add cron job: %w", err)
	}
	e.cron.Start()
	slog.Info("alert evaluator started")
	return nil
}

func (e *Evaluator) Stop() {
	e.cron.Stop()
}

func (e *Evaluator) evaluateAll() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rules, err := e.queries.ListActiveAlertRules(ctx)
	if err != nil {
		slog.Error("evaluator: failed to list rules", "error", err)
		return
	}

	for _, rule := range rules {
		e.evaluateRule(ctx, rule)
	}
}

func (e *Evaluator) evaluateRule(ctx context.Context, rule queries.AlertRule) {
	// Parse query from rule
	var queryFilter map[string]string
	if err := json.Unmarshal(rule.Query, &queryFilter); err != nil {
		slog.Error("evaluator: invalid rule query", "rule_id", rule.ID, "error", err)
		return
	}

	// Build count query over the window period
	now := time.Now().UTC()
	windowStart := now.Add(-time.Duration(rule.WindowSeconds) * time.Second)

	must := []map[string]any{
		{"term": map[string]any{"tenant_id": rule.TenantID.String()}},
		{"range": map[string]any{
			"timestamp": map[string]any{
				"gte": windowStart.Format(time.RFC3339Nano),
				"lte": now.Format(time.RFC3339Nano),
			},
		}},
	}

	for field, value := range queryFilter {
		if field == "query" || field == "message" {
			must = append(must, map[string]any{
				"match": map[string]any{"message": value},
			})
		} else {
			must = append(must, map[string]any{
				"term": map[string]any{field: value},
			})
		}
	}

	osQuery := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{"must": must},
		},
		"size": 0,
	}

	indices := []string{fmt.Sprintf("mintlog-%s-*", rule.TenantID.String())}
	result, err := e.searcher.Search(ctx, indices, osQuery)
	if err != nil {
		slog.Error("evaluator: search failed", "rule_id", rule.ID, "error", err)
		return
	}

	count := int32(result.Total)

	// Get current state
	currentState := StateOK
	existingState, err := e.queries.GetAlertState(ctx, rule.ID)
	if err == nil {
		currentState = existingState.State
	}

	exceeded := count >= rule.Threshold
	newState := NextState(currentState, exceeded)

	// Update state
	_, err = e.queries.UpsertAlertState(ctx, rule.ID, rule.TenantID, newState, count)
	if err != nil {
		slog.Error("evaluator: failed to upsert state", "rule_id", rule.ID, "error", err)
		return
	}

	// Publish event if state changed
	if newState != currentState && (newState == StateFiring || (currentState == StateFiring && newState == StateOK)) {
		eventState := newState
		if currentState == StateFiring && newState == StateOK {
			eventState = StateResolved
		}
		event := AlertEvent{
			RuleID:    rule.ID,
			TenantID:  rule.TenantID,
			RuleName:  rule.Name,
			State:     eventState,
			Value:     count,
			Threshold: rule.Threshold,
			Timestamp: time.Now().UTC(),
		}
		subject := fmt.Sprintf("alerts.events.%s", rule.TenantID.String())
		if err := e.pub.Publish(subject, event); err != nil {
			slog.Error("evaluator: failed to publish alert event", "rule_id", rule.ID, "error", err)
		} else {
			slog.Info("alert state changed", "rule_id", rule.ID, "rule_name", rule.Name, "from", currentState, "to", eventState, "value", count)
		}
	}
}
