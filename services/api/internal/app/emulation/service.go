// Package emulation runs catalog attack tests against twins, gated by the
// preflight policy engine (docs/MASTER_PLAN.md §5.12, §6.4).
package emulation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

var (
	// ErrPreflightFailed: the run was recorded but did not start; the
	// returned run carries every check and its result.
	ErrPreflightFailed = errors.New("emulation: preflight failed")
	ErrUnknownScenario = errors.New("emulation: unknown scenario")
	ErrNotRunning      = errors.New("emulation: run is not running")
)

type Request struct {
	TwinID       uuid.UUID
	ScenarioID   string
	TestGUIDs    []string
	Acknowledged bool
}

type activeRun struct {
	cancel  context.CancelFunc
	aborted bool
	done    chan struct{} // closed once the run has finished and been recorded
}

// abortSettleTimeout is how long Abort waits for a cancelled run to finish
// reverting before responding; a slow real revert just means the response
// still says "running" and the final status lands shortly after.
const abortSettleTimeout = 10 * time.Second

type Service struct {
	catalog      ports.AttackCatalog
	runs         ports.RunRepository
	twins        ports.TwinRepository
	provider     ports.LabProvider
	providerName string
	audit        ports.AuditRepository
	now          func() time.Time

	mu     sync.Mutex
	active map[uuid.UUID]*activeRun
	wg     sync.WaitGroup
}

func NewService(catalog ports.AttackCatalog, runs ports.RunRepository, twins ports.TwinRepository, provider ports.LabProvider, providerName string, audit ports.AuditRepository) *Service {
	return &Service{
		catalog: catalog, runs: runs, twins: twins, provider: provider, providerName: providerName,
		audit: audit, now: time.Now, active: map[uuid.UUID]*activeRun{},
	}
}

func (s *Service) Catalog() ports.AttackCatalog { return s.catalog }

// Start resolves the request, runs preflight, records the run either way,
// and — only if every check passed — starts executing in the background.
// approver is whether the caller holds the approver capability; RBAC
// already blocks non-approvers at the HTTP layer, and preflight checks it
// again so the policy engine never relies on a caller having done so.
func (s *Service) Start(ctx context.Context, actorID, workspaceID uuid.UUID, approver bool, req Request) (emulation.Run, error) {
	// §6.4 invariant 1: the target is resolved as a twin in this
	// workspace. There is no other kind of target to resolve.
	tw, err := s.twins.Get(ctx, workspaceID, req.TwinID)
	if err != nil {
		return emulation.Run{}, err
	}

	guids := append([]string(nil), req.TestGUIDs...)
	if req.ScenarioID != "" {
		sc, ok := s.catalog.Scenario(req.ScenarioID)
		if !ok {
			return emulation.Run{}, ErrUnknownScenario
		}
		// Copy first: appending to sc.TestGUIDs directly could write into
		// the shared catalog's backing array.
		guids = append(append([]string(nil), sc.TestGUIDs...), guids...)
	}
	guids = dedupe(guids)

	var tests []emulation.Test
	var unknown []string
	for _, g := range guids {
		if t, ok := s.catalog.Test(g); ok {
			tests = append(tests, t)
		} else {
			unknown = append(unknown, g)
		}
	}

	iso, err := s.provider.VerifyIsolation(ctx, tw.ID)
	if err != nil {
		iso = emulation.IsolationReport{Verified: false, Detail: "isolation check failed: " + err.Error()}
	}
	catalogOK, catalogDetail := s.catalog.Verified()

	checks, ok := emulation.Preflight(emulation.PreflightInput{
		TwinState: tw.State, Isolation: iso,
		CatalogVerified: catalogOK, CatalogDetail: catalogDetail,
		Tests: tests, UnknownTests: unknown,
		Approver: approver, Acknowledged: req.Acknowledged,
	})

	run := emulation.Run{
		WorkspaceID: workspaceID, TwinID: tw.ID, ScenarioID: req.ScenarioID, TestGUIDs: guids,
		Status: emulation.RunRunning, Preflight: checks,
		CatalogSource: s.catalog.Source(), CatalogCommit: s.catalog.Commit(),
		Provider: s.providerName, AuthorisedBy: actorID,
	}
	if req.Acknowledged {
		run.AuthorisationStatement = emulation.AuthorisationStatement
	}
	if !ok {
		run.Status = emulation.RunPreflightFailed
		run.Detail = "preflight failed; nothing was executed"
	}

	stored, err := s.runs.Create(ctx, run)
	if err != nil {
		return emulation.Run{}, fmt.Errorf("emulation: record run: %w", err)
	}
	if !ok {
		now := s.now()
		stored.FinishedAt = &now
		_ = s.runs.Finish(ctx, workspaceID, stored.ID, emulation.RunPreflightFailed, run.Detail, now)
	}

	if _, err := s.audit.Record(ctx, workspaceID, &actorID, "emulation.run.create", "emulation_run", stored.ID.String(), map[string]any{
		"twin_id": tw.ID.String(), "scenario_id": req.ScenarioID, "tests": len(guids),
		"status": string(stored.Status), "acknowledged": req.Acknowledged, "catalog_commit": s.catalog.Commit(),
	}); err != nil {
		return emulation.Run{}, err
	}

	if !ok {
		return stored, ErrPreflightFailed
	}

	runCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), emulation.MaxRunDuration)
	ar := &activeRun{cancel: cancel, done: make(chan struct{})}
	s.mu.Lock()
	s.active[stored.ID] = ar
	s.mu.Unlock()

	s.wg.Add(1)
	go s.execute(runCtx, ar, actorID, stored, tests)
	return stored, nil
}

