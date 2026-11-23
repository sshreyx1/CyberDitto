package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgdb "github.com/sshreyx1/cyberditto/services/api/internal/adapters/postgres/gen"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

type AssetRepo struct {
	pool *pgxpool.Pool
}

func NewAssetRepo(pool *pgxpool.Pool) *AssetRepo {
	return &AssetRepo{pool: pool}
}

// notFound maps "no row" to ports.ErrNotFound. Under RLS, a row in another
// workspace also comes back as no row, which is exactly the
// indistinguishability ports.ErrNotFound promises.
func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ErrNotFound
	}
	return err
}

func (r *AssetRepo) CreateAsset(ctx context.Context, workspaceID uuid.UUID, name string) (asset.Asset, error) {
	var row pgdb.Asset
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		row, err = q.CreateAsset(ctx, pgdb.CreateAssetParams{WorkspaceID: workspaceID, Name: name})
		return err
	})
	if err != nil {
		return asset.Asset{}, err
	}
	return assetFromRow(row), nil
}

func (r *AssetRepo) ListAssets(ctx context.Context, workspaceID uuid.UUID) ([]asset.Asset, error) {
	var rows []pgdb.Asset
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		rows, err = q.ListAssets(ctx, workspaceID)
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]asset.Asset, 0, len(rows))
	for _, row := range rows {
		out = append(out, assetFromRow(row))
	}
	return out, nil
}

func (r *AssetRepo) GetAsset(ctx context.Context, workspaceID, id uuid.UUID) (asset.Asset, error) {
	var row pgdb.Asset
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		row, err = q.GetAsset(ctx, pgdb.GetAssetParams{ID: id, WorkspaceID: workspaceID})
		return err
	})
	if err != nil {
		return asset.Asset{}, notFound(err)
	}
	return assetFromRow(row), nil
}

// CreateSnapshot takes a per-asset advisory lock so concurrent uploads for
// the same asset can't both read the same "next version" (the UNIQUE
// (asset_id, version) constraint would reject the loser anyway; the lock
// turns that into a clean serialisation instead of an error).
func (r *AssetRepo) CreateSnapshot(ctx context.Context, workspaceID, assetID uuid.UUID, data []byte, sha256 string) (asset.Snapshot, error) {
	var row pgdb.Snapshot
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		if _, err := q.GetAsset(ctx, pgdb.GetAssetParams{ID: assetID, WorkspaceID: workspaceID}); err != nil {
			return notFound(err)
		}
		if err := advisoryLock(ctx, q, "snapshot:"+assetID.String()); err != nil {
			return err
		}
		version, err := q.NextSnapshotVersion(ctx, assetID)
		if err != nil {
			return err
		}
		row, err = q.CreateSnapshot(ctx, pgdb.CreateSnapshotParams{
			WorkspaceID: workspaceID,
			AssetID:     assetID,
			Version:     version,
			Data:        data,
			Sha256:      sha256,
		})
		return err
	})
	if err != nil {
		return asset.Snapshot{}, err
	}
	return snapshotFromRow(row), nil
}

func (r *AssetRepo) ListSnapshots(ctx context.Context, workspaceID, assetID uuid.UUID) ([]asset.Snapshot, error) {
	var rows []pgdb.ListSnapshotSummariesForAssetRow
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		rows, err = q.ListSnapshotSummariesForAsset(ctx, pgdb.ListSnapshotSummariesForAssetParams{AssetID: assetID, WorkspaceID: workspaceID})
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]asset.Snapshot, 0, len(rows))
	for _, row := range rows {
		out = append(out, asset.Snapshot{
			ID: row.ID, WorkspaceID: row.WorkspaceID, AssetID: row.AssetID,
			Version: int(row.Version), SHA256: row.Sha256, CreatedAt: row.CreatedAt.Time,
		})
	}
	return out, nil
}

func (r *AssetRepo) GetSnapshot(ctx context.Context, workspaceID, id uuid.UUID) (asset.Snapshot, error) {
	var row pgdb.Snapshot
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		row, err = q.GetSnapshot(ctx, pgdb.GetSnapshotParams{ID: id, WorkspaceID: workspaceID})
		return err
	})
	if err != nil {
		return asset.Snapshot{}, notFound(err)
	}
	return snapshotFromRow(row), nil
}

func assetFromRow(row pgdb.Asset) asset.Asset {
	return asset.Asset{
		ID: row.ID, WorkspaceID: row.WorkspaceID, Name: row.Name,
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func snapshotFromRow(row pgdb.Snapshot) asset.Snapshot {
	return asset.Snapshot{
		ID: row.ID, WorkspaceID: row.WorkspaceID, AssetID: row.AssetID,
		Version: int(row.Version), SHA256: row.Sha256, Data: row.Data,
		CreatedAt: row.CreatedAt.Time,
	}
}
