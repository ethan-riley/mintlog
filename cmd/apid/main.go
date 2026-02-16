package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nats-io/nats.go"

	"github.com/felipemonteiro/mintlog/internal/alerting"
	"github.com/felipemonteiro/mintlog/internal/auth"
	"github.com/felipemonteiro/mintlog/internal/bus"
	"github.com/felipemonteiro/mintlog/internal/config"
	"github.com/felipemonteiro/mintlog/internal/incident"
	mw "github.com/felipemonteiro/mintlog/internal/middleware"
	"github.com/felipemonteiro/mintlog/internal/notification"
	"github.com/felipemonteiro/mintlog/internal/search"
	osstore "github.com/felipemonteiro/mintlog/internal/storage/opensearch"
	"github.com/felipemonteiro/mintlog/internal/storage/postgres"
	"github.com/felipemonteiro/mintlog/internal/storage/postgres/queries"
	redisstore "github.com/felipemonteiro/mintlog/internal/storage/redis"
	"github.com/felipemonteiro/mintlog/internal/tenant"
	"github.com/felipemonteiro/mintlog/pkg/apierror"
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

	// OpenSearch
	osClient, err := osstore.NewClient(cfg.OpenSearch.URL, cfg.OpenSearch.User, cfg.OpenSearch.Password)
	if err != nil {
		slog.Error("opensearch connect failed", "error", err)
		os.Exit(1)
	}

	if err := osstore.EnsureIndexTemplate(ctx, osClient); err != nil {
		slog.Warn("failed to ensure index template (may already exist)", "error", err)
	}

	// Start OpenSearch indexer (consumes logs.parsed from NATS)
	indexer := osstore.NewIndexer(osClient, js)
	if err := indexer.Start(ctx); err != nil {
		slog.Error("failed to start indexer", "error", err)
		os.Exit(1)
	}
	defer indexer.Stop()

	// Auth
	resolver := auth.NewKeyResolver(q, cache)

	// Search
	searcher := osstore.NewSearcher(osClient)
	searchHandler := search.NewHandler(searcher)

	// Alerting
	alertHandler := alerting.NewHandler(q)

	// Notifications
	notifHandler := notification.NewHandler(q)

	// Incidents
	incidentSvc := incident.NewService(q)
	incidentHandler := incident.NewHandler(incidentSvc)

	// Start incident auto-creator (consumes incidents.events from NATS)
	go startIncidentConsumer(js, incidentSvc)

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

		// Search
		r.With(auth.RequireScope(auth.ScopeSearchLogs)).Post("/logs/search", searchHandler.Search)
		r.With(auth.RequireScope(auth.ScopeSearchLogs)).Post("/logs/tail", searchHandler.Tail)
		r.With(auth.RequireScope(auth.ScopeSearchLogs)).Post("/logs/aggregate", searchHandler.Aggregate)

		// Alert Rules
		r.Route("/alerts/rules", func(r chi.Router) {
			r.With(auth.RequireScope(auth.ScopeAlertRead)).Get("/", alertHandler.List)
			r.With(auth.RequireScope(auth.ScopeAlertWrite)).Post("/", alertHandler.Create)
			r.With(auth.RequireScope(auth.ScopeAlertRead)).Get("/{id}", alertHandler.Get)
			r.With(auth.RequireScope(auth.ScopeAlertWrite)).Put("/{id}", alertHandler.Update)
			r.With(auth.RequireScope(auth.ScopeAlertWrite)).Delete("/{id}", alertHandler.Delete)
		})

		// Notification Channels
		r.Route("/notifications/channels", func(r chi.Router) {
			r.With(auth.RequireScope(auth.ScopeNotifRead)).Get("/", notifHandler.List)
			r.With(auth.RequireScope(auth.ScopeNotifWrite)).Post("/", notifHandler.Create)
			r.With(auth.RequireScope(auth.ScopeNotifRead)).Get("/{id}", notifHandler.Get)
			r.With(auth.RequireScope(auth.ScopeNotifWrite)).Delete("/{id}", notifHandler.Delete)
		})

		// Incidents
		r.Route("/incidents", func(r chi.Router) {
			r.With(auth.RequireScope(auth.ScopeIncidentRead)).Get("/", incidentHandler.List)
			r.With(auth.RequireScope(auth.ScopeIncidentWrite)).Post("/", incidentHandler.Create)
			r.With(auth.RequireScope(auth.ScopeIncidentRead)).Get("/{id}", incidentHandler.Get)
			r.With(auth.RequireScope(auth.ScopeIncidentWrite)).Patch("/{id}", incidentHandler.Patch)
			r.With(auth.RequireScope(auth.ScopeIncidentWrite)).Post("/{id}/timeline", incidentHandler.AddTimeline)
		})

		// Admin
		r.Route("/admin", func(r chi.Router) {
			r.Use(auth.RequireScope(auth.ScopeAdmin))
			r.Post("/tenants", adminCreateTenant(q))
			r.Post("/tenants/{id}/keys", adminCreateKey(q))
		})
	})

	srv := &http.Server{
		Addr:         cfg.API.Addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("apid starting", "addr", cfg.API.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down apid")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}

// startIncidentConsumer listens on incidents.events and auto-creates incidents from alerts.
func startIncidentConsumer(js nats.JetStreamContext, svc *incident.Service) {
	_, err := js.QueueSubscribe(
		"incidents.events.>",
		"incident-creators",
		func(msg *nats.Msg) {
			var event struct {
				Type        string    `json:"type"`
				TenantID    uuid.UUID `json:"tenant_id"`
				Title       string    `json:"title"`
				Severity    string    `json:"severity"`
				AlertRuleID uuid.UUID `json:"alert_rule_id"`
			}
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				slog.Error("incident consumer: unmarshal failed", "error", err)
				msg.Nak()
				return
			}

			if event.Type != "auto_create" {
				msg.Ack()
				return
			}

			alertRuleID := pgtype.UUID{Bytes: event.AlertRuleID, Valid: event.AlertRuleID != uuid.Nil}
			_, err := svc.Create(context.Background(), event.TenantID, event.Title, event.Severity, alertRuleID)
			if err != nil {
				slog.Error("incident consumer: failed to create incident", "error", err)
				msg.Nak()
				return
			}

			slog.Info("auto-created incident", "tenant_id", event.TenantID, "title", event.Title)
			msg.Ack()
		},
		nats.Durable("incident-creator"),
		nats.ManualAck(),
		nats.AckWait(30_000_000_000),
		nats.MaxDeliver(3),
	)
	if err != nil {
		slog.Error("failed to subscribe incidents.events", "error", err)
	}
}

// Admin handlers

type createTenantReq struct {
	Name          string `json:"name"`
	Plan          string `json:"plan"`
	RetentionDays int32  `json:"retention_days"`
}

func adminCreateTenant(q *queries.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createTenantReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, apierror.BadRequest("invalid JSON"))
			return
		}
		if req.Plan == "" {
			req.Plan = "free"
		}
		if req.RetentionDays <= 0 {
			req.RetentionDays = 30
		}

		t, err := q.CreateTenant(r.Context(), req.Name, req.Plan, req.RetentionDays)
		if err != nil {
			apierror.Write(w, apierror.Internal("failed to create tenant: "+err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)
	}
}

