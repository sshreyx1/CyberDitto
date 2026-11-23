package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
)

// AttackCatalog is the pinned, verified set of executable tests and the
// curated scenarios built from them (docs/adr/0006).
type AttackCatalog interface {
	Source() string
	Repository() string
	Commit() string
	CommitDate() string
	License() string
	Verified() (bool, string)
	Tests() []emulation.Test
	Test(guid string) (emulation.Test, bool)
	Scenarios() []emulation.Scenario
	Scenario(id string) (emulation.Scenario, bool)
}

// RunRepository persists emulation runs and their per-test executions.
type RunRepository interface {
	Create(ctx context.Context, r emulation.Run) (emulation.Run, error)
	AddExecution(ctx context.Context, workspaceID, runID uuid.UUID, e emulation.Execution) error
	Finish(ctx context.Context, workspaceID, runID uuid.UUID, status emulation.RunStatus, detail string, at time.Time) error
	Get(ctx context.Context, workspaceID, id uuid.UUID) (emulation.Run, error)
	// List returns runs without executions, newest first.
	List(ctx context.Context, workspaceID uuid.UUID, limit int) ([]emulation.Run, error)
}
