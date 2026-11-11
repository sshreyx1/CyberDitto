package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	pgdb "github.com/sshreyx1/cyberditto/services/api/internal/adapters/postgres/gen"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/audit"
)

type AuditRepo struct {
	pool *pgxpool.Pool
}

func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{pool: pool}
}

// Record appends one event to a workspace's hash chain. The
// pg_advisory_xact_lock call serializes concurrent Record calls for the
// same workspace for the lifetime of the transaction, which is what makes
// "read the last hash, then insert the next one" race-free without a
// table-level lock (docs/MASTER_PLAN.md §5.6: the chain must never fork).
func (r *AuditRepo) Record(ctx context.Context, workspaceID uuid.UUID, actorUserID *uuid.UUID, action, targetType, targetID string, metadata map[string]any) (audit.Event, error) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return audit.Event{}, fmt.Errorf("audit: marshal metadata: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return audit.Event{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `SELECT set_config('app.current_workspace_id', $1, true)`, workspaceID.String()); err != nil {
		return audit.Event{}, err
	}
	// Always set app.current_user_id explicitly (to uuid.Nil if there's no
	// actor) rather than leaving it unset — see the comment on
	// withWorkspaceContext in tx.go for why that matters.
	actorForGUC := uuid.Nil
	if actorUserID != nil {
		actorForGUC = *actorUserID
	}
	if _, err := tx.Exec(ctx, `SELECT set_config('app.current_user_id', $1, true)`, actorForGUC.String()); err != nil {
		return audit.Event{}, err
	}
	// Serializes concurrent Record calls for this workspace until the
	// transaction ends (commit or rollback); hashtext gives a stable
	// int32 lock key from the workspace UUID's text form.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, workspaceID.String()); err != nil {
		return audit.Event{}, err
	}

	q := pgdb.New(tx)

	prevHash, err := q.GetLastAuditHash(ctx, pgtype.UUID{Bytes: workspaceID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			prevHash = audit.GenesisHash
		} else {
			return audit.Event{}, err
		}
	}

	hash, err := audit.ComputeHash(workspaceID, actorUserID, action, targetType, targetID, metadata, prevHash)
	if err != nil {
		return audit.Event{}, err
	}

	var targetIDPtr *string
	if targetID != "" {
		targetIDPtr = &targetID
	}
	var actorPg pgtype.UUID
	if actorUserID != nil {
		actorPg = pgtype.UUID{Bytes: *actorUserID, Valid: true}
	}

	row, err := q.InsertAuditEvent(ctx, pgdb.InsertAuditEventParams{
		WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
		ActorUserID: actorPg,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetIDPtr,
		Metadata:    metaJSON,
		PrevHash:    prevHash,
		Hash:        hash,
	})
	if err != nil {
		return audit.Event{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return audit.Event{}, err
	}

	return eventFromRow(row)
}

func (r *AuditRepo) List(ctx context.Context, workspaceID uuid.UUID, limit, offset int32) ([]audit.Event, error) {
	var rows []pgdb.AuditEvent
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		rows, err = q.ListAuditEvents(ctx, pgdb.ListAuditEventsParams{
			WorkspaceID: pgtype.UUID{Bytes: workspaceID, Valid: true},
			Limit:       limit,
			Offset:      offset,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]audit.Event, 0, len(rows))
	for _, row := range rows {
		e, err := eventFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func eventFromRow(row pgdb.AuditEvent) (audit.Event, error) {
	var metadata map[string]any
	if len(row.Metadata) > 0 {
		if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
			return audit.Event{}, fmt.Errorf("audit: unmarshal metadata: %w", err)
		}
	}
	e := audit.Event{
		ID:         row.ID,
		Action:     row.Action,
		TargetType: row.TargetType,
		Metadata:   metadata,
		PrevHash:   row.PrevHash,
		Hash:       row.Hash,
		CreatedAt:  row.CreatedAt.Time,
	}
	if row.WorkspaceID.Valid {
		e.WorkspaceID = uuid.UUID(row.WorkspaceID.Bytes)
	}
	if row.ActorUserID.Valid {
		id := uuid.UUID(row.ActorUserID.Bytes)
		e.ActorUserID = &id
	}
	if row.TargetID != nil {
		e.TargetID = *row.TargetID
	}
	return e, nil
}
