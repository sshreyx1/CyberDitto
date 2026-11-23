//go:build integration

package postgres

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

const integrationDoc = `{"schema_version":"1","hostname":"h","collected_at":"2026-11-23T06:00:00Z","fields":{"os.version":{"source":"os.version","collected_at":"2026-11-23T06:00:00Z","value":{"product_name":"x"}}}}`

func TestAssetRepo_SnapshotsVersionedAndIsolated(t *testing.T) {
	appPool, _ := testPools(t)
	ctx := context.Background()
	iamRepo := NewIAMRepo(appPool)
	repo := NewAssetRepo(appPool)

	wsA, _ := iamRepo.CreateWorkspace(ctx, "A")
	wsB, _ := iamRepo.CreateWorkspace(ctx, "B")

	a, err := repo.CreateAsset(ctx, wsA.ID, "laptop")
	if err != nil {
		t.Fatalf("CreateAsset: %v", err)
	}

	// Concurrent uploads for the same asset must get distinct, gapless
	// versions (the advisory lock serialises them).
	const n = 8
	var wg sync.WaitGroup
	versions := make(chan int, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := repo.CreateSnapshot(ctx, wsA.ID, a.ID, []byte(integrationDoc), "sha")
			if err != nil {
				t.Errorf("CreateSnapshot: %v", err)
				return
			}
			versions <- s.Version
		}()
	}
	wg.Wait()
	close(versions)
	seen := map[int]bool{}
	for v := range versions {
		if seen[v] {
			t.Fatalf("duplicate snapshot version %d", v)
		}
		seen[v] = true
	}
	for v := 1; v <= n; v++ {
		if !seen[v] {
			t.Fatalf("missing snapshot version %d (got %v)", v, seen)
		}
	}

	// Workspace B can see none of it — not the asset, not its snapshots,
	// and can't upload into it.
	if _, err := repo.GetAsset(ctx, wsB.ID, a.ID); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("GetAsset from workspace B = %v, want ErrNotFound", err)
	}
	if list, _ := repo.ListAssets(ctx, wsB.ID); len(list) != 0 {
		t.Errorf("ListAssets(B) = %d assets, want 0", len(list))
	}
	if _, err := repo.CreateSnapshot(ctx, wsB.ID, a.ID, []byte(integrationDoc), "sha"); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("CreateSnapshot into another workspace's asset = %v, want ErrNotFound", err)
	}

	summaries, err := repo.ListSnapshots(ctx, wsA.ID, a.ID)
	if err != nil || len(summaries) != n || summaries[0].Data != nil || summaries[0].Version != n {
		t.Fatalf("ListSnapshots = %d (first %+v), %v", len(summaries), summaries[0], err)
	}
	full, err := repo.GetSnapshot(ctx, wsA.ID, summaries[0].ID)
	if err != nil || len(full.Data) == 0 {
		t.Fatalf("GetSnapshot = %+v, %v", full, err)
	}
	if _, err := repo.GetSnapshot(ctx, wsB.ID, summaries[0].ID); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("GetSnapshot from workspace B = %v, want ErrNotFound", err)
	}
}

func TestTwinRepo_StateAndEventsPersistAndAreIsolated(t *testing.T) {
	appPool, _ := testPools(t)
	ctx := context.Background()
	iamRepo := NewIAMRepo(appPool)
	assets := NewAssetRepo(appPool)
	repo := NewTwinRepo(appPool)

	wsA, _ := iamRepo.CreateWorkspace(ctx, "A")
	wsB, _ := iamRepo.CreateWorkspace(ctx, "B")
	a, _ := assets.CreateAsset(ctx, wsA.ID, "laptop")
	snap, _ := assets.CreateSnapshot(ctx, wsA.ID, a.ID, []byte(integrationDoc), "sha")

	tw, err := repo.Create(ctx, wsA.ID, snap.ID, "fake")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tw.State != twin.StateRequested {
		t.Fatalf("initial state = %s", tw.State)
	}

	updated, err := repo.UpdateState(ctx, wsA.ID, tw.ID, twin.StateImaging)
	if err != nil || updated.State != twin.StateImaging {
		t.Fatalf("UpdateState = %+v, %v", updated, err)
	}
	if _, err := repo.UpdateState(ctx, wsB.ID, tw.ID, twin.StateFailed); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("UpdateState from workspace B = %v, want ErrNotFound", err)
	}

	for i, s := range []twin.State{twin.StateImaging, twin.StateProvisioning, twin.StateApplying} {
		ev, err := repo.AppendEvent(ctx, wsA.ID, twin.Event{TwinID: tw.ID, State: s, Message: string(s)})
		if err != nil {
			t.Fatalf("AppendEvent: %v", err)
		}
		if ev.Sequence != i+1 {
			t.Fatalf("event sequence = %d, want %d", ev.Sequence, i+1)
		}
	}
	after1, err := repo.ListEvents(ctx, wsA.ID, tw.ID, 1)
	if err != nil || len(after1) != 2 || after1[0].Sequence != 2 {
		t.Fatalf("ListEvents after 1 = %+v, %v", after1, err)
	}
	if fromB, _ := repo.ListEvents(ctx, wsB.ID, tw.ID, 0); len(fromB) != 0 {
		t.Errorf("ListEvents from workspace B returned %d events, want 0", len(fromB))
	}
	if _, err := repo.Get(ctx, wsB.ID, tw.ID); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("Get from workspace B = %v, want ErrNotFound", err)
	}

	// A twin can't reference a snapshot id that doesn't exist (FK).
	if _, err := repo.Create(ctx, wsA.ID, uuid.New(), "fake"); err == nil {
		t.Error("Create with a dangling snapshot id should fail the foreign key")
	}
}
