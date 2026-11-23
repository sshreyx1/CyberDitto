package twins

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/fake"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
)

func testService() *Service {
	provider := fake.LabProvider{StepDelay: time.Millisecond}
	return NewService(fake.NewTwinRepo(), provider, fake.NewAuditRepo())
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
	svc := testService()
	ctx := context.Background()
	actorID, workspaceID, snapshotID := uuid.New(), uuid.New(), uuid.New()

	created, err := svc.Create(ctx, actorID, workspaceID, snapshotID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.State != twin.StateRequested {
		t.Fatalf("initial state = %s, want %s", created.State, twin.StateRequested)
	}

	ready := waitForState(t, svc, workspaceID, created.ID, twin.StateReady)
	if ready.WorkspaceID != workspaceID {
		t.Fatalf("WorkspaceID = %s, want %s", ready.WorkspaceID, workspaceID)
	}

	events, err := svc.Events(ctx, created.ID, 0)
	if err != nil {
		t.Fatalf("Events: %v", err)
	}
	wantStates := []twin.State{
		twin.StateImaging, twin.StateProvisioning, twin.StateApplying,
		twin.StateVerifying, twin.StateReady,
	}
	if len(events) != len(wantStates) {
		t.Fatalf("got %d events, want %d: %+v", len(events), len(wantStates), events)
	}
	for i, ev := range events {
		if ev.State != wantStates[i] {
			t.Errorf("event %d state = %s, want %s", i, ev.State, wantStates[i])
		}
		if ev.Sequence != i+1 {
			t.Errorf("event %d sequence = %d, want %d", i, ev.Sequence, i+1)
		}
	}
}

func TestSubscribe_ReceivesLiveEvents(t *testing.T) {
	svc := testService()
	ctx := context.Background()
	actorID, workspaceID, snapshotID := uuid.New(), uuid.New(), uuid.New()

	created, err := svc.Create(ctx, actorID, workspaceID, snapshotID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ch, cancel := svc.Subscribe(created.ID)
	defer cancel()

	seen := map[twin.State]bool{}
	deadline := time.After(2 * time.Second)
	for len(seen) < 5 {
		select {
		case ev := <-ch:
			seen[ev.State] = true
		case <-deadline:
			t.Fatalf("timed out waiting for live events; saw %v", seen)
		}
	}
	if !seen[twin.StateReady] {
		t.Error("never observed a Ready event over the subscription")
	}
}

func TestStartStopDestroy_EnforceStateMachine(t *testing.T) {
	svc := testService()
	ctx := context.Background()
	actorID, workspaceID, snapshotID := uuid.New(), uuid.New(), uuid.New()

	created, err := svc.Create(ctx, actorID, workspaceID, snapshotID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Starting before the build has reached Ready is rejected by the state
	// machine, not silently accepted.
	if _, err := svc.Start(ctx, actorID, workspaceID, created.ID); err == nil {
		t.Error("Start on a Requested twin should fail")
	}

	waitForState(t, svc, workspaceID, created.ID, twin.StateReady)

	running, err := svc.Start(ctx, actorID, workspaceID, created.ID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if running.State != twin.StateRunning {
		t.Fatalf("state after Start = %s, want %s", running.State, twin.StateRunning)
	}

	stopped, err := svc.Stop(ctx, actorID, workspaceID, created.ID)
	if err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if stopped.State != twin.StateStopped {
		t.Fatalf("state after Stop = %s, want %s", stopped.State, twin.StateStopped)
	}

	destroyed, err := svc.Destroy(ctx, actorID, workspaceID, created.ID)
	if err != nil {
		t.Fatalf("Destroy: %v", err)
	}
	if destroyed.State != twin.StateDestroyed {
		t.Fatalf("state after Destroy = %s, want %s", destroyed.State, twin.StateDestroyed)
	}

	// Destroyed is terminal: nothing further is allowed.
	if _, err := svc.Start(ctx, actorID, workspaceID, created.ID); err == nil {
		t.Error("Start on a Destroyed twin should fail")
	}
}

func TestGet_IsWorkspaceScoped(t *testing.T) {
	svc := testService()
	ctx := context.Background()
	actorID, workspaceID, otherWorkspaceID, snapshotID := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	created, err := svc.Create(ctx, actorID, workspaceID, snapshotID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if _, err := svc.Get(ctx, otherWorkspaceID, created.ID); err == nil {
		t.Error("Get from a different workspace should fail, not return another workspace's twin")
	}
}
