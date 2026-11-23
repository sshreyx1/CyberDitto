// Package reports exports an assessment, with its prioritised gaps, as
// JSON, SARIF 2.1.0 or an ATT&CK Navigator layer (docs/MASTER_PLAN.md §10
// Phase 5). Every export is audit-logged: a report describes a machine's
// weaknesses, so who took one out of the system matters.
package reports

import (
	"context"
	"time"

	"github.com/google/uuid"
	appgaps "github.com/sshreyx1/cyberditto/services/api/internal/app/gaps"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/report"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

type Service struct {
	gaps   *appgaps.Service
	assets ports.AssetRepository
	audit  ports.AuditRepository
	now    func() time.Time
}

func NewService(gaps *appgaps.Service, assets ports.AssetRepository, audit ports.AuditRepository) *Service {
	return &Service{gaps: gaps, assets: assets, audit: audit, now: time.Now}
}

// Export is a rendered report ready to be written to the response.
type Export struct {
	Filename string
	Body     any
}

func (s *Service) Export(ctx context.Context, actorID, workspaceID, assessmentID uuid.UUID, format report.Format) (Export, error) {
	res, err := s.gaps.ForAssessment(ctx, workspaceID, assessmentID)
	if err != nil {
		return Export{}, err
	}
	snap, err := s.assets.GetSnapshot(ctx, workspaceID, res.Assessment.SnapshotID)
	if err != nil {
		return Export{}, err
	}
	a, err := s.assets.GetAsset(ctx, workspaceID, snap.AssetID)
	if err != nil {
		return Export{}, err
	}

	in := report.Input{AssetName: a.Name, Assessment: res.Assessment, Gaps: res.Gaps, GeneratedAt: s.now()}
	var body any
	switch format {
	case report.FormatSARIF:
		body = report.SARIF(in, compliance.Catalog())
	case report.FormatNavigator:
		body = report.Navigator(in)
	default:
		format = report.FormatJSON
		body = report.JSON(in)
	}
	filename := report.Filename(in, format)

	if _, err := s.audit.Record(ctx, workspaceID, &actorID, "report.export", "assessment", assessmentID.String(),
		map[string]any{"format": string(format), "filename": filename}); err != nil {
		return Export{}, err
	}
	return Export{Filename: filename, Body: body}, nil
}
