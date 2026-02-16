# Mintlog

Backend platform for deep log management (audits, troubleshooting, long retention) and fast observability/incident flow (uptime, alerts, on-call). API-first, no GUI.

## Architecture

```
App --> POST /v1/ingest/logs --> Ingest Gateway (ingestd :8080)
  --> NATS logs.raw.{tenant}
  --> Pipeline Worker (pipelined) -- parse, normalize
  --> NATS logs.parsed.{tenant}
  --> OpenSearch Indexer --> mintlog-{tenant}-YYYY.MM.DD
  --> Query API (apid :8081) reads OpenSearch
  --> Alert Evaluator (alertd) queries OpenSearch on schedule
  --> NATS alerts.events.{tenant}
  --> Notification Dispatcher (notifierd) --> Webhook + auto-create Incident
```

### Services

| Service | Port | Description |
|---------|------|-------------|
| **ingestd** | 8080 | Ingest gateway — accepts log events, publishes to NATS |
| **pipelined** | — | Pipeline worker — parses, normalizes, enriches log events |
| **apid** | 8081 | Query + Management API — search, alerts, notifications, incidents, admin |
| **alertd** | — | Alert evaluator — cron-based query evaluation against OpenSearch |
| **notifierd** | — | Notification dispatcher — webhook delivery with HMAC + retry |

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.23 |
| HTTP Router | chi v5 |
| Message Bus | NATS JetStream |
| Search Store | OpenSearch 2.x |
| Metadata DB | PostgreSQL 16 |
| Cache | Redis 7 |
| Object Storage | MinIO |
| Config | Viper |
| Logging | slog (stdlib) |
| Scheduled Jobs | robfig/cron |

## Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (for migrations)

### 1. Start Infrastructure

```bash
docker compose up -d
```

This starts NATS, PostgreSQL, OpenSearch, Redis, and MinIO.

### 2. Run Migrations

```bash
make migrate-up
```

### 3. Build All Services

```bash
make build
```

Binaries are placed in `./bin/`.

### 4. Seed Test Data

```bash
./scripts/seed.sh
```

This creates a test tenant and API key. Export the key:

```bash
export KEY=<printed API key>
```

### 5. Start Services

In separate terminals (or use `&`):

```bash
./bin/ingestd
./bin/pipelined
./bin/apid
./bin/alertd
./bin/notifierd
```

## API Reference

All authenticated endpoints require the `X-API-Key` header.

### Ingest (ingestd :8080)

#### POST /v1/ingest/logs

Accepts a batch of log events. Returns 202 with accepted/rejected counts.

```bash
curl -X POST http://localhost:8080/v1/ingest/logs \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $KEY" \
  -d '{
    "events": [
      {"service": "web", "level": "error", "message": "connection timeout"},
      {"service": "api", "level": "info", "message": "user logged in", "fields": {"user_id": "u123"}}
    ]
  }'
```

### Query & Management API (apid :8081)

#### Log Search

```bash
# Full-text search with filters
curl -X POST http://localhost:8081/v1/logs/search \
  -H "X-API-Key: $KEY" \
  -d '{"query": "timeout", "level": "error", "size": 10}'

# Real-time tail (SSE stream)
curl -X POST http://localhost:8081/v1/logs/tail \
  -H "X-API-Key: $KEY" \
  -d '{"level": "error"}'

# Aggregations
curl -X POST http://localhost:8081/v1/logs/aggregate \
  -H "X-API-Key: $KEY" \
  -d '{"group_by": "level", "interval": "1h"}'
```

#### Alert Rules

```bash
# Create alert rule
curl -X POST http://localhost:8081/v1/alerts/rules \
  -H "X-API-Key: $KEY" \
  -d '{"name": "high-errors", "query": {"level": "error"}, "threshold": 5, "window_seconds": 300}'

# List rules
curl http://localhost:8081/v1/alerts/rules -H "X-API-Key: $KEY"

# Get rule
curl http://localhost:8081/v1/alerts/rules/{id} -H "X-API-Key: $KEY"

# Update rule
curl -X PUT http://localhost:8081/v1/alerts/rules/{id} \
  -H "X-API-Key: $KEY" \
  -d '{"name": "high-errors", "threshold": 10, "is_active": true}'

# Delete rule
curl -X DELETE http://localhost:8081/v1/alerts/rules/{id} -H "X-API-Key: $KEY"
```

#### Notification Channels

```bash
# Create webhook channel
curl -X POST http://localhost:8081/v1/notifications/channels \
  -H "X-API-Key: $KEY" \
  -d '{"name": "slack-webhook", "channel_type": "webhook", "config": {"url": "https://hooks.example.com/xxx", "secret": "mysecret"}}'

# List channels
curl http://localhost:8081/v1/notifications/channels -H "X-API-Key: $KEY"

# Delete channel
curl -X DELETE http://localhost:8081/v1/notifications/channels/{id} -H "X-API-Key: $KEY"
```