func (s *Service) execute(ctx context.Context, ar *activeRun, actorID uuid.UUID, run emulation.Run, tests []emulation.Test) {
	defer s.wg.Done()
	defer close(ar.done)
	defer ar.cancel()
	bg := context.WithoutCancel(ctx)

	for i, t := range tests {
		if ctx.Err() != nil {
			for _, rest := range tests[i:] {
				now := s.now()
				_ = s.runs.AddExecution(bg, run.WorkspaceID, run.ID, emulation.Execution{
					TestGUID: rest.GUID, TechniqueID: rest.TechniqueID, TestName: rest.Name, RiskTier: rest.RiskTier,
					Outcome: emulation.OutcomeSkipped, Detail: "run stopped before this test started",
					StartedAt: now, FinishedAt: now,
				})
			}
			break
		}
		tctx, tcancel := context.WithTimeout(ctx, emulation.MaxTestDuration)
		e := s.provider.Execute(tctx, run.TwinID, t)
		tcancel()
		_ = s.runs.AddExecution(bg, run.WorkspaceID, run.ID, e)
	}

	// Every run ends with a revert to the clean snapshot, whatever happened.
	revertCtx, rcancel := context.WithTimeout(bg, 2*time.Minute)
	revertErr := s.provider.Revert(revertCtx, run.TwinID)
	rcancel()

	s.mu.Lock()
	aborted := ar.aborted
	delete(s.active, run.ID)
	s.mu.Unlock()

	status, detail := emulation.RunCompleted, "all tests attempted; twin reverted"
	switch {
	case aborted:
		status, detail = emulation.RunAborted, "aborted by the kill switch; twin reverted"
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		status, detail = emulation.RunFailed, fmt.Sprintf("exceeded the %s run limit; twin reverted", emulation.MaxRunDuration)
	}
	if revertErr != nil {
		status, detail = emulation.RunFailed, "revert after the run failed: "+revertErr.Error()
	}
	_ = s.runs.Finish(bg, run.WorkspaceID, run.ID, status, detail, s.now())
	_, _ = s.audit.Record(bg, run.WorkspaceID, &actorID, "emulation.run.finish", "emulation_run", run.ID.String(), map[string]any{"status": string(status)})
}

// Abort is the kill switch (§6.4 invariant 5): stop the run now and revert
// the twin. A run left "running" by a process that no longer exists (e.g.
// after a restart) is marked aborted and its twin reverted directly.
func (s *Service) Abort(ctx context.Context, actorID, workspaceID, runID uuid.UUID) (emulation.Run, error) {
	run, err := s.runs.Get(ctx, workspaceID, runID)
	if err != nil {
		return emulation.Run{}, err
	}
	if run.Status != emulation.RunRunning {
		return run, ErrNotRunning
	}
	if _, err := s.audit.Record(ctx, workspaceID, &actorID, "emulation.run.abort", "emulation_run", runID.String(), nil); err != nil {
		return emulation.Run{}, err
	}

	s.mu.Lock()
	ar, live := s.active[runID]
	if live {
		ar.aborted = true
		ar.cancel()
	}
	s.mu.Unlock()

	if live {
		select {
		case <-ar.done:
		case <-time.After(abortSettleTimeout):
		case <-ctx.Done():
		}
	} else {
		revertErr := s.provider.Revert(ctx, run.TwinID)
		detail := "aborted (run was orphaned by a restart); twin reverted"
		status := emulation.RunAborted
		if revertErr != nil {
			status, detail = emulation.RunFailed, "aborted, but the revert failed: "+revertErr.Error()
		}
		if err := s.runs.Finish(ctx, workspaceID, runID, status, detail, s.now()); err != nil {
			return emulation.Run{}, err
		}
	}
	return s.runs.Get(ctx, workspaceID, runID)
}

// Wait blocks until every background run has finished (tests, shutdown).
func (s *Service) Wait() { s.wg.Wait() }

func (s *Service) Get(ctx context.Context, workspaceID, id uuid.UUID) (emulation.Run, error) {
	return s.runs.Get(ctx, workspaceID, id)
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID) ([]emulation.Run, error) {
	return s.runs.List(ctx, workspaceID, 100)
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
