package http

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

func TestRemediation_PlanAndExport(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)
	doc, err := os.ReadFile("../../domain/compliance/testdata/weak.json")
	if err != nil {
		t.Fatal(err)
	}
	rec := doJSON(t, router, http.MethodPost, "/api/v1/assets", createAssetRequest{Name: "weak box"}, cookies)
	var asset struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&asset)
	rec = doRaw(t, router, http.MethodPost, "/api/v1/assets/"+asset.ID+"/snapshots", doc, nil, cookies)
	var snap struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&snap)
	rec = doJSON(t, router, http.MethodPost, "/api/v1/assessments", runAssessmentRequest{SnapshotID: snap.ID}, cookies)
	var a struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&a)

	rec = doJSON(t, router, http.MethodGet, "/api/v1/assessments/"+a.ID+"/remediation", nil, cookies)
	var plan struct {
		Filename  string           `json:"filename"`
		Automated []map[string]any `json:"automated"`
		Manual    []map[string]any `json:"manual"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&plan)
	if rec.Code != http.StatusOK || len(plan.Automated) == 0 || len(plan.Manual) == 0 || !strings.HasSuffix(plan.Filename, ".ps1") {
		t.Fatalf("plan status %d = %+v", rec.Code, plan)
	}

	rec = doJSON(t, router, http.MethodGet, "/api/v1/assessments/"+a.ID+"/remediation.ps1", nil, cookies)
	if rec.Code != http.StatusOK {
		t.Fatalf("script status = %d", rec.Code)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, plan.Filename) {
		t.Errorf("Content-Disposition = %q", cd)
	}
	if body := rec.Body.String(); !strings.Contains(body, "[switch]$Apply") || !strings.Contains(body, a.ID) {
		t.Error("script body is not the bundle for this assessment")
	}

	if rec := doJSON(t, router, http.MethodGet, "/api/v1/assessments/"+uuid.NewString()+"/remediation.ps1", nil, cookies); rec.Code != http.StatusNotFound {
		t.Errorf("unknown assessment: status %d, want 404", rec.Code)
	}
}
