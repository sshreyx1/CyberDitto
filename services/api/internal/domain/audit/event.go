// Package audit holds the pure, hash-chained audit-log domain logic
// (docs/MASTER_PLAN.md §5.6 invariants, §6.4). Persisting events and
// serializing the chain under concurrency is an adapter concern; this
// package only computes the hash a given event must carry.
package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// GenesisHash is the PrevHash recorded on the first event in a chain (a
// workspace's very first audit event has no predecessor to chain from).
const GenesisHash = "genesis"

type Event struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	ActorUserID *uuid.UUID
	Action      string
	TargetType  string
	TargetID    string
	Metadata    map[string]any
	PrevHash    string
	Hash        string
	CreatedAt   time.Time
}

// payload is the deterministic subset of an Event that gets hashed. Deriving
// it as its own type (rather than hashing Event directly) keeps the hash
// stable even if unrelated Event fields (e.g. ID, CreatedAt precision) change
// representation later.
type payload struct {
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	ActorUserID *uuid.UUID     `json:"actor_user_id"`
	Action      string         `json:"action"`
	TargetType  string         `json:"target_type"`
	TargetID    string         `json:"target_id"`
	Metadata    map[string]any `json:"metadata"`
	PrevHash    string         `json:"prev_hash"`
}

// ComputeHash deterministically hashes an event given the hash of the
// preceding event in its workspace's chain (or GenesisHash for the first).
func ComputeHash(workspaceID uuid.UUID, actorUserID *uuid.UUID, action, targetType, targetID string, metadata map[string]any, prevHash string) (string, error) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	b, err := json.Marshal(payload{
		WorkspaceID: workspaceID,
		ActorUserID: actorUserID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Metadata:    metadata,
		PrevHash:    prevHash,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// Verify reports whether an event's Hash field actually matches its own
// content and PrevHash — used by chain-integrity tests/tooling, never on the
// hot path (the adapter computes the hash once, at insert time).
func Verify(e Event) (bool, error) {
	want, err := ComputeHash(e.WorkspaceID, e.ActorUserID, e.Action, e.TargetType, e.TargetID, e.Metadata, e.PrevHash)
	if err != nil {
		return false, err
	}
	return want == e.Hash, nil
}
