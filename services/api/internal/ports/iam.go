// Package ports defines the interfaces application use cases depend on;
// concrete I/O lives in internal/adapters (docs/MASTER_PLAN.md §5.1, §5.8).
package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

// IAMRepository persists identity/access state. Workspace-scoped methods
// take the workspace explicitly so the adapter can enforce Postgres RLS
// (docs/MASTER_PLAN.md §6.2) rather than trusting the caller.
type IAMRepository interface {
	CountUsers(ctx context.Context) (int64, error)
	CreateWorkspace(ctx context.Context, name string) (iam.Workspace, error)
	CreateUser(ctx context.Context, email, passwordHash string) (iam.User, error)
	GetUserByEmail(ctx context.Context, email string) (iam.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (iam.User, error)
	CreateMembership(ctx context.Context, workspaceID, userID uuid.UUID, role iam.Role, approver bool) (iam.Membership, error)
	GetMembershipsForUser(ctx context.Context, userID uuid.UUID) ([]iam.Membership, error)
	CreateSession(ctx context.Context, userID uuid.UUID, tokenHash, userAgent string, expiresAt time.Time) (iam.Session, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (iam.Session, error)
	RevokeSession(ctx context.Context, id uuid.UUID) error
}

// PasswordHasher hashes and verifies passwords. The adapter owns the
// algorithm choice (Argon2id, docs/MASTER_PLAN.md §6.2) and its parameters;
// the app layer only ever sees opaque hash strings.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
}

// TokenGenerator produces cryptographically random, URL-safe session
// tokens. The app layer never sees raw bytes, only the encoded string.
type TokenGenerator interface {
	Generate() (string, error)
}
