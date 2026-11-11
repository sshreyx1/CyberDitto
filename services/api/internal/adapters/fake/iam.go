// Package fake provides in-memory ports implementations for fast unit
// tests and demo mode (docs/MASTER_PLAN.md §5.1: "a Fake adapter for each
// [port] makes CI and demo mode possible without a hypervisor" — the same
// principle applies to any port, not just LabProvider).
package fake

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

type IAMRepo struct {
	mu          sync.Mutex
	workspaces  map[uuid.UUID]iam.Workspace
	users       map[uuid.UUID]iam.User
	usersByMail map[string]uuid.UUID
	memberships map[uuid.UUID]iam.Membership
	sessions    map[string]iam.Session // keyed by token hash
}

func NewIAMRepo() *IAMRepo {
	return &IAMRepo{
		workspaces:  map[uuid.UUID]iam.Workspace{},
		users:       map[uuid.UUID]iam.User{},
		usersByMail: map[string]uuid.UUID{},
		memberships: map[uuid.UUID]iam.Membership{},
		sessions:    map[string]iam.Session{},
	}
}

func (r *IAMRepo) CountUsers(ctx context.Context) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(len(r.users)), nil
}

func (r *IAMRepo) CreateWorkspace(ctx context.Context, name string) (iam.Workspace, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w := iam.Workspace{ID: uuid.New(), Name: name, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	r.workspaces[w.ID] = w
	return w, nil
}

func (r *IAMRepo) CreateUser(ctx context.Context, email, passwordHash string) (iam.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u := iam.User{ID: uuid.New(), Email: email, PasswordHash: passwordHash, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	r.users[u.ID] = u
	r.usersByMail[email] = u.ID
	return u, nil
}

func (r *IAMRepo) GetUserByEmail(ctx context.Context, email string) (iam.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.usersByMail[email]
	if !ok {
		return iam.User{}, errNotFound
	}
	return r.users[id], nil
}

func (r *IAMRepo) GetUserByID(ctx context.Context, id uuid.UUID) (iam.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		return iam.User{}, errNotFound
	}
	return u, nil
}

func (r *IAMRepo) CreateMembership(ctx context.Context, workspaceID, userID uuid.UUID, role iam.Role, approver bool) (iam.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := iam.Membership{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID,
		Role: role, Approver: approver, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	r.memberships[m.ID] = m
	return m, nil
}

func (r *IAMRepo) GetMembershipsForUser(ctx context.Context, userID uuid.UUID) ([]iam.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []iam.Membership
	for _, m := range r.memberships {
		if m.UserID == userID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *IAMRepo) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash, userAgent string, expiresAt time.Time) (iam.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := iam.Session{
		ID: uuid.New(), UserID: userID, TokenHash: tokenHash, UserAgent: userAgent,
		CreatedAt: time.Now(), ExpiresAt: expiresAt,
	}
	r.sessions[tokenHash] = s
	return s, nil
}

func (r *IAMRepo) GetSessionByTokenHash(ctx context.Context, tokenHash string) (iam.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[tokenHash]
	if !ok || !s.Active(time.Now()) {
		return iam.Session{}, errNotFound
	}
	return s, nil
}

func (r *IAMRepo) RevokeSession(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for hash, s := range r.sessions {
		if s.ID == id {
			now := time.Now()
			s.RevokedAt = &now
			r.sessions[hash] = s
		}
	}
	return nil
}

type notFoundError struct{}

func (notFoundError) Error() string { return "fake: not found" }

var errNotFound error = notFoundError{}
