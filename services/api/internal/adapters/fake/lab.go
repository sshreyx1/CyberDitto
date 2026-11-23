package fake

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

// buildSteps is the fake provider's build order; it must stay in lockstep
// with the state graph in internal/domain/twin (CanTransition rejects the
// step otherwise — the provider doesn't get to invent its own ordering).
var buildSteps = []twin.State{
	twin.StateImaging,
	twin.StateProvisioning,
	twin.StateApplying,
	twin.StateVerifying,
	twin.StateReady,
}

const defaultStepDelay = 300 * time.Millisecond

// NotExecutedDetail is what every fake execution reports.
const NotExecutedDetail = "not executed: the fake lab provider has no virtual machine to run tests in. Real outcomes (blocked/detected/logged/missed) need the Hyper-V provider."

// LabProvider is a ports.LabProvider with no hypervisor behind it, for CI,
// unit tests and demo mode (docs/MASTER_PLAN.md §5.1). It never executes
// anything and never reports a guessed attack outcome (docs/adr/0014):
// Execute always returns OutcomeNotExecuted.
type LabProvider struct {
	// StepDelay overrides defaultStepDelay for builds.
	StepDelay time.Duration
	// ExecDelay simulates time spent per test, so tests can abort a run
	// mid-flight. Zero means instant.
	ExecDelay time.Duration
	// IsolationBroken makes VerifyIsolation fail, for testing that
	// preflight blocks a run when isolation can't be confirmed.
	IsolationBroken bool

	reverts atomic.Int64
}

func (p *LabProvider) delay() time.Duration {
	if p.StepDelay > 0 {
		return p.StepDelay
	}
	return defaultStepDelay
}

func (p *LabProvider) Build(ctx context.Context, twinID uuid.UUID, step ports.StepFunc) {
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

func (p *LabProvider) Start(ctx context.Context, twinID uuid.UUID) error   { return nil }
func (p *LabProvider) Stop(ctx context.Context, twinID uuid.UUID) error    { return nil }
func (p *LabProvider) Destroy(ctx context.Context, twinID uuid.UUID) error { return nil }

func (p *LabProvider) VerifyIsolation(ctx context.Context, twinID uuid.UUID) (emulation.IsolationReport, error) {
	if p.IsolationBroken {
		return emulation.IsolationReport{Verified: false, Detail: "fake provider: isolation deliberately reported broken (test hook)"}, nil
	}
	return emulation.IsolationReport{
		Verified: true,
		Detail:   "fake provider: no virtual machine or network exists, so nothing can leave the lab",
	}, nil
}

func (p *LabProvider) Execute(ctx context.Context, twinID uuid.UUID, test emulation.Test) emulation.Execution {
	e := emulation.Execution{
		TestGUID: test.GUID, TechniqueID: test.TechniqueID, TestName: test.Name,
		RiskTier: test.RiskTier, StartedAt: time.Now(),
	}
	if p.ExecDelay > 0 {
		select {
		case <-ctx.Done():
			e.Outcome, e.Detail, e.FinishedAt = emulation.OutcomeSkipped, "run stopped before this test finished", time.Now()
			return e
		case <-time.After(p.ExecDelay):
		}
	}
	e.Outcome, e.Detail, e.FinishedAt = emulation.OutcomeNotExecuted, NotExecutedDetail, time.Now()
	return e
}

func (p *LabProvider) Revert(ctx context.Context, twinID uuid.UUID) error {
	p.reverts.Add(1)
	return nil
}

// Reverts reports how many times Revert was called (kill-switch tests).
func (p *LabProvider) Reverts() int64 { return p.reverts.Load() }
