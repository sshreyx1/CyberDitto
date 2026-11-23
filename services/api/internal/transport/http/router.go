// Package http wires HTTP transport adapters to application use cases.
// It must not contain business logic (see docs/MASTER_PLAN.md §5.3).
package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	appassessments "github.com/sshreyx1/cyberditto/services/api/internal/app/assessments"
	appassets "github.com/sshreyx1/cyberditto/services/api/internal/app/assets"
	appiam "github.com/sshreyx1/cyberditto/services/api/internal/app/iam"
	apptwins "github.com/sshreyx1/cyberditto/services/api/internal/app/twins"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

// maxRequestBodyBytes caps JSON request bodies (docs/MASTER_PLAN.md §6.3:
// "request size/time limits"). Snapshot uploads get asset.MaxDocumentBytes
// instead (see bodyLimit).
const maxRequestBodyBytes = 1 << 20 // 1 MiB

// authRateLimit bounds setup/login attempts per caller IP
// (docs/MASTER_PLAN.md §6.3, docs/PROGRESS.md's "Phase 1 next action" item).
// 10 attempts per minute is generous enough for a real user who mistypes a
// password a few times, tight enough to make online password guessing or
// setup-race attempts expensive.
const (
	authRateLimitAttempts = 10
	authRateLimitWindow   = time.Minute
)

// Deps are the application services the HTTP transport depends on. Deps
// hold ports/app-layer types only — no adapters are constructed here
// (docs/MASTER_PLAN.md §5.1: transport depends on app, not on adapters).
type Deps struct {
	IAM            *appiam.Service
	Assets         *appassets.Service
	Twins          *apptwins.Service
	Assessments    *appassessments.Service
	Secure         bool // true in prod (HTTPS); controls cookie Secure/__Host- (see cookies.go)
	AllowedOrigins []string
}

// NewRouter builds the control-plane's HTTP routes.
func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)

	auth := &authHandlers{svc: deps.IAM, secure: deps.Secure}
	authLimiter := newRateLimiter(authRateLimitAttempts, authRateLimitWindow)
	mux.HandleFunc("GET /api/v1/auth/setup-status", handleSetupStatus(deps.IAM))
	mux.HandleFunc("POST /api/v1/auth/setup", rateLimit(authLimiter, auth.setup))
	mux.HandleFunc("POST /api/v1/auth/login", rateLimit(authLimiter, auth.login))
	mux.HandleFunc("POST /api/v1/auth/logout", auth.logout)
	mux.Handle("GET /api/v1/me", requireAuth(http.HandlerFunc(auth.me)))

	as := &assetHandlers{svc: deps.Assets}
	mux.Handle("POST /api/v1/assets", requireCapability(iam.CapManageAssets)(http.HandlerFunc(as.create)))
	mux.Handle("GET /api/v1/assets", requireCapability(iam.CapViewAssets)(http.HandlerFunc(as.list)))
	mux.Handle("GET /api/v1/assets/{id}", requireCapability(iam.CapViewAssets)(http.HandlerFunc(as.get)))
	mux.Handle("POST /api/v1/assets/{id}/snapshots", requireCapability(iam.CapManageAssets)(http.HandlerFunc(as.uploadSnapshot)))
	mux.Handle("GET /api/v1/assets/{id}/snapshots", requireCapability(iam.CapViewAssets)(http.HandlerFunc(as.listSnapshots)))
	mux.Handle("GET /api/v1/snapshots/{id}", requireCapability(iam.CapViewAssets)(http.HandlerFunc(as.getSnapshot)))

	ah := &assessmentHandlers{svc: deps.Assessments}
	mux.Handle("POST /api/v1/assessments", requireCapability(iam.CapRunAssessment)(http.HandlerFunc(ah.run)))
	mux.Handle("GET /api/v1/assessments", requireCapability(iam.CapViewAssessments)(http.HandlerFunc(ah.list)))
	mux.Handle("GET /api/v1/assessments/{id}", requireCapability(iam.CapViewAssessments)(http.HandlerFunc(ah.get)))
	mux.Handle("GET /api/v1/controls", requireCapability(iam.CapViewAssessments)(http.HandlerFunc(ah.controls)))

	tw := &twinHandlers{svc: deps.Twins}
	mux.Handle("POST /api/v1/twins", requireCapability(iam.CapManageTwins)(http.HandlerFunc(tw.create)))
	mux.Handle("GET /api/v1/twins", requireCapability(iam.CapViewTwins)(http.HandlerFunc(tw.list)))
	mux.Handle("GET /api/v1/twins/{id}", requireCapability(iam.CapViewTwins)(http.HandlerFunc(tw.get)))
	mux.Handle("GET /api/v1/twins/{id}/events", requireCapability(iam.CapViewTwins)(http.HandlerFunc(tw.events)))
	mux.Handle("POST /api/v1/twins/{id}/start", requireCapability(iam.CapManageTwins)(http.HandlerFunc(tw.start)))
	mux.Handle("POST /api/v1/twins/{id}/stop", requireCapability(iam.CapManageTwins)(http.HandlerFunc(tw.stop)))
	mux.Handle("POST /api/v1/twins/{id}/destroy", requireCapability(iam.CapManageTwins)(http.HandlerFunc(tw.destroy)))

	var handler http.Handler = mux
	handler = sessionAuth(deps.IAM, deps.Secure)(handler)
	handler = csrfOriginCheck(deps.AllowedOrigins)(handler)
	handler = limitRequestBody(handler)
	handler = securityHeaders(handler)
	handler = cors(deps.AllowedOrigins)(handler)
	return handler
}

func limitRequestBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, bodyLimit(r))
		next.ServeHTTP(w, r)
	})
}

// bodyLimit is the one place a route gets a larger body cap than the
// default: snapshot uploads (POST /api/v1/assets/{id}/snapshots).
func bodyLimit(r *http.Request) int64 {
	if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/v1/assets/") && strings.HasSuffix(r.URL.Path, "/snapshots") {
		return asset.MaxDocumentBytes
	}
	return maxRequestBodyBytes
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
