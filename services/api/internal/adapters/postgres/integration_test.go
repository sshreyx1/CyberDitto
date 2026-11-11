//go:build integration

package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/argon2"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/token"
	appiam "github.com/sshreyx1/cyberditto/services/api/internal/app/iam"
	domainaudit "github.com/sshreyx1/cyberditto/services/api/internal/domain/audit"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

// These tests need two DSNs (docs/adr/0012-integration-test-strategy.md):
// CD_TEST_DATABASE_URL (the cyberditto_app role; exercises the real
// production privilege set, including RLS) and CD_TEST_ADMIN_DATABASE_URL
// (the owner role; used only to reset state between test runs — resetting
// audit_events/snapshots relies on the owner's privileges precisely because
// cyberditto_app is deliberately never granted DELETE on those immutable
// tables). Both are unset by default, so `go test ./...` skips these
// without a configured Postgres; CI sets them (ci.yml).
func testPools(t *testing.T) (app *pgxpool.Pool, admin *pgxpool.Pool) {
	t.Helper()
	appDSN := os.Getenv("CD_TEST_DATABASE_URL")
	adminDSN := os.Getenv("CD_TEST_ADMIN_DATABASE_URL")
	if appDSN == "" || adminDSN == "" {
		t.Skip("CD_TEST_DATABASE_URL/CD_TEST_ADMIN_DATABASE_URL not set; skipping Postgres integration test")
	}

	ctx := context.Background()
	appPool, err := NewPool(ctx, appDSN)
	if err != nil {
		t.Fatalf("connect app pool: %v", err)
	}
	t.Cleanup(appPool.Close)

	adminPool, err := NewPool(ctx, adminDSN)
	if err != nil {
		t.Fatalf("connect admin pool: %v", err)
	}
	t.Cleanup(adminPool.Close)

	// Reset state (FK ON DELETE CASCADE takes care of dependent rows).
	if _, err := adminPool.Exec(ctx, "DELETE FROM workspaces"); err != nil {
		t.Fatalf("reset workspaces: %v", err)
	}
	if _, err := adminPool.Exec(ctx, "DELETE FROM users"); err != nil {
		t.Fatalf("reset users: %v", err)
	}

	return appPool, adminPool
}

func TestIAMRepo_SetupLoginAuthenticate(t *testing.T) {
	appPool, _ := testPools(t)
	ctx := context.Background()

	repo := NewIAMRepo(appPool)
	auditRepo := NewAuditRepo(appPool)
	svc := appiam.NewService(repo, auditRepo, argon2.New(), token.New())

	const email = "owner@example.com"
	const password = "correct horse battery staple"

	setup, err := svc.Setup(ctx, email, password)
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if setup.Membership.Role != iam.RoleOwner || !setup.Membership.Approver {
		t.Fatalf("Setup membership = %+v, want owner+approver", setup.Membership)
	}

	if _, err := svc.Setup(ctx, "someone-else@example.com", password); !errors.Is(err, appiam.ErrAlreadyInitialized) {
		t.Fatalf("second Setup error = %v, want ErrAlreadyInitialized", err)
	}

	login, err := svc.Login(ctx, email, password, "integration-test")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if login.User.ID != setup.User.ID {
		t.Fatalf("Login user = %s, want %s", login.User.ID, setup.User.ID)
	}

	if _, err := svc.Login(ctx, email, "wrong password", ""); !errors.Is(err, appiam.ErrInvalidCredentials) {
		t.Fatalf("wrong-password Login error = %v, want ErrInvalidCredentials", err)
	}
	if _, err := svc.Login(ctx, "nobody@example.com", password, ""); !errors.Is(err, appiam.ErrInvalidCredentials) {
		t.Fatalf("unknown-email Login error = %v, want ErrInvalidCredentials (must not reveal the email doesn't exist)", err)
	}

	authed, err := svc.Authenticate(ctx, login.Token)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if authed.User.Email != email {
		t.Fatalf("Authenticate email = %s, want %s", authed.User.Email, email)
	}

	if err := svc.Logout(ctx, login.Token); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if _, err := svc.Authenticate(ctx, login.Token); !errors.Is(err, appiam.ErrSessionNotFound) {
		t.Fatalf("Authenticate after Logout error = %v, want ErrSessionNotFound", err)
	}
}

