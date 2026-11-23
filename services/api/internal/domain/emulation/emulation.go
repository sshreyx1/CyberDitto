// Package emulation holds the attack-emulation domain: catalog tests,
// scenarios, runs, the destructive-test denylist, and the preflight policy
// engine that enforces the lab safety invariants (docs/MASTER_PLAN.md
// §5.12, §6.4). Pure: no I/O. The policy engine here is the final
// authority on whether a run may start — nothing upstream (UI, API, a
// future AI copilot) can override it.
package emulation

import (
	"time"

	"github.com/google/uuid"
)

type RiskTier string

const (
	TierBenign      RiskTier = "benign"
	TierIntrusive   RiskTier = "intrusive"
	TierDestructive RiskTier = "destructive"
)

// Origin says where a test came from. Only catalog tests may run; drafts
// (e.g. AI-proposed tests) are never executable (§6.4 invariant 8).
type Origin string

const (
	OriginCatalog Origin = "catalog"
	OriginDraft   Origin = "draft"
)

type Input struct {
	Description string `json:"description"`
	Type        string `json:"type"`
	Default     string `json:"default"`
}

// Test is one executable catalog entry (an Atomic Red Team test, pinned).
type Test struct {
	GUID              string           `json:"guid"`
	TechniqueID       string           `json:"technique_id"`
	TechniqueName     string           `json:"technique_name"`
	Name              string           `json:"name"`
	Description       string           `json:"description"`
	Executor          string           `json:"executor"` // powershell | command_prompt
	ElevationRequired bool             `json:"elevation_required"`
	Command           string           `json:"command"`
	Cleanup           string           `json:"cleanup,omitempty"`
	Inputs            map[string]Input `json:"inputs,omitempty"`
	RiskTier          RiskTier         `json:"risk_tier"`
	Origin            Origin           `json:"origin"`
	SourcePath        string           `json:"source_path"`
	CurationNote      string           `json:"curation_note,omitempty"`
}

type Scenario struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TestGUIDs   []string `json:"test_guids"`
}

type RunStatus string

const (
	RunPreflightFailed RunStatus = "preflight_failed"
	RunRunning         RunStatus = "running"
	RunCompleted       RunStatus = "completed"
	RunAborted         RunStatus = "aborted"
	RunFailed          RunStatus = "failed"
)

// Outcome is what happened when a test ran (§4.4 outcome taxonomy).
// NotExecuted is what a provider that cannot really execute anything (the
// fake provider) must report: it never invents a blocked/detected/logged/
// missed outcome (docs/MASTER_PLAN.md §0.4.3, docs/adr/0014).
type Outcome string

const (
	OutcomeBlocked     Outcome = "blocked"
	OutcomeDetected    Outcome = "detected"
	OutcomeLogged      Outcome = "logged"
	OutcomeMissed      Outcome = "missed"
	OutcomeNotExecuted Outcome = "not_executed"
	OutcomeError       Outcome = "error"
	OutcomeSkipped     Outcome = "skipped"
)

type Execution struct {
	TestGUID    string
	TechniqueID string
	TestName    string
	RiskTier    RiskTier
	Outcome     Outcome
	Detail      string
	StartedAt   time.Time
	FinishedAt  time.Time
}

// AuthorisationStatement is what the approver acknowledges for every run
// (§6.4 invariant 4). It is stored verbatim with the run.
const AuthorisationStatement = "I am authorised to run these attack tests against this lab twin, and I understand they execute real attacker techniques inside the isolated lab."

type Run struct {
	ID                     uuid.UUID
	WorkspaceID            uuid.UUID
	TwinID                 uuid.UUID
	ScenarioID             string
	TestGUIDs              []string
	Status                 RunStatus
	Detail                 string
	Preflight              []Check
	CatalogSource          string
	CatalogCommit          string
	Provider               string
	AuthorisedBy           uuid.UUID
	AuthorisationStatement string
	Executions             []Execution
	CreatedAt              time.Time
	FinishedAt             *time.Time
}
