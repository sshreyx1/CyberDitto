//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

func TestAssessmentRepo_RoundTripAndIsolation(t *testing.T) {
	appPool, _ := testPools(t)
	ctx := context.Background()
	iamRepo := NewIAMRepo(appPool)
	assets := NewAssetRepo(appPool)
	repo := NewAssessmentRepo(appPool)

	wsA, _ := iamRepo.CreateWorkspace(ctx, "A")
	wsB, _ := iamRepo.CreateWorkspace(ctx, "B")
	user, _ := iamRepo.CreateUser(ctx, "assessor@example.com", "hash")

	data, err := os.ReadFile("../../domain/compliance/testdata/weak.json")
	if err != nil {
		t.Fatal(err)
	}
	a, _ := assets.CreateAsset(ctx, wsA.ID, "laptop")
	snap, err := assets.CreateSnapshot(ctx, wsA.ID, a.ID, data, "sha")
	if err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}
	doc, _ := asset.ParseDocument(data)
	findings := compliance.Evaluate(compliance.Catalog(), compliance.NewFacts(doc, time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)))
	// Mix in a measurable pass so the stored score isn't trivially 0.
	findings[0].Status = compliance.StatusPass
	score, counts := compliance.Score(findings)

	created, err := repo.Create(ctx, compliance.Assessment{
		WorkspaceID: wsA.ID, SnapshotID: snap.ID, Engine: compliance.EngineName,
		EngineVersion: compliance.EngineVersion, Score: score, Counts: counts, Findings: findings,
	}, user.ID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.Get(ctx, wsA.ID, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Score == nil || *got.Score != *score {
		t.Errorf("stored score = %v, want %v", got.Score, *score)
	}
	if got.Counts != counts {
		t.Errorf("stored counts = %+v, want %+v", got.Counts, counts)
	}
	if len(got.Findings) != len(findings) {
		t.Fatalf("stored findings = %d, want %d", len(got.Findings), len(findings))
	}
	for _, f := range got.Findings {
		if f.Status == compliance.StatusFail && (f.Observed == nil || f.Reason == "") {
			t.Errorf("%s: failed finding lost its evidence", f.ControlID)
		}
	}

	if _, err := repo.Get(ctx, wsB.ID, created.ID); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("Get from workspace B = %v, want ErrNotFound", err)
	}
	if list, _ := repo.List(ctx, wsB.ID, 10); len(list) != 0 {
		t.Errorf("List(B) = %d, want 0", len(list))
	}
	if list, _ := repo.List(ctx, wsA.ID, 10); len(list) != 1 {
		t.Errorf("List(A) = %d, want 1", len(list))
	}
}
