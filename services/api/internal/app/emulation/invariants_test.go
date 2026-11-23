package emulation

// One test per lab safety invariant in docs/MASTER_PLAN.md §6.4. Each must
// fail the build if the invariant is ever weakened. Invariant 9 (Caldera
// isolation) has no code yet — Caldera isn't integrated — and invariant 10
// (persistent "authorised testing only" notice + AUP) is covered by the
// console layout and docs/ACCEPTABLE_USE.md; both are noted in
// docs/PROGRESS.md.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/atomic"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/fake"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

type fixture struct {
	svc      *Service
	provider *fake.LabProvider
	twins    *fake.TwinRepo
	assets   *fake.AssetRepo
	catalog  ports.AttackCatalog
	ws       uuid.UUID
	actor    uuid.UUID
	twinID   uuid.UUID
}

func newFixture(t *testing.T, provider *fake.LabProvider, catalog ports.AttackCatalog) fixture {
	t.Helper()
	ctx := context.Background()
	if catalog == nil {
		catalog = atomic.Load()
	}
	assets := fake.NewAssetRepo()
	twins := fake.NewTwinRepo()
	ws := uuid.New()
	a, _ := assets.CreateAsset(ctx, ws, "host")
	snap, _ := assets.CreateSnapshot(ctx, ws, a.ID, []byte(`{}`), "sha")
	tw, _ := twins.Create(ctx, ws, snap.ID, "fake")
	_, _ = twins.UpdateState(ctx, ws, tw.ID, twin.StateReady)
	svc := NewService(catalog, fake.NewRunRepo(), twins, provider, "fake", fake.NewAuditRepo())
	t.Cleanup(svc.Wait)
	return fixture{svc: svc, provider: provider, twins: twins, assets: assets, catalog: catalog, ws: ws, actor: uuid.New(), twinID: tw.ID}
}

func (f fixture) start(t *testing.T, req Request) (emulation.Run, error) {
	t.Helper()
	if req.TwinID == uuid.Nil {
		req.TwinID = f.twinID
	}
	return f.svc.Start(context.Background(), f.actor, f.ws, true, req)
}

func waitFinished(t *testing.T, f fixture, id uuid.UUID) emulation.Run {
	t.Helper()
	f.svc.Wait()
	run, err := f.svc.Get(context.Background(), f.ws, id)
	if err != nil {
		t.Fatal(err)
	}
	return run
}

func failed(run emulation.Run) map[string]bool {
	out := map[string]bool{}
	for _, c := range run.Preflight {
		if !c.Passed {
			out[c.ID] = true
		}
	}
	return out
}

// Invariant 1: attack execution targets only TWIN records. An asset,
// snapshot, or arbitrary ID is not a twin, so nothing starts.
func TestInvariant1_OnlyTwinsCanBeTargeted(t *testing.T) {
	f := newFixture(t, &fake.LabProvider{}, nil)
	a, _ := f.assets.CreateAsset(context.Background(), f.ws, "real machine")
	for name, id := range map[string]uuid.UUID{"asset id": a.ID, "random id": uuid.New()} {
		if _, err := f.start(t, Request{TwinID: id, ScenarioID: "discovery", Acknowledged: true}); !errors.Is(err, ports.ErrNotFound) {
			t.Errorf("%s as target: err = %v, want ErrNotFound", name, err)
		}
	}
	// And a twin in another workspace is not this workspace's twin.
	if _, err := f.svc.Start(context.Background(), f.actor, uuid.New(), true, Request{TwinID: f.twinID, ScenarioID: "discovery", Acknowledged: true}); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("another workspace's twin: err = %v, want ErrNotFound", err)
	}
}

// Invariant 2: lab isolation is verified before every run.
func TestInvariant2_UnverifiedIsolationBlocksTheRun(t *testing.T) {
	f := newFixture(t, &fake.LabProvider{IsolationBroken: true}, nil)
	run, err := f.start(t, Request{ScenarioID: "discovery", Acknowledged: true})
	if !errors.Is(err, ErrPreflightFailed) || !failed(run)["lab.isolated"] {
		t.Fatalf("broken isolation: err = %v, failed checks = %v", err, failed(run))
	}
	if got := waitFinished(t, f, run.ID); len(got.Executions) != 0 || got.Status != emulation.RunPreflightFailed {
		t.Fatalf("a run with unverified isolation executed %d tests (status %s)", len(got.Executions), got.Status)
	}
}

