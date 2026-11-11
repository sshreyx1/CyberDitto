package fake

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/audit"
)

type AuditRepo struct {
	mu     sync.Mutex
	events map[uuid.UUID][]audit.Event // keyed by workspace
}

func NewAuditRepo() *AuditRepo {
	return &AuditRepo{events: map[uuid.UUID][]audit.Event{}}
}

func (r *AuditRepo) Record(ctx context.Context, workspaceID uuid.UUID, actorUserID *uuid.UUID, action, targetType, targetID string, metadata map[string]any) (audit.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	prevHash := audit.GenesisHash
	if existing := r.events[workspaceID]; len(existing) > 0 {
		prevHash = existing[len(existing)-1].Hash
	}

	hash, err := audit.ComputeHash(workspaceID, actorUserID, action, targetType, targetID, metadata, prevHash)
	if err != nil {
		return audit.Event{}, err
	}

	e := audit.Event{
		ID: uuid.New(), WorkspaceID: workspaceID, ActorUserID: actorUserID,
		Action: action, TargetType: targetType, TargetID: targetID,
		Metadata: metadata, PrevHash: prevHash, Hash: hash, CreatedAt: time.Now(),
	}
	r.events[workspaceID] = append(r.events[workspaceID], e)
	return e, nil
}

func (r *AuditRepo) List(ctx context.Context, workspaceID uuid.UUID, limit, offset int32) ([]audit.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	all := r.events[workspaceID]
	start := int(offset)
	if start > len(all) {
		start = len(all)
	}
	end := start + int(limit)
	if end > len(all) {
		end = len(all)
	}
	// Newest first, matching the Postgres adapter's ORDER BY created_at DESC.
	out := make([]audit.Event, 0, end-start)
	for i := len(all) - 1; i >= 0; i-- {
		out = append(out, all[i])
	}
	if start > len(out) {
		return nil, nil
	}
	if end > len(out) {
		end = len(out)
	}
	return out[start:end], nil
}
