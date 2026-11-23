package ports

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
)

// ErrNotFound is what repositories return when a row doesn't exist in the
// caller's workspace — including when it exists in a different one, which
// must be indistinguishable from not existing at all.
var ErrNotFound = errors.New("not found")

// AssetRepository persists assets and their immutable snapshots.
type AssetRepository interface {
	CreateAsset(ctx context.Context, workspaceID uuid.UUID, name string) (asset.Asset, error)
	ListAssets(ctx context.Context, workspaceID uuid.UUID) ([]asset.Asset, error)
	GetAsset(ctx context.Context, workspaceID, id uuid.UUID) (asset.Asset, error)
	// CreateSnapshot assigns the next version for assetID atomically.
	CreateSnapshot(ctx context.Context, workspaceID, assetID uuid.UUID, data []byte, sha256 string) (asset.Snapshot, error)
	// ListSnapshots returns summaries (Data is nil).
	ListSnapshots(ctx context.Context, workspaceID, assetID uuid.UUID) ([]asset.Snapshot, error)
	GetSnapshot(ctx context.Context, workspaceID, id uuid.UUID) (asset.Snapshot, error)
}