// Invariant 3: destructive tests cannot be scheduled, regardless of what a
// catalog claims about them.
func TestInvariant3_DestructiveTestsCannotRun(t *testing.T) {
	bad := emulation.Test{GUID: "bad-1", TechniqueID: "T1490", Name: "inhibit recovery", Command: "vssadmin delete shadows /all /quiet", RiskTier: emulation.TierIntrusive, Origin: emulation.OriginCatalog}
	f := newFixture(t, &fake.LabProvider{}, stubCatalog{tests: []emulation.Test{bad}, verified: true})
	run, err := f.start(t, Request{TestGUIDs: []string{"bad-1"}, Acknowledged: true})
	if !errors.Is(err, ErrPreflightFailed) || !failed(run)["tests.not_destructive"] {
		t.Fatalf("destructive test: err = %v, failed checks = %v", err, failed(run))
	}
}

// Invariant 4: every run needs a recorded acknowledgement by an approver.
func TestInvariant4_AuthorisationRequired(t *testing.T) {
	f := newFixture(t, &fake.LabProvider{}, nil)
	run, err := f.start(t, Request{ScenarioID: "discovery", Acknowledged: false})
	if !errors.Is(err, ErrPreflightFailed) || !failed(run)["run.authorised"] {
		t.Fatalf("unacknowledged run: err = %v, failed = %v", err, failed(run))
	}
	run, err = f.svc.Start(context.Background(), f.actor, f.ws, false, Request{TwinID: f.twinID, ScenarioID: "discovery", Acknowledged: true})
	if !errors.Is(err, ErrPreflightFailed) || !failed(run)["run.authorised"] {
		t.Fatalf("non-approver run: err = %v, failed = %v", err, failed(run))
	}
	ok, err := f.start(t, Request{ScenarioID: "discovery", Acknowledged: true})
	if err != nil || ok.AuthorisationStatement != emulation.AuthorisationStatement || ok.AuthorisedBy != f.actor {
		t.Fatalf("approved run did not record who acknowledged what: %+v, %v", ok, err)
	}
}

// Invariant 5: limits, and a kill switch that stops the run and reverts.
func TestInvariant5_KillSwitchStopsAndReverts(t *testing.T) {
	p := &fake.LabProvider{ExecDelay: 200 * time.Millisecond}
	f := newFixture(t, p, nil)
	run, err := f.start(t, Request{ScenarioID: "persistence", Acknowledged: true})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	aborted, err := f.svc.Abort(context.Background(), f.actor, f.ws, run.ID)
	if err != nil {
		t.Fatalf("Abort: %v", err)
	}
	if aborted.Status != emulation.RunAborted {
		t.Errorf("Abort returned status %s; it should wait for the run to settle", aborted.Status)
	}
	got := waitFinished(t, f, run.ID)
	if got.Status != emulation.RunAborted {
		t.Fatalf("status after abort = %s, want aborted", got.Status)
	}
	if p.Reverts() < 1 {
		t.Fatal("the kill switch did not revert the twin")
	}
	for _, e := range got.Executions {
		if e.Outcome != emulation.OutcomeSkipped {
			t.Errorf("after an abort, %s has outcome %s, want skipped", e.TestGUID, e.Outcome)
		}
	}
	if len(got.Executions) != len(got.TestGUIDs) {
		t.Errorf("every requested test must be accounted for: %d executions for %d tests", len(got.Executions), len(got.TestGUIDs))
	}
}

func TestInvariant5_TestCountLimit(t *testing.T) {
	f := newFixture(t, &fake.LabProvider{}, nil)
	var all []string
	for _, tc := range f.catalog.Tests() {
		all = append(all, tc.GUID)
	}
	if len(all) <= emulation.MaxTestsPerRun {
		t.Skipf("catalog has %d tests, not enough to exceed the %d limit", len(all), emulation.MaxTestsPerRun)
	}
	run, err := f.start(t, Request{TestGUIDs: all, Acknowledged: true})
	if !errors.Is(err, ErrPreflightFailed) || !failed(run)["run.limits"] {
		t.Fatalf("over-limit run: err = %v, failed = %v", err, failed(run))
	}
}

