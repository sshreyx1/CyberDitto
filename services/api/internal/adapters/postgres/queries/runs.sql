-- name: InsertRun :one
INSERT INTO emulation_runs
    (workspace_id, twin_id, scenario_id, test_guids, status, detail, preflight,
     catalog_source, catalog_commit, provider, authorised_by, authorisation_statement)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: FinishRun :execrows
UPDATE emulation_runs SET status = $3, detail = $4, finished_at = $5
WHERE id = $1 AND workspace_id = $2;

-- name: InsertExecution :exec
INSERT INTO run_executions
    (workspace_id, run_id, test_guid, technique_id, test_name, risk_tier, outcome, detail, started_at, finished_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetRun :one
SELECT * FROM emulation_runs WHERE id = $1 AND workspace_id = $2;

-- name: ListRuns :many
SELECT * FROM emulation_runs WHERE workspace_id = $1 ORDER BY created_at DESC LIMIT $2;

-- name: ListExecutions :many
SELECT * FROM run_executions WHERE run_id = $1 AND workspace_id = $2 ORDER BY seq;
