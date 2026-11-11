package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/audit"
)

// AuditRepository appends to the hash-chained audit log. Implementations
// must serialize concurrent appends per workspace (e.g. an advisory lock)
// so the chain can never fork (docs/MASTER_PLAN.md §5.6, §6.4).
type AuditRepository interface {
	Record(ctx context.Context, workspaceID uuid.UUID, actorUserID *uuid.UUID, action, targetType, targetID string, metadata map[string]any) (audit.Event, error)
	List(ctx context.Context, workspaceID uuid.UUID, limit, offset int32) ([]audit.Event, error)
}
