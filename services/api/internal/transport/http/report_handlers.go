package http

import (
	"errors"
	"net/http"

	appreports "github.com/sshreyx1/cyberditto/services/api/internal/app/reports"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/report"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

type reportHandlers struct {
	svc *appreports.Service
}

// export serves GET /api/v1/assessments/{id}/report?format=json|sarif|navigator
// as a download.
func (h *reportHandlers) export(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	raw := r.URL.Query().Get("format")
	if raw == "" {
		raw = string(report.FormatJSON)
	}
	format, ok := report.ParseFormat(raw)
	if !ok {
		writeProblem(w, http.StatusBadRequest, "unsupported format", "format must be json, sarif or navigator")
		return
	}
	exp, err := h.svc.Export(r.Context(), ac.User.ID, ac.Membership.WorkspaceID, id, format)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			writeProblem(w, http.StatusNotFound, "assessment not found", "")
			return
		}
		writeProblem(w, http.StatusInternalServerError, "request failed", "")
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+exp.Filename+`"`)
	writeJSON(w, http.StatusOK, exp.Body)
}
