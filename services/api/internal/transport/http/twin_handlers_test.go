package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

func TestTwinRoutes_RequireAuthentication(t *testing.T) {
	router, _ := setupAs(t, iam.RoleOwner, true)
	rec := doJSON(t, router, http.MethodGet, "/api/v1/twins", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /api/v1/twins with no session: status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestTwinCreate_RequiresRealSnapshot(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/twins", createTwinRequest{}, cookies)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("create without snapshot_id: status = %d, want 400", rec.Code)
	}
	rec = doJSON(t, router, http.MethodPost, "/api/v1/twins", createTwinRequest{SnapshotID: uuid.NewString()}, cookies)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("create with unknown snapshot_id: status = %d, want 422", rec.Code)
	}
}

func createTwin(t *testing.T, router http.Handler, cookies []*http.Cookie) string {
	t.Helper()
	_, snapshotID := createAssetWithSnapshot(t, router, cookies)
	rec := doJSON(t, router, http.MethodPost, "/api/v1/twins", createTwinRequest{SnapshotID: snapshotID}, cookies)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create twin status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var created struct {
		ID         string `json:"id"`
		State      string `json:"state"`
		SnapshotID string `json:"snapshot_id"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.State != "requested" || created.SnapshotID != snapshotID {
		t.Fatalf("created twin = %+v", created)
	}
	return created.ID
}

func TestTwinLifecycle_CreateListGetStartStopDestroy(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)
	id := createTwin(t, router, cookies)

	rec := doJSON(t, router, http.MethodPost, "/api/v1/twins/"+id+"/start", nil, cookies)
	if rec.Code != http.StatusConflict {
		t.Fatalf("premature start status = %d, want %d", rec.Code, http.StatusConflict)
	}

	if state := pollTwinState(t, router, cookies, id, "ready"); state != "ready" {
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

	for _, action := range []string{"start", "stop", "destroy"} {
		rec = doJSON(t, router, http.MethodPost, "/api/v1/twins/"+id+"/"+action, nil, cookies)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status = %d, body = %s", action, rec.Code, rec.Body.String())
		}
	}

	rec = doJSON(t, router, http.MethodPost, "/api/v1/twins/"+id+"/start", nil, cookies)
	if rec.Code != http.StatusConflict {
		t.Fatalf("start after destroy status = %d, want %d", rec.Code, http.StatusConflict)
	}

	rec = doJSON(t, router, http.MethodGet, "/api/v1/twins/"+uuid.NewString(), nil, cookies)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get unknown twin status = %d, want 404", rec.Code)
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

func streamUntil(t *testing.T, router http.Handler, cookies []*http.Cookie, path string, headers map[string]string, contains string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, path, nil).WithContext(ctx)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := newSSERecorder()
	done := make(chan struct{})
	go func() {
		router.ServeHTTP(rec, req)
		close(done)
	}()
	for ctx.Err() == nil && !strings.Contains(rec.snapshot(), contains) {
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done
	return rec.snapshot()
}

func TestTwinEvents_SSEStreamsBuildProgress(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)
	id := createTwin(t, router, cookies)

	out := streamUntil(t, router, cookies, "/api/v1/twins/"+id+"/events", nil, `"state":"ready"`)
	if !strings.Contains(out, `"state":"ready"`) {
		t.Fatalf("SSE stream never reported ready; got:\n%s", out)
	}
	if !strings.Contains(out, "id: 1\n") {
		t.Errorf("SSE events must carry ids for Last-Event-ID resume; got:\n%s", out)
	}
}

func TestTwinEvents_ResumesAfterLastEventID(t *testing.T) {
	router, cookies := setupAs(t, iam.RoleOwner, true)
	id := createTwin(t, router, cookies)
	pollTwinState(t, router, cookies, id, "ready")

	out := streamUntil(t, router, cookies, "/api/v1/twins/"+id+"/events", map[string]string{"Last-Event-ID": "3"}, `"state":"ready"`)
	if strings.Contains(out, "id: 1\n") || strings.Contains(out, "id: 3\n") {
		t.Errorf("resume after Last-Event-ID 3 replayed events 1-3:\n%s", out)
	}
	if !strings.Contains(out, "id: 4\n") || !strings.Contains(out, "id: 5\n") {
		t.Errorf("resume after Last-Event-ID 3 should replay 4 and 5:\n%s", out)
	}
}
