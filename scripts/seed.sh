#!/usr/bin/env bash
set -euo pipefail

PG_URL="${PG_URL:-postgres://mintlog:mintlog@localhost:6543/mintlog?sslmode=disable}"
TENANT_NAME="${TENANT_NAME:-testcorp}"
KEY_NAME="${KEY_NAME:-dev-key}"
RAW_KEY="${RAW_KEY:-mlk_testkey_$(openssl rand -hex 16)}"

echo "==> Creating tenant: $TENANT_NAME"
TENANT_ID=$(psql "$PG_URL" -qtAc "
  INSERT INTO tenants (name, plan, retention_days)
  VALUES ('$TENANT_NAME', 'pro', 90)
  ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
  RETURNING id;
")
echo "    Tenant ID: $TENANT_ID"

KEY_HASH=$(printf '%s' "$RAW_KEY" | shasum -a 256 | cut -d' ' -f1)
KEY_PREFIX="${RAW_KEY:0:8}"

echo "==> Creating API key: $KEY_NAME"
psql "$PG_URL" -c "
  INSERT INTO api_keys (tenant_id, key_hash, key_prefix, name, scopes, rate_limit)
  VALUES ('$TENANT_ID', '$KEY_HASH', '$KEY_PREFIX', '$KEY_NAME',
    ARRAY['ingest:logs','search:logs','alerts:read','alerts:write','incidents:read','incidents:write','notifications:read','notifications:write','admin'],
    10000)
  ON CONFLICT (key_hash) DO NOTHING;
"

echo ""
echo "==> Done!"
echo "    API Key: $RAW_KEY"
echo "    Export it: export KEY=$RAW_KEY"
