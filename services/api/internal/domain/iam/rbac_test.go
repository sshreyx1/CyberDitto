package iam

import "testing"

// allCapabilities must be kept in sync with the const block in rbac.go; the
// test fails loudly (via the matrix below) if a capability is ever added
// there and forgotten here, rather than silently skipping it.
var allCapabilities = []Capability{
	CapViewWorkspace, CapManageMembers, CapManageAssets, CapViewAssets,
	CapManageTwins, CapViewTwins, CapRunAssessment, CapViewAssessments,
	CapViewEmulation, CapRunEmulation, CapApplyRemediation, CapViewAuditLog,
}

var allRoles = []Role{RoleOwner, RoleAdmin, RoleAnalyst, RoleViewer}

// TestAuthZMatrix is the "every route (capability) × every role" table the
// MASTER_PLAN requires (§6.2, §9.4). Expected grants are spelled out
// explicitly so a change in rbac.go's tables shows up as a readable diff
// here, not a silent behavior change.
func TestAuthZMatrix(t *testing.T) {
	expected := map[Role]map[Capability]bool{
		RoleOwner: {
			CapViewWorkspace: true, CapManageMembers: true, CapManageAssets: true,
			CapViewAssets: true, CapManageTwins: true, CapViewTwins: true,
			CapRunAssessment: true, CapViewAssessments: true, CapViewEmulation: true,
			CapRunEmulation: true, CapApplyRemediation: true, CapViewAuditLog: true,
		},
		RoleAdmin: {
			CapViewWorkspace: true, CapManageMembers: true, CapManageAssets: true,
			CapViewAssets: true, CapManageTwins: true, CapViewTwins: true,
			CapRunAssessment: true, CapViewAssessments: true, CapViewEmulation: true,
			CapRunEmulation: true, CapApplyRemediation: true, CapViewAuditLog: true,
		},
		RoleAnalyst: {
			CapViewWorkspace: true, CapManageMembers: false, CapManageAssets: true,
			CapViewAssets: true, CapManageTwins: true, CapViewTwins: true,
			CapRunAssessment: true, CapViewAssessments: true, CapViewEmulation: true,
			CapRunEmulation: true, CapApplyRemediation: true, CapViewAuditLog: false,
		},
		RoleViewer: {
			CapViewWorkspace: true, CapManageMembers: false, CapManageAssets: false,
			CapViewAssets: true, CapManageTwins: false, CapViewTwins: true,
			CapRunAssessment: false, CapViewAssessments: true, CapViewEmulation: true,
			CapRunEmulation: false, CapApplyRemediation: false, CapViewAuditLog: false,
		},
	}

	for _, role := range allRoles {
		for _, cap := range allCapabilities {
			want, ok := expected[role][cap]
			if !ok {
				t.Fatalf("no expectation recorded for role=%s capability=%s; update the matrix", role, cap)
			}
			m := Membership{Role: role, Approver: true} // approver=true isolates the role grant from the approver overlay, tested separately below
			if got := Can(m, cap); got != want {
				t.Errorf("Can(role=%s, approver=true, %s) = %v, want %v", role, cap, got, want)
			}
		}
	}
}

// TestApproverGate verifies §6.4 invariant 4: emulation runs and remediation
// applies require Approver=true regardless of role, including the owner.
func TestApproverGate(t *testing.T) {
	gated := []Capability{CapRunEmulation, CapApplyRemediation}
	for _, role := range allRoles {
		for _, cap := range gated {
			m := Membership{Role: role, Approver: false}
			if Can(m, cap) {
				t.Errorf("Can(role=%s, approver=false, %s) = true, want false (approver-gated capability)", role, cap)
			}
		}
	}
}

func TestCan_InvalidRoleDenied(t *testing.T) {
	m := Membership{Role: Role("made-up"), Approver: true}
	for _, cap := range allCapabilities {
		if Can(m, cap) {
			t.Errorf("Can(invalid role, %s) = true, want false", cap)
		}
	}
}

func TestCan_UnknownCapabilityDenied(t *testing.T) {
	m := Membership{Role: RoleOwner, Approver: true}
	if Can(m, Capability("made-up:capability")) {
		t.Error("Can(owner, unknown capability) = true, want false")
	}
}
