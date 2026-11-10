// Package domain holds pure business logic with no I/O (see docs/MASTER_PLAN.md §5.1).
package domain

// Health reports the liveness of the control plane. It carries no
// infrastructure dependencies so it can be unit tested without a server.
type Health struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// Version is set at build time via -ldflags; it defaults to "dev" so local
// builds and tests still produce a deterministic value.
var Version = "dev"

// NewHealth returns the current health snapshot.
func NewHealth() Health {
	return Health{Status: "ok", Version: Version}
}
