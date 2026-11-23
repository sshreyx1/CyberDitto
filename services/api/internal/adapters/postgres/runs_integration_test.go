//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

func TestRunRepo_LifecycleAndIsolation(t *testing.T) {
	appPool, _ := testPools(t)
	ctx := context.Background()
	iamRepo := NewIAMRepo(appPool)
	assets := NewAssetRepo(appPool)
	twins := NewTwinRepo(appPool)
	repo := NewRunRepo(appPool)

	wsA, _ := iamRepo.CreateWorkspace(ctx, "A")
	wsB, _ := iamRepo.CreateWorkspace(ctx, "B")
	user, _ := iamRepo.CreateUser(ctx, "approver@example.com", "hash")
	a, _ := assets.CreateAsset(ctx, wsA.ID, "laptop")
	snap, _ := assets.CreateSnapshot(ctx, wsA.ID, a.ID, []byte(integrationDoc), "sha")
	tw, _ := twins.Create(ctx, wsA.ID, snap.ID, "fake")

	run, err := repo.Create(ctx, emulation.Run{
		WorkspaceID: wsA.ID, TwinID: tw.ID, ScenarioID: "discovery", TestGUIDs: []string{"g1", "g2"},
		Status:        emulation.RunRunning,
		Preflight:     []emulation.Check{{ID: "lab.isolated", Name: "x", Passed: true, Detail: "ok"}},
		CatalogSource: "atomic-red-team", CatalogCommit: "abc", Provider: "fake",
		AuthorisedBy: user.ID, AuthorisationStatement: emulation.AuthorisationStatement,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	now := time.Now()
	for _, g := range []string{"g1", "g2"} {
		if err := repo.AddExecution(ctx, wsA.ID, run.ID, emulation.Execution{
			TestGUID: g, TechniqueID: "T1087.001", TestName: g, RiskTier: emulation.TierBenign,
			Outcome: emulation.OutcomeNotExecuted, Detail: "fake", StartedAt: now, FinishedAt: now,
		}); err != nil {
			t.Fatalf("AddExecution: %v", err)
		}
	}
	if err := repo.Finish(ctx, wsA.ID, run.ID, emulation.RunCompleted, "done", now); err != nil {
		t.Fatalf("Finish: %v", err)
	}

	got, err := repo.Get(ctx, wsA.ID, run.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != emulation.RunCompleted || got.FinishedAt == nil || len(got.Executions) != 2 ||
		got.Executions[0].TestGUID != "g1" || len(got.Preflight) != 1 || got.AuthorisedBy != user.ID {
		t.Fatalf("round trip = %+v", got)
	}

	// The schema itself refuses a destructive execution row.
	err = repo.AddExecution(ctx, wsA.ID, run.ID, emulation.Execution{
		TestGUID: "bad", TechniqueID: "T1490", TestName: "bad", RiskTier: emulation.TierDestructive,
		Outcome: emulation.OutcomeMissed, StartedAt: now, FinishedAt: now,
	})
	if err == nil {
		t.Error("a destructive-tier execution row was accepted by the database")
	}

	if _, err := repo.Get(ctx, wsB.ID, run.ID); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("Get from workspace B = %v, want ErrNotFound", err)
	}
	if err := repo.Finish(ctx, wsB.ID, run.ID, emulation.RunAborted, "x", now); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("Finish from workspace B = %v, want ErrNotFound", err)
	}
	if list, _ := repo.List(ctx, wsB.ID, 10); len(list) != 0 {
		t.Errorf("List(B) = %d", len(list))
	}
}
