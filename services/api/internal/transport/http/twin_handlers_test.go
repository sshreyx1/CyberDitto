package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/argon2"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/fake"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/token"
	appiam "github.com/sshreyx1/cyberditto/services/api/internal/app/iam"
	apptwins "github.com/sshreyx1/cyberditto/services/api/internal/app/twins"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

// setupAs bootstraps a router with a single owner account, then (if role
// isn't owner, or approver is false) rewrites that membership's role
// directly in the fake repo, so tests can exercise every role against the
// real middleware chain (session auth + requireCapability) rather than
// calling iam.Can() in isolation — Phase 1 has no "manage members" handler
// yet to do this through the API itself.
func setupAs(t *testing.T, role iam.Role, approver bool) (http.Handler, []*http.Cookie) {
	t.Helper()

	iamRepo := fake.NewIAMRepo()
	iamSvc := appiam.NewService(iamRepo, fake.NewAuditRepo(), argon2.New(), token.New())
	twinsSvc := apptwins.NewService(fake.NewTwinRepo(), fake.LabProvider{StepDelay: time.Millisecond}, fake.NewAuditRepo())
	router := NewRouter(Deps{IAM: iamSvc, Twins: twinsSvc, Secure: false, AllowedOrigins: []string{"http://localhost:3000"}})

	rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/setup", setupRequest{Email: "owner@example.com", Password: "correct horse battery staple"}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup status = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, router, http.MethodPost, "/api/v1/auth/login", loginRequest{Email: "owner@example.com", Password: "correct horse battery staple"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()

	if role != iam.RoleOwner || !approver {
		var me struct {
			UserID      string `json:"user_id"`
			WorkspaceID string `json:"workspace_id"`
		}
		rec = doJSON(t, router, http.MethodGet, "/api/v1/me", nil, cookies)
		if err := json.NewDecoder(rec.Body).Decode(&me); err != nil {
			t.Fatalf("decode /me: %v", err)
		}
		userID, err := uuid.Parse(me.UserID)
		if err != nil {
			t.Fatalf("parse user id: %v", err)
		}
		workspaceID, err := uuid.Parse(me.WorkspaceID)
		if err != nil {
			t.Fatalf("parse workspace id: %v", err)
		}
		if err := iamRepo.SetMembershipRole(workspaceID, userID, role, approver); err != nil {
			t.Fatalf("SetMembershipRole: %v", err)
		}
	}

	return router, cookies
}

// TestTwinRoutes_AuthZMatrix is the HTTP-level "every twins route × every
// role" matrix (docs/MASTER_PLAN.md §6.2, §9.4) — exercised against a real
// capability-gated route for the first time, through the actual middleware
// chain rather than calling iam.Can() directly, so it would also catch a
// route wired to the wrong capability.
func TestTwinRoutes_AuthZMatrix(t *testing.T) {
	expected := map[iam.Role]map[string]bool{
		iam.RoleOwner:   {"POST /api/v1/twins": true, "GET /api/v1/twins": true},
		iam.RoleAdmin:   {"POST /api/v1/twins": true, "GET /api/v1/twins": true},
		iam.RoleAnalyst: {"POST /api/v1/twins": true, "GET /api/v1/twins": true},
		iam.RoleViewer:  {"POST /api/v1/twins": false, "GET /api/v1/twins": true},
	}

	for role, want := range expected {
		router, cookies := setupAs(t, role, true)

		rec := doJSON(t, router, http.MethodPost, "/api/v1/twins", createTwinRequest{}, cookies)
		if allowed := rec.Code != http.StatusForbidden; allowed != want["POST /api/v1/twins"] {
			t.Errorf("role=%s POST /api/v1/twins: allowed=%v (status %d), want allowed=%v", role, allowed, rec.Code, want["POST /api/v1/twins"])
		}

		rec = doJSON(t, router, http.MethodGet, "/api/v1/twins", nil, cookies)
		if allowed := rec.Code != http.StatusForbidden; allowed != want["GET /api/v1/twins"] {
			t.Errorf("role=%s GET /api/v1/twins: allowed=%v (status %d), want allowed=%v", role, allowed, rec.Code, want["GET /api/v1/twins"])
		}
	}
}

