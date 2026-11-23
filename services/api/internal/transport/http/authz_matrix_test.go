package http

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

// routeCase is one capability-gated route. Path placeholders {id} are
// filled with a random UUID: authorisation runs before the handler looks
// anything up, so a forbidden role gets 403 regardless, and an allowed
// role gets whatever the handler returns (400/404/…) — anything but 403.
type routeCase struct {
	method, path string
	capability   iam.Capability
}

// authzRoutes must list every capability-gated route registered in
// NewRouter; add a row whenever a gated route is added. The capability
// column is restated here on purpose rather than read from the router, so
// a route wired to the wrong capability shows up as a failing cell.
var authzRoutes = []routeCase{
	{http.MethodPost, "/api/v1/assets", iam.CapManageAssets},
	{http.MethodGet, "/api/v1/assets", iam.CapViewAssets},
	{http.MethodGet, "/api/v1/assets/{id}", iam.CapViewAssets},
	{http.MethodPost, "/api/v1/assets/{id}/snapshots", iam.CapManageAssets},
	{http.MethodGet, "/api/v1/assets/{id}/snapshots", iam.CapViewAssets},
	{http.MethodGet, "/api/v1/snapshots/{id}", iam.CapViewAssets},
	{http.MethodPost, "/api/v1/assessments", iam.CapRunAssessment},
	{http.MethodGet, "/api/v1/assessments", iam.CapViewAssessments},
	{http.MethodGet, "/api/v1/assessments/{id}", iam.CapViewAssessments},
	{http.MethodGet, "/api/v1/controls", iam.CapViewAssessments},
	{http.MethodPost, "/api/v1/twins", iam.CapManageTwins},
	{http.MethodGet, "/api/v1/twins", iam.CapViewTwins},
	{http.MethodGet, "/api/v1/twins/{id}", iam.CapViewTwins},
	{http.MethodGet, "/api/v1/twins/{id}/events", iam.CapViewTwins},
	{http.MethodPost, "/api/v1/twins/{id}/start", iam.CapManageTwins},
	{http.MethodPost, "/api/v1/twins/{id}/stop", iam.CapManageTwins},
	{http.MethodPost, "/api/v1/twins/{id}/destroy", iam.CapManageTwins},
}

// TestAuthZMatrix is the HTTP-level "every route × every role" matrix
// (docs/MASTER_PLAN.md §6.2, §9.4), driven through the real middleware
// chain so it also catches a route wired to the wrong capability. The
// expected outcome per cell comes from iam.Can, whose own table is pinned
// separately by internal/domain/iam's TestAuthZMatrix.
func TestAuthZMatrix(t *testing.T) {
	for _, role := range []iam.Role{iam.RoleOwner, iam.RoleAdmin, iam.RoleAnalyst, iam.RoleViewer} {
		for _, approver := range []bool{true, false} {
			router, cookies := setupAs(t, role, approver)
			m := iam.Membership{Role: role, Approver: approver}
			for _, rc := range authzRoutes {
				path := strings.ReplaceAll(rc.path, "{id}", uuid.NewString())
				var rec = doRaw(t, router, rc.method, path, []byte(`{}`), nil, cookies)
				allowed := rec.Code != http.StatusForbidden
				if want := iam.Can(m, rc.capability); allowed != want {
					t.Errorf("role=%s approver=%v %s %s: status %d (allowed=%v), want allowed=%v",
						role, approver, rc.method, rc.path, rec.Code, allowed, want)
				}
			}
		}
	}
}

// TestAuthZMatrix_UnauthenticatedGetsUnauthorized: no session → 401 on
// every gated route, never 403 or a handler response.
func TestAuthZMatrix_UnauthenticatedGetsUnauthorized(t *testing.T) {
	router, _ := setupAs(t, iam.RoleOwner, true)
	for _, rc := range authzRoutes {
		path := strings.ReplaceAll(rc.path, "{id}", uuid.NewString())
		if rec := doRaw(t, router, rc.method, path, []byte(`{}`), nil, nil); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without a session: status %d, want 401", rc.method, rc.path, rec.Code)
		}
	}
}
