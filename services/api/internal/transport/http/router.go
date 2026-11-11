// Package http wires HTTP transport adapters to application use cases.
// It must not contain business logic (see docs/MASTER_PLAN.md §5.3).
package http

import (
	"encoding/json"
	"net/http"

	appiam "github.com/sshreyx1/cyberditto/services/api/internal/app/iam"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain"
)

// maxRequestBodyBytes caps JSON request bodies (docs/MASTER_PLAN.md §6.3:
// "request size/time limits"). Large payloads (snapshot uploads) get their
// own, higher limit once that endpoint exists.
const maxRequestBodyBytes = 1 << 20 // 1 MiB

// Deps are the application services the HTTP transport depends on. Deps
// hold ports/app-layer types only — no adapters are constructed here
// (docs/MASTER_PLAN.md §5.1: transport depends on app, not on adapters).
type Deps struct {
	IAM            *appiam.Service
	Secure         bool // true in prod (HTTPS); controls cookie Secure/__Host- (see cookies.go)
	AllowedOrigins []string
}

// NewRouter builds the control-plane's HTTP routes.
func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)

	auth := &authHandlers{svc: deps.IAM, secure: deps.Secure}
	mux.HandleFunc("GET /api/v1/auth/setup-status", handleSetupStatus(deps.IAM))
	mux.HandleFunc("POST /api/v1/auth/setup", auth.setup)
	mux.HandleFunc("POST /api/v1/auth/login", auth.login)
	mux.HandleFunc("POST /api/v1/auth/logout", auth.logout)
	mux.Handle("GET /api/v1/me", requireAuth(http.HandlerFunc(auth.me)))

	var handler http.Handler = mux
	handler = sessionAuth(deps.IAM, deps.Secure)(handler)
	handler = csrfOriginCheck(deps.AllowedOrigins)(handler)
	handler = limitRequestBody(handler)
	handler = securityHeaders(handler)
	return handler
}

func limitRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
		next.ServeHTTP(w, r)
	})
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(domain.NewHealth())
}

// handleSetupStatus lets the web app decide whether to show the first-run
// wizard or the login page, without needing a session.
func handleSetupStatus(svc *appiam.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		initialized, err := svc.IsInitialized(r.Context())
		if err != nil {
			writeProblem(w, http.StatusInternalServerError, "setup status check failed", "")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"initialized": initialized})
	}
}
