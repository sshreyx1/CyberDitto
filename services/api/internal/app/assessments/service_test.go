package assessments

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/fake"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
)

type fixture struct {
	svc        *Service
	twins      *fake.TwinRepo
	ws, actor  uuid.UUID
	snapshotID uuid.UUID
}

func newFixture(t *testing.T, docPath string) fixture {
	t.Helper()
	data, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatal(err)
	}
	assets := fake.NewAssetRepo()
	twins := fake.NewTwinRepo()
	ws := uuid.New()
	a, _ := assets.CreateAsset(context.Background(), ws, "host")
	snap, _ := assets.CreateSnapshot(context.Background(), ws, a.ID, data, "sha")
	svc := NewService(fake.NewAssessmentRepo(assets), assets, twins, fake.NewAuditRepo())
	svc.now = func() time.Time { return time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC) }
	return fixture{svc: svc, twins: twins, ws: ws, actor: uuid.New(), snapshotID: snap.ID}
}

const weakDoc = "../../domain/compliance/testdata/weak.json"

func TestRun_SnapshotTargetStoresScoredFindings(t *testing.T) {
	f := newFixture(t, weakDoc)
	got, err := f.svc.Run(context.Background(), f.actor, f.ws, Target{SnapshotID: &f.snapshotID})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got.Score == nil || *got.Score != 0 || got.Counts.Failed != len(compliance.Catalog()) {
		t.Fatalf("assessment = score %v counts %+v", got.Score, got.Counts)
	}
	if got.Engine != compliance.EngineName || got.EngineVersion != compliance.EngineVersion {
		t.Errorf("engine not recorded: %s %s", got.Engine, got.EngineVersion)
	}
	stored, err := f.svc.Get(context.Background(), f.ws, got.ID)
	if err != nil || len(stored.Findings) != len(compliance.Catalog()) {
		t.Fatalf("Get = %d findings, %v", len(stored.Findings), err)
	}
	if _, err := f.svc.Get(context.Background(), uuid.New(), got.ID); err == nil {
		t.Error("another workspace must not see the assessment")
	}
}

func TestRun_TwinTargetRequiresBuiltTwin(t *testing.T) {
	f := newFixture(t, weakDoc)
	ctx := context.Background()
	tw, _ := f.twins.Create(ctx, f.ws, f.snapshotID, "fake")

	if _, err := f.svc.Run(ctx, f.actor, f.ws, Target{TwinID: &tw.ID}); !errors.Is(err, ErrTwinNotReady) {
		t.Fatalf("assessing a requested twin = %v, want ErrTwinNotReady", err)
	}

	_, _ = f.twins.UpdateState(ctx, f.ws, tw.ID, twin.StateReady)
	got, err := f.svc.Run(ctx, f.actor, f.ws, Target{TwinID: &tw.ID})
	if err != nil {
		t.Fatalf("Run on ready twin: %v", err)
	}
	if got.TwinID == nil || *got.TwinID != tw.ID || got.SnapshotID != f.snapshotID {
		t.Fatalf("twin assessment target = twin %v snapshot %s", got.TwinID, got.SnapshotID)
	}
}

func TestRun_TargetValidation(t *testing.T) {
	f := newFixture(t, weakDoc)
	ctx := context.Background()
	if _, err := f.svc.Run(ctx, f.actor, f.ws, Target{}); !errors.Is(err, ErrInvalidTarget) {
		t.Errorf("no target = %v", err)
	}
	id := uuid.New()
	if _, err := f.svc.Run(ctx, f.actor, f.ws, Target{SnapshotID: &f.snapshotID, TwinID: &id}); !errors.Is(err, ErrInvalidTarget) {
		t.Errorf("both targets = %v", err)
	}
	if _, err := f.svc.Run(ctx, f.actor, uuid.New(), Target{SnapshotID: &f.snapshotID}); err == nil {
		t.Error("assessing another workspace's snapshot must fail")
	}
}
