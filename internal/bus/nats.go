package bus

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

func Connect(url string) (*nats.Conn, nats.JetStreamContext, error) {
	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			slog.Warn("nats disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			slog.Info("nats reconnected", "url", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, nil, fmt.Errorf("jetstream: %w", err)
	}

	return nc, js, nil
}
