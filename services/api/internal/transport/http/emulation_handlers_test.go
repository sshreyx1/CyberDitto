package http

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

type runBody struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Preflight []struct {
		ID     string `json:"id"`
		Passed bool   `json:"passed"`
	} `json:"preflight"`
	Executions []struct {
		TestGUID string `json:"test_guid"`
		Outcome  string `json:"outcome"`
	} `json:"executions"`
	TestGUIDs []string `json:"test_guids"`
}

func readyTwin(t *testing.T, router http.Handler, cookies []*http.Cookie) string {
	t.Helper()
	id := createTwin(t, router, cookies)
	if pollTwinState(t, router, cookies, id, "ready") != "ready" {
		t.Fatal("twin never became ready")
	}
	return id
}

func TestRuns_StartCompleteAndRecordNotExecuted(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)
	twinID := readyTwin(t, router, cookies)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/attack/catalog", nil, cookies)
	var cat struct {
		Verified bool `json:"verified"`
		Tests    []any
	}
	_ = json.NewDecoder(rec.Body).Decode(&cat)
	if !cat.Verified || len(cat.Tests) == 0 {
		t.Fatalf("catalog verified=%v tests=%d", cat.Verified, len(cat.Tests))
	}

	rec = doJSON(t, router, http.MethodPost, "/api/v1/runs", startRunRequest{TwinID: twinID, ScenarioID: "discovery", AuthorisationAcknowledged: true}, cookies)
	if rec.Code != http.StatusCreated {
		t.Fatalf("start status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var run runBody
	_ = json.NewDecoder(rec.Body).Decode(&run)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		rec = doJSON(t, router, http.MethodGet, "/api/v1/runs/"+run.ID, nil, cookies)
		_ = json.NewDecoder(rec.Body).Decode(&run)
		if run.Status != "running" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if run.Status != "completed" || len(run.Executions) != len(run.TestGUIDs) {
		t.Fatalf("run = %s with %d/%d executions", run.Status, len(run.Executions), len(run.TestGUIDs))
	}
	for _, e := range run.Executions {
		if e.Outcome != "not_executed" {
			t.Errorf("%s outcome = %s; the fake provider must never report a real-looking outcome", e.TestGUID, e.Outcome)
		}
	}

	if rec := doJSON(t, router, http.MethodPost, "/api/v1/runs/"+run.ID+"/abort", nil, cookies); rec.Code != http.StatusConflict {
		t.Errorf("abort a finished run: status %d, want 409", rec.Code)
	}
	rec = doJSON(t, router, http.MethodGet, "/api/v1/runs", nil, cookies)
	var list struct {
		Runs []runBody `json:"runs"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&list)
	if len(list.Runs) != 1 {
		t.Errorf("runs listed = %d, want 1", len(list.Runs))
	}
}

func TestRuns_PreflightRefusalIsRecordedAndExplained(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)
	twinID := readyTwin(t, router, cookies)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/runs", startRunRequest{TwinID: twinID, ScenarioID: "discovery"}, cookies)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unacknowledged run status = %d, want 422", rec.Code)
	}
	var run runBody
	_ = json.NewDecoder(rec.Body).Decode(&run)
	if run.Status != "preflight_failed" || run.ID == "" {
		t.Fatalf("refused run = %+v", run)
	}
	for _, c := range run.Preflight {
		if c.ID == "run.authorised" && c.Passed {
			t.Error("run.authorised passed without acknowledgement")
		}
	}

	if rec := doJSON(t, router, http.MethodPost, "/api/v1/runs", startRunRequest{TwinID: uuid.NewString(), ScenarioID: "discovery", AuthorisationAcknowledged: true}, cookies); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("run against a non-twin id: status %d, want 422", rec.Code)
	}
	if rec := doJSON(t, router, http.MethodPost, "/api/v1/runs", startRunRequest{TwinID: twinID, ScenarioID: "nope", AuthorisationAcknowledged: true}, cookies); rec.Code != http.StatusBadRequest {
		t.Errorf("unknown scenario: status %d, want 400", rec.Code)
	}
}

// An owner without the approver flag is refused at the RBAC layer before
// any run is created.
func TestRuns_NonApproverForbidden(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, false)
	if rec := doJSON(t, router, http.MethodPost, "/api/v1/runs", startRunRequest{TwinID: uuid.NewString(), ScenarioID: "discovery", AuthorisationAcknowledged: true}, cookies); rec.Code != http.StatusForbidden {
		t.Fatalf("non-approver start: status %d, want 403", rec.Code)
	}
}
