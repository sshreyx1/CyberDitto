package fake

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

// buildSteps is the fake provider's build order; it must stay in lockstep
// with the state graph in internal/domain/twin (CanTransition rejects the
// step otherwise — that's the point: the provider doesn't get to invent its
// own ordering, the domain still polices it).
var buildSteps = []twin.State{
	twin.StateImaging,
	twin.StateProvisioning,
	twin.StateApplying,
	twin.StateVerifying,
	twin.StateReady,
}

// defaultStepDelay keeps a fake build fast in CI/unit tests while still
// being slow enough for a demo/console UI to show real, visible progress
// rather than everything completing in the same tick.
const defaultStepDelay = 300 * time.Millisecond

// LabProvider is a ports.LabProvider that simulates a hypervisor
// (docs/MASTER_PLAN.md §5.1, §10 Phase 1: "fake LabProvider: twin lifecycle
// state machine with progress events over SSE"). It does no real I/O —
// Start/Stop/Destroy are no-ops; the app layer owns the actual state
// transition and persistence for those (see internal/app/twins.Service) —
// and Build only times out a fixed step sequence.
type LabProvider struct {
	// StepDelay overrides defaultStepDelay; tests set this to ~0 so the
	// state machine can be exercised without waiting on wall-clock time.
	StepDelay time.Duration
}

func (p LabProvider) delay() time.Duration {
	if p.StepDelay > 0 {
		return p.StepDelay
	}
	return defaultStepDelay
}

func (p LabProvider) Build(ctx context.Context, twinID uuid.UUID, step ports.StepFunc) {
	for _, state := range buildSteps {
		select {
		case <-ctx.Done():
			return
		case <-time.After(p.delay()):
		}
		msg := fmt.Sprintf("fake provider: twin reached %s", state)
		if err := step(ctx, state, msg); err != nil {
			return
		}
	}
}

func (p LabProvider) Start(ctx context.Context, twinID uuid.UUID) error   { return nil }
func (p LabProvider) Stop(ctx context.Context, twinID uuid.UUID) error    { return nil }
func (p LabProvider) Destroy(ctx context.Context, twinID uuid.UUID) error { return nil }
