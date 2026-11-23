package twins

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/fake"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
)

type fixture struct {
	svc         *Service
	actorID     uuid.UUID
	workspaceID uuid.UUID
	snapshotID  uuid.UUID
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	assets := fake.NewAssetRepo()
	workspaceID := uuid.New()
	a, err := assets.CreateAsset(context.Background(), workspaceID, "host")
	if err != nil {
		t.Fatal(err)
	}
	snap, err := assets.CreateSnapshot(context.Background(), workspaceID, a.ID, []byte(`{}`), "sha")
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(fake.NewTwinRepo(), assets, fake.LabProvider{StepDelay: time.Millisecond}, "fake", fake.NewAuditRepo())
	return fixture{svc: svc, actorID: uuid.New(), workspaceID: workspaceID, snapshotID: snap.ID}
}

func waitForState(t *testing.T, svc *Service, workspaceID, twinID uuid.UUID, want twin.State) twin.Twin {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, err := svc.Get(context.Background(), workspaceID, twinID)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.State == want {
			return got
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("twin never reached state %s", want)
	return twin.Twin{}
}

func TestCreate_BuildsToReadyAndEmitsOrderedEvents(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	created, err := f.svc.Create(ctx, f.actorID, f.workspaceID, f.snapshotID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.State != twin.StateRequested || created.Provider != "fake" {
		t.Fatalf("created = %+v", created)
	}

	waitForState(t, f.svc, f.workspaceID, created.ID, twin.StateReady)

	events, err := f.svc.Events(ctx, f.workspaceID, created.ID, 0)
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	want := []twin.State{twin.StateImaging, twin.StateProvisioning, twin.StateApplying, twin.StateVerifying, twin.StateReady}
	if len(events) != len(want) {
		t.Fatalf("got %d events, want %d: %+v", len(events), len(want), events)
	}
	for i, ev := range events {
		if ev.State != want[i] || ev.Sequence != i+1 {
			t.Errorf("event %d = %+v, want state %s sequence %d", i, ev, want[i], i+1)
		}
	}
}

func TestCreate_RejectsUnknownSnapshot(t *testing.T) {
	f := newFixture(t)
	_, err := f.svc.Create(context.Background(), f.actorID, f.workspaceID, uuid.New())
	if !errors.Is(err, ErrSnapshotNotFound) {
		t.Fatalf("Create with unknown snapshot error = %v, want ErrSnapshotNotFound", err)
	}
}

func TestCreate_RejectsOtherWorkspacesSnapshot(t *testing.T) {
	f := newFixture(t)
	_, err := f.svc.Create(context.Background(), f.actorID, uuid.New(), f.snapshotID)
	if !errors.Is(err, ErrSnapshotNotFound) {
		t.Fatalf("Create with another workspace's snapshot error = %v, want ErrSnapshotNotFound", err)
	}
}

func TestSubscribe_ReceivesLiveEvents(t *testing.T) {
	f := newFixture(t)
	created, err := f.svc.Create(context.Background(), f.actorID, f.workspaceID, f.snapshotID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ch, cancel := f.svc.Subscribe(created.ID)
	defer cancel()

	deadline := time.After(2 * time.Second)
	for {
		select {
		case ev := <-ch:
			if ev.State == twin.StateReady {
				return
			}
		case <-deadline:
			t.Fatal("timed out waiting for a live Ready event")
		}
	}
}

func TestStartStopDestroy_EnforceStateMachine(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	created, err := f.svc.Create(ctx, f.actorID, f.workspaceID, f.snapshotID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := f.svc.Start(ctx, f.actorID, f.workspaceID, created.ID); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("Start before Ready error = %v, want ErrInvalidTransition", err)
	}

	waitForState(t, f.svc, f.workspaceID, created.ID, twin.StateReady)

	for _, step := range []struct {
		op   func(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (twin.Twin, error)
		want twin.State
	}{
		{f.svc.Start, twin.StateRunning},
		{f.svc.Stop, twin.StateStopped},
		{f.svc.Destroy, twin.StateDestroyed},
	} {
		got, err := step.op(ctx, f.actorID, f.workspaceID, created.ID)
		if err != nil {
			t.Fatalf("transition to %s: %v", step.want, err)
		}
		if got.State != step.want {
			t.Fatalf("state = %s, want %s", got.State, step.want)
		}
	}

	if _, err := f.svc.Start(ctx, f.actorID, f.workspaceID, created.ID); !errors.Is(err, ErrInvalidTransition) {
		t.Error("Start on a Destroyed twin should fail with ErrInvalidTransition")
	}
}

func TestGet_IsWorkspaceScoped(t *testing.T) {
	f := newFixture(t)
	created, err := f.svc.Create(context.Background(), f.actorID, f.workspaceID, f.snapshotID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := f.svc.Get(context.Background(), uuid.New(), created.ID); err == nil {
		t.Error("Get from a different workspace should fail")
	}
}
