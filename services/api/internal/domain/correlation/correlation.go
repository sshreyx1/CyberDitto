// Package correlation joins failed compliance controls with the ATT&CK
// techniques they make easier and with any observed emulation outcomes
// for those techniques, and ranks the result (docs/MASTER_PLAN.md §4.4,
// §5.13). Pure: no I/O.
//
// Priority = severity × evidence × prevalence × exposure. Every factor is
// returned with its value and an explanation; a factor with no data
// behind it is 1.0 and says so, rather than a made-up number.
package correlation

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
)

// Evidence factor values (§4.4: 1 if the technique succeeded on the twin,
// else 0.3).
const (
	evidenceObserved = 1.0
	evidenceUnproven = 0.3
)

// EvidenceState summarises what emulation has shown about one technique.
type EvidenceState string

const (
	// Untested: no run on a twin with this configuration has executed a
	// test for the technique (this includes fake-provider runs, which
	// execute nothing — docs/adr/0014).
	EvidenceUntested EvidenceState = "untested"
	// Succeeded: a test ran and wasn't blocked (missed, logged, or
	// detected — detected still means the technique worked, it was just
	// seen).
	EvidenceSucceeded EvidenceState = "succeeded"
	// Blocked: every test that ran for the technique was blocked.
	EvidenceBlocked EvidenceState = "blocked"
)

// Observation is one executed test's outcome for one technique.
type Observation struct {
	TechniqueID string
	Outcome     emulation.Outcome
	RunID       uuid.UUID
	TestName    string
}

type TechniqueEvidence struct {
	TechniqueID string
	State       EvidenceState
	// Observations that informed State (empty when untested).
	Observations []Observation
}

type Factor struct {
	Name        string
	Value       float64
	Explanation string
}

type Gap struct {
	ControlID   string
	Title       string
	Category    string
	Severity    compliance.Severity
	Reason      string
	Remediation string
	Techniques  []TechniqueEvidence
	Priority    float64
	Factors     []Factor
}

// Correlate builds one Gap per failed finding. observations may include
// outcomes for techniques no finding mentions; those are ignored.
func Correlate(findings []compliance.Finding, observations []Observation) []Gap {
	byTechnique := map[string][]Observation{}
	for _, o := range observations {
		byTechnique[o.TechniqueID] = append(byTechnique[o.TechniqueID], o)
	}

	var gaps []Gap
	for _, f := range findings {
		if f.Status != compliance.StatusFail {
			continue
		}
		g := Gap{
			ControlID: f.ControlID, Title: f.Title, Category: f.Category, Severity: f.Severity,
			Reason: f.Reason, Remediation: f.Remediation,
		}
		anySucceeded := false
		for _, tech := range f.Techniques {
			te := classify(tech, byTechnique[tech])
			anySucceeded = anySucceeded || te.State == EvidenceSucceeded
			g.Techniques = append(g.Techniques, te)
		}

		severity := Factor{Name: "severity", Value: f.Severity.Weight(),
			Explanation: fmt.Sprintf("%s-severity control (weight %g)", f.Severity, f.Severity.Weight())}
		evidence := Factor{Name: "evidence", Value: evidenceUnproven,
			Explanation: "no related technique has been observed succeeding on a twin with this configuration yet"}
		if anySucceeded {
			evidence = Factor{Name: "evidence", Value: evidenceObserved,
				Explanation: "a related technique was observed succeeding on a twin with this configuration"}
		}
		prevalence := Factor{Name: "prevalence", Value: 1,
			Explanation: "unweighted: no technique-prevalence data source is configured"}
		exposure := Factor{Name: "exposure", Value: 1,
			Explanation: "unweighted: asset criticality is not rated"}

		g.Factors = []Factor{severity, evidence, prevalence, exposure}
		g.Priority = round2(severity.Value * evidence.Value * prevalence.Value * exposure.Value)
		gaps = append(gaps, g)
	}

	sort.SliceStable(gaps, func(i, j int) bool {
		if gaps[i].Priority != gaps[j].Priority {
			return gaps[i].Priority > gaps[j].Priority
		}
		return gaps[i].ControlID < gaps[j].ControlID
	})
	return gaps
}

func classify(technique string, obs []Observation) TechniqueEvidence {
	te := TechniqueEvidence{TechniqueID: technique, State: EvidenceUntested}
	executed, blocked := 0, 0
	for _, o := range obs {
		switch o.Outcome {
		case emulation.OutcomeMissed, emulation.OutcomeLogged, emulation.OutcomeDetected:
			executed++
			te.Observations = append(te.Observations, o)
		case emulation.OutcomeBlocked:
			executed++
			blocked++
			te.Observations = append(te.Observations, o)
		}
		// not_executed, skipped and error say nothing about the technique.
	}
	switch {
	case executed == 0:
		te.State = EvidenceUntested
	case blocked == executed:
		te.State = EvidenceBlocked
	default:
		te.State = EvidenceSucceeded
	}
	return te
}

func round2(v float64) float64 { return float64(int64(v*100+0.5)) / 100 }
