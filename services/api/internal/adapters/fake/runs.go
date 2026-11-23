package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
)

// RunRepo is an in-memory ports.RunRepository.
type RunRepo struct {
	mu   sync.Mutex
	runs map[uuid.UUID]emulation.Run
}

func NewRunRepo() *RunRepo { return &RunRepo{runs: map[uuid.UUID]emulation.Run{}} }

func (r *RunRepo) Create(ctx context.Context, run emulation.Run) (emulation.Run, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	run.ID = uuid.New()
	run.CreatedAt = time.Now()
	run.Executions = nil
	r.runs[run.ID] = run
	return run, nil
}

func (r *RunRepo) AddExecution(ctx context.Context, workspaceID, runID uuid.UUID, e emulation.Execution) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	run, ok := r.runs[runID]
	if !ok || run.WorkspaceID != workspaceID {
		return errNotFound
	}
	run.Executions = append(run.Executions, e)
	r.runs[runID] = run
	return nil
}

func (r *RunRepo) Finish(ctx context.Context, workspaceID, runID uuid.UUID, status emulation.RunStatus, detail string, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	run, ok := r.runs[runID]
	if !ok || run.WorkspaceID != workspaceID {
		return errNotFound
	}
	run.Status, run.Detail, run.FinishedAt = status, detail, &at
	r.runs[runID] = run
	return nil
}

func (r *RunRepo) Get(ctx context.Context, workspaceID, id uuid.UUID) (emulation.Run, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	run, ok := r.runs[id]
	if !ok || run.WorkspaceID != workspaceID {
		return emulation.Run{}, errNotFound
	}
	run.Executions = append([]emulation.Execution(nil), run.Executions...)
	return run, nil
}

func (r *RunRepo) List(ctx context.Context, workspaceID uuid.UUID, limit int) ([]emulation.Run, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]emulation.Run, 0)
	for _, run := range r.runs {
		if run.WorkspaceID == workspaceID {
			run.Executions = nil
			out = append(out, run)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
