-- name: CountUsers :one
SELECT count(*) FROM users;

-- name: CreateWorkspace :one
INSERT INTO workspaces (name) VALUES ($1) RETURNING *;

-- name: CreateUser :one
INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateMembership :one
INSERT INTO memberships (workspace_id, user_id, role, approver)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetMembership :one
SELECT * FROM memberships WHERE workspace_id = $1 AND user_id = $2;

-- name: GetMembershipsForUser :many
SELECT * FROM memberships WHERE user_id = $1 ORDER BY created_at;

-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, user_agent, expires_at)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetSessionByTokenHash :one
SELECT * FROM sessions
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now();

-- name: RevokeSession :exec
UPDATE sessions SET revoked_at = now() WHERE id = $1;

-- name: RevokeAllSessionsForUser :exec
UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL;
