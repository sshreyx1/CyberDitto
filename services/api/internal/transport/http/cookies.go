package http

import (
	"net/http"
	"time"
)

// Browsers reject a __Host- prefixed cookie outright unless Secure is set,
// so dev-over-plain-HTTP (CD_ENV=dev, no TLS) uses an unprefixed name
// instead of silently failing to set the cookie at all. Production (secure
// transport) always uses the __Host- prefix, the strongest cookie-scoping
// guarantee available — no Domain attribute, Path=/, Secure required by the
// browser itself (docs/MASTER_PLAN.md §6.2).
const (
	sessionCookieNameSecure   = "__Host-cd_session"
	sessionCookieNameInsecure = "cd_session"
)

func sessionCookieName(secure bool) string {
	if secure {
		return sessionCookieNameSecure
	}
	return sessionCookieNameInsecure
}

func setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName(secure),
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName(secure),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func readSessionCookie(r *http.Request, secure bool) (string, bool) {
	c, err := r.Cookie(sessionCookieName(secure))
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}
