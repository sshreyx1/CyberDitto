package collect

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeCollector struct {
	id, description string
	value           any
	err             error
}

func (f fakeCollector) ID() string          { return f.id }
func (f fakeCollector) Description() string { return f.description }
func (f fakeCollector) Collect(ctx context.Context) (any, error) {
	return f.value, f.err
}

func TestManifest(t *testing.T) {
	collectors := []Collector{
		fakeCollector{id: "a", description: "collects a"},
		fakeCollector{id: "b", description: "collects b"},
	}
	got := Manifest(collectors)
	want := []CategoryInfo{{ID: "a", Description: "collects a"}, {ID: "b", Description: "collects b"}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("Manifest() = %+v, want %+v", got, want)
	}
}

// TestRun_GoldenSnapshot is the collector's golden-file test
// (docs/MASTER_PLAN.md §5.10): a fixed set of collectors — one succeeding,
// one failing — run through Run with a fixed clock, and the resulting
// Snapshot JSON must match the recorded fixture byte-for-byte. The failing
// collector is the point of the test: its category must come back as
// "unknown" (an Error string, Value omitted), never a zero/default value
// standing in for a real reading.
func TestRun_GoldenSnapshot(t *testing.T) {
	fixedTime := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	now := func() time.Time { return fixedTime }

	collectors := []Collector{
		fakeCollector{id: "os.version", description: "d1", value: OSVersion{ProductName: "Windows 11 Pro", CurrentBuild: "26200"}},
		fakeCollector{id: "firewall.profiles", description: "d2", err: errFakeRegistry},
	}

	snap := Run(context.Background(), collectors, now)
	// Hostname varies by machine; pin it for a stable golden comparison.
	snap.Hostname = "test-host"

	got, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}

	goldenPath := filepath.Join("testdata", "run_golden.json")
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.WriteFile(goldenPath, append(got, '\n'), 0o644); err != nil {
			t.Fatalf("write golden file: %v", err)
		}
	}

	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	if string(got)+"\n" != string(want) {
		t.Errorf("snapshot JSON does not match testdata/run_golden.json:\ngot:\n%s\nwant:\n%s", got, want)
	}
}
