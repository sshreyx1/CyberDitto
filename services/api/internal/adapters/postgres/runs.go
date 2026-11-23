package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	pgdb "github.com/sshreyx1/cyberditto/services/api/internal/adapters/postgres/gen"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

type RunRepo struct {
	pool *pgxpool.Pool
}

func NewRunRepo(pool *pgxpool.Pool) *RunRepo { return &RunRepo{pool: pool} }

func ts(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: true} }

func (r *RunRepo) Create(ctx context.Context, run emulation.Run) (emulation.Run, error) {
	preflight, err := json.Marshal(run.Preflight)
	if err != nil {
		return emulation.Run{}, fmt.Errorf("runs: encode preflight: %w", err)
	}
	guids := run.TestGUIDs
	if guids == nil {
		guids = []string{}
	}
	var row pgdb.EmulationRun
	actor := run.AuthorisedBy
	err = withWorkspaceContext(ctx, r.pool, run.WorkspaceID, &actor, func(q *pgdb.Queries) error {
		var err error
		row, err = q.InsertRun(ctx, pgdb.InsertRunParams{
			WorkspaceID: run.WorkspaceID, TwinID: run.TwinID, ScenarioID: run.ScenarioID,
			TestGuids: guids, Status: string(run.Status), Detail: run.Detail, Preflight: preflight,
			CatalogSource: run.CatalogSource, CatalogCommit: run.CatalogCommit, Provider: run.Provider,
			AuthorisedBy:           pgtype.UUID{Bytes: run.AuthorisedBy, Valid: run.AuthorisedBy != uuid.Nil},
			AuthorisationStatement: run.AuthorisationStatement,
		})
		return err
	})
	if err != nil {
		return emulation.Run{}, err
	}
	return runFromRow(row)
}

func (r *RunRepo) AddExecution(ctx context.Context, workspaceID, runID uuid.UUID, e emulation.Execution) error {
	return withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		return q.InsertExecution(ctx, pgdb.InsertExecutionParams{
			WorkspaceID: workspaceID, RunID: runID, TestGuid: e.TestGUID, TechniqueID: e.TechniqueID,
			TestName: e.TestName, RiskTier: string(e.RiskTier), Outcome: string(e.Outcome), Detail: e.Detail,
			StartedAt: ts(e.StartedAt), FinishedAt: ts(e.FinishedAt),
		})
	})
}

func (r *RunRepo) Finish(ctx context.Context, workspaceID, runID uuid.UUID, status emulation.RunStatus, detail string, at time.Time) error {
	return withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		n, err := q.FinishRun(ctx, pgdb.FinishRunParams{ID: runID, WorkspaceID: workspaceID, Status: string(status), Detail: detail, FinishedAt: ts(at)})
		if err != nil {
			return err
		}
		if n == 0 {
			return ports.ErrNotFound
		}
		return nil
	})
}

func (r *RunRepo) Get(ctx context.Context, workspaceID, id uuid.UUID) (emulation.Run, error) {
	var row pgdb.EmulationRun
	var execs []pgdb.RunExecution
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		if row, err = q.GetRun(ctx, pgdb.GetRunParams{ID: id, WorkspaceID: workspaceID}); err != nil {
			return err
		}
		execs, err = q.ListExecutions(ctx, pgdb.ListExecutionsParams{RunID: id, WorkspaceID: workspaceID})
		return err
	})
	if err != nil {
		return emulation.Run{}, notFound(err)
	}
	run, err := runFromRow(row)
	if err != nil {
		return emulation.Run{}, err
	}
	for _, e := range execs {
		run.Executions = append(run.Executions, emulation.Execution{
			TestGUID: e.TestGuid, TechniqueID: e.TechniqueID, TestName: e.TestName,
			RiskTier: emulation.RiskTier(e.RiskTier), Outcome: emulation.Outcome(e.Outcome), Detail: e.Detail,
			StartedAt: e.StartedAt.Time, FinishedAt: e.FinishedAt.Time,
		})
	}
	return run, nil
}

func (r *RunRepo) List(ctx context.Context, workspaceID uuid.UUID, limit int) ([]emulation.Run, error) {
	var rows []pgdb.EmulationRun
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		rows, err = q.ListRuns(ctx, pgdb.ListRunsParams{WorkspaceID: workspaceID, Limit: int32(limit)})
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]emulation.Run, 0, len(rows))
	for _, row := range rows {
		run, err := runFromRow(row)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, nil
}

func runFromRow(row pgdb.EmulationRun) (emulation.Run, error) {
	var checks []emulation.Check
	if err := json.Unmarshal(row.Preflight, &checks); err != nil {
		return emulation.Run{}, fmt.Errorf("runs: decode preflight: %w", err)
	}
	run := emulation.Run{
		ID: row.ID, WorkspaceID: row.WorkspaceID, TwinID: row.TwinID, ScenarioID: row.ScenarioID,
		TestGUIDs: row.TestGuids, Status: emulation.RunStatus(row.Status), Detail: row.Detail,
		Preflight: checks, CatalogSource: row.CatalogSource, CatalogCommit: row.CatalogCommit,
		Provider: row.Provider, AuthorisationStatement: row.AuthorisationStatement,
		CreatedAt: row.CreatedAt.Time,
	}
	if row.AuthorisedBy.Valid {
		run.AuthorisedBy = uuid.UUID(row.AuthorisedBy.Bytes)
	}
	if row.FinishedAt.Valid {
		t := row.FinishedAt.Time
		run.FinishedAt = &t
	}
	return run, nil
}
