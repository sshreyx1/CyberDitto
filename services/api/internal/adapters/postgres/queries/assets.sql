-- name: CreateAsset :one
INSERT INTO assets (workspace_id, name) VALUES ($1, $2) RETURNING *;

-- name: ListAssets :many
SELECT * FROM assets WHERE workspace_id = $1 ORDER BY created_at DESC;

-- name: GetAsset :one
SELECT * FROM assets WHERE id = $1 AND workspace_id = $2;

-- name: CreateSnapshot :one
INSERT INTO snapshots (workspace_id, asset_id, version, data, sha256)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: NextSnapshotVersion :one
SELECT coalesce(max(version), 0) + 1 FROM snapshots WHERE asset_id = $1;

-- name: ListSnapshotsForAsset :many
SELECT * FROM snapshots WHERE asset_id = $1 AND workspace_id = $2 ORDER BY version DESC;

-- name: GetSnapshot :one
SELECT * FROM snapshots WHERE id = $1 AND workspace_id = $2;
