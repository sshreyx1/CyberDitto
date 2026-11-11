// Package iam holds pure identity/access domain logic: roles, capabilities,
// and the RBAC policy function. No I/O, no infrastructure imports
// (docs/MASTER_PLAN.md §5.1).
package iam

import (
	"time"

	"github.com/google/uuid"
)

// Role is a membership's role within a workspace (docs/MASTER_PLAN.md §6.2).
type Role string

const (
	RoleOwner   Role = "owner"
	RoleAdmin   Role = "admin"
	RoleAnalyst Role = "analyst"
	RoleViewer  Role = "viewer"
)

// Valid reports whether r is one of the known roles.
func (r Role) Valid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleAnalyst, RoleViewer:
		return true
	default:
		return false
	}
}

type Workspace struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Membership grants a User a Role within a Workspace. Approver is a
// capability layered on top of role: it gates attack-run authorisation and
// remediation-apply actions regardless of role (docs/MASTER_PLAN.md §6.2).
type Membership struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Role        Role
	Approver    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Session is a server-side session record. The plaintext token never
// persists server-side — only TokenHash does (docs/MASTER_PLAN.md §6.2).
type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	UserAgent string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func (s Session) Active(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}
