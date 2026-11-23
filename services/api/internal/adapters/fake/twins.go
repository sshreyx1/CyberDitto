package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
)

// TwinRepo is an in-memory ports.TwinRepository for unit tests and demo
// mode; production uses postgres.TwinRepo.
type TwinRepo struct {
	mu     sync.Mutex
	twins  map[uuid.UUID]twin.Twin
	events map[uuid.UUID][]twin.Event // keyed by twin ID, append-only
}

func NewTwinRepo() *TwinRepo {
	return &TwinRepo{
		twins:  map[uuid.UUID]twin.Twin{},
		events: map[uuid.UUID][]twin.Event{},
	}
}

func (r *TwinRepo) Create(ctx context.Context, workspaceID, snapshotID uuid.UUID, provider string) (twin.Twin, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	t := twin.Twin{
		ID: uuid.New(), WorkspaceID: workspaceID, SnapshotID: snapshotID,
		Provider: provider, State: twin.StateRequested,
		CreatedAt: now, UpdatedAt: now,
	}
	r.twins[t.ID] = t
	return t, nil
}

func (r *TwinRepo) Get(ctx context.Context, workspaceID, id uuid.UUID) (twin.Twin, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.twins[id]
	if !ok || t.WorkspaceID != workspaceID {
		return twin.Twin{}, errNotFound
	}
	return t, nil
}

func (r *TwinRepo) List(ctx context.Context, workspaceID uuid.UUID) ([]twin.Twin, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]twin.Twin, 0)
	for _, t := range r.twins {
		if t.WorkspaceID == workspaceID {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *TwinRepo) UpdateState(ctx context.Context, workspaceID, id uuid.UUID, state twin.State) (twin.Twin, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.twins[id]
	if !ok || t.WorkspaceID != workspaceID {
		return twin.Twin{}, errNotFound
	}
	t.State = state
	t.UpdatedAt = time.Now()
	r.twins[id] = t
	return t, nil
}

func (r *TwinRepo) AppendEvent(ctx context.Context, workspaceID uuid.UUID, e twin.Event) (twin.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.twins[e.TwinID]; !ok || t.WorkspaceID != workspaceID {
		return twin.Event{}, errNotFound
	}
	events := r.events[e.TwinID]
	e.Sequence = len(events) + 1
	e.CreatedAt = time.Now()
	r.events[e.TwinID] = append(events, e)
	return e, nil
}

func (r *TwinRepo) ListEvents(ctx context.Context, workspaceID, twinID uuid.UUID, afterSequence int) ([]twin.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.twins[twinID]; !ok || t.WorkspaceID != workspaceID {
		return nil, errNotFound
	}
	out := make([]twin.Event, 0)
	for _, e := range r.events[twinID] {
		if e.Sequence > afterSequence {
			out = append(out, e)
		}
	}
	return out, nil
}