// TestIAMRepo_RLSIsolation exercises the real adapter code path (not raw
// SQL) to confirm a user's GetMembershipsForUser call can never surface
// another workspace's membership row, which is the property
// docs/MASTER_PLAN.md §6.2's RLS-as-defence-in-depth exists to guarantee
// even if an app-layer bug ever asked for the wrong workspace.
func TestIAMRepo_RLSIsolation(t *testing.T) {
	appPool, _ := testPools(t)
	ctx := context.Background()
	repo := NewIAMRepo(appPool)

	wsA, err := repo.CreateWorkspace(ctx, "WS-A")
	if err != nil {
		t.Fatalf("CreateWorkspace A: %v", err)
	}
	wsB, err := repo.CreateWorkspace(ctx, "WS-B")
	if err != nil {
		t.Fatalf("CreateWorkspace B: %v", err)
	}

	userA, err := repo.CreateUser(ctx, "a@example.com", "hash")
	if err != nil {
		t.Fatalf("CreateUser A: %v", err)
	}
	userB, err := repo.CreateUser(ctx, "b@example.com", "hash")
	if err != nil {
		t.Fatalf("CreateUser B: %v", err)
	}

	if _, err := repo.CreateMembership(ctx, wsA.ID, userA.ID, iam.RoleOwner, true); err != nil {
		t.Fatalf("CreateMembership A: %v", err)
	}
	if _, err := repo.CreateMembership(ctx, wsB.ID, userB.ID, iam.RoleOwner, true); err != nil {
		t.Fatalf("CreateMembership B: %v", err)
	}

	membershipsA, err := repo.GetMembershipsForUser(ctx, userA.ID)
	if err != nil {
		t.Fatalf("GetMembershipsForUser A: %v", err)
	}
	if len(membershipsA) != 1 || membershipsA[0].WorkspaceID != wsA.ID {
		t.Fatalf("GetMembershipsForUser(A) = %+v, want exactly [workspace A]", membershipsA)
	}

	membershipsB, err := repo.GetMembershipsForUser(ctx, userB.ID)
	if err != nil {
		t.Fatalf("GetMembershipsForUser B: %v", err)
	}
	if len(membershipsB) != 1 || membershipsB[0].WorkspaceID != wsB.ID {
		t.Fatalf("GetMembershipsForUser(B) = %+v, want exactly [workspace B]", membershipsB)
	}
}

func TestAuditRepo_HashChain(t *testing.T) {
	appPool, _ := testPools(t)
	ctx := context.Background()

	iamRepo := NewIAMRepo(appPool)
	auditRepo := NewAuditRepo(appPool)

	ws, err := iamRepo.CreateWorkspace(ctx, "chain-test")
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	// actor_user_id has a real FK to users(id), so the test needs an actual
	// row there, not an arbitrary UUID.
	actorUser, err := iamRepo.CreateUser(ctx, "chain-actor@example.com", "hash")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	actor := actorUser.ID

	first, err := auditRepo.Record(ctx, ws.ID, &actor, "workspace.setup", "workspace", ws.ID.String(), nil)
	if err != nil {
		t.Fatalf("Record first: %v", err)
	}
	if first.PrevHash != domainaudit.GenesisHash {
		t.Fatalf("first.PrevHash = %q, want GenesisHash", first.PrevHash)
	}

	second, err := auditRepo.Record(ctx, ws.ID, &actor, "session.login", "user", actor.String(), map[string]any{"note": "test"})
	if err != nil {
		t.Fatalf("Record second: %v", err)
	}
	if second.PrevHash != first.Hash {
		t.Fatalf("second.PrevHash = %q, want first.Hash = %q", second.PrevHash, first.Hash)
	}

	for _, e := range []domainaudit.Event{first, second} {
		ok, err := domainaudit.Verify(e)
		if err != nil {
			t.Fatalf("Verify(%s): %v", e.Action, err)
		}
		if !ok {
			t.Fatalf("Verify(%s) = false, want true", e.Action)
		}
	}

	events, err := auditRepo.List(ctx, ws.ID, 10, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("List returned %d events, want 2", len(events))
	}
}