func TestTwinRoutes_RequireAuthentication(t *testing.T) {
	router, _ := setupAs(t, iam.RoleOwner, true)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/twins", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/v1/twins with no session: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTwinLifecycle_CreateListGetStartStopDestroy(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/twins", createTwinRequest{}, cookies)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created struct {
		ID    string `json:"id"`
		State string `json:"state"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.State != "requested" {
		t.Fatalf("initial state = %q, want %q", created.State, "requested")
	}

	// Starting before the fake build reaches Ready is a conflict, not a 500.
	rec = doJSON(t, router, http.MethodPost, "/api/v1/twins/"+created.ID+"/start", nil, cookies)
	if rec.Code != http.StatusConflict {
		t.Fatalf("premature start status = %d, want %d", rec.Code, http.StatusConflict)
	}

	state := pollTwinState(t, router, cookies, created.ID, "ready")
	if state != "ready" {
		t.Fatalf("twin never reached ready, last state = %q", state)
	}

	rec = doJSON(t, router, http.MethodGet, "/api/v1/twins", nil, cookies)
	var list struct {
		Twins []map[string]any `json:"twins"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&list)
	if len(list.Twins) != 1 {
		t.Fatalf("list returned %d twins, want 1", len(list.Twins))
	}

	rec = doJSON(t, router, http.MethodPost, "/api/v1/twins/"+created.ID+"/start", nil, cookies)
	if rec.Code != http.StatusOK {
		t.Fatalf("start status = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, router, http.MethodPost, "/api/v1/twins/"+created.ID+"/stop", nil, cookies)
	if rec.Code != http.StatusOK {
		t.Fatalf("stop status = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, router, http.MethodPost, "/api/v1/twins/"+created.ID+"/destroy", nil, cookies)
	if rec.Code != http.StatusOK {
		t.Fatalf("destroy status = %d, body = %s", rec.Code, rec.Body.String())
	}

	rec = doJSON(t, router, http.MethodPost, "/api/v1/twins/"+created.ID+"/start", nil, cookies)
	if rec.Code != http.StatusConflict {
		t.Fatalf("start after destroy status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func pollTwinState(t *testing.T, router http.Handler, cookies []*http.Cookie, id, want string) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var state string
	for time.Now().Before(deadline) {
		rec := doJSON(t, router, http.MethodGet, "/api/v1/twins/"+id, nil, cookies)
		if rec.Code != http.StatusOK {
			t.Fatalf("get status = %d, body = %s", rec.Code, rec.Body.String())
		}
		var got struct {
			State string `json:"state"`
		}
		_ = json.NewDecoder(rec.Body).Decode(&got)
		state = got.State
		if state == want {
			return state
		}
		time.Sleep(10 * time.Millisecond)
	}
	return state
}

func TestTwinEvents_SSEStreamsBuildProgress(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/twins", createTwinRequest{}, cookies)
	var created struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&created)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/twins/"+created.ID+"/events", nil).WithContext(ctx)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	sseRec := newSSERecorder()
	done := make(chan struct{})
	go func() {
		router.ServeHTTP(sseRec, req)
		close(done)
	}()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(sseRec.snapshot(), `"state":"ready"`) {
			cancel()
			<-done
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done
	t.Fatalf("SSE stream never reported ready; got:\n%s", sseRec.snapshot())
}

// sseRecorder is a minimal, concurrency-safe http.ResponseWriter +
// http.Flusher. httptest.ResponseRecorder's Body is a plain *bytes.Buffer,
// which the race detector correctly flags when a test goroutine reads it
// while the handler goroutine is still writing SSE events to it; this
// serializes both sides behind one mutex instead.
type sseRecorder struct {
	mu     sync.Mutex
	header http.Header
	code   int
	body   bytes.Buffer
}

func newSSERecorder() *sseRecorder {
	return &sseRecorder{header: http.Header{}}
}

func (r *sseRecorder) Header() http.Header { return r.header }

func (r *sseRecorder) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.body.Write(p)
}

func (r *sseRecorder) WriteHeader(code int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.code = code
}

func (r *sseRecorder) Flush() {}

func (r *sseRecorder) snapshot() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.body.String()
}
