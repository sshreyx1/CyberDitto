package http

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// rateLimiter is a simple in-memory sliding-window limiter, keyed by caller
// IP. It is intentionally not distributed: Phase 1 runs a single
// control-plane instance (docs/MASTER_PLAN.md §5.1, "modular monolith
// first"), so there is nothing to coordinate with yet. If/when the control
// plane is ever run with more than one replica, this must move to a shared
// store (Postgres/River already in the stack) before it's load-bearing —
// tracked in docs/PROGRESS.md, not forgotten.
//
// docs/MASTER_PLAN.md §6.3 calls for rate limiting on auth endpoints
// specifically because they are the one place an unauthenticated caller can
// make the server do repeated, meaningful work (a password hash comparison)
// — this is what closes that gap.
type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
	now      func() time.Time
	// sweepEvery bounds how often allow() also sweeps the whole map for
	// keys that have gone fully idle, so long-lived processes don't
	// accumulate one slice per IP ever seen, forever.
	sweepEvery int
	calls      int
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		attempts:   map[string][]time.Time{},
		limit:      limit,
		window:     window,
		now:        time.Now,
		sweepEvery: 1000,
	}
}

// allow reports whether key may make another attempt right now, and records
// the attempt if so. It always prunes key's own history to the current
// window first, so a caller that stops retrying stops counting against
// itself immediately rather than after some global sweep.
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	cutoff := now.Add(-rl.window)
	kept := pruneBefore(rl.attempts[key], cutoff)

	allowed := len(kept) < rl.limit
	if allowed {
		kept = append(kept, now)
	}
	rl.attempts[key] = kept

	rl.calls++
	if rl.calls%rl.sweepEvery == 0 {
		rl.sweepLocked(cutoff)
	}
	return allowed
}

func (rl *rateLimiter) sweepLocked(cutoff time.Time) {
	for k, times := range rl.attempts {
		kept := pruneBefore(times, cutoff)
		if len(kept) == 0 {
			delete(rl.attempts, k)
		} else {
			rl.attempts[k] = kept
		}
	}
}

func pruneBefore(times []time.Time, cutoff time.Time) []time.Time {
	kept := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	return kept
}

// clientKey extracts the caller's IP for rate-limiting purposes.
// Deliberately uses only RemoteAddr, never X-Forwarded-For: this process is
// not told it sits behind a trusted proxy (no such config exists yet), and
// trusting a client-supplied header for rate-limit keys would let any caller
// claim a fresh identity on every request and bypass the limiter entirely.
// A reverse-proxy deployment needs to either forward the real source IP at
// the TCP layer or this must be revisited alongside that deployment's config.
func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// rateLimit wraps next so repeated requests from the same caller beyond the
// limiter's budget get a 429 with Retry-After, instead of reaching next at
// all (docs/MASTER_PLAN.md §6.3).
func rateLimit(rl *rateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientKey(r)) {
			w.Header().Set("Retry-After", strconv.Itoa(int(rl.window.Seconds())))
			writeProblem(w, http.StatusTooManyRequests, "too many requests", "Rate limit exceeded for this endpoint. Try again later.")
			return
		}
		next(w, r)
	}
}
