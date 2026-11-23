package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
)

// AssessmentRepo is an in-memory ports.AssessmentRepository. It reads the
// fake AssetRepo to resolve a snapshot's asset, standing in for the SQL
// join the Postgres adapter does.
type AssessmentRepo struct {
	mu     sync.Mutex
	rows   map[uuid.UUID]compliance.Assessment
	assets *AssetRepo
}

func NewAssessmentRepo(assets *AssetRepo) *AssessmentRepo {
	return &AssessmentRepo{rows: map[uuid.UUID]compliance.Assessment{}, assets: assets}
}

func (r *AssessmentRepo) ListForAsset(ctx context.Context, workspaceID, assetID uuid.UUID, limit int) ([]compliance.Assessment, error) {
	all, err := r.List(ctx, workspaceID, len(r.rows)+1)
	if err != nil {
		return nil, err
	}
	out := make([]compliance.Assessment, 0)
	for _, a := range all {
		if s, err := r.assets.GetSnapshot(ctx, workspaceID, a.SnapshotID); err == nil && s.AssetID == assetID {
			out = append(out, a)
		}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *AssessmentRepo) Create(ctx context.Context, a compliance.Assessment, createdBy uuid.UUID) (compliance.Assessment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a.ID = uuid.New()
	a.CreatedAt = time.Now()
	r.rows[a.ID] = a
	return a, nil
}

func (r *AssessmentRepo) List(ctx context.Context, workspaceID uuid.UUID, limit int) ([]compliance.Assessment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]compliance.Assessment, 0)
	for _, a := range r.rows {
		if a.WorkspaceID == workspaceID {
			a.Findings = []compliance.Finding{}
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (r *AssessmentRepo) Get(ctx context.Context, workspaceID, id uuid.UUID) (compliance.Assessment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.rows[id]
	if !ok || a.WorkspaceID != workspaceID {
		return compliance.Assessment{}, errNotFound
	}
	return a, nil
}
