package iam

// Capability is a single authorizable action. Handlers check capabilities,
// never roles directly, so the mapping stays in one place
// (docs/MASTER_PLAN.md §6.2: "check on every handler via a single policy
// function"). This table is also what the authZ-matrix tests enumerate.
type Capability string

const (
	CapViewWorkspace    Capability = "workspace:view"
	CapManageMembers    Capability = "members:manage"
	CapManageAssets     Capability = "assets:manage"
	CapViewAssets       Capability = "assets:view"
	CapManageTwins      Capability = "twins:manage"
	CapViewTwins        Capability = "twins:view"
	CapRunAssessment    Capability = "assessments:run"
	CapViewAssessments  Capability = "assessments:view"
	CapViewEmulation    Capability = "emulation:view"
	CapRunEmulation     Capability = "emulation:run"     // requires Approver, see Can
	CapApplyRemediation Capability = "remediation:apply" // requires Approver, see Can
	CapViewAuditLog     Capability = "audit:view"
)

// roleCapabilities is the base grant per role, before the Approver overlay.
var roleCapabilities = map[Role]map[Capability]bool{
	RoleOwner: {
		CapViewWorkspace: true, CapManageMembers: true,
		CapManageAssets: true, CapViewAssets: true,
		CapManageTwins: true, CapViewTwins: true,
		CapRunAssessment: true, CapViewAssessments: true, CapViewEmulation: true,
		CapRunEmulation: true, CapApplyRemediation: true,
		CapViewAuditLog: true,
	},
	RoleAdmin: {
		CapViewWorkspace: true, CapManageMembers: true,
		CapManageAssets: true, CapViewAssets: true,
		CapManageTwins: true, CapViewTwins: true,
		CapRunAssessment: true, CapViewAssessments: true, CapViewEmulation: true,
		CapRunEmulation: true, CapApplyRemediation: true,
		CapViewAuditLog: true,
	},
	RoleAnalyst: {
		CapViewWorkspace: true,
		CapManageAssets:  true, CapViewAssets: true,
		CapManageTwins: true, CapViewTwins: true,
		CapRunAssessment: true, CapViewAssessments: true, CapViewEmulation: true,
		CapRunEmulation: true, CapApplyRemediation: true,
	},
	RoleViewer: {
		CapViewWorkspace:   true,
		CapViewAssets:      true,
		CapViewTwins:       true,
		CapViewAssessments: true,
		CapViewEmulation:   true,
	},
}

// approverGated capabilities additionally require Membership.Approver = true,
// regardless of role (docs/MASTER_PLAN.md §6.4 invariant 4: every attack run
// and remediation apply needs a recorded authorisation acknowledgement by an
// approver).
var approverGated = map[Capability]bool{
	CapRunEmulation:     true,
	CapApplyRemediation: true,
}

// Can is the single authZ policy function. Every HTTP handler that mutates
// or reads workspace-scoped state must call this rather than inspecting
// Role directly.
func Can(m Membership, cap Capability) bool {
	if !m.Role.Valid() {
		return false
	}
	caps := roleCapabilities[m.Role]
	if caps == nil || !caps[cap] {
		return false
	}
	if approverGated[cap] && !m.Approver {
		return false
	}
	return true
}
