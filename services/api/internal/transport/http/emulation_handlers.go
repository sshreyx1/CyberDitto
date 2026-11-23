package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	appemulation "github.com/sshreyx1/cyberditto/services/api/internal/app/emulation"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

type emulationHandlers struct {
	svc *appemulation.Service
}

func (h *emulationHandlers) catalog(w http.ResponseWriter, r *http.Request) {
	c := h.svc.Catalog()
	ok, detail := c.Verified()
	writeJSON(w, http.StatusOK, map[string]any{
		"source":                  c.Source(),
		"repository":              c.Repository(),
		"commit":                  c.Commit(),
		"commit_date":             c.CommitDate(),
		"license":                 c.License(),
		"verified":                ok,
		"verification_detail":     detail,
		"authorisation_statement": emulation.AuthorisationStatement,
		"limits": map[string]any{
			"max_tests_per_run": emulation.MaxTestsPerRun,
			"max_run_seconds":   emulation.MaxRunDuration.Seconds(),
			"max_test_seconds":  emulation.MaxTestDuration.Seconds(),
		},
		"tests":     c.Tests(),
		"scenarios": c.Scenarios(),
	})
}

func runResponse(run emulation.Run) map[string]any {
	execs := make([]map[string]any, 0, len(run.Executions))
	for _, e := range run.Executions {
		execs = append(execs, map[string]any{
			"test_guid": e.TestGUID, "technique_id": e.TechniqueID, "test_name": e.TestName,
			"risk_tier": e.RiskTier, "outcome": e.Outcome, "detail": e.Detail,
			"started_at": e.StartedAt, "finished_at": e.FinishedAt,
		})
	}
	preflight := run.Preflight
	if preflight == nil {
		preflight = []emulation.Check{}
	}
	return map[string]any{
		"id":                      run.ID,
		"twin_id":                 run.TwinID,
		"scenario_id":             run.ScenarioID,
		"test_guids":              run.TestGUIDs,
		"status":                  run.Status,
		"detail":                  run.Detail,
		"preflight":               preflight,
		"catalog_source":          run.CatalogSource,
		"catalog_commit":          run.CatalogCommit,
		"provider":                run.Provider,
		"authorised_by":           run.AuthorisedBy,
		"authorisation_statement": run.AuthorisationStatement,
		"executions":              execs,
		"created_at":              run.CreatedAt,
		"finished_at":             run.FinishedAt,
	}
}

type startRunRequest struct {
	TwinID                    string   `json:"twin_id"`
	ScenarioID                string   `json:"scenario_id,omitempty"`
	TestGUIDs                 []string `json:"test_guids,omitempty"`
	AuthorisationAcknowledged bool     `json:"authorisation_acknowledged"`
}

func (h *emulationHandlers) start(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	var req startRunRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}
	twinID, err := uuid.Parse(req.TwinID)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid twin_id", "twin_id is required and must be a UUID")
		return
	}

	run, err := h.svc.Start(r.Context(), ac.User.ID, ac.Membership.WorkspaceID,
		iam.Can(ac.Membership, iam.CapRunEmulation),
		appemulation.Request{TwinID: twinID, ScenarioID: req.ScenarioID, TestGUIDs: req.TestGUIDs, Acknowledged: req.AuthorisationAcknowledged})
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, runResponse(run))
	case errors.Is(err, appemulation.ErrPreflightFailed):
		// The refused run is recorded (audit trail) and returned so the
		// console can show exactly which checks failed.
		writeJSON(w, http.StatusUnprocessableEntity, runResponse(run))
	case errors.Is(err, appemulation.ErrUnknownScenario):
		writeProblem(w, http.StatusBadRequest, "unknown scenario", "")
	case errors.Is(err, ports.ErrNotFound):
		writeProblem(w, http.StatusUnprocessableEntity, "twin not found", "runs can only target a twin in this workspace")
	default:
		writeProblem(w, http.StatusInternalServerError, "could not start the run", "")
	}
}

func (h *emulationHandlers) list(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	runs, err := h.svc.List(r.Context(), ac.Membership.WorkspaceID)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "list runs failed", "")
		return
	}
	out := make([]map[string]any, 0, len(runs))
	for _, run := range runs {
		out = append(out, runResponse(run))
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": out})
}

func (h *emulationHandlers) get(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	run, err := h.svc.Get(r.Context(), ac.Membership.WorkspaceID, id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "run not found", "")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "request failed", "")
		return
	}
	writeJSON(w, http.StatusOK, runResponse(run))
}

// abort is the kill switch.
func (h *emulationHandlers) abort(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	run, err := h.svc.Abort(r.Context(), ac.User.ID, ac.Membership.WorkspaceID, id)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, runResponse(run))
	case errors.Is(err, appemulation.ErrNotRunning):
		writeProblem(w, http.StatusConflict, "run is not running", "status "+string(run.Status))
	case errors.Is(err, ports.ErrNotFound):
		writeProblem(w, http.StatusNotFound, "run not found", "")
	default:
		writeProblem(w, http.StatusInternalServerError, "abort failed", "")
	}
}
