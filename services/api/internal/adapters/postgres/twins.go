package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	pgdb "github.com/sshreyx1/cyberditto/services/api/internal/adapters/postgres/gen"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
)

type TwinRepo struct {
	pool *pgxpool.Pool
}

func NewTwinRepo(pool *pgxpool.Pool) *TwinRepo {
	return &TwinRepo{pool: pool}
}

func (r *TwinRepo) Create(ctx context.Context, workspaceID, snapshotID uuid.UUID, provider string) (twin.Twin, error) {
	var row pgdb.Twin
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		row, err = q.CreateTwin(ctx, pgdb.CreateTwinParams{
			WorkspaceID: workspaceID,
			SnapshotID:  snapshotID,
			Provider:    provider,
			State:       string(twin.StateRequested),
		})
		return err
	})
	if err != nil {
		return twin.Twin{}, err
	}
	return twinFromRow(row), nil
}

func (r *TwinRepo) Get(ctx context.Context, workspaceID, id uuid.UUID) (twin.Twin, error) {
	var row pgdb.Twin
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		row, err = q.GetTwin(ctx, pgdb.GetTwinParams{ID: id, WorkspaceID: workspaceID})
		return err
	})
	if err != nil {
		return twin.Twin{}, notFound(err)
	}
	return twinFromRow(row), nil
}

func (r *TwinRepo) List(ctx context.Context, workspaceID uuid.UUID) ([]twin.Twin, error) {
	var rows []pgdb.Twin
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		rows, err = q.ListTwins(ctx, workspaceID)
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]twin.Twin, 0, len(rows))
	for _, row := range rows {
		out = append(out, twinFromRow(row))
	}
	return out, nil
}

func (r *TwinRepo) UpdateState(ctx context.Context, workspaceID, id uuid.UUID, state twin.State) (twin.Twin, error) {
	var row pgdb.Twin
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		row, err = q.UpdateTwinState(ctx, pgdb.UpdateTwinStateParams{ID: id, WorkspaceID: workspaceID, State: string(state)})
		return err
	})
	if err != nil {
		return twin.Twin{}, notFound(err)
	}
	return twinFromRow(row), nil
}

func (r *TwinRepo) AppendEvent(ctx context.Context, workspaceID uuid.UUID, e twin.Event) (twin.Event, error) {
	var row pgdb.TwinEvent
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		if err := advisoryLock(ctx, q, "twin-events:"+e.TwinID.String()); err != nil {
			return err
		}
		seq, err := q.NextTwinEventSequence(ctx, e.TwinID)
		if err != nil {
			return err
		}
		row, err = q.InsertTwinEvent(ctx, pgdb.InsertTwinEventParams{
			WorkspaceID: workspaceID,
			TwinID:      e.TwinID,
			Sequence:    seq,
			State:       string(e.State),
			Message:     e.Message,
		})
		return err
	})
	if err != nil {
		return twin.Event{}, err
	}
	return twinEventFromRow(row), nil
}

func (r *TwinRepo) ListEvents(ctx context.Context, workspaceID, twinID uuid.UUID, afterSequence int) ([]twin.Event, error) {
	var rows []pgdb.TwinEvent
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		rows, err = q.ListTwinEvents(ctx, pgdb.ListTwinEventsParams{TwinID: twinID, WorkspaceID: workspaceID, Sequence: int32(afterSequence)})
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]twin.Event, 0, len(rows))
	for _, row := range rows {
		out = append(out, twinEventFromRow(row))
	}
	return out, nil
}

func twinFromRow(row pgdb.Twin) twin.Twin {
	t := twin.Twin{
		ID: row.ID, WorkspaceID: row.WorkspaceID, SnapshotID: row.SnapshotID,
		Provider: row.Provider, State: twin.State(row.State),
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
	if row.FidelityScore.Valid {
		if f, err := row.FidelityScore.Float64Value(); err == nil && f.Valid {
			v := f.Float64
			t.FidelityScore = &v
		}
	}
	return t
}

func twinEventFromRow(row pgdb.TwinEvent) twin.Event {
	return twin.Event{
		TwinID: row.TwinID, Sequence: int(row.Sequence),
		State: twin.State(row.State), Message: row.Message,
		CreatedAt: row.CreatedAt.Time,
	}
}
