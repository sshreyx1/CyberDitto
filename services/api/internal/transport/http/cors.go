package http

import (
	"net/http"
	"slices"
)

// cors answers cross-origin requests from the web app (a separate Next.js
// origin, e.g. http://localhost:3000, from the API's own origin) with an
// explicit allow-list — never a wildcard, because credentials (the session
// cookie) are involved (docs/MASTER_PLAN.md §6.3: "CORS: explicit allow-list
// from config (no wildcards with credentials)"). It must run before
// csrfOriginCheck/sessionAuth so an OPTIONS preflight (which carries no
// cookie and isn't a "real" request) gets a clean response instead of being
// treated as an unauthenticated or cross-origin request itself.
func cors(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && slices.Contains(allowedOrigins, origin) {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Set("Access-Control-Allow-Credentials", "true")
				// Vary: Origin so a cache in front of this API (CDN, shared
				// proxy) never serves one caller's CORS headers to another.
				h.Add("Vary", "Origin")
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, "+snapshotDigestHeader)
				w.Header().Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
