package emulation

import (
	"fmt"
	"strings"
	"time"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
)

// Run limits (§6.4 invariant 5).
const (
	MaxTestsPerRun = 25
	MaxRunDuration = 30 * time.Minute
	// MaxTestDuration bounds one test; the provider must enforce it.
	MaxTestDuration = 2 * time.Minute
)

// IsolationReport is the lab provider's answer to "can anything in this
// twin reach outside the lab right now?" (§6.4 invariant 2), produced
// fresh before every run.
type IsolationReport struct {
	Verified bool
	Detail   string
}

type Check struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

// PreflightInput is everything the policy engine needs. The run target is
// a twin state, not an ID the caller supplies: by the time preflight runs,
// the app layer has already resolved the target as a twin in this
// workspace (invariant 1 — there is no code path that accepts an asset or
// a host).
type PreflightInput struct {
	TwinState       twin.State
	Isolation       IsolationReport
	CatalogVerified bool
	CatalogDetail   string
	Tests           []Test
	UnknownTests    []string // requested GUIDs not found in the pinned catalog
	Approver        bool
	Acknowledged    bool
}

// Preflight evaluates every safety check and returns all of them (so the
// console can show which passed and which didn't), plus whether the run
// may start. Every check must pass.
func Preflight(in PreflightInput) ([]Check, bool) {
	checks := []Check{
		checkTwinBuilt(in.TwinState),
		{
			ID: "lab.isolated", Name: "Lab isolation verified for this run",
			Passed: in.Isolation.Verified, Detail: in.Isolation.Detail,
		},
		{
			ID: "catalog.verified", Name: "Attack catalog is the pinned, hash-verified build",
			Passed: in.CatalogVerified, Detail: in.CatalogDetail,
		},
		checkTestsFromCatalog(in),
		checkNotDestructive(in.Tests),
		checkTestCount(len(in.Tests) + len(in.UnknownTests)),
		checkAuthorisation(in.Approver, in.Acknowledged),
	}
	ok := true
	for _, c := range checks {
		ok = ok && c.Passed
	}
	return checks, ok
}

func checkTwinBuilt(s twin.State) Check {
	c := Check{ID: "twin.ready", Name: "Target is a built twin"}
	switch s {
	case twin.StateReady, twin.StateRunning, twin.StateStopped:
		c.Passed = true
		c.Detail = "twin state " + string(s)
	default:
		c.Detail = "twin state " + string(s) + "; it must be ready, running, or stopped"
	}
	return c
}

func checkTestsFromCatalog(in PreflightInput) Check {
	c := Check{ID: "tests.catalog_only", Name: "Every test comes from the pinned catalog (no drafts, nothing ad hoc)"}
	var bad []string
	bad = append(bad, in.UnknownTests...)
	for _, t := range in.Tests {
		if t.Origin != OriginCatalog {
			bad = append(bad, t.GUID+" ("+string(t.Origin)+")")
		}
	}
	if len(bad) > 0 {
		c.Detail = "not in the pinned catalog: " + strings.Join(bad, ", ")
		return c
	}
	c.Passed = true
	c.Detail = fmt.Sprintf("%d tests, all from the catalog", len(in.Tests))
	return c
}

func checkNotDestructive(tests []Test) Check {
	c := Check{ID: "tests.not_destructive", Name: "No destructive or denylisted test"}
	var bad []string
	for _, t := range tests {
		if t.RiskTier == TierDestructive {
			bad = append(bad, t.GUID+": risk tier destructive")
			continue
		}
		if t.RiskTier != TierBenign && t.RiskTier != TierIntrusive {
			bad = append(bad, t.GUID+": unknown risk tier "+string(t.RiskTier))
			continue
		}
		if denied, why := Denied(t); denied {
			bad = append(bad, t.GUID+": "+why)
		}
	}
	if len(bad) > 0 {
		c.Detail = strings.Join(bad, "; ")
		return c
	}
	c.Passed = true
	c.Detail = "all tests are benign or intrusive and pass the denylist"
	return c
}

func checkTestCount(n int) Check {
	c := Check{ID: "run.limits", Name: fmt.Sprintf("Between 1 and %d tests; %s wall-clock limit", MaxTestsPerRun, MaxRunDuration)}
	c.Passed = n >= 1 && n <= MaxTestsPerRun
	c.Detail = fmt.Sprintf("%d tests requested", n)
	return c
}

func checkAuthorisation(approver, acknowledged bool) Check {
	c := Check{ID: "run.authorised", Name: "Authorisation acknowledged by an approver"}
	switch {
	case !approver:
		c.Detail = "the requesting user does not have the approver capability"
	case !acknowledged:
		c.Detail = "the authorisation statement was not acknowledged"
	default:
		c.Passed = true
		c.Detail = "acknowledged: " + AuthorisationStatement
	}
	return c
}
