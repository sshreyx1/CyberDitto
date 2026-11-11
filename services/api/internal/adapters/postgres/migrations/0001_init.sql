-- +goose Up
-- +goose StatementBegin

-- Dedicated non-superuser application role (docs/MASTER_PLAN.md §6.2: the app
-- connects as a non-superuser; RLS is defence in depth on top of app-layer authZ).
-- The role is created without LOGIN privileges here; a deployment sets its
-- password out of band (never in migrations) and grants LOGIN separately.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'cyberditto_app') THEN
    CREATE ROLE cyberditto_app NOSUPERUSER NOCREATEDB NOCREATEROLE NOLOGIN;
  END IF;
END
$$;

CREATE TABLE workspaces (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email          text NOT NULL UNIQUE,
    password_hash  text NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE memberships (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role          text NOT NULL CHECK (role IN ('owner', 'admin', 'analyst', 'viewer')),
    approver      boolean NOT NULL DEFAULT false,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, user_id)
);

-- Sessions store only a hash of the session token (docs/MASTER_PLAN.md §6.2);
-- the plaintext token exists only in the httpOnly cookie on the client.
CREATE TABLE sessions (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash   text NOT NULL UNIQUE,
    user_agent   text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL,
    revoked_at   timestamptz
);

-- Append-only, hash-chained audit log (docs/MASTER_PLAN.md §5.6 invariants,
-- §6.4). workspace_id is nullable because some events (failed login, setup)
-- happen before a workspace/membership exists.
CREATE TABLE audit_events (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    uuid REFERENCES workspaces(id) ON DELETE CASCADE,
    actor_user_id   uuid REFERENCES users(id) ON DELETE SET NULL,
    action          text NOT NULL,
    target_type     text NOT NULL,
    target_id       text,
    metadata        jsonb NOT NULL DEFAULT '{}'::jsonb,
    prev_hash       text NOT NULL,
    hash            text NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_events_workspace_id_created_at_idx ON audit_events (workspace_id, created_at);

CREATE TABLE assets (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name          text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- Snapshots are immutable (docs/MASTER_PLAN.md §5.6): no updated_at, no UPDATE
-- grants in app-role privileges; a new version is a new row.
CREATE TABLE snapshots (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id  uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    asset_id      uuid NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    version       integer NOT NULL,
    data          jsonb NOT NULL,
    sha256        text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (asset_id, version)
);

CREATE TABLE twins (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id    uuid NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    snapshot_id     uuid NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE,
    provider        text NOT NULL DEFAULT 'fake',
    state           text NOT NULL DEFAULT 'requested'
                      CHECK (state IN ('requested', 'imaging', 'provisioning', 'applying',
                                        'verifying', 'ready', 'running', 'stopped',
                                        'failed', 'destroyed')),
    fidelity_score  numeric,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

-- Row-level security: workspace-scoped tables only (users/workspaces/sessions
-- are identity tables keyed by their own id, not workspace-scoped).
ALTER TABLE memberships  ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE assets       ENABLE ROW LEVEL SECURITY;
ALTER TABLE snapshots    ENABLE ROW LEVEL SECURITY;
ALTER TABLE twins        ENABLE ROW LEVEL SECURITY;

-- Memberships additionally allow a user to see their own membership rows
-- across workspaces (needed to list "which workspaces am I in" before one
-- is selected as the active context) without exposing other users' rows.
CREATE POLICY workspace_isolation ON memberships
    USING (
        workspace_id = current_setting('app.current_workspace_id', true)::uuid
        OR user_id = current_setting('app.current_user_id', true)::uuid
    );
CREATE POLICY workspace_isolation ON audit_events
    USING (workspace_id = current_setting('app.current_workspace_id', true)::uuid);
CREATE POLICY workspace_isolation ON assets
    USING (workspace_id = current_setting('app.current_workspace_id', true)::uuid);
CREATE POLICY workspace_isolation ON snapshots
    USING (workspace_id = current_setting('app.current_workspace_id', true)::uuid);
CREATE POLICY workspace_isolation ON twins
    USING (workspace_id = current_setting('app.current_workspace_id', true)::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON workspaces, users, sessions TO cyberditto_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON memberships, assets, twins TO cyberditto_app;
GRANT SELECT, INSERT ON audit_events, snapshots TO cyberditto_app;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS twins;
DROP TABLE IF EXISTS snapshots;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS audit_events;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS workspaces;
DROP ROLE IF EXISTS cyberditto_app;
-- +goose StatementEnd
