package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
)

// AssetRepo is an in-memory ports.AssetRepository.
type AssetRepo struct {
	mu        sync.Mutex
	assets    map[uuid.UUID]asset.Asset
	snapshots map[uuid.UUID]asset.Snapshot
}

func NewAssetRepo() *AssetRepo {
	return &AssetRepo{assets: map[uuid.UUID]asset.Asset{}, snapshots: map[uuid.UUID]asset.Snapshot{}}
}

func (r *AssetRepo) CreateAsset(ctx context.Context, workspaceID uuid.UUID, name string) (asset.Asset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	a := asset.Asset{ID: uuid.New(), WorkspaceID: workspaceID, Name: name, CreatedAt: now, UpdatedAt: now}
	r.assets[a.ID] = a
	return a, nil
}

func (r *AssetRepo) ListAssets(ctx context.Context, workspaceID uuid.UUID) ([]asset.Asset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]asset.Asset, 0)
	for _, a := range r.assets {
		if a.WorkspaceID == workspaceID {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *AssetRepo) GetAsset(ctx context.Context, workspaceID, id uuid.UUID) (asset.Asset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.assets[id]
	if !ok || a.WorkspaceID != workspaceID {
		return asset.Asset{}, errNotFound
	}
	return a, nil
}

func (r *AssetRepo) CreateSnapshot(ctx context.Context, workspaceID, assetID uuid.UUID, data []byte, sha256 string) (asset.Snapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a, ok := r.assets[assetID]; !ok || a.WorkspaceID != workspaceID {
		return asset.Snapshot{}, errNotFound
	}
	version := 1
	for _, s := range r.snapshots {
		if s.AssetID == assetID && s.Version >= version {
			version = s.Version + 1
		}
	}
	s := asset.Snapshot{
		ID: uuid.New(), WorkspaceID: workspaceID, AssetID: assetID, Version: version,
		SHA256: sha256, Data: append([]byte(nil), data...), CreatedAt: time.Now(),
	}
	r.snapshots[s.ID] = s
	return s, nil
}

func (r *AssetRepo) ListSnapshots(ctx context.Context, workspaceID, assetID uuid.UUID) ([]asset.Snapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]asset.Snapshot, 0)
	for _, s := range r.snapshots {
		if s.AssetID == assetID && s.WorkspaceID == workspaceID {
			s.Data = nil
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Version > out[j].Version })
	return out, nil
}

func (r *AssetRepo) GetSnapshot(ctx context.Context, workspaceID, id uuid.UUID) (asset.Snapshot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.snapshots[id]
	if !ok || s.WorkspaceID != workspaceID {
		return asset.Snapshot{}, errNotFound
	}
	return s, nil
}
