package http

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

func TestGaps_ForAssessment(t *testing.T) {
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

	// A twin of this snapshot plus a completed fake run: the run executes
	// nothing, so it must contribute no evidence.
	rec = doJSON(t, router, http.MethodPost, "/api/v1/twins", createTwinRequest{SnapshotID: snap.ID}, cookies)
	var tw struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&tw)
	pollTwinState(t, router, cookies, tw.ID, "ready")
	rec = doJSON(t, router, http.MethodPost, "/api/v1/runs", startRunRequest{TwinID: tw.ID, ScenarioID: "credential-access", AuthorisationAcknowledged: true}, cookies)
	if rec.Code != http.StatusCreated {
		t.Fatalf("run status = %d", rec.Code)
	}

	rec = doJSON(t, router, http.MethodPost, "/api/v1/assessments", runAssessmentRequest{SnapshotID: snap.ID}, cookies)
	var a struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&a)

	// Let the fake run finish so it's part of the evidence base.
	var run struct {
		Status string `json:"status"`
	}
	for i := 0; i < 200 && run.Status != "completed"; i++ {
		rec = doJSON(t, router, http.MethodGet, "/api/v1/runs", nil, cookies)
		var list struct {
			Runs []struct {
				Status string `json:"status"`
			} `json:"runs"`
		}
		_ = json.NewDecoder(rec.Body).Decode(&list)
		if len(list.Runs) > 0 {
			run.Status = list.Runs[0].Status
		}
	}

	rec = doJSON(t, router, http.MethodGet, "/api/v1/assessments/"+a.ID+"/gaps", nil, cookies)
	if rec.Code != http.StatusOK {
		t.Fatalf("gaps status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body struct {
		EvidenceBase struct {
			Twins        []string `json:"twins"`
			Runs         int      `json:"runs"`
			Observations int      `json:"observations"`
		} `json:"evidence_base"`
		Gaps []struct {
			ControlID  string  `json:"control_id"`
			Priority   float64 `json:"priority"`
			Techniques []struct {
				Evidence string `json:"evidence"`
			} `json:"techniques"`
			Factors []struct {
				Name  string  `json:"name"`
				Value float64 `json:"value"`
			} `json:"factors"`
		} `json:"gaps"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&body)

	if len(body.Gaps) != 28 {
		t.Fatalf("gaps = %d, want 28 (every control fails on the weak fixture)", len(body.Gaps))
	}
	if len(body.EvidenceBase.Twins) != 1 || body.EvidenceBase.Runs != 1 || body.EvidenceBase.Observations == 0 {
		t.Errorf("evidence base = %+v, want the one twin and its one run", body.EvidenceBase)
	}
	for _, g := range body.Gaps {
		for _, te := range g.Techniques {
			if te.Evidence != "untested" {
				t.Errorf("%s: fake-provider run produced evidence %q", g.ControlID, te.Evidence)
			}
		}
		if g.Factors[1].Name != "evidence" || g.Factors[1].Value != 0.3 {
			t.Errorf("%s evidence factor = %+v", g.ControlID, g.Factors[1])
		}
	}
	if body.Gaps[0].Priority < body.Gaps[len(body.Gaps)-1].Priority {
		t.Error("gaps are not sorted by priority")
	}

	if rec := doJSON(t, router, http.MethodGet, "/api/v1/assessments/"+uuid.NewString()+"/gaps", nil, cookies); rec.Code != http.StatusNotFound {
		t.Errorf("unknown assessment: status %d, want 404", rec.Code)
	}
}
