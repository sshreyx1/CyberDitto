// Package collect implements the collector's categories
// (docs/MASTER_PLAN.md §5.10). The collector is single-direction and
// read-only: nothing in this package makes a network call, and nothing
// here ever collects file contents, documents, browser data, credentials,
// tokens, event-log contents, clipboard data, or screenshots — only the L1
// configuration categories MASTER_PLAN §5.10 lists.
package collect

import "context"

// Collector gathers exactly one category. Collect must never fabricate a
// value on failure: returning an error is how a category becomes "unknown"
// in the resulting snapshot (snapshot.Field.Error), which is the entire
// point of this interface existing rather than every category being a
// plain function — it makes "I don't know" a representable, tested outcome
// instead of something a collector can quietly paper over.
type Collector interface {
	// ID is a stable, dotted category name (e.g. "os.version"), used as
	// both the snapshot's field key and the manifest entry's id.
	ID() string
	// Description is shown to the operator in the pre-run consent
	// manifest and in --dry-run output, in plain language, before
	// anything is collected.
	Description() string
	Collect(ctx context.Context) (any, error)
}

// CategoryInfo is one manifest entry: what a category is, without having
// run it.
type CategoryInfo struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// Manifest describes every collector that would run, for --dry-run and the
// consent prompt (docs/MASTER_PLAN.md §5.10: "requires --consent and
// prints a manifest of every category it will collect before running").
func Manifest(collectors []Collector) []CategoryInfo {
	out := make([]CategoryInfo, 0, len(collectors))
	for _, c := range collectors {
		out = append(out, CategoryInfo{ID: c.ID(), Description: c.Description()})
	}
	return out
}
