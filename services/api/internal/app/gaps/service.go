// Package gaps computes prioritised gaps for an assessment: its failed
// controls joined with emulation outcomes from twins built from the same
// configuration (docs/MASTER_PLAN.md §5.13).
package gaps

import (
	"context"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/correlation"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

// maxRunsConsidered bounds how much run history is read per request.
const maxRunsConsidered = 200

type Service struct {
	assessments ports.AssessmentRepository
	twins       ports.TwinRepository
	runs        ports.RunRepository
}

func NewService(assessments ports.AssessmentRepository, twins ports.TwinRepository, runs ports.RunRepository) *Service {
	return &Service{assessments: assessments, twins: twins, runs: runs}
}

// Result is an assessment's gaps plus the evidence base they were built
// from, so the console can say how much emulation evidence exists.
type Result struct {
	Assessment   compliance.Assessment
	Gaps         []correlation.Gap
	TwinsUsed    []uuid.UUID
	RunsUsed     int
	Observations int
}

// ForAssessment computes gaps for one assessment. Evidence comes from runs
// on any twin built from the assessment's snapshot: they share the same
// configuration, so what a technique did on one is evidence for all.
func (s *Service) ForAssessment(ctx context.Context, workspaceID, assessmentID uuid.UUID) (Result, error) {
	a, err := s.assessments.Get(ctx, workspaceID, assessmentID)
	if err != nil {
		return Result{}, err
	}

	twins, err := s.twins.List(ctx, workspaceID)
	if err != nil {
		return Result{}, err
	}
	relevant := map[uuid.UUID]bool{}
	var used []uuid.UUID
	for _, t := range twins {
		if t.SnapshotID == a.SnapshotID {
			relevant[t.ID] = true
			used = append(used, t.ID)
		}
	}

	var observations []correlation.Observation
	runsUsed := 0
	if len(relevant) > 0 {
		runs, err := s.runs.List(ctx, workspaceID, maxRunsConsidered)
		if err != nil {
			return Result{}, err
		}
		for _, r := range runs {
			if !relevant[r.TwinID] || r.Status == emulation.RunPreflightFailed || r.Status == emulation.RunRunning {
				continue
			}
			full, err := s.runs.Get(ctx, workspaceID, r.ID)
			if err != nil {
				return Result{}, err
			}
			runsUsed++
			for _, e := range full.Executions {
				observations = append(observations, correlation.Observation{
					TechniqueID: e.TechniqueID, Outcome: e.Outcome, RunID: r.ID, TestName: e.TestName,
				})
			}
		}
	}

	return Result{
		Assessment:   a,
		Gaps:         correlation.Correlate(a.Findings, observations),
		TwinsUsed:    used,
		RunsUsed:     runsUsed,
		Observations: len(observations),
	}, nil
}
