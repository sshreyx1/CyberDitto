-- +goose Up
-- +goose StatementBegin

-- Append-only lifecycle event log per twin (docs/MASTER_PLAN.md §5.1:
-- event-sourced runs/builds, streamed over SSE). sequence is per twin and
-- strictly increasing so a reconnecting client can resume "after N".
CREATE TABLE twin_events (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    twin_id       uuid NOT NULL REFERENCES twins(id) ON DELETE CASCADE,
    sequence      integer NOT NULL,
    state         text NOT NULL,
    message       text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (twin_id, sequence)
);

ALTER TABLE twin_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY workspace_isolation ON twin_events
    USING (workspace_id = current_setting('app.current_workspace_id', true)::uuid);

-- Events are history: insert and read only, like audit_events.
GRANT SELECT, INSERT ON twin_events TO cyberditto_app;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS twin_events;
-- +goose StatementEnd
