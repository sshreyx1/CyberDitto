package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

type assessmentBody struct {
	ID       string   `json:"id"`
	TwinID   *string  `json:"twin_id"`
	Score    *float64 `json:"score"`
	Counts   map[string]int
	Findings []struct {
		ControlID string          `json:"control_id"`
		Status    string          `json:"status"`
		Observed  json.RawMessage `json:"observed"`
		Reason    string          `json:"reason"`
	} `json:"findings"`
}

func TestAssessments_RunOnSnapshotAndTwin(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)
	_, snapshotID := createAssetWithSnapshot(t, router, cookies)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/assessments", runAssessmentRequest{SnapshotID: snapshotID}, cookies)
	if rec.Code != http.StatusCreated {
		t.Fatalf("run status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var a assessmentBody
	_ = json.NewDecoder(rec.Body).Decode(&a)
	if len(a.Findings) != len(compliance.Catalog()) {
		t.Fatalf("findings = %d, want %d", len(a.Findings), len(compliance.Catalog()))
	}
	// testSnapshotDoc only carries os.version, so nearly everything is
	// unknown and no control may pretend otherwise.
	if a.Counts["unknown"] == 0 || a.Counts["failed"]+a.Counts["passed"] != 0 {
		t.Errorf("counts on a near-empty snapshot = %v, want everything unknown or n/a", a.Counts)
	}
	if a.Score != nil {
		t.Errorf("score = %v, want null when nothing is measurable", *a.Score)
	}

	rec = doJSON(t, router, http.MethodGet, "/api/v1/assessments/"+a.ID, nil, cookies)
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d", rec.Code)
	}
	rec = doJSON(t, router, http.MethodGet, "/api/v1/assessments", nil, cookies)
	var list struct {
		Assessments []assessmentBody `json:"assessments"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&list)
	if len(list.Assessments) != 1 || list.Assessments[0].Findings != nil {
		t.Fatalf("list = %+v, want one summary without findings", list.Assessments)
	}

	// A twin that hasn't finished building can't be assessed yet.
	rec = doJSON(t, router, http.MethodPost, "/api/v1/twins", createTwinRequest{SnapshotID: snapshotID}, cookies)
	var tw struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&tw)
	rec = doJSON(t, router, http.MethodPost, "/api/v1/assessments", runAssessmentRequest{TwinID: tw.ID}, cookies)
	if rec.Code != http.StatusConflict {
		t.Fatalf("assess building twin status = %d, want 409", rec.Code)
	}
	pollTwinState(t, router, cookies, tw.ID, "ready")
	rec = doJSON(t, router, http.MethodPost, "/api/v1/assessments", runAssessmentRequest{TwinID: tw.ID}, cookies)
	if rec.Code != http.StatusCreated {
		t.Fatalf("assess ready twin status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var ta assessmentBody
	_ = json.NewDecoder(rec.Body).Decode(&ta)
	if ta.TwinID == nil || *ta.TwinID != tw.ID {
		t.Errorf("twin assessment twin_id = %v, want %s", ta.TwinID, tw.ID)
	}
}

func TestAssessments_Validation(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)
	cases := []struct {
		body runAssessmentRequest
		want int
	}{
		{runAssessmentRequest{}, http.StatusBadRequest},
		{runAssessmentRequest{SnapshotID: uuid.NewString(), TwinID: uuid.NewString()}, http.StatusBadRequest},
		{runAssessmentRequest{SnapshotID: "nope"}, http.StatusBadRequest},
		{runAssessmentRequest{SnapshotID: uuid.NewString()}, http.StatusUnprocessableEntity},
	}
	for _, c := range cases {
		if rec := doJSON(t, router, http.MethodPost, "/api/v1/assessments", c.body, cookies); rec.Code != c.want {
			t.Errorf("body %+v: status %d, want %d", c.body, rec.Code, c.want)
		}
	}
}

func TestControls_ListsCatalog(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleViewer, false)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/controls", nil, cookies)
	var body struct {
		EngineVersion string           `json:"engine_version"`
		Controls      []map[string]any `json:"controls"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if rec.Code != http.StatusOK || len(body.Controls) != len(compliance.Catalog()) || body.EngineVersion == "" {
		t.Fatalf("controls status %d, %d controls, version %q", rec.Code, len(body.Controls), body.EngineVersion)
	}
}
