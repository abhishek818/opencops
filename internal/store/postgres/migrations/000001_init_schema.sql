-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    slack_channel TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    owner_team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    repo_url TEXT NOT NULL DEFAULT '',
    runbook_url TEXT NOT NULL DEFAULT '',
    slack_channel TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    severity TEXT NOT NULL CHECK (severity IN ('SEV1', 'SEV2', 'SEV3', 'SEV4', 'SEV5')),
    status TEXT NOT NULL CHECK (status IN ('triggered', 'acknowledged', 'investigating', 'mitigated', 'resolved', 'closed')),
    service_id UUID REFERENCES services(id) ON DELETE SET NULL,
    owner_team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    commander_user_id TEXT,
    dedup_key TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    acknowledged_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source TEXT NOT NULL,
    external_alert_id TEXT NOT NULL DEFAULT '',
    fingerprint TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    severity TEXT NOT NULL CHECK (severity IN ('SEV1', 'SEV2', 'SEV3', 'SEV4', 'SEV5')),
    status TEXT NOT NULL CHECK (status IN ('firing', 'resolved')),
    service_id UUID REFERENCES services(id) ON DELETE SET NULL,
    incident_id UUID REFERENCES incidents(id) ON DELETE SET NULL,
    labels_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    annotations_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    starts_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE incident_timeline_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    message TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'system',
    actor_user_id TEXT,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER teams_set_updated_at
BEFORE UPDATE ON teams
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER services_set_updated_at
BEFORE UPDATE ON services
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER incidents_set_updated_at
BEFORE UPDATE ON incidents
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER alerts_set_updated_at
BEFORE UPDATE ON alerts
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_services_owner_team_id
ON services(owner_team_id);

CREATE INDEX idx_incidents_status
ON incidents(status);

CREATE INDEX idx_incidents_severity
ON incidents(severity);

CREATE INDEX idx_incidents_service_id
ON incidents(service_id);

CREATE INDEX idx_incidents_owner_team_id
ON incidents(owner_team_id);

CREATE INDEX idx_incidents_created_at
ON incidents(created_at DESC);

CREATE UNIQUE INDEX idx_incidents_active_dedup_key
ON incidents(dedup_key)
WHERE status IN ('triggered', 'acknowledged', 'investigating', 'mitigated');

CREATE INDEX idx_alerts_source
ON alerts(source);

CREATE INDEX idx_alerts_status
ON alerts(status);

CREATE INDEX idx_alerts_service_id
ON alerts(service_id);

CREATE INDEX idx_alerts_incident_id
ON alerts(incident_id);

CREATE INDEX idx_alerts_fingerprint
ON alerts(fingerprint);

CREATE INDEX idx_alerts_received_at
ON alerts(received_at DESC);

CREATE INDEX idx_alerts_labels_json
ON alerts USING GIN(labels_json);

CREATE INDEX idx_alerts_annotations_json
ON alerts USING GIN(annotations_json);

CREATE INDEX idx_incident_timeline_events_incident_id
ON incident_timeline_events(incident_id);

CREATE INDEX idx_incident_timeline_events_created_at
ON incident_timeline_events(created_at ASC);

CREATE INDEX idx_incident_timeline_events_event_type
ON incident_timeline_events(event_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_incident_timeline_events_event_type;
DROP INDEX IF EXISTS idx_incident_timeline_events_created_at;
DROP INDEX IF EXISTS idx_incident_timeline_events_incident_id;

DROP INDEX IF EXISTS idx_alerts_annotations_json;
DROP INDEX IF EXISTS idx_alerts_labels_json;
DROP INDEX IF EXISTS idx_alerts_received_at;
DROP INDEX IF EXISTS idx_alerts_fingerprint;
DROP INDEX IF EXISTS idx_alerts_incident_id;
DROP INDEX IF EXISTS idx_alerts_service_id;
DROP INDEX IF EXISTS idx_alerts_status;
DROP INDEX IF EXISTS idx_alerts_source;

DROP INDEX IF EXISTS idx_incidents_active_dedup_key;
DROP INDEX IF EXISTS idx_incidents_created_at;
DROP INDEX IF EXISTS idx_incidents_owner_team_id;
DROP INDEX IF EXISTS idx_incidents_service_id;
DROP INDEX IF EXISTS idx_incidents_severity;
DROP INDEX IF EXISTS idx_incidents_status;

DROP INDEX IF EXISTS idx_services_owner_team_id;

DROP TABLE IF EXISTS incident_timeline_events;
DROP TABLE IF EXISTS alerts;
DROP TABLE IF EXISTS incidents;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS teams;

DROP FUNCTION IF EXISTS set_updated_at();

-- +goose StatementEnd