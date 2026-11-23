package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_AllowsThenBlocksWithinWindow(t *testing.T) {
	rl := newRateLimiter(3, time.Minute)
	fixed := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rl.now = func() time.Time { return fixed }

	for i := 0; i < 3; i++ {
		if !rl.allow("1.2.3.4") {
			t.Fatalf("attempt %d: expected allow", i)
		}
	}
	if rl.allow("1.2.3.4") {
		t.Fatal("4th attempt within the window should be blocked")
	}
	// A different key has its own budget.
	if !rl.allow("5.6.7.8") {
		t.Fatal("a different key must not share the exhausted key's budget")
	}
}

func TestRateLimiter_ResetsAfterWindow(t *testing.T) {
	rl := newRateLimiter(1, time.Minute)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rl.now = func() time.Time { return now }

	if !rl.allow("1.2.3.4") {
		t.Fatal("first attempt should be allowed")
	}
	if rl.allow("1.2.3.4") {
		t.Fatal("second attempt before the window elapses should be blocked")
	}

	now = now.Add(time.Minute + time.Second)
	rl.now = func() time.Time { return now }
	if !rl.allow("1.2.3.4") {
		t.Fatal("attempt after the window elapses should be allowed again")
	}
}

func TestAuthEndpoints_RateLimitedPerCaller(t *testing.T) {
	router := NewRouter(testDeps())

	var last *httptest.ResponseRecorder
	for i := 0; i < authRateLimitAttempts; i++ {
		last = doJSON(t, router, http.MethodPost, "/api/v1/auth/login", loginRequest{Email: "nobody@example.com", Password: "wrong"}, nil)
		if last.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: status = %d, want %d", i, last.Code, http.StatusUnauthorized)
		}
	}

	limited := doJSON(t, router, http.MethodPost, "/api/v1/auth/login", loginRequest{Email: "nobody@example.com", Password: "wrong"}, nil)
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("status after exceeding the limit = %d, want %d", limited.Code, http.StatusTooManyRequests)
	}
	if limited.Header().Get("Retry-After") == "" {
		t.Error("429 response is missing a Retry-After header")
	}

	// /auth/setup shares the same per-IP budget as /auth/login, so it is
	// also blocked now — an attacker can't dodge the limiter by alternating
	// between the two sensitive endpoints.
	setupLimited := doJSON(t, router, http.MethodPost, "/api/v1/auth/setup", setupRequest{Email: "owner@example.com", Password: "correct horse battery staple"}, nil)
	if setupLimited.Code != http.StatusTooManyRequests {
		t.Fatalf("setup status after exhausting the shared budget = %d, want %d", setupLimited.Code, http.StatusTooManyRequests)
	}
}
