package reports

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/fake"
	appassessments "github.com/sshreyx1/cyberditto/services/api/internal/app/assessments"
	appgaps "github.com/sshreyx1/cyberditto/services/api/internal/app/gaps"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/report"
)

func TestExport_IsAudited(t *testing.T) {
	ctx := context.Background()
	ws, actor := uuid.New(), uuid.New()
	assets := fake.NewAssetRepo()
	audit := fake.NewAuditRepo()
	twins := fake.NewTwinRepo()
	repo := fake.NewAssessmentRepo(assets)

	doc, err := os.ReadFile("../../domain/compliance/testdata/weak.json")
	if err != nil {
		t.Fatal(err)
	}
	a, _ := assets.CreateAsset(ctx, ws, "weak box")
	snap, err := assets.CreateSnapshot(ctx, ws, a.ID, doc, "sha")
	if err != nil {
		t.Fatal(err)
	}
	assessment, err := appassessments.NewService(repo, assets, twins, audit).Run(ctx, actor, ws, appassessments.Target{SnapshotID: &snap.ID})
	if err != nil {
		t.Fatal(err)
	}

	svc := NewService(appgaps.NewService(repo, twins, fake.NewRunRepo()), assets, audit)
	exp, err := svc.Export(ctx, actor, ws, assessment.ID, report.FormatSARIF)
	if err != nil {
		t.Fatal(err)
	}
	if exp.Filename == "" || exp.Body == nil {
		t.Fatalf("empty export: %+v", exp)
	}

	events, _ := audit.List(ctx, ws, 50, 0)
	for _, e := range events {
		if e.Action == "report.export" && e.TargetID == assessment.ID.String() && e.Metadata["format"] == "sarif" {
			return
		}
	}
	t.Fatalf("no report.export audit event for the export; got %+v", events)
}

func TestExport_OtherWorkspaceIsNotFound(t *testing.T) {
	assets := fake.NewAssetRepo()
	repo := fake.NewAssessmentRepo(assets)
	svc := NewService(appgaps.NewService(repo, fake.NewTwinRepo(), fake.NewRunRepo()), assets, fake.NewAuditRepo())
	if _, err := svc.Export(context.Background(), uuid.New(), uuid.New(), uuid.New(), report.FormatJSON); err == nil {
		t.Fatal("export of a nonexistent assessment succeeded")
	}
}
