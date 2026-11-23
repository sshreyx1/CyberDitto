// Package assets implements asset registration and snapshot upload
// (docs/MASTER_PLAN.md §4.2 step 2, §6.3 file-upload rules).
package assets

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

// ErrDigestMismatch means the uploader declared a SHA-256 for the body
// (the digest the collector prints) and the received bytes don't match it:
// the snapshot was altered or truncated in transit.
var ErrDigestMismatch = errors.New("assets: snapshot digest does not match the declared sha256")

type Service struct {
	repo  ports.AssetRepository
	audit ports.AuditRepository
}

func NewService(repo ports.AssetRepository, audit ports.AuditRepository) *Service {
	return &Service{repo: repo, audit: audit}
}

func (s *Service) Create(ctx context.Context, actorID, workspaceID uuid.UUID, name string) (asset.Asset, error) {
	name, err := asset.ValidateName(name)
	if err != nil {
		return asset.Asset{}, err
	}
	a, err := s.repo.CreateAsset(ctx, workspaceID, name)
	if err != nil {
		return asset.Asset{}, fmt.Errorf("assets: create: %w", err)
	}
	if _, err := s.audit.Record(ctx, workspaceID, &actorID, "asset.create", "asset", a.ID.String(), map[string]any{"name": name}); err != nil {
		return asset.Asset{}, err
	}
	return a, nil
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID) ([]asset.Asset, error) {
	return s.repo.ListAssets(ctx, workspaceID)
}

func (s *Service) Get(ctx context.Context, workspaceID, id uuid.UUID) (asset.Asset, error) {
	return s.repo.GetAsset(ctx, workspaceID, id)
}

// UploadSnapshot validates a collector document and stores it as the next
// immutable version for the asset. declaredSHA256, if non-empty, must match
// the body's digest — this is an integrity check (corruption/truncation in
// transit), not an authenticity check; collector output signing is still
// to come (docs/PROGRESS.md).
func (s *Service) UploadSnapshot(ctx context.Context, actorID, workspaceID, assetID uuid.UUID, body []byte, declaredSHA256 string) (asset.Snapshot, error) {
	if _, err := s.repo.GetAsset(ctx, workspaceID, assetID); err != nil {
		return asset.Snapshot{}, err
	}
	if _, err := asset.ParseDocument(body); err != nil {
		return asset.Snapshot{}, err
	}

	sum := sha256.Sum256(body)
	digest := hex.EncodeToString(sum[:])
	if declaredSHA256 != "" && !strings.EqualFold(declaredSHA256, digest) {
		return asset.Snapshot{}, ErrDigestMismatch
	}

	snap, err := s.repo.CreateSnapshot(ctx, workspaceID, assetID, body, digest)
	if err != nil {
		return asset.Snapshot{}, fmt.Errorf("assets: store snapshot: %w", err)
	}
	if _, err := s.audit.Record(ctx, workspaceID, &actorID, "snapshot.upload", "snapshot", snap.ID.String(),
		map[string]any{"asset_id": assetID.String(), "version": snap.Version, "sha256": digest}); err != nil {
		return asset.Snapshot{}, err
	}
	return snap, nil
}

func (s *Service) ListSnapshots(ctx context.Context, workspaceID, assetID uuid.UUID) ([]asset.Snapshot, error) {
	if _, err := s.repo.GetAsset(ctx, workspaceID, assetID); err != nil {
		return nil, err
	}
	return s.repo.ListSnapshots(ctx, workspaceID, assetID)
}

func (s *Service) GetSnapshot(ctx context.Context, workspaceID, id uuid.UUID) (asset.Snapshot, error) {
	return s.repo.GetSnapshot(ctx, workspaceID, id)
}
