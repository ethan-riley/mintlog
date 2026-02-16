package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/felipemonteiro/mintlog/internal/alerting"
	"github.com/felipemonteiro/mintlog/internal/bus"
	"github.com/felipemonteiro/mintlog/internal/storage/postgres/queries"
)

type Dispatcher struct {
	js      nats.JetStreamContext
	pub     *bus.Publisher
	queries *queries.Queries
	webhook *WebhookSender
	sub     *nats.Subscription
}

func NewDispatcher(js nats.JetStreamContext, pub *bus.Publisher, q *queries.Queries) *Dispatcher {
	return &Dispatcher{
		js:      js,
		pub:     pub,
		queries: q,
		webhook: NewWebhookSender(),
	}
}

func (d *Dispatcher) Start() error {
	sub, err := d.js.QueueSubscribe(
		"alerts.events.>",
		"notifier-workers",
		d.handleMessage,
		nats.Durable("notifier-worker"),
		nats.ManualAck(),
		nats.AckWait(60_000_000_000), // 60s for retries
		nats.MaxDeliver(3),
	)
	if err != nil {
		return fmt.Errorf("subscribe alerts.events: %w", err)
	}
	d.sub = sub
	slog.Info("notification dispatcher started")
	return nil
}

func (d *Dispatcher) Stop() {
	if d.sub != nil {
		d.sub.Unsubscribe()
	}
}

func (d *Dispatcher) handleMessage(msg *nats.Msg) {
	var event alerting.AlertEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		slog.Error("dispatcher: unmarshal failed", "error", err)
		msg.Nak()
		return
	}

	ctx := context.Background()

	// Get active channels for this tenant
	channels, err := d.queries.ListActiveChannelsByTenant(ctx, event.TenantID)
	if err != nil {
		slog.Error("dispatcher: failed to list channels", "error", err, "tenant_id", event.TenantID)
		msg.Nak()
		return
	}

	for _, ch := range channels {
		if ch.ChannelType == "webhook" {
			var cfg WebhookConfig
			if err := json.Unmarshal(ch.Config, &cfg); err != nil {
				slog.Error("dispatcher: invalid webhook config", "channel_id", ch.ID, "error", err)
				continue
			}
			if err := d.webhook.Send(cfg, event); err != nil {
				slog.Error("dispatcher: webhook delivery failed", "channel_id", ch.ID, "error", err)
			} else {
				slog.Info("webhook delivered", "channel_id", ch.ID, "rule_id", event.RuleID, "state", event.State)
			}
		}
	}

	// Auto-create incident for firing alerts
	if event.State == alerting.StateFiring {
		d.createIncident(ctx, event)
	}

	msg.Ack()
}

func (d *Dispatcher) createIncident(ctx context.Context, event alerting.AlertEvent) {
	incidentEvent := map[string]any{
		"type":          "auto_create",
		"tenant_id":     event.TenantID,
		"title":         fmt.Sprintf("Alert: %s", event.RuleName),
		"severity":      "high",
		"alert_rule_id": event.RuleID,
	}
	subject := fmt.Sprintf("incidents.events.%s", event.TenantID.String())
	if err := d.pub.Publish(subject, incidentEvent); err != nil {
		slog.Error("dispatcher: failed to publish incident event", "error", err)
	}
}

// EnsureAlertsConsumer creates the consumer for the ALERTS_EVENTS stream needed by the dispatcher.
func EnsureAlertsConsumer(js nats.JetStreamContext) error {
	_, err := js.AddConsumer("ALERTS_EVENTS", &nats.ConsumerConfig{
		Durable:       "notifier-worker",
		DeliverPolicy: nats.DeliverAllPolicy,
		AckPolicy:     nats.AckExplicitPolicy,
		FilterSubject: "alerts.events.>",
		DeliverGroup:  "notifier-workers",
	})
	if err != nil {
		// Ignore "already exists" errors
		_ = uuid.Nil // use uuid to avoid import cycle in a stupid way
		return nil
	}
	return nil
}