#### Incidents

```bash
# Create incident
curl -X POST http://localhost:8081/v1/incidents \
  -H "X-API-Key: $KEY" \
  -d '{"title": "Database outage", "severity": "critical"}'

# List incidents (optionally filter by status)
curl "http://localhost:8081/v1/incidents?status=triggered" -H "X-API-Key: $KEY"

# Get incident with timeline
curl http://localhost:8081/v1/incidents/{id} -H "X-API-Key: $KEY"

# Acknowledge/resolve incident
curl -X PATCH http://localhost:8081/v1/incidents/{id} \
  -H "X-API-Key: $KEY" \
  -d '{"status": "acknowledged"}'

# Add timeline comment
curl -X POST http://localhost:8081/v1/incidents/{id}/timeline \
  -H "X-API-Key: $KEY" \
  -d '{"event_type": "comment", "content": "Investigating root cause"}'
```

#### Admin

```bash
# Create tenant
curl -X POST http://localhost:8081/v1/admin/tenants \
  -H "X-API-Key: $KEY" \
  -d '{"name": "acme-corp", "plan": "pro", "retention_days": 90}'

# Create API key for tenant
curl -X POST http://localhost:8081/v1/admin/tenants/{tenant_id}/keys \
  -H "X-API-Key: $KEY" \
  -d '{"name": "production-key", "rate_limit": 5000}'
```

### Health Check

Both services expose an unauthenticated health endpoint:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8081/healthz
```

## Authentication

All API endpoints use API key authentication via the `X-API-Key` header.

**Flow:** API key -> SHA-256 hash -> Redis cache (5min TTL) -> Postgres fallback -> tenant context injected into request.

**Scopes:** `ingest:logs`, `search:logs`, `alerts:read`, `alerts:write`, `incidents:read`, `incidents:write`, `notifications:read`, `notifications:write`, `admin`

## Data Model

### NATS JetStream Streams

| Stream | Subjects | Purpose |
|--------|----------|---------|
| LOGS_RAW | `logs.raw.>` | Raw ingested events |
| LOGS_PARSED | `logs.parsed.>` | Normalized events |
| ALERTS_EVENTS | `alerts.events.>` | Alert state changes |
| INCIDENTS_EVENTS | `incidents.events.>` | Incident lifecycle |

### Postgres Tables

1. **tenants** — id, name, plan, retention_days
2. **api_keys** — key_hash (SHA-256), scopes, rate_limit, expiry
3. **alert_rules** + **alert_states** — rule config + state machine (ok/firing/resolved)
4. **incidents** + **incident_timeline** — status machine (triggered/acknowledged/resolved)
5. **notification_channels** — webhook config (url, headers, HMAC secret)

### OpenSearch Indices

Per-tenant, date-partitioned: `mintlog-{tenant_id}-YYYY.MM.DD`

## Project Structure

```
new_log/
├── cmd/                           # Service entrypoints
│   ├── ingestd/main.go
│   ├── pipelined/main.go
│   ├── apid/main.go
│   ├── alertd/main.go
│   └── notifierd/main.go
├── internal/
│   ├── config/                    # Viper-based configuration
│   ├── auth/                      # API key auth + scope middleware
│   ├── tenant/                    # Tenant context helpers
│   ├── ingest/                    # Ingest handler + validation + NATS publishing
│   ├── pipeline/                  # Parse, normalize, enrich log events
│   ├── storage/
│   │   ├── opensearch/            # Client, indexer, searcher, mappings
│   │   ├── postgres/              # Pool, migrations, query layer
│   │   └── redis/                 # Client, cache, rate limiter
│   ├── search/                    # Search API handlers + query builder
│   ├── alerting/                  # Alert rules, evaluator, state machine
│   ├── notification/              # Webhook sender, dispatcher, channel CRUD
│   ├── incident/                  # Incident service, timeline, CRUD
│   ├── bus/                       # NATS connection, streams, publisher
│   └── middleware/                # Logging, recovery, request ID, rate limit
├── pkg/
│   ├── logmodel/                  # Canonical LogEvent struct
│   └── apierror/                  # Standard API error format
├── sql/                           # sqlc config + query source
├── scripts/                       # Seed + test scripts
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Makefile Targets

```bash
make build          # Build all 5 binaries
make test           # Run unit tests
make lint           # Run golangci-lint
make migrate-up     # Apply all migrations
make migrate-down   # Roll back 1 migration
make sqlc           # Regenerate sqlc queries
make infra          # docker compose up -d
make infra-down     # docker compose down
make seed           # Run seed script
```

## Configuration

Configuration is loaded from environment variables (or a `.env` file). All values have sensible defaults for local development. See `.env.example` for the full list.

## License

Private.
