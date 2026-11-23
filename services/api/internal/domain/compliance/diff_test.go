package compliance

import "testing"

func TestDiff(t *testing.T) {
	before := Assessment{Findings: []Finding{
		{ControlID: "CD-A-001", Severity: SeverityLow, Status: StatusPass},
		{ControlID: "CD-B-001", Severity: SeverityHigh, Status: StatusFail},
		{ControlID: "CD-C-001", Severity: SeverityHigh, Status: StatusPass},
		{ControlID: "CD-D-001", Severity: SeverityMedium, Status: StatusUnknown},
		{ControlID: "CD-GONE-001", Severity: SeverityLow, Status: StatusPass},
	}}
	after := Assessment{Findings: []Finding{
		{ControlID: "CD-A-001", Severity: SeverityLow, Status: StatusPass},  // unchanged
		{ControlID: "CD-B-001", Severity: SeverityHigh, Status: StatusPass}, // improved
		{ControlID: "CD-C-001", Severity: SeverityHigh, Status: StatusFail}, // regressed
		{ControlID: "CD-D-001", Severity: SeverityMedium, Status: StatusPass},
		{ControlID: "CD-NEW-001", Severity: SeverityLow, Status: StatusFail},
	}}

	changes := Diff(before, after)
	got := map[string]Change{}
	for _, c := range changes {
		got[c.ControlID] = c
	}
	if _, ok := got["CD-A-001"]; ok {
		t.Error("unchanged control reported as a change")
	}
	if got["CD-B-001"].Direction() != "improved" || got["CD-C-001"].Direction() != "regressed" {
		t.Errorf("directions: B=%s C=%s", got["CD-B-001"].Direction(), got["CD-C-001"].Direction())
	}
	if got["CD-D-001"].Direction() != "changed" {
		t.Error("unknown -> pass is a change in what was measured, not an improvement")
	}
	if got["CD-NEW-001"].Before != "" || got["CD-GONE-001"].After != "" {
		t.Error("added/removed controls must have an empty side")
	}
	if len(changes) != 5 || changes[0].ControlID != "CD-C-001" {
		t.Errorf("regressions must sort first; got %+v", changes)
	}
}
