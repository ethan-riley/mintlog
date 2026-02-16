package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/felipemonteiro/mintlog/internal/auth"
	"github.com/felipemonteiro/mintlog/internal/bus"
	"github.com/felipemonteiro/mintlog/internal/config"
	"github.com/felipemonteiro/mintlog/internal/ingest"
	mw "github.com/felipemonteiro/mintlog/internal/middleware"
	"github.com/felipemonteiro/mintlog/internal/storage/postgres"
	"github.com/felipemonteiro/mintlog/internal/storage/postgres/queries"
	redisstore "github.com/felipemonteiro/mintlog/internal/storage/redis"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Postgres
	pool, err := postgres.NewPool(ctx, cfg.Postgres.DSN())
	if err != nil {
		slog.Error("postgres connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	q := queries.New(pool)

	// Redis
	rdb, err := redisstore.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		slog.Error("redis connect failed", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()
	cache := redisstore.NewCache(rdb, 5*time.Minute)
	limiter := redisstore.NewRateLimiter(rdb, 1*time.Minute)

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

	// Auth
	resolver := auth.NewKeyResolver(q, cache)

	// Ingest
	pub := bus.NewPublisher(js)
	logPub := ingest.NewLogPublisher(pub)
	ingestHandler := ingest.NewHandler(logPub)

	// Router
	r := chi.NewRouter()
	r.Use(mw.RequestID)
	r.Use(mw.Recovery)
	r.Use(mw.Logging)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.Middleware(resolver))
		r.Use(mw.RateLimit(limiter))

		r.With(auth.RequireScope(auth.ScopeIngestLogs)).
			Post("/ingest/logs", ingestHandler.IngestLogs)
	})

	srv := &http.Server{
		Addr:         cfg.Ingest.Addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("ingestd starting", "addr", cfg.Ingest.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down ingestd")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}
