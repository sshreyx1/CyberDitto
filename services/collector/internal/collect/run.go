package collect

import (
	"context"
	"os"
	"time"

	"github.com/sshreyx1/cyberditto/services/collector/internal/snapshot"
)

// Run executes every collector and assembles the canonical Snapshot. One
// collector failing never stops the others — an "unknown" firewall reading
// must not cost you the OS version too (docs/MASTER_PLAN.md §5.10).
func Run(ctx context.Context, collectors []Collector, now func() time.Time) snapshot.Snapshot {
	// Best-effort: an empty hostname is a visibly odd value an operator can
	// notice, so this doesn't need the Field{Error} treatment the way a
	// Collector's result does.
	hostname, _ := os.Hostname()

	snap := snapshot.Snapshot{
		SchemaVersion: snapshot.SchemaVersion,
		Hostname:      hostname,
		CollectedAt:   now(),
		Fields:        make(map[string]snapshot.Field, len(collectors)),
	}

	for _, c := range collectors {
		collectedAt := now()
		value, err := c.Collect(ctx)
		field := snapshot.Field{Source: c.ID(), CollectedAt: collectedAt}
		if err != nil {
			field.Error = err.Error()
		} else {
			field.Value = value
		}
		snap.Fields[c.ID()] = field
	}

	return snap
}
