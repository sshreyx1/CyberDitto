// Package snapshot defines the collector's canonical output shape
// (docs/MASTER_PLAN.md §5.10). One JSON document, one Field per category.
// Every field carries its own source and collection time; a failed read is
// recorded as an error, never silently defaulted or omitted — the
// difference between "we checked and it's off" and "we couldn't check"
// must never be lost (this is the exact class of defect v1 shipped:
// MASTER_PLAN §2.3 #3, a collector that hardcoded `firewall.enabled =
// true` regardless of the real state).
package snapshot

import "time"

// SchemaVersion is bumped whenever the Field/Snapshot shape changes in a
// way that could break a consumer parsing stored snapshots.
const SchemaVersion = "1"

// Field is one collected category's result.
type Field struct {
	// Source identifies what was actually read (a registry key, a
	// PowerShell cmdlet, …) so a surprising value can be traced back to
	// its evidence, not just trusted.
	Source      string    `json:"source"`
	CollectedAt time.Time `json:"collected_at"`
	// Value is omitted entirely when Error is set — there is no
	// zero-value/default standing in for "unknown".
	Value any    `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

// Snapshot is the collector's entire output for one run.
type Snapshot struct {
	SchemaVersion string           `json:"schema_version"`
	Hostname      string           `json:"hostname"`
	CollectedAt   time.Time        `json:"collected_at"`
	Fields        map[string]Field `json:"fields"`
}
