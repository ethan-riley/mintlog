# Mintlog — Claude Code Context

## Project
Go monorepo for a backend log management + observability platform. API-first, no GUI.

## Build & Run
```bash
docker compose up -d          # Start NATS, Postgres, OpenSearch, Redis, MinIO
make migrate-up               # Run Postgres migrations
make build                    # Build all 5 binaries to ./bin/
make test                     # Run unit tests
./scripts/seed.sh             # Create test tenant + API key
```

## Services
| Binary | Port | Purpose |
|--------|------|---------|
| ingestd | 8080 | Ingest gateway (POST /v1/ingest/logs) |
| pipelined | — | NATS consumer: parse/normalize logs |
| apid | 8081 | Query + Management API |
| alertd | — | Cron-based alert evaluator |
| notifierd | — | Webhook notification dispatcher |

## Code Layout
- `cmd/` — Service entrypoints
- `internal/` — All business logic (auth, ingest, pipeline, search, alerting, notification, incident, bus, middleware, storage)
- `pkg/` — Shared types (logmodel, apierror)
- `sql/` — sqlc config + query source files
- `internal/storage/postgres/queries/` — Hand-written sqlc-style Go query files
- `internal/storage/postgres/migrations/` — 5 migration pairs (.up.sql/.down.sql)

## Key Conventions
- Multi-tenant: all queries scoped by tenant_id
- Auth: X-API-Key header → SHA-256 → Redis cache → Postgres
- OpenSearch indices: `mintlog-{tenant_id}-YYYY.MM.DD`
- NATS subjects: `logs.raw.{tenant}`, `logs.parsed.{tenant}`, `alerts.events.{tenant}`, `incidents.events.{tenant}`
- Config via env vars or .env file (Viper), all have sensible defaults for local dev

## Dependencies
- Go 1.23, chi v5, pgx/v5, nats.go, opensearch-go/v4, go-redis/v9, viper, robfig/cron
