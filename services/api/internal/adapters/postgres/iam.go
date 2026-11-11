package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	pgdb "github.com/sshreyx1/cyberditto/services/api/internal/adapters/postgres/gen"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/iam"
)

type IAMRepo struct {
	pool *pgxpool.Pool
}

func NewIAMRepo(pool *pgxpool.Pool) *IAMRepo {
	return &IAMRepo{pool: pool}
}

func (r *IAMRepo) CountUsers(ctx context.Context) (int64, error) {
	return pgdb.New(r.pool).CountUsers(ctx)
}

func (r *IAMRepo) CreateWorkspace(ctx context.Context, name string) (iam.Workspace, error) {
	w, err := pgdb.New(r.pool).CreateWorkspace(ctx, name)
	if err != nil {
		return iam.Workspace{}, err
	}
	return workspaceFromRow(w), nil
}

func (r *IAMRepo) CreateUser(ctx context.Context, email, passwordHash string) (iam.User, error) {
	u, err := pgdb.New(r.pool).CreateUser(ctx, pgdb.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return iam.User{}, err
	}
	return userFromRow(u), nil
}

func (r *IAMRepo) GetUserByEmail(ctx context.Context, email string) (iam.User, error) {
	u, err := pgdb.New(r.pool).GetUserByEmail(ctx, email)
	if err != nil {
		return iam.User{}, err
	}
	return userFromRow(u), nil
}

func (r *IAMRepo) GetUserByID(ctx context.Context, id uuid.UUID) (iam.User, error) {
	u, err := pgdb.New(r.pool).GetUserByID(ctx, id)
	if err != nil {
		return iam.User{}, err
	}
	return userFromRow(u), nil
}

// CreateMembership runs inside a workspace-scoped transaction: the new
// membership row's workspace_id must equal the RLS session variable or the
// insert is rejected by the policy, which is the point (docs/MASTER_PLAN.md
// §6.2 RLS as defence in depth).
func (r *IAMRepo) CreateMembership(ctx context.Context, workspaceID, userID uuid.UUID, role iam.Role, approver bool) (iam.Membership, error) {
	var m pgdb.Membership
	err := withWorkspaceContext(ctx, r.pool, workspaceID, &userID, func(q *pgdb.Queries) error {
		var err error
		m, err = q.CreateMembership(ctx, pgdb.CreateMembershipParams{
			WorkspaceID: workspaceID,
			UserID:      userID,
			Role:        string(role),
			Approver:    approver,
		})
		return err
	})
	if err != nil {
		return iam.Membership{}, err
	}
	return membershipFromRow(m), nil
}

func (r *IAMRepo) GetMembershipsForUser(ctx context.Context, userID uuid.UUID) ([]iam.Membership, error) {
	var rows []pgdb.Membership
	err := withUserContext(ctx, r.pool, userID, func(q *pgdb.Queries) error {
		var err error
		rows, err = q.GetMembershipsForUser(ctx, userID)
		return err
	})
	if err != nil {
		return nil, err
	}
	out := make([]iam.Membership, 0, len(rows))
	for _, m := range rows {
		out = append(out, membershipFromRow(m))
	}
	return out, nil
}

func (r *IAMRepo) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash, userAgent string, expiresAt time.Time) (iam.Session, error) {
	var ua *string
	if userAgent != "" {
		ua = &userAgent
	}
	s, err := pgdb.New(r.pool).CreateSession(ctx, pgdb.CreateSessionParams{
		UserID:    userID,
		TokenHash: tokenHash,
		UserAgent: ua,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return iam.Session{}, err
	}
	return sessionFromRow(s), nil
}

func (r *IAMRepo) GetSessionByTokenHash(ctx context.Context, tokenHash string) (iam.Session, error) {
	s, err := pgdb.New(r.pool).GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return iam.Session{}, err
	}
	return sessionFromRow(s), nil
}

func (r *IAMRepo) RevokeSession(ctx context.Context, id uuid.UUID) error {
	return pgdb.New(r.pool).RevokeSession(ctx, id)
}

func workspaceFromRow(w pgdb.Workspace) iam.Workspace {
	return iam.Workspace{
		ID:        w.ID,
		Name:      w.Name,
		CreatedAt: w.CreatedAt.Time,
		UpdatedAt: w.UpdatedAt.Time,
	}
}

func userFromRow(u pgdb.User) iam.User {
	return iam.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}
}

func membershipFromRow(m pgdb.Membership) iam.Membership {
	return iam.Membership{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		UserID:      m.UserID,
		Role:        iam.Role(m.Role),
		Approver:    m.Approver,
		CreatedAt:   m.CreatedAt.Time,
		UpdatedAt:   m.UpdatedAt.Time,
	}
}

func sessionFromRow(s pgdb.Session) iam.Session {
	var ua string
	if s.UserAgent != nil {
		ua = *s.UserAgent
	}
	sess := iam.Session{
		ID:        s.ID,
		UserID:    s.UserID,
		TokenHash: s.TokenHash,
		UserAgent: ua,
		CreatedAt: s.CreatedAt.Time,
		ExpiresAt: s.ExpiresAt.Time,
	}
	if s.RevokedAt.Valid {
		t := s.RevokedAt.Time
		sess.RevokedAt = &t
	}
	return sess
}
