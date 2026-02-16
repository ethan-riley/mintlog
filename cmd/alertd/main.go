package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/felipemonteiro/mintlog/internal/alerting"
	"github.com/felipemonteiro/mintlog/internal/bus"
	"github.com/felipemonteiro/mintlog/internal/config"
	osstore "github.com/felipemonteiro/mintlog/internal/storage/opensearch"
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

	// OpenSearch
	osClient, err := osstore.NewClient(cfg.OpenSearch.URL, cfg.OpenSearch.User, cfg.OpenSearch.Password)
	if err != nil {
		slog.Error("opensearch connect failed", "error", err)
		os.Exit(1)
	}

	searcher := osstore.NewSearcher(osClient)
	pub := bus.NewPublisher(js)

	evaluator := alerting.NewEvaluator(q, searcher, pub)
	if err := evaluator.Start(); err != nil {
		slog.Error("failed to start evaluator", "error", err)
		os.Exit(1)
	}
	defer evaluator.Stop()

	slog.Info("alertd running")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	slog.Info("shutting down alertd")
}
