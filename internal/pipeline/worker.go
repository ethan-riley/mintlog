package pipeline

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"

	"github.com/felipemonteiro/mintlog/internal/bus"
	"github.com/felipemonteiro/mintlog/pkg/logmodel"
)

type Worker struct {
	js  nats.JetStreamContext
	pub *bus.Publisher
	sub *nats.Subscription
}

func NewWorker(js nats.JetStreamContext, pub *bus.Publisher) *Worker {
	return &Worker{js: js, pub: pub}
}

func (w *Worker) Start() error {
	sub, err := w.js.QueueSubscribe(
		"logs.raw.>",
		"pipeline-workers",
		w.handleMessage,
		nats.Durable("pipeline-worker"),
		nats.ManualAck(),
		nats.AckWait(30_000_000_000), // 30s
		nats.MaxDeliver(3),
	)
	if err != nil {
		return fmt.Errorf("subscribe logs.raw: %w", err)
	}
	w.sub = sub
	slog.Info("pipeline worker started", "subject", "logs.raw.>")
	return nil
}

func (w *Worker) Stop() {
	if w.sub != nil {
		w.sub.Unsubscribe()
	}
}

func (w *Worker) handleMessage(msg *nats.Msg) {
	var event logmodel.LogEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		slog.Error("pipeline: failed to unmarshal event", "error", err)
		msg.Nak()
		return
	}

	// Parse, normalize, enrich
	ParseJSON(&event)
	Normalize(&event)
	Enrich(&event)

	// Publish to logs.parsed
	subject := fmt.Sprintf("logs.parsed.%s", event.TenantID)
	if err := w.pub.Publish(subject, &event); err != nil {
		slog.Error("pipeline: failed to publish parsed event", "error", err)
		msg.Nak()
		return
	}

	msg.Ack()
}
