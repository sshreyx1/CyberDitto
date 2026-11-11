package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	pgdb "github.com/sshreyx1/cyberditto/services/api/internal/adapters/postgres/gen"
)

// withWorkspaceContext runs fn inside a transaction with the Postgres
// session variables that the RLS policies in 0001_init.sql check
// (app.current_workspace_id, app.current_user_id) set via the parameterized
// set_config() so a UUID can never be used to inject SQL
// (docs/MASTER_PLAN.md §6.3: no shell/SQL-string building from validated
// typed values is still done via parameters, not string formatting).
// Both GUCs are always set explicitly, even to the nil UUID, rather than
// leaving either one unset: a pooled connection's custom GUC does not
// reliably read back as SQL NULL via current_setting(..., true) once it has
// been touched earlier in the connection's lifetime by another request (a
// bare SET LOCAL/set_config revert can leave it as '' instead of unset on
// some Postgres versions), and casting '' to uuid raises a hard error. The
// nil UUID can never match a real workspace_id/user_id, so it is a safe
// "no match intended" sentinel for whichever of the two is not in scope.
func withWorkspaceContext(ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, userID *uuid.UUID, fn func(q *pgdb.Queries) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op once committed

	if _, err := tx.Exec(ctx, `SELECT set_config('app.current_workspace_id', $1, true)`, workspaceID.String()); err != nil {
		return err
	}
	u := uuid.Nil
	if userID != nil {
		u = *userID
	}
	if _, err := tx.Exec(ctx, `SELECT set_config('app.current_user_id', $1, true)`, u.String()); err != nil {
		return err
	}

	if err := fn(pgdb.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// withUserContext is for RLS lookups that are scoped by user rather than by
// an already-known workspace (e.g. "which workspaces am I a member of").
func withUserContext(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID, fn func(q *pgdb.Queries) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `SELECT set_config('app.current_workspace_id', $1, true)`, uuid.Nil.String()); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `SELECT set_config('app.current_user_id', $1, true)`, userID.String()); err != nil {
		return err
	}
	if err := fn(pgdb.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
