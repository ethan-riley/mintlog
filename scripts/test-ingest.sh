#!/usr/bin/env bash
set -euo pipefail

KEY="${KEY:?Set KEY env var to your API key}"
INGEST_URL="${INGEST_URL:-http://localhost:8080}"

echo "==> Sending test log events"
curl -s -X POST "$INGEST_URL/v1/ingest/logs" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $KEY" \
  -d '{
    "events": [
      {"service":"web","level":"info","message":"User logged in","fields":{"user_id":"u123"}},
      {"service":"web","level":"error","message":"Database connection timeout","fields":{"db":"primary"}},
      {"service":"api","level":"warn","message":"Slow query detected","fields":{"duration_ms":2500}}
    ]
  }' | jq .

echo ""
echo "==> Done"
