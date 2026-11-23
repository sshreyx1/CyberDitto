package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/argon2"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/fake"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/token"
	appassessments "github.com/sshreyx1/cyberditto/services/api/internal/app/assessments"
	appassets "github.com/sshreyx1/cyberditto/services/api/internal/app/assets"
	appiam "github.com/sshreyx1/cyberditto/services/api/internal/app/iam"
	apptwins "github.com/sshreyx1/cyberditto/services/api/internal/app/twins"
)

func testDeps() Deps {
	deps, _ := testDepsWithIAMRepo()
	return deps
}

// testDepsWithIAMRepo wires every service against fake adapters and also
// returns the fake IAM repo, so tests can change a membership's role.
func testDepsWithIAMRepo() (Deps, *fake.IAMRepo) {
	iamRepo := fake.NewIAMRepo()
	audit := fake.NewAuditRepo()
	assetRepo := fake.NewAssetRepo()
	twinRepo := fake.NewTwinRepo()
	return Deps{
		IAM:            appiam.NewService(iamRepo, audit, argon2.New(), token.New()),
		Assets:         appassets.NewService(assetRepo, audit),
		Twins:          apptwins.NewService(twinRepo, assetRepo, fake.LabProvider{StepDelay: time.Millisecond}, "fake", audit),
		Assessments:    appassessments.NewService(fake.NewAssessmentRepo(assetRepo), assetRepo, twinRepo, audit),
		Secure:         false,
		AllowedOrigins: []string{"http://localhost:3000"},
	}, iamRepo
}

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	NewRouter(testDeps()).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("status field = %q, want %q", body.Status, "ok")
	}
}

func doJSON(t *testing.T, router http.Handler, method, path string, body any, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAuthFlow_SetupLoginMeLogout(t *testing.T) {
	router := NewRouter(testDeps())

	// Before setup: setup-status says not initialized.
	rec := doJSON(t, router, http.MethodGet, "/api/v1/auth/setup-status", nil, nil)
	var status struct {
		Initialized bool `json:"initialized"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&status)
	if status.Initialized {
		t.Fatal("setup-status.initialized = true before any setup call")
	}

	// Setup creates the owner.
	rec = doJSON(t, router, http.MethodPost, "/api/v1/auth/setup", setupRequest{Email: "owner@example.com", Password: "correct horse battery staple"}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup status = %d, body = %s", rec.Code, rec.Body.String())
	}

	// Second setup is rejected.
	rec = doJSON(t, router, http.MethodPost, "/api/v1/auth/setup", setupRequest{Email: "other@example.com", Password: "correct horse battery staple"}, nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("second setup status = %d, want %d", rec.Code, http.StatusConflict)
	}

	// /me with no session is unauthenticated.
	rec = doJSON(t, router, http.MethodGet, "/api/v1/me", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("/me (no session) status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	// Wrong password is rejected without revealing which part was wrong.
	rec = doJSON(t, router, http.MethodPost, "/api/v1/auth/login", loginRequest{Email: "owner@example.com", Password: "wrong"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad login status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	// Correct login sets a session cookie.
	rec = doJSON(t, router, http.MethodPost, "/api/v1/auth/login", loginRequest{Email: "owner@example.com", Password: "correct horse battery staple"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", rec.Code, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("login did not set a session cookie")
	}
	for _, c := range cookies {
		if !c.HttpOnly {
			t.Errorf("session cookie %q is not HttpOnly", c.Name)
		}
		if c.SameSite != http.SameSiteLaxMode {
			t.Errorf("session cookie %q SameSite = %v, want Lax", c.Name, c.SameSite)
		}
	}

	// /me with the session cookie succeeds.
	rec = doJSON(t, router, http.MethodGet, "/api/v1/me", nil, cookies)
	if rec.Code != http.StatusOK {
		t.Fatalf("/me (with session) status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var me struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&me); err != nil {
		t.Fatal(err)
	}
	if me.Email != "owner@example.com" || me.Role != "owner" {
		t.Fatalf("/me body = %+v, want owner@example.com/owner", me)
	}

	// Logout clears the session.
	rec = doJSON(t, router, http.MethodPost, "/api/v1/auth/logout", nil, cookies)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	rec = doJSON(t, router, http.MethodGet, "/api/v1/me", nil, cookies)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("/me after logout status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestCSRF_CrossOriginStateChangeRejected(t *testing.T) {
	router := NewRouter(testDeps())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://evil.example")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin POST status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
