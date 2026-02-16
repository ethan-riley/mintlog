package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/felipemonteiro/mintlog/internal/bus"
	"github.com/felipemonteiro/mintlog/internal/config"
	"github.com/felipemonteiro/mintlog/internal/notification"
	"github.com/felipemonteiro/mintlog/internal/storage/postgres"
	"github.com/felipemonteiro/mintlog/internal/storage/postgres/queries"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Postgres
	pool, err := postgres.NewPool(ctx, cfg.Postgres.DSN())
	if err != nil {
		slog.Error("postgres connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	q := queries.New(pool)

	// NATS
	nc, js, err := bus.Connect(cfg.NATS.URL)
	if err != nil {
		slog.Error("nats connect failed", "error", err)
		os.Exit(1)
	}
	defer nc.Close()

	if err := bus.EnsureStreams(js); err != nil {
		slog.Error("failed to ensure streams", "error", err)
		os.Exit(1)
	}

	pub := bus.NewPublisher(js)
	dispatcher := notification.NewDispatcher(js, pub, q)

	if err := dispatcher.Start(); err != nil {
		slog.Error("failed to start dispatcher", "error", err)
		os.Exit(1)
	}
	defer dispatcher.Stop()

	slog.Info("notifierd running")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	slog.Info("shutting down notifierd")
}
