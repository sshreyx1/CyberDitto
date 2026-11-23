// Package compliance is the `native` ComplianceEngine's pure core
// (docs/MASTER_PLAN.md §5.4, §10 Phase 3): a catalog of controls evaluated
// against a collector snapshot, and the posture score (§4.4). No I/O.
//
// Control IDs and wording are CyberDitto's own. They cover ground that
// public hardening guidance also covers, but no benchmark text is copied
// and nothing here claims to implement or certify against any benchmark
// (docs/adr/0008).
package compliance

import (
	"time"

	"github.com/google/uuid"
)

// EngineName and EngineVersion are stored with every assessment so a
// score can always be traced to the exact control set that produced it.
// Bump EngineVersion whenever a control is added, removed, or changes
// what it accepts.
const (
	EngineName    = "native"
	EngineVersion = "2026.11.1"
)

type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
)

// Weight is the severity's contribution to the posture score
// (docs/MASTER_PLAN.md §4.4: High=5, Medium=3, Low=1).
func (s Severity) Weight() float64 {
	switch s {
	case SeverityHigh:
		return 5
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 1
	}
	return 0
}

type Status string

const (
	StatusPass          Status = "pass"
	StatusFail          Status = "fail"
	StatusUnknown       Status = "unknown"
	StatusNotApplicable Status = "not_applicable"
)

// Result is one control's verdict on one snapshot. Observed is the
// evidence — the actual values read from the snapshot — so every verdict
// can be checked by a human rather than trusted.
type Result struct {
	Status   Status
	Observed any
	Expected string
	Reason   string
}

// Control is one check. Check must never guess: if the data it needs
// wasn't collected, the result is unknown, not pass or fail.
type Control struct {
	ID          string
	Title       string
	Category    string
	Severity    Severity
	Rationale   string
	Remediation string
	// Techniques are MITRE ATT&CK technique IDs this control makes harder;
	// Phase 5's gap correlation joins on them.
	Techniques []string
	// Reads lists the snapshot categories the check uses.
	Reads []string
	Check func(Facts) Result
}

// Finding is a control's result plus the control metadata it was judged
// against, as stored with an assessment.
type Finding struct {
	ControlID   string
	Title       string
	Category    string
	Severity    Severity
	Status      Status
	Observed    any
	Expected    string
	Reason      string
	Remediation string
	Techniques  []string
}

type Counts struct {
	Passed        int
	Failed        int
	Unknown       int
	NotApplicable int
}

// Assessment is one evaluation of one snapshot (optionally via a twin).
type Assessment struct {
	ID            uuid.UUID
	WorkspaceID   uuid.UUID
	SnapshotID    uuid.UUID
	TwinID        *uuid.UUID
	Engine        string
	EngineVersion string
	// Score is nil when no control produced a pass or fail (everything
	// unknown or not applicable): there is nothing to score, and 100 or
	// 0 would both be lies.
	Score     *float64
	Counts    Counts
	Findings  []Finding // empty in list results
	CreatedAt time.Time
}

// Score computes the posture score over pass/fail findings only
// (docs/MASTER_PLAN.md §4.4): 100 × (1 − Σ w·fail / Σ w). Unknown and
// not-applicable are excluded and reported separately in Counts.
func Score(findings []Finding) (*float64, Counts) {
	var total, failed float64
	var c Counts
	for _, f := range findings {
		switch f.Status {
		case StatusPass:
			c.Passed++
			total += f.Severity.Weight()
		case StatusFail:
			c.Failed++
			total += f.Severity.Weight()
			failed += f.Severity.Weight()
		case StatusUnknown:
			c.Unknown++
		case StatusNotApplicable:
			c.NotApplicable++
		}
	}
	if total == 0 {
		return nil, c
	}
	s := 100 * (1 - failed/total)
	// Two decimals: stable across runs and storage (numeric(5,2)).
	s = float64(int64(s*100+0.5)) / 100
	return &s, c
}

// Evaluate runs every control in catalog against facts.
func Evaluate(catalog []Control, facts Facts) []Finding {
	out := make([]Finding, 0, len(catalog))
	for _, c := range catalog {
		r := c.Check(facts)
		out = append(out, Finding{
			ControlID: c.ID, Title: c.Title, Category: c.Category, Severity: c.Severity,
			Status: r.Status, Observed: r.Observed, Expected: r.Expected, Reason: r.Reason,
			Remediation: c.Remediation, Techniques: c.Techniques,
		})
	}
	return out
}

func pass(observed any, expected string) Result {
	return Result{Status: StatusPass, Observed: observed, Expected: expected}
}

func fail(observed any, expected, reason string) Result {
	return Result{Status: StatusFail, Observed: observed, Expected: expected, Reason: reason}
}

func unknown(reason string) Result {
	return Result{Status: StatusUnknown, Reason: reason}
}

func notApplicable(reason string) Result {
	return Result{Status: StatusNotApplicable, Reason: reason}
}
