// Package iam implements the identity/access use cases: first-run setup,
// login/logout, and session resolution (docs/MASTER_PLAN.md §5.3 "app"
// layer: depends on domain + ports, no concrete infrastructure imports).
package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
	"github.com/sshreyx1/cyberditto/services/api/internal/ports"
)

var (
	ErrAlreadyInitialized = errors.New("iam: workspace already initialized")
	ErrInvalidCredentials = errors.New("iam: invalid credentials")
	ErrWeakPassword       = errors.New("iam: password must be at least 12 characters")
	ErrSessionNotFound    = errors.New("iam: session not found or expired")
	ErrNoMembership       = errors.New("iam: user has no workspace membership")
)

const (
	minPasswordLen = 12
	// SessionTTL is the absolute session lifetime (docs/MASTER_PLAN.md §6.2:
	// "idle + absolute timeouts"). Idle timeout is a near-term follow-up;
	// tracked in docs/PROGRESS.md rather than half-implemented here.
	SessionTTL = 30 * 24 * time.Hour
)

type Service struct {
	repo   ports.IAMRepository
	audit  ports.AuditRepository
	hasher ports.PasswordHasher
	tokens ports.TokenGenerator
	now    func() time.Time
}

func NewService(repo ports.IAMRepository, auditRepo ports.AuditRepository, hasher ports.PasswordHasher, tokens ports.TokenGenerator) *Service {
	return &Service{repo: repo, audit: auditRepo, hasher: hasher, tokens: tokens, now: time.Now}
}

type SetupResult struct {
	Workspace  iam.Workspace
	User       iam.User
	Membership iam.Membership
}

// Setup creates the single owner account and its default workspace. It is
// only ever valid once: if any user already exists, CyberDitto refuses to
// run in an "open" state (docs/MASTER_PLAN.md §6.2).
func (s *Service) Setup(ctx context.Context, email, password string) (*SetupResult, error) {
	count, err := s.repo.CountUsers(ctx)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrAlreadyInitialized
	}
	if len(password) < minPasswordLen {
		return nil, ErrWeakPassword
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	ws, err := s.repo.CreateWorkspace(ctx, "Default")
	if err != nil {
		return nil, err
	}

	user, err := s.repo.CreateUser(ctx, email, hash)
	if err != nil {
		return nil, err
	}

	membership, err := s.repo.CreateMembership(ctx, ws.ID, user.ID, iam.RoleOwner, true)
	if err != nil {
		return nil, err
	}

	if _, err := s.audit.Record(ctx, ws.ID, &user.ID, "workspace.setup", "workspace", ws.ID.String(), nil); err != nil {
		return nil, err
	}

	return &SetupResult{Workspace: ws, User: user, Membership: membership}, nil
}

type LoginResult struct {
	Token      string
	ExpiresAt  time.Time
	User       iam.User
	Membership iam.Membership
}

// Login verifies credentials and creates a new session. It returns the same
// ErrInvalidCredentials whether the email doesn't exist or the password is
// wrong, so the API can't be used to enumerate registered emails.
func (s *Service) Login(ctx context.Context, email, password, userAgent string) (*LoginResult, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	ok, err := s.hasher.Verify(password, user.PasswordHash)
	if err != nil || !ok {
		return nil, ErrInvalidCredentials
	}

	membership, err := s.primaryMembership(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	rawToken, err := s.tokens.Generate()
	if err != nil {
		return nil, err
	}
	expiresAt := s.now().Add(SessionTTL)

	if _, err := s.repo.CreateSession(ctx, user.ID, hashToken(rawToken), userAgent, expiresAt); err != nil {
		return nil, err
	}

	if _, err := s.audit.Record(ctx, membership.WorkspaceID, &user.ID, "session.login", "user", user.ID.String(), nil); err != nil {
		return nil, err
	}

	return &LoginResult{Token: rawToken, ExpiresAt: expiresAt, User: user, Membership: membership}, nil
}

// Logout revokes the session behind rawToken. It is intentionally
// idempotent/quiet about an unknown token: callers always clear the cookie
// either way, and a missing session isn't an actionable error for the user.
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	sess, err := s.repo.GetSessionByTokenHash(ctx, hashToken(rawToken))
	if err != nil {
		return nil
	}
	if err := s.repo.RevokeSession(ctx, sess.ID); err != nil {
		return err
	}

	membership, mErr := s.primaryMembership(ctx, sess.UserID)
	if mErr == nil {
		_, _ = s.audit.Record(ctx, membership.WorkspaceID, &sess.UserID, "session.logout", "user", sess.UserID.String(), nil)
	}
	return nil
}

// AuthContext is what middleware attaches to an authenticated request.
type AuthContext struct {
	User       iam.User
	Membership iam.Membership
}

// Authenticate resolves a raw session-cookie token to the user and their
// (Phase 1: single, primary) workspace membership.
func (s *Service) Authenticate(ctx context.Context, rawToken string) (*AuthContext, error) {
	sess, err := s.repo.GetSessionByTokenHash(ctx, hashToken(rawToken))
	if err != nil {
		return nil, ErrSessionNotFound
	}
	if !sess.Active(s.now()) {
		return nil, ErrSessionNotFound
	}

	user, err := s.repo.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}

	membership, err := s.primaryMembership(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthContext{User: user, Membership: membership}, nil
}

// primaryMembership returns the first workspace a user belongs to. Phase 1
// assumes one workspace per user (the one created at Setup); a workspace
// switcher is future scope once multi-workspace membership is exercised by
// the product, not before.
func (s *Service) primaryMembership(ctx context.Context, userID uuid.UUID) (iam.Membership, error) {
	memberships, err := s.repo.GetMembershipsForUser(ctx, userID)
	if err != nil {
		return iam.Membership{}, err
	}
	if len(memberships) == 0 {
		return iam.Membership{}, ErrNoMembership
	}
	return memberships[0], nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
