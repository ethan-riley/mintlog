package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/felipemonteiro/mintlog/internal/bus"
	"github.com/felipemonteiro/mintlog/internal/config"
	"github.com/felipemonteiro/mintlog/internal/pipeline"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

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
	worker := pipeline.NewWorker(js, pub)

	if err := worker.Start(); err != nil {
		slog.Error("failed to start pipeline worker", "error", err)
		os.Exit(1)
	}
	defer worker.Stop()

	slog.Info("pipelined running")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	slog.Info("shutting down pipelined")
}
