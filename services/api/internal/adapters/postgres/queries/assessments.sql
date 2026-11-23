-- name: InsertAssessment :one
INSERT INTO assessments
    (workspace_id, snapshot_id, twin_id, engine, engine_version, posture_score,
     passed, failed, unknown, not_applicable, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: InsertFinding :exec
INSERT INTO findings
    (workspace_id, assessment_id, control_id, title, category, severity, status,
     observed, expected, reason, remediation, techniques)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: ListAssessments :many
SELECT * FROM assessments WHERE workspace_id = $1 ORDER BY created_at DESC LIMIT $2;

-- name: GetAssessment :one
SELECT * FROM assessments WHERE id = $1 AND workspace_id = $2;

-- name: ListFindings :many
SELECT * FROM findings WHERE assessment_id = $1 AND workspace_id = $2 ORDER BY control_id;
