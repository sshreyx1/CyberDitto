package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
)

// AssessmentRepo is an in-memory ports.AssessmentRepository.
type AssessmentRepo struct {
	mu   sync.Mutex
	rows map[uuid.UUID]compliance.Assessment
}

func NewAssessmentRepo() *AssessmentRepo {
	return &AssessmentRepo{rows: map[uuid.UUID]compliance.Assessment{}}
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
