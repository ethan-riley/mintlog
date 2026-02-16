CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    query JSONB NOT NULL DEFAULT '{}',
    threshold INT NOT NULL DEFAULT 1,
    window_seconds INT NOT NULL DEFAULT 300,
    eval_interval VARCHAR(20) NOT NULL DEFAULT '1m',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_alert_rules_tenant_id ON alert_rules(tenant_id);

CREATE TABLE IF NOT EXISTS alert_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL UNIQUE REFERENCES alert_rules(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    state VARCHAR(20) NOT NULL DEFAULT 'ok',
    last_value INT NOT NULL DEFAULT 0,
    last_eval_at TIMESTAMPTZ,
    fired_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_alert_states_rule_id ON alert_states(rule_id);
