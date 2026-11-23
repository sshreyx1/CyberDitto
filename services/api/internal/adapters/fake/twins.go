package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
)

// TwinRepo is an in-memory ports.TwinRepository (docs/MASTER_PLAN.md §5.1:
// a Fake adapter for every port). Phase 1 deliberately has no
// Postgres-backed twin repository yet: the roadmap's Phase 1 item is "fake
// LabProvider: twin lifecycle state machine with progress events over SSE",
// not real persistence — that lands with the Phase 2 real lab provider,
// alongside the `twins` table this mirrors
// (internal/adapters/postgres/migrations/0001_init.sql).
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
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *TwinRepo) UpdateState(ctx context.Context, id uuid.UUID, state twin.State) (twin.Twin, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.twins[id]
	if !ok {
		return twin.Twin{}, errNotFound
	}
	t.State = state
	t.UpdatedAt = time.Now()
	r.twins[id] = t
	return t, nil
}

func (r *TwinRepo) AppendEvent(ctx context.Context, e twin.Event) (twin.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	events := r.events[e.TwinID]
	e.Sequence = len(events) + 1
	e.CreatedAt = time.Now()
	r.events[e.TwinID] = append(events, e)
	return e, nil
}

func (r *TwinRepo) ListEvents(ctx context.Context, twinID uuid.UUID, afterSequence int) ([]twin.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]twin.Event, 0)
	for _, e := range r.events[twinID] {
		if e.Sequence > afterSequence {
			out = append(out, e)
		}
	}
	return out, nil
}
