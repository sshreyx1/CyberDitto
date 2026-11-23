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

func TestReports_Export(t *testing.T) {
	// Viewer access is covered by the authZ matrix; the fixture needs an
	// owner because viewers can't create assets.
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

	cases := []struct {
		query, suffix, marker string
	}{
		{"", "-report.json", `"report":"cyberditto-assessment"`},
		{"?format=json", "-report.json", `"asset":"weak box"`},
		{"?format=sarif", ".sarif", `"version":"2.1.0"`},
		{"?format=navigator", "-attack-layer.json", `"domain":"enterprise-attack"`},
	}
	for _, c := range cases {
		rec := doJSON(t, router, http.MethodGet, "/api/v1/assessments/"+a.ID+"/report"+c.query, nil, cookies)
		if rec.Code != http.StatusOK {
			t.Fatalf("%q: status = %d, body = %s", c.query, rec.Code, rec.Body.String())
		}
		cd := rec.Header().Get("Content-Disposition")
		if !strings.Contains(cd, "attachment") || !strings.Contains(cd, c.suffix) {
			t.Errorf("%q: Content-Disposition = %q", c.query, cd)
		}
		if body := rec.Body.String(); !strings.Contains(body, c.marker) {
			t.Errorf("%q: body is missing %s", c.query, c.marker)
		}
	}

	if rec := doJSON(t, router, http.MethodGet, "/api/v1/assessments/"+a.ID+"/report?format=pdf", nil, cookies); rec.Code != http.StatusBadRequest {
		t.Errorf("format=pdf: status %d, want 400", rec.Code)
	}
	if rec := doJSON(t, router, http.MethodGet, "/api/v1/assessments/"+uuid.NewString()+"/report", nil, cookies); rec.Code != http.StatusNotFound {
		t.Errorf("unknown assessment: status %d, want 404", rec.Code)
	}
}