// Invariant 6: only catalog tests run — an unknown GUID can't smuggle in
// anything ad hoc, and the provider is only ever handed catalog tests.
func TestInvariant6_OnlyCatalogTestsRun(t *testing.T) {
	f := newFixture(t, &fake.LabProvider{}, nil)
	run, err := f.start(t, Request{TestGUIDs: []string{"powershell -c whoami"}, Acknowledged: true})
	if !errors.Is(err, ErrPreflightFailed) || !failed(run)["tests.catalog_only"] {
		t.Fatalf("ad hoc test: err = %v, failed = %v", err, failed(run))
	}
}

// Invariant 7: the catalog must be the pinned, hash-verified build.
func TestInvariant7_UnverifiedCatalogBlocksEveryRun(t *testing.T) {
	good := atomic.Load()
	f := newFixture(t, &fake.LabProvider{}, stubCatalog{tests: good.Tests(), scenarios: good.Scenarios(), verified: false})
	run, err := f.start(t, Request{ScenarioID: "discovery", Acknowledged: true})
	if !errors.Is(err, ErrPreflightFailed) || !failed(run)["catalog.verified"] {
		t.Fatalf("unverified catalog: err = %v, failed = %v", err, failed(run))
	}
}

// Invariant 8: drafts (e.g. AI-generated tests) are never executable.
func TestInvariant8_DraftsNeverRun(t *testing.T) {
	draft := emulation.Test{GUID: "draft-1", TechniqueID: "T1087.001", Command: "net user", RiskTier: emulation.TierBenign, Origin: emulation.OriginDraft}
	f := newFixture(t, &fake.LabProvider{}, stubCatalog{tests: []emulation.Test{draft}, verified: true})
	run, err := f.start(t, Request{TestGUIDs: []string{"draft-1"}, Acknowledged: true})
	if !errors.Is(err, ErrPreflightFailed) || !failed(run)["tests.catalog_only"] {
		t.Fatalf("draft test: err = %v, failed = %v", err, failed(run))
	}
}

// The fake provider must never invent an outcome (docs/adr/0014).
func TestFakeProvider_RecordsNotExecutedNeverAGuess(t *testing.T) {
	f := newFixture(t, &fake.LabProvider{}, nil)
	run, err := f.start(t, Request{ScenarioID: "credential-access", Acknowledged: true})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	got := waitFinished(t, f, run.ID)
	if got.Status != emulation.RunCompleted || len(got.Executions) != len(got.TestGUIDs) {
		t.Fatalf("run = status %s, %d/%d executions", got.Status, len(got.Executions), len(got.TestGUIDs))
	}
	for _, e := range got.Executions {
		if e.Outcome != emulation.OutcomeNotExecuted {
			t.Errorf("%s: fake provider reported %s", e.TestGUID, e.Outcome)
		}
	}
	if f.provider.Reverts() != 1 {
		t.Errorf("reverts = %d, want exactly one after the run", f.provider.Reverts())
	}
}

func TestStart_UnknownScenario(t *testing.T) {
	f := newFixture(t, &fake.LabProvider{}, nil)
	if _, err := f.start(t, Request{ScenarioID: "nope", Acknowledged: true}); !errors.Is(err, ErrUnknownScenario) {
		t.Fatalf("err = %v", err)
	}
}

type stubCatalog struct {
	tests     []emulation.Test
	scenarios []emulation.Scenario
	verified  bool
}

func (c stubCatalog) Source() string     { return "stub" }
func (c stubCatalog) Repository() string { return "stub/stub" }
func (c stubCatalog) Commit() string     { return "0000000000000000000000000000000000000000" }
func (c stubCatalog) CommitDate() string { return "" }
func (c stubCatalog) License() string    { return "" }
func (c stubCatalog) Verified() (bool, string) {
	if c.verified {
		return true, "stub"
	}
	return false, "stub catalog marked unverified"
}
func (c stubCatalog) Tests() []emulation.Test { return c.tests }
func (c stubCatalog) Test(g string) (emulation.Test, bool) {
	for _, t := range c.tests {
		if t.GUID == g {
			return t, true
		}
	}
	return emulation.Test{}, false
}
func (c stubCatalog) Scenarios() []emulation.Scenario { return c.scenarios }
func (c stubCatalog) Scenario(id string) (emulation.Scenario, bool) {
	for _, s := range c.scenarios {
		if s.ID == id {
			return s, true
		}
	}
	return emulation.Scenario{}, false
}
