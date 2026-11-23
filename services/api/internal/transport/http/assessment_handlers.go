package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	appassessments "github.com/sshreyx1/cyberditto/services/api/internal/app/assessments"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

type assessmentHandlers struct {
	svc *appassessments.Service
}

func assessmentResponse(a compliance.Assessment, withFindings bool) map[string]any {
	out := map[string]any{
		"id":             a.ID,
		"snapshot_id":    a.SnapshotID,
		"twin_id":        a.TwinID,
		"engine":         a.Engine,
		"engine_version": a.EngineVersion,
		"score":          a.Score,
		"counts": map[string]int{
			"passed": a.Counts.Passed, "failed": a.Counts.Failed,
			"unknown": a.Counts.Unknown, "not_applicable": a.Counts.NotApplicable,
		},
		"created_at": a.CreatedAt,
	}
	if withFindings {
		findings := make([]map[string]any, 0, len(a.Findings))
		for _, f := range a.Findings {
			techniques := f.Techniques
			if techniques == nil {
				techniques = []string{}
			}
			findings = append(findings, map[string]any{
				"control_id":  f.ControlID,
				"title":       f.Title,
				"category":    f.Category,
				"severity":    f.Severity,
				"status":      f.Status,
				"observed":    f.Observed,
				"expected":    f.Expected,
				"reason":      f.Reason,
				"remediation": f.Remediation,
				"techniques":  techniques,
			})
		}
		out["findings"] = findings
	}
	return out
}

type runAssessmentRequest struct {
	SnapshotID string `json:"snapshot_id,omitempty"`
	TwinID     string `json:"twin_id,omitempty"`
}

func optionalUUID(s string) (*uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (h *assessmentHandlers) run(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	var req runAssessmentRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}
	snapshotID, err1 := optionalUUID(req.SnapshotID)
	twinID, err2 := optionalUUID(req.TwinID)
	if err1 != nil || err2 != nil {
		writeProblem(w, http.StatusBadRequest, "invalid id", "snapshot_id and twin_id must be UUIDs")
		return
	}

	a, err := h.svc.Run(r.Context(), ac.User.ID, ac.Membership.WorkspaceID, appassessments.Target{SnapshotID: snapshotID, TwinID: twinID})
	if err != nil {
		switch {
		case errors.Is(err, appassessments.ErrInvalidTarget):
			writeProblem(w, http.StatusBadRequest, "invalid target", err.Error())
		case errors.Is(err, appassessments.ErrTwinNotReady):
			writeProblem(w, http.StatusConflict, "twin not ready", err.Error())
		case errors.Is(err, ports.ErrNotFound):
			writeProblem(w, http.StatusUnprocessableEntity, "target not found", "the snapshot or twin does not exist in this workspace")
		default:
			writeProblem(w, http.StatusInternalServerError, "assessment failed", "")
		}
		return
	}
	writeJSON(w, http.StatusCreated, assessmentResponse(a, true))
}

func (h *assessmentHandlers) list(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	as, err := h.svc.List(r.Context(), ac.Membership.WorkspaceID)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "list assessments failed", "")
		return
	}
	out := make([]map[string]any, 0, len(as))
	for _, a := range as {
		out = append(out, assessmentResponse(a, false))
	}
	writeJSON(w, http.StatusOK, map[string]any{"assessments": out})
}

func (h *assessmentHandlers) get(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	a, err := h.svc.Get(r.Context(), ac.Membership.WorkspaceID, id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "assessment not found", "")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "request failed", "")
		return
	}
	writeJSON(w, http.StatusOK, assessmentResponse(a, true))
}

func (h *assessmentHandlers) listForAsset(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	as, err := h.svc.ListForAsset(r.Context(), ac.Membership.WorkspaceID, id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "asset not found", "")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "request failed", "")
		return
	}
	out := make([]map[string]any, 0, len(as))
	for _, a := range as {
		out = append(out, assessmentResponse(a, false))
	}
	writeJSON(w, http.StatusOK, map[string]any{"assessments": out})
}

// diff compares assessment {id} (the later one) against ?against= (the
// earlier one).
func (h *assessmentHandlers) diff(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	afterID, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	beforeID, err := uuid.Parse(r.URL.Query().Get("against"))
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid against", "?against= must be the earlier assessment's UUID")
		return
	}
	before, after, changes, err := h.svc.Diff(r.Context(), ac.Membership.WorkspaceID, beforeID, afterID)
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrNotFound):
			writeProblem(w, http.StatusNotFound, "assessment not found", "")
		case errors.Is(err, appassessments.ErrEngineMismatch):
			writeProblem(w, http.StatusConflict, "not comparable", err.Error())
		default:
			writeProblem(w, http.StatusInternalServerError, "request failed", "")
		}
		return
	}
	out := make([]map[string]any, 0, len(changes))
	for _, c := range changes {
		out = append(out, map[string]any{
			"control_id": c.ControlID, "title": c.Title, "severity": c.Severity,
			"before": c.Before, "after": c.After, "direction": c.Direction(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"before":  assessmentResponse(before, false),
		"after":   assessmentResponse(after, false),
		"changes": out,
	})
}

// controls lists the catalog: what each control checks, its weight, and
// the ATT&CK techniques it relates to.
func (h *assessmentHandlers) controls(w http.ResponseWriter, r *http.Request) {
	catalog := h.svc.Catalog()
	out := make([]map[string]any, 0, len(catalog))
	for _, c := range catalog {
		out = append(out, map[string]any{
			"id":          c.ID,
			"title":       c.Title,
			"category":    c.Category,
			"severity":    c.Severity,
			"weight":      c.Severity.Weight(),
			"rationale":   c.Rationale,
			"remediation": c.Remediation,
			"techniques":  c.Techniques,
			"reads":       c.Reads,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"engine":         compliance.EngineName,
		"engine_version": compliance.EngineVersion,
		"controls":       out,
	})
}
