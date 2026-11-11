package http

import (
	"encoding/json"
	"errors"
	"net/http"

	appiam "github.com/sshreyx1/cyberditto/services/api/internal/app/iam"
)

type authHandlers struct {
	svc    *appiam.Service
	secure bool
}

type setupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *authHandlers) setup(w http.ResponseWriter, r *http.Request) {
	var req setupRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	result, err := h.svc.Setup(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, appiam.ErrAlreadyInitialized):
			writeProblem(w, http.StatusConflict, "already initialized", "An owner account already exists.")
		case errors.Is(err, appiam.ErrWeakPassword):
			writeProblem(w, http.StatusBadRequest, "weak password", err.Error())
		default:
			writeProblem(w, http.StatusInternalServerError, "setup failed", "")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"workspace_id": result.Workspace.ID,
		"user_id":      result.User.ID,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *authHandlers) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	result, err := h.svc.Login(r.Context(), req.Email, req.Password, r.UserAgent())
	if err != nil {
		// Deliberately the same response whether the email doesn't exist
		// or the password is wrong (docs/MASTER_PLAN.md §6.2: no
		// enumeration). Rate limiting on this route is tracked in
		// docs/PROGRESS.md as a near-term follow-up.
		writeProblem(w, http.StatusUnauthorized, "invalid credentials", "")
		return
	}

	setSessionCookie(w, result.Token, result.ExpiresAt, h.secure)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user_id": result.User.ID,
		"email":   result.User.Email,
		"role":    result.Membership.Role,
	})
}

func (h *authHandlers) logout(w http.ResponseWriter, r *http.Request) {
	if token, ok := readSessionCookie(r, h.secure); ok {
		_ = h.svc.Logout(r.Context(), token)
	}
	clearSessionCookie(w, h.secure)
	w.WriteHeader(http.StatusNoContent)
}

func (h *authHandlers) me(w http.ResponseWriter, r *http.Request) {
	ac := authFromContext(r.Context())
	if ac == nil {
		writeProblem(w, http.StatusUnauthorized, "authentication required", "No valid session.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user_id":      ac.User.ID,
		"email":        ac.User.Email,
		"workspace_id": ac.Membership.WorkspaceID,
		"role":         ac.Membership.Role,
		"approver":     ac.Membership.Approver,
	})
}
