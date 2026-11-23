// Package twin holds the pure twin lifecycle state machine: the valid state
// graph and the entities that move through it. No I/O
// (docs/MASTER_PLAN.md §5.1, §5.7).
package twin

import (
	"time"

	"github.com/google/uuid"
)

// State mirrors the CHECK constraint on the `twins` table
// (services/api/internal/adapters/postgres/migrations/0001_init.sql) so the
// domain and schema can never silently drift apart.
type State string

const (
	StateRequested    State = "requested"
	StateImaging      State = "imaging"
	StateProvisioning State = "provisioning"
	StateApplying     State = "applying"
	StateVerifying    State = "verifying"
	StateReady        State = "ready"
	StateRunning      State = "running"
	StateStopped      State = "stopped"
	StateFailed       State = "failed"
	StateDestroyed    State = "destroyed"
)

// terminal states have no further transitions out.
var terminal = map[State]bool{
	StateDestroyed: true,
}

// transitions is the twin lifecycle's allowed state graph
// (docs/MASTER_PLAN.md §5.7). Build walks Requested→Ready one step at a
// time; Ready and Stopped are where a twin sits between runs; Failed can be
// reached from any non-terminal state and can only be cleaned up by
// destroying it — there is deliberately no "retry build" edge back out of
// Failed, so a twin that failed mid-build can't silently resume from a
// half-applied state.
var transitions = map[State][]State{
	StateRequested:    {StateImaging, StateFailed},
	StateImaging:      {StateProvisioning, StateFailed},
	StateProvisioning: {StateApplying, StateFailed},
	StateApplying:     {StateVerifying, StateFailed},
	StateVerifying:    {StateReady, StateFailed},
	StateReady:        {StateRunning, StateDestroyed, StateFailed},
	StateRunning:      {StateStopped, StateDestroyed, StateFailed},
	StateStopped:      {StateRunning, StateDestroyed, StateFailed},
	StateFailed:       {StateDestroyed},
	StateDestroyed:    {},
}

// Valid reports whether s is one of the known states.
func (s State) Valid() bool {
	_, ok := transitions[s]
	return ok
}

// Terminal reports whether a twin in state s can ever leave it.
func (s State) Terminal() bool {
	return terminal[s]
}

// CanTransition is the single policy function for twin state changes
// (docs/MASTER_PLAN.md §6.2's "single policy function" pattern, applied here
// to the lifecycle rather than RBAC). Every state write — build steps,
// start/stop/destroy, failure handling — must go through this rather than
// assigning State directly.
func CanTransition(from, to State) bool {
	for _, s := range transitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

// Twin is one digital twin VM. SnapshotID is a foreign key into the
// snapshots table; Phase 1's fake provider doesn't require a real snapshot
// to exist yet (Assets/Snapshots upload is still unbuilt — see
// docs/PROGRESS.md), so callers may pass any UUID until that lands.
type Twin struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	SnapshotID    uuid.UUID
	Provider      string
	State         State
	FidelityScore *float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Event is one lifecycle step, persisted and streamed to the browser over
// SSE (docs/MASTER_PLAN.md §5.1: "event-sourced where it matters").
// Sequence is per-twin and strictly increasing, so a reconnecting client can
// ask for everything "after sequence N" instead of replaying from scratch.
type Event struct {
	TwinID    uuid.UUID
	Sequence  int
	State     State
	Message   string
	CreatedAt time.Time
}
