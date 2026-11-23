-- +goose Up
-- +goose StatementBegin

-- An attack-emulation run (docs/MASTER_PLAN.md §5.12). twin_id is NOT NULL
-- and references twins: at the schema level, a run can only ever target a
-- twin row (§6.4 invariant 1). Preflight results and the authorisation
-- statement are stored with the run, including runs preflight refused.
CREATE TABLE emulation_runs (
    id                        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id              uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    twin_id                   uuid NOT NULL REFERENCES twins(id) ON DELETE CASCADE,
    scenario_id               text NOT NULL DEFAULT '',
    test_guids                text[] NOT NULL,
    status                    text NOT NULL CHECK (status IN ('preflight_failed', 'running', 'completed', 'aborted', 'failed')),
    detail                    text NOT NULL DEFAULT '',
    preflight                 jsonb NOT NULL,
    catalog_source            text NOT NULL,
    catalog_commit            text NOT NULL,
    provider                  text NOT NULL,
    authorised_by             uuid REFERENCES users(id) ON DELETE SET NULL,
    authorisation_statement   text NOT NULL DEFAULT '',
    created_at                timestamptz NOT NULL DEFAULT now(),
    finished_at               timestamptz
);
CREATE INDEX emulation_runs_workspace_created_idx ON emulation_runs (workspace_id, created_at DESC);

CREATE TABLE run_executions (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id   uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    run_id         uuid NOT NULL REFERENCES emulation_runs(id) ON DELETE CASCADE,
    test_guid      text NOT NULL,
    technique_id   text NOT NULL,
    test_name      text NOT NULL,
    risk_tier      text NOT NULL CHECK (risk_tier IN ('benign', 'intrusive')),
    outcome        text NOT NULL CHECK (outcome IN ('blocked', 'detected', 'logged', 'missed', 'not_executed', 'error', 'skipped')),
    detail         text NOT NULL DEFAULT '',
    started_at     timestamptz NOT NULL,
    finished_at    timestamptz NOT NULL,
    seq            bigserial
);
CREATE INDEX run_executions_run_idx ON run_executions (run_id, seq);

ALTER TABLE emulation_runs ENABLE ROW LEVEL SECURITY;
ALTER TABLE run_executions ENABLE ROW LEVEL SECURITY;
CREATE POLICY workspace_isolation ON emulation_runs
    USING (workspace_id = current_setting('app.current_workspace_id', true)::uuid);
CREATE POLICY workspace_isolation ON run_executions
    USING (workspace_id = current_setting('app.current_workspace_id', true)::uuid);

-- Runs change status as they progress; executions are evidence and never
-- change once written. Note the risk_tier CHECK: 'destructive' cannot even
-- be stored as an execution.
GRANT SELECT, INSERT, UPDATE ON emulation_runs TO cyberditto_app;
GRANT SELECT, INSERT ON run_executions TO cyberditto_app;
GRANT USAGE ON SEQUENCE run_executions_seq_seq TO cyberditto_app;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS run_executions;
DROP TABLE IF EXISTS emulation_runs;
-- +goose StatementEnd
