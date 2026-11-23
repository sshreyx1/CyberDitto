package correlation

import (
	"testing"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
)

func findings() []compliance.Finding {
	return []compliance.Finding{
		{ControlID: "CD-LSA-001", Severity: compliance.SeverityHigh, Status: compliance.StatusFail, Techniques: []string{"T1003.001"}},
		{ControlID: "CD-UAC-001", Severity: compliance.SeverityHigh, Status: compliance.StatusFail, Techniques: []string{"T1548.002"}},
		{ControlID: "CD-LOG-001", Severity: compliance.SeverityMedium, Status: compliance.StatusFail, Techniques: []string{"T1059.001"}},
		{ControlID: "CD-FW-001", Severity: compliance.SeverityHigh, Status: compliance.StatusPass, Techniques: []string{"T1562.004"}},
		{ControlID: "CD-DISK-001", Severity: compliance.SeverityHigh, Status: compliance.StatusUnknown, Techniques: []string{"T1005"}},
	}
}

func TestCorrelate_OnlyFailedControlsBecomeGaps(t *testing.T) {
	gaps := Correlate(findings(), nil)
	if len(gaps) != 3 {
		t.Fatalf("gaps = %d, want 3 (pass and unknown are not gaps)", len(gaps))
	}
}

func TestCorrelate_UntestedEvidenceIsPointThree(t *testing.T) {
	for _, g := range Correlate(findings(), nil) {
		if g.Factors[1].Name != "evidence" || g.Factors[1].Value != evidenceUnproven {
			t.Errorf("%s evidence = %+v, want %g untested", g.ControlID, g.Factors[1], evidenceUnproven)
		}
		for _, te := range g.Techniques {
			if te.State != EvidenceUntested {
				t.Errorf("%s/%s = %s with no observations", g.ControlID, te.TechniqueID, te.State)
			}
		}
	}
}

// not_executed (the fake provider), skipped and error are not evidence.
func TestCorrelate_NonExecutionsAreNotEvidence(t *testing.T) {
	obs := []Observation{
		{TechniqueID: "T1003.001", Outcome: emulation.OutcomeNotExecuted, RunID: uuid.New()},
		{TechniqueID: "T1003.001", Outcome: emulation.OutcomeSkipped, RunID: uuid.New()},
		{TechniqueID: "T1003.001", Outcome: emulation.OutcomeError, RunID: uuid.New()},
	}
	for _, g := range Correlate(findings(), obs) {
		if g.ControlID == "CD-LSA-001" && (g.Techniques[0].State != EvidenceUntested || g.Factors[1].Value != evidenceUnproven) {
			t.Fatalf("non-executions counted as evidence: %+v", g)
		}
	}
}

func TestCorrelate_ObservedSuccessRaisesPriorityAndRanksFirst(t *testing.T) {
	obs := []Observation{
		{TechniqueID: "T1059.001", Outcome: emulation.OutcomeMissed, RunID: uuid.New()},
		{TechniqueID: "T1548.002", Outcome: emulation.OutcomeBlocked, RunID: uuid.New()},
	}
	gaps := Correlate(findings(), obs)
	// CD-LOG-001: medium (3) × observed (1.0) = 3.0 outranks the high (5)
	// controls at 5 × 0.3 = 1.5 — proven beats theoretical (§1.1 #2).
	if gaps[0].ControlID != "CD-LOG-001" || gaps[0].Priority != 3 {
		t.Fatalf("top gap = %s (%g), want CD-LOG-001 (3)", gaps[0].ControlID, gaps[0].Priority)
	}
	for _, g := range gaps {
		if g.ControlID == "CD-UAC-001" {
			if g.Techniques[0].State != EvidenceBlocked || g.Priority != 1.5 {
				t.Errorf("blocked-only technique: state %s priority %g", g.Techniques[0].State, g.Priority)
			}
		}
	}
}

func TestCorrelate_DetectedStillCountsAsSucceeded(t *testing.T) {
	obs := []Observation{{TechniqueID: "T1003.001", Outcome: emulation.OutcomeDetected, RunID: uuid.New()}}
	for _, g := range Correlate(findings(), obs) {
		if g.ControlID == "CD-LSA-001" && g.Techniques[0].State != EvidenceSucceeded {
			t.Fatalf("detected = %s, want succeeded (it ran; it was seen)", g.Techniques[0].State)
		}
	}
}

func TestCorrelate_FactorsMultiplyToPriority(t *testing.T) {
	for _, g := range Correlate(findings(), nil) {
		product := 1.0
		for _, f := range g.Factors {
			product *= f.Value
			if f.Explanation == "" {
				t.Errorf("%s: factor %s has no explanation", g.ControlID, f.Name)
			}
		}
		if round2(product) != g.Priority {
			t.Errorf("%s: priority %g != product of factors %g", g.ControlID, g.Priority, product)
		}
	}
}
