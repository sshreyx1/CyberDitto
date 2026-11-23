package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

// testSnapshotDoc is a minimal valid collector document (schema v1).
const testSnapshotDoc = `{"schema_version":"1","hostname":"HOST-test","collected_at":"2026-11-23T06:00:00Z","fields":{"os.version":{"source":"os.version","collected_at":"2026-11-23T06:00:00Z","value":{"product_name":"Windows 11 Pro","current_build":"26200"}}}}`

// setupAs bootstraps a router with a single owner account, then (if role
// isn't owner, or approver is false) rewrites that membership's role in the
// fake repo, so every role can be exercised through the real middleware
// chain (there is no "manage members" handler yet to do it via the API).
func setupAs(t *testing.T, role iam.Role, approver bool) (http.Handler, []*http.Cookie) {
	t.Helper()
	deps, iamRepo := testDepsWithIAMRepo()
	router := NewRouter(deps)

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
		if err := iamRepo.SetMembershipRole(uuid.MustParse(me.WorkspaceID), uuid.MustParse(me.UserID), role, approver); err != nil {
			t.Fatalf("SetMembershipRole: %v", err)
		}
	}
	return router, cookies
}

// doRaw sends body bytes as-is (snapshot uploads are the raw collector
// file, not a JSON envelope).
func doRaw(t *testing.T, router http.Handler, method, path string, body []byte, headers map[string]string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// createAssetWithSnapshot registers an asset and uploads testSnapshotDoc to
// it through the API, returning (assetID, snapshotID).
func createAssetWithSnapshot(t *testing.T, router http.Handler, cookies []*http.Cookie) (string, string) {
	t.Helper()
	rec := doJSON(t, router, http.MethodPost, "/api/v1/assets", createAssetRequest{Name: "laptop-01"}, cookies)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create asset status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var a struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&a)

	rec = doRaw(t, router, http.MethodPost, "/api/v1/assets/"+a.ID+"/snapshots", []byte(testSnapshotDoc), nil, cookies)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload snapshot status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var s struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&s)
	return a.ID, s.ID
}

// sseRecorder is a concurrency-safe http.ResponseWriter + http.Flusher:
// httptest.ResponseRecorder's Body would race between the handler goroutine
// writing SSE events and the test goroutine reading them.
type sseRecorder struct {
	mu     sync.Mutex
	header http.Header
	code   int
	body   bytes.Buffer
}

func newSSERecorder() *sseRecorder { return &sseRecorder{header: http.Header{}} }

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
