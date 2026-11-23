-- +goose Up
-- +goose StatementBegin

-- One compliance evaluation of one snapshot (docs/MASTER_PLAN.md §10
-- Phase 3). twin_id is set when the assessment was run against a twin;
-- engine/engine_version pin the exact control set so a score is always
-- reproducible and comparable only with like-for-like runs.
CREATE TABLE assessments (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id     uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    snapshot_id      uuid NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE,
    twin_id          uuid REFERENCES twins(id) ON DELETE SET NULL,
    engine           text NOT NULL,
    engine_version   text NOT NULL,
    posture_score    numeric(5,2) CHECK (posture_score BETWEEN 0 AND 100),
    passed           integer NOT NULL,
    failed           integer NOT NULL,
    unknown          integer NOT NULL,
    not_applicable   integer NOT NULL,
    created_by       uuid REFERENCES users(id) ON DELETE SET NULL,
    created_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX assessments_workspace_created_idx ON assessments (workspace_id, created_at DESC);

CREATE TABLE findings (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id   uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    assessment_id  uuid NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    control_id     text NOT NULL,
    title          text NOT NULL,
    category       text NOT NULL,
    severity       text NOT NULL CHECK (severity IN ('high', 'medium', 'low')),
    status         text NOT NULL CHECK (status IN ('pass', 'fail', 'unknown', 'not_applicable')),
    observed       jsonb,
    expected       text NOT NULL,
    reason         text NOT NULL,
    remediation    text NOT NULL,
    techniques     text[] NOT NULL DEFAULT '{}',
    UNIQUE (assessment_id, control_id)
);

ALTER TABLE assessments ENABLE ROW LEVEL SECURITY;
ALTER TABLE findings    ENABLE ROW LEVEL SECURITY;
CREATE POLICY workspace_isolation ON assessments
    USING (workspace_id = current_setting('app.current_workspace_id', true)::uuid);
CREATE POLICY workspace_isolation ON findings
    USING (workspace_id = current_setting('app.current_workspace_id', true)::uuid);

-- Results are evidence: insert and read only.
GRANT SELECT, INSERT ON assessments, findings TO cyberditto_app;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS findings;
DROP TABLE IF EXISTS assessments;
-- +goose StatementEnd
