-- name: CreateTwin :one
INSERT INTO twins (workspace_id, snapshot_id, provider, state)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetTwin :one
SELECT * FROM twins WHERE id = $1 AND workspace_id = $2;

-- name: ListTwins :many
SELECT * FROM twins WHERE workspace_id = $1 ORDER BY created_at DESC;

-- name: UpdateTwinState :one
UPDATE twins SET state = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: UpdateTwinFidelity :one
UPDATE twins SET fidelity_score = $2, updated_at = now() WHERE id = $1 RETURNING *;
