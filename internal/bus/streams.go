package bus

import (
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

var StreamConfigs = []nats.StreamConfig{
	{
		Name:      "LOGS_RAW",
		Subjects:  []string{"logs.raw.>"},
		Retention: nats.WorkQueuePolicy,
		MaxAge:    24 * time.Hour,
		Storage:   nats.FileStorage,
	},
	{
		Name:      "LOGS_PARSED",
		Subjects:  []string{"logs.parsed.>"},
		Retention: nats.WorkQueuePolicy,
		MaxAge:    24 * time.Hour,
		Storage:   nats.FileStorage,
	},
	{
		Name:      "ALERTS_EVENTS",
		Subjects:  []string{"alerts.events.>"},
		Retention: nats.InterestPolicy,
		MaxAge:    72 * time.Hour,
		Storage:   nats.FileStorage,
	},
	{
		Name:      "INCIDENTS_EVENTS",
		Subjects:  []string{"incidents.events.>"},
		Retention: nats.InterestPolicy,
		MaxAge:    72 * time.Hour,
		Storage:   nats.FileStorage,
	},
}

func EnsureStreams(js nats.JetStreamContext) error {
	for _, cfg := range StreamConfigs {
		_, err := js.StreamInfo(cfg.Name)
		if err == nats.ErrStreamNotFound {
			_, err = js.AddStream(&cfg)
			if err != nil {
				return err
			}
			slog.Info("created stream", "name", cfg.Name)
		} else if err != nil {
			return err
		} else {
			_, err = js.UpdateStream(&cfg)
			if err != nil {
				slog.Warn("failed to update stream", "name", cfg.Name, "error", err)
			}
		}
	}
	return nil
}
