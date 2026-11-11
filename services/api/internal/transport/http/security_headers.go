package http

import "net/http"

// securityHeaders sets the baseline headers from docs/MASTER_PLAN.md §6.3.
// CSP intentionally has no "unsafe-inline"; the web app is a separate
// origin/process (Next.js) and does not render anything the API returns as
// HTML, so a strict default-src 'none' is correct here — this API only
// emits JSON and SSE.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("X-Frame-Options", "DENY")
		if r.TLS != nil {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}
