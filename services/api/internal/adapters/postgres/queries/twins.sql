-- name: CreateTwin :one
INSERT INTO twins (workspace_id, snapshot_id, provider, state)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetTwin :one
SELECT * FROM twins WHERE id = $1 AND workspace_id = $2;

-- name: ListTwins :many
SELECT * FROM twins WHERE workspace_id = $1 ORDER BY created_at DESC;

-- name: UpdateTwinState :one
UPDATE twins SET state = $3, updated_at = now()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: UpdateTwinFidelity :one
UPDATE twins SET fidelity_score = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: NextTwinEventSequence :one
SELECT (coalesce(max(sequence), 0) + 1)::integer FROM twin_events WHERE twin_id = $1;

-- name: InsertTwinEvent :one
INSERT INTO twin_events (workspace_id, twin_id, sequence, state, message)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: ListTwinEvents :many
SELECT * FROM twin_events
WHERE twin_id = $1 AND workspace_id = $2 AND sequence > $3
ORDER BY sequence;
