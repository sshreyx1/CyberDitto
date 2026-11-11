package http

import (
	"encoding/json"
	"net/http"
	"slices"

	appiam "github.com/sshreyx1/cyberditto/services/api/internal/app/iam"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

// writeProblem writes an RFC 9457 problem+json error body
// (docs/MASTER_PLAN.md §5.15).
func writeProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":   "about:blank",
		"title":  title,
		"status": status,
		"detail": detail,
	})
}

var stateChangingMethods = []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

// csrfOriginCheck rejects cross-origin state-changing requests
// (docs/MASTER_PLAN.md §6.3: "SameSite + Origin/Referer check"). SameSite=Lax
// on the session cookie already blocks most cross-site form/script
// submissions from carrying it; this is defence in depth for the cases it
// doesn't cover (e.g. a same-site Lax exemption window, or a misconfigured
// subdomain). Requests with no Origin header (same-origin fetches in some
// browsers, non-browser API clients) are allowed through — they can't be
// forged by a browser-hosted attacker page in the first place.
func csrfOriginCheck(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if slices.Contains(stateChangingMethods, r.Method) {
				origin := r.Header.Get("Origin")
				if origin != "" && !slices.Contains(allowedOrigins, origin) {
					writeProblem(w, http.StatusForbidden, "cross-origin request rejected", "Origin header did not match an allowed origin.")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// sessionAuth resolves the session cookie to an iam.AuthContext and attaches
// it to the request context. It does not itself enforce authentication —
// requireAuth/requireCapability do — so public routes (health, setup-check)
// can share the same middleware chain.
func sessionAuth(svc *appiam.Service, secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := readSessionCookie(r, secure)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			ac, err := svc.Authenticate(r.Context(), token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r.WithContext(withAuthContext(r.Context(), ac)))
		})
	}
}

// requireAuth rejects a request that sessionAuth could not attach an
// AuthContext to.
func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authFromContext(r.Context()) == nil {
			writeProblem(w, http.StatusUnauthorized, "authentication required", "No valid session.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireCapability is the single authZ checkpoint every mutating/sensitive
// route goes through (docs/MASTER_PLAN.md §6.2: "check on every handler via
// a single policy function"). It implies requireAuth.
func requireCapability(cap iam.Capability) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ac := authFromContext(r.Context())
			if !iam.Can(ac.Membership, cap) {
				writeProblem(w, http.StatusForbidden, "forbidden", "Your role does not grant this capability.")
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}
