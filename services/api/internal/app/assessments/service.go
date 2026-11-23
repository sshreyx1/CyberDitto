// Package assessments runs the native compliance engine against a snapshot
// or a twin and stores the result (docs/MASTER_PLAN.md §10 Phase 3).
package assessments

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

var (
	// ErrInvalidTarget: exactly one of snapshot or twin must be given.
	ErrInvalidTarget = errors.New("assessments: specify exactly one of snapshot_id or twin_id")
	// ErrTwinNotReady: a twin can only be assessed once it has been built.
	ErrTwinNotReady = errors.New("assessments: twin is not built (must be ready, running, or stopped)")
)

// Target is what to assess.
type Target struct {
	SnapshotID *uuid.UUID
	TwinID     *uuid.UUID
}

type Service struct {
	repo   ports.AssessmentRepository
	assets ports.AssetRepository
	twins  ports.TwinRepository
	audit  ports.AuditRepository
	now    func() time.Time
}

func NewService(repo ports.AssessmentRepository, assets ports.AssetRepository, twins ports.TwinRepository, audit ports.AuditRepository) *Service {
	return &Service{repo: repo, assets: assets, twins: twins, audit: audit, now: time.Now}
}

// Catalog exposes the control set so the console can show exactly what is
// checked and how it's weighted (docs/MASTER_PLAN.md §4.4: scoring is
// documented in-app).
func (s *Service) Catalog() []compliance.Control { return compliance.Catalog() }

// Run evaluates the native control catalog and stores the result.
//
// Assessing a twin evaluates the configuration the twin was built from.
// With the fake lab provider that is exactly the twin's configuration;
// once a real provider exists, a twin assessment should evaluate a snapshot
// re-collected from inside the twin instead (the Phase 2 Fidelity Score
// work) — that swap happens here and nowhere else.
func (s *Service) Run(ctx context.Context, actorID, workspaceID uuid.UUID, target Target) (compliance.Assessment, error) {
	if (target.SnapshotID == nil) == (target.TwinID == nil) {
		return compliance.Assessment{}, ErrInvalidTarget
	}

	snapshotID := target.SnapshotID
	if target.TwinID != nil {
		t, err := s.twins.Get(ctx, workspaceID, *target.TwinID)
		if err != nil {
			return compliance.Assessment{}, err
		}
		switch t.State {
		case twin.StateReady, twin.StateRunning, twin.StateStopped:
		default:
			return compliance.Assessment{}, fmt.Errorf("%w (state %s)", ErrTwinNotReady, t.State)
		}
		snapshotID = &t.SnapshotID
	}

	snap, err := s.assets.GetSnapshot(ctx, workspaceID, *snapshotID)
	if err != nil {
		return compliance.Assessment{}, err
	}
	doc, err := asset.ParseDocument(snap.Data)
	if err != nil {
		// Stored snapshots were validated on upload; failing here means the
		// stored data and this build's parser disagree.
		return compliance.Assessment{}, fmt.Errorf("assessments: stored snapshot %s no longer parses: %w", snap.ID, err)
	}

	findings := compliance.Evaluate(compliance.Catalog(), compliance.NewFacts(doc, s.now()))
	score, counts := compliance.Score(findings)

	stored, err := s.repo.Create(ctx, compliance.Assessment{
		WorkspaceID:   workspaceID,
		SnapshotID:    snap.ID,
		TwinID:        target.TwinID,
		Engine:        compliance.EngineName,
		EngineVersion: compliance.EngineVersion,
		Score:         score,
		Counts:        counts,
		Findings:      findings,
	}, actorID)
	if err != nil {
		return compliance.Assessment{}, fmt.Errorf("assessments: store: %w", err)
	}

	meta := map[string]any{
		"snapshot_id": snap.ID.String(), "engine_version": compliance.EngineVersion,
		"passed": counts.Passed, "failed": counts.Failed, "unknown": counts.Unknown,
	}
	if target.TwinID != nil {
		meta["twin_id"] = target.TwinID.String()
	}
	if score != nil {
		meta["score"] = *score
	}
	if _, err := s.audit.Record(ctx, workspaceID, &actorID, "assessment.run", "assessment", stored.ID.String(), meta); err != nil {
		return compliance.Assessment{}, err
	}
	return stored, nil
}

func (s *Service) List(ctx context.Context, workspaceID uuid.UUID) ([]compliance.Assessment, error) {
	return s.repo.List(ctx, workspaceID, 100)
}

func (s *Service) Get(ctx context.Context, workspaceID, id uuid.UUID) (compliance.Assessment, error) {
	return s.repo.Get(ctx, workspaceID, id)
}
