package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
)

// StepFunc is supplied by the app layer to a LabProvider's Build call so the
// provider can drive a twin through its build steps without owning
// persistence, the state-machine rules, or event delivery itself
// (docs/MASTER_PLAN.md §5.1: adapters are swappable; business rules live in
// domain/app, never in an adapter). Returning a non-nil error stops the
// build early.
type StepFunc func(ctx context.Context, state twin.State, message string) error

// LabProvider builds, starts, stops, and destroys twin VMs. The fake
// adapter (internal/adapters/fake) simulates this with timers for CI, unit
// tests, and demo mode; real adapters (Hyper-V, libvirt, Proxmox) are
// Phase 2 (docs/MASTER_PLAN.md §5.4, §5.9, §10).
type LabProvider interface {
	// Build drives twinID from Requested toward Ready, calling step once
	// per intermediate state in order. It returns once the build reaches a
	// terminal outcome for this attempt (step returns an error, e.g.
	// because the twin was deleted mid-build) or ctx is cancelled; it never
	// itself decides that the twin has failed — the caller does, based on
	// the state it was left in.
	Build(ctx context.Context, twinID uuid.UUID, step StepFunc)
	Start(ctx context.Context, twinID uuid.UUID) error
	Stop(ctx context.Context, twinID uuid.UUID) error
	Destroy(ctx context.Context, twinID uuid.UUID) error
}

// TwinRepository persists twin records and their append-only event log.
// Workspace-scoped methods take the workspace explicitly so an adapter can
// enforce it (Postgres RLS, in a future real adapter) rather than trusting
// the caller (docs/MASTER_PLAN.md §6.2).
type TwinRepository interface {
	Create(ctx context.Context, workspaceID, snapshotID uuid.UUID, provider string) (twin.Twin, error)
	Get(ctx context.Context, workspaceID, id uuid.UUID) (twin.Twin, error)
	List(ctx context.Context, workspaceID uuid.UUID) ([]twin.Twin, error)
	UpdateState(ctx context.Context, id uuid.UUID, state twin.State) (twin.Twin, error)
	AppendEvent(ctx context.Context, e twin.Event) (twin.Event, error)
	ListEvents(ctx context.Context, twinID uuid.UUID, afterSequence int) ([]twin.Event, error)
}
