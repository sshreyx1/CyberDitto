package http

import (
	"errors"
	"net/http"

	appgaps "github.com/sshreyx1/cyberditto/services/api/internal/app/gaps"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

type gapHandlers struct {
	svc *appgaps.Service
}

func (h *gapHandlers) forAssessment(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	res, err := h.svc.ForAssessment(r.Context(), ac.Membership.WorkspaceID, id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "assessment not found", "")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "request failed", "")
		return
	}

	gaps := make([]map[string]any, 0, len(res.Gaps))
	for _, g := range res.Gaps {
		techniques := make([]map[string]any, 0, len(g.Techniques))
		for _, te := range g.Techniques {
			obs := make([]map[string]any, 0, len(te.Observations))
			for _, o := range te.Observations {
				obs = append(obs, map[string]any{"run_id": o.RunID, "test_name": o.TestName, "outcome": o.Outcome})
			}
			techniques = append(techniques, map[string]any{"technique_id": te.TechniqueID, "evidence": te.State, "observations": obs})
		}
		factors := make([]map[string]any, 0, len(g.Factors))
		for _, f := range g.Factors {
			factors = append(factors, map[string]any{"name": f.Name, "value": f.Value, "explanation": f.Explanation})
		}
		gaps = append(gaps, map[string]any{
			"control_id": g.ControlID, "title": g.Title, "category": g.Category, "severity": g.Severity,
			"reason": g.Reason, "remediation": g.Remediation, "priority": g.Priority,
			"factors": factors, "techniques": techniques,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"assessment_id": res.Assessment.ID,
		"snapshot_id":   res.Assessment.SnapshotID,
		"evidence_base": map[string]any{
			"twins": res.TwinsUsed, "runs": res.RunsUsed, "observations": res.Observations,
		},
		"formula": "priority = severity × evidence × prevalence × exposure (docs/MASTER_PLAN.md §4.4)",
		"gaps":    gaps,
	})
}