type createKeyReq struct {
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	RateLimit int32    `json:"rate_limit"`
}

type createKeyResp struct {
	Key    string `json:"key"`
	Prefix string `json:"prefix"`
	ID     string `json:"id"`
}

func adminCreateKey(q *queries.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantIDStr := chi.URLParam(r, "id")
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			apierror.Write(w, apierror.BadRequest("invalid tenant ID"))
			return
		}

		// Verify tenant belongs to authenticated user's tenant (admin scope already checked)
		info := tenant.FromContext(r.Context())
		if info == nil {
			apierror.Write(w, apierror.Unauthorized("not authenticated"))
			return
		}

		var req createKeyReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			apierror.Write(w, apierror.BadRequest("invalid JSON"))
			return
		}
		if req.RateLimit <= 0 {
			req.RateLimit = 1000
		}
		if len(req.Scopes) == 0 {
			req.Scopes = auth.AllScopes
		}

		rawKey := "mlk_" + uuid.New().String()
		h := sha256.Sum256([]byte(rawKey))
		keyHash := hex.EncodeToString(h[:])
		prefix := rawKey[:8]

		k, err := q.CreateAPIKey(r.Context(), tenantID, keyHash, prefix, req.Name, req.Scopes, req.RateLimit, pgtype.Timestamptz{})
		if err != nil {
			apierror.Write(w, apierror.Internal("failed to create API key: "+err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createKeyResp{
			Key:    rawKey,
			Prefix: prefix,
			ID:     k.ID.String(),
		})
	}
}
