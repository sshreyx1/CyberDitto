package http

import (
	"errors"
	"net/http"

	appremediation "github.com/sshreyx1/cyberditto/services/api/internal/app/remediation"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

type remediationHandlers struct {
	svc *appremediation.Service
}

func (h *remediationHandlers) plan(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	b, err := h.svc.Plan(r.Context(), ac.Membership.WorkspaceID, id)
	if err != nil {
		writeRemediationError(w, err)
		return
	}
	automated := make([]map[string]any, 0, len(b.Automated))
	for _, t := range b.Automated {
		automated = append(automated, map[string]any{
			"control_id": t.ControlID, "summary": t.Summary, "reboot_required": t.RebootRequired, "risk": t.Risk,
		})
	}
	manual := make([]map[string]any, 0, len(b.Manual))
	for _, m := range b.Manual {
		manual = append(manual, map[string]any{"control_id": m.ControlID, "title": m.Title, "remediation": m.Remediation})
	}
	writeJSON(w, http.StatusOK, map[string]any{"filename": b.Filename, "automated": automated, "manual": manual})
}

// script returns the bundle itself as a download.
func (h *remediationHandlers) script(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	b, err := h.svc.Export(r.Context(), ac.User.ID, ac.Membership.WorkspaceID, id)
	if err != nil {
		writeRemediationError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+b.Filename+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(b.Script))
}

func writeRemediationError(w http.ResponseWriter, err error) {
	if errors.Is(err, ports.ErrNotFound) {
		writeProblem(w, http.StatusNotFound, "assessment not found", "")
		return
	}
	writeProblem(w, http.StatusInternalServerError, "request failed", "")
}
