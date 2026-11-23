package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	pgdb "github.com/sshreyx1/cyberditto/services/api/internal/adapters/postgres/gen"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
)

type AssessmentRepo struct {
	pool *pgxpool.Pool
}

func NewAssessmentRepo(pool *pgxpool.Pool) *AssessmentRepo {
	return &AssessmentRepo{pool: pool}
}

func (r *AssessmentRepo) Create(ctx context.Context, a compliance.Assessment, createdBy uuid.UUID) (compliance.Assessment, error) {
	var score pgtype.Numeric
	if a.Score != nil {
		if err := score.Scan(strconv.FormatFloat(*a.Score, 'f', 2, 64)); err != nil {
			return compliance.Assessment{}, fmt.Errorf("assessments: encode score: %w", err)
		}
	}
	var twinID pgtype.UUID
	if a.TwinID != nil {
		twinID = pgtype.UUID{Bytes: *a.TwinID, Valid: true}
	}

	var row pgdb.Assessment
	err := withWorkspaceContext(ctx, r.pool, a.WorkspaceID, &createdBy, func(q *pgdb.Queries) error {
		var err error
		row, err = q.InsertAssessment(ctx, pgdb.InsertAssessmentParams{
			WorkspaceID:   a.WorkspaceID,
			SnapshotID:    a.SnapshotID,
			TwinID:        twinID,
			Engine:        a.Engine,
			EngineVersion: a.EngineVersion,
			PostureScore:  score,
			Passed:        int32(a.Counts.Passed),
			Failed:        int32(a.Counts.Failed),
			Unknown:       int32(a.Counts.Unknown),
			NotApplicable: int32(a.Counts.NotApplicable),
			CreatedBy:     pgtype.UUID{Bytes: createdBy, Valid: true},
		})
		if err != nil {
			return err
		}
		for _, f := range a.Findings {
			var observed []byte
			if f.Observed != nil {
				if observed, err = json.Marshal(f.Observed); err != nil {
					return fmt.Errorf("assessments: encode evidence for %s: %w", f.ControlID, err)
				}
			}
			techniques := f.Techniques
			if techniques == nil {
				techniques = []string{}
			}
			if err := q.InsertFinding(ctx, pgdb.InsertFindingParams{
				WorkspaceID:  a.WorkspaceID,
				AssessmentID: row.ID,
				ControlID:    f.ControlID,
				Title:        f.Title,
				Category:     f.Category,
				Severity:     string(f.Severity),
				Status:       string(f.Status),
				Observed:     observed,
				Expected:     f.Expected,
				Reason:       f.Reason,
				Remediation:  f.Remediation,
				Techniques:   techniques,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return compliance.Assessment{}, err
	}
	out := assessmentFromRow(row)
	out.Findings = a.Findings
	return out, nil
}

func (r *AssessmentRepo) List(ctx context.Context, workspaceID uuid.UUID, limit int) ([]compliance.Assessment, error) {
	var rows []pgdb.Assessment
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		rows, err = q.ListAssessments(ctx, pgdb.ListAssessmentsParams{WorkspaceID: workspaceID, Limit: int32(limit)})
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]compliance.Assessment, 0, len(rows))
	for _, row := range rows {
		out = append(out, assessmentFromRow(row))
	}
	return out, nil
}

func (r *AssessmentRepo) ListForAsset(ctx context.Context, workspaceID, assetID uuid.UUID, limit int) ([]compliance.Assessment, error) {
	var rows []pgdb.Assessment
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		rows, err = q.ListAssessmentsForAsset(ctx, pgdb.ListAssessmentsForAssetParams{AssetID: assetID, WorkspaceID: workspaceID, Limit: int32(limit)})
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]compliance.Assessment, 0, len(rows))
	for _, row := range rows {
		out = append(out, assessmentFromRow(row))
	}
	return out, nil
}

func (r *AssessmentRepo) Get(ctx context.Context, workspaceID, id uuid.UUID) (compliance.Assessment, error) {
	var row pgdb.Assessment
	var findings []pgdb.Finding
	err := withWorkspaceContext(ctx, r.pool, workspaceID, nil, func(q *pgdb.Queries) error {
		var err error
		if row, err = q.GetAssessment(ctx, pgdb.GetAssessmentParams{ID: id, WorkspaceID: workspaceID}); err != nil {
			return err
		}
		findings, err = q.ListFindings(ctx, pgdb.ListFindingsParams{AssessmentID: id, WorkspaceID: workspaceID})
		return err
	})
	if err != nil {
		return compliance.Assessment{}, notFound(err)
	}
	a := assessmentFromRow(row)
	a.Findings = make([]compliance.Finding, 0, len(findings))
	for _, f := range findings {
		var observed any
		if len(f.Observed) > 0 {
			observed = json.RawMessage(f.Observed)
		}
		a.Findings = append(a.Findings, compliance.Finding{
			ControlID: f.ControlID, Title: f.Title, Category: f.Category,
			Severity: compliance.Severity(f.Severity), Status: compliance.Status(f.Status),
			Observed: observed, Expected: f.Expected, Reason: f.Reason,
			Remediation: f.Remediation, Techniques: f.Techniques,
		})
	}
	return a, nil
}

func assessmentFromRow(row pgdb.Assessment) compliance.Assessment {
	a := compliance.Assessment{
		ID: row.ID, WorkspaceID: row.WorkspaceID, SnapshotID: row.SnapshotID,
		Engine: row.Engine, EngineVersion: row.EngineVersion,
		Counts: compliance.Counts{
			Passed: int(row.Passed), Failed: int(row.Failed),
			Unknown: int(row.Unknown), NotApplicable: int(row.NotApplicable),
		},
		Findings:  []compliance.Finding{},
		CreatedAt: row.CreatedAt.Time,
	}
	if row.TwinID.Valid {
		id := uuid.UUID(row.TwinID.Bytes)
		a.TwinID = &id
	}
	if row.PostureScore.Valid {
		if f, err := row.PostureScore.Float64Value(); err == nil && f.Valid {
			v := f.Float64
			a.Score = &v
		}
	}
	return a
}
