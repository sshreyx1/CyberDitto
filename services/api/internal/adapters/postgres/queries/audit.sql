-- name: GetLastAuditHash :one
SELECT hash FROM audit_events
WHERE workspace_id = $1
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: InsertAuditEvent :one
INSERT INTO audit_events
    (workspace_id, actor_user_id, action, target_type, target_id, metadata, prev_hash, hash)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListAuditEvents :many
SELECT * FROM audit_events
WHERE workspace_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3;
