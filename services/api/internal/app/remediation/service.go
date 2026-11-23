// Package remediation exports reviewed fix bundles for an assessment's
// failed controls (docs/MASTER_PLAN.md §5.14 "Prod path: export bundle...
// never auto-applied by CyberDitto").
package remediation

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/remediation"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

type Service struct {
	assessments ports.AssessmentRepository
	audit       ports.AuditRepository
	now         func() time.Time
}

func NewService(assessments ports.AssessmentRepository, audit ports.AuditRepository) *Service {
	return &Service{assessments: assessments, audit: audit, now: time.Now}
}

// Plan returns what a bundle for this assessment would contain.
func (s *Service) Plan(ctx context.Context, workspaceID, assessmentID uuid.UUID) (remediation.Bundle, error) {
	a, err := s.assessments.Get(ctx, workspaceID, assessmentID)
	if err != nil {
		return remediation.Bundle{}, err
	}
	return remediation.Render(remediation.Meta{
		AssessmentID: a.ID.String(), SnapshotID: a.SnapshotID.String(),
		EngineVersion: a.EngineVersion, GeneratedAt: s.now(),
	}, a.Findings), nil
}

// Export renders the bundle and records who exported it: a script that
// changes security settings leaving the system is worth an audit entry.
func (s *Service) Export(ctx context.Context, actorID, workspaceID, assessmentID uuid.UUID) (remediation.Bundle, error) {
	b, err := s.Plan(ctx, workspaceID, assessmentID)
	if err != nil {
		return remediation.Bundle{}, err
	}
	ids := make([]string, 0, len(b.Automated))
	for _, t := range b.Automated {
		ids = append(ids, t.ControlID)
	}
	if _, err := s.audit.Record(ctx, workspaceID, &actorID, "remediation.export", "assessment", assessmentID.String(),
		map[string]any{"filename": b.Filename, "automated": ids, "manual": len(b.Manual)}); err != nil {
		return remediation.Bundle{}, err
	}
	return b, nil
}
