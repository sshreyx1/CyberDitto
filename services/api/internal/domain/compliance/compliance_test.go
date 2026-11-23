package compliance

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
)

var evalTime = time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)

// loadFixture reads a synthetic snapshot (testdata/*.json — never real
// scan output, per .gitignore and docs/MASTER_PLAN.md §2.3 #4).
func loadFixture(t *testing.T, name string) Facts {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	doc, err := asset.ParseDocument(data)
	if err != nil {
		t.Fatalf("fixture %s is not a valid collector document: %v", name, err)
	}
	return NewFacts(doc, evalTime)
}

func byID(findings []Finding) map[string]Finding {
	out := map[string]Finding{}
	for _, f := range findings {
		out[f.ControlID] = f
	}
	return out
}

func TestCatalog_Integrity(t *testing.T) {
	idPattern := regexp.MustCompile(`^CD-[A-Z]+-\d{3}$`)
	techPattern := regexp.MustCompile(`^T\d{4}(\.\d{3})?$`)
	seen := map[string]bool{}
	for _, c := range Catalog() {
		if !idPattern.MatchString(c.ID) {
			t.Errorf("%s: id format", c.ID)
		}
		if seen[c.ID] {
			t.Errorf("%s: duplicate id", c.ID)
		}
		seen[c.ID] = true
		if c.Severity.Weight() == 0 {
			t.Errorf("%s: invalid severity %q", c.ID, c.Severity)
		}
		if c.Title == "" || c.Rationale == "" || c.Remediation == "" || c.Category == "" || len(c.Reads) == 0 || c.Check == nil {
			t.Errorf("%s: incomplete control definition", c.ID)
		}
		for _, tech := range c.Techniques {
			if !techPattern.MatchString(tech) {
				t.Errorf("%s: technique %q is not an ATT&CK id", c.ID, tech)
			}
		}
	}
}

// Every control must answer "unknown" — never pass or fail — when none of
// the data it reads was collected. This is the engine's core honesty rule.
func TestCatalog_MissingDataIsUnknownNotAGuess(t *testing.T) {
	doc := asset.Document{SchemaVersion: "1", Fields: map[string]asset.Field{
		"unrelated.category": {Source: "x", Value: json.RawMessage(`1`)},
	}}
	for _, f := range Evaluate(Catalog(), NewFacts(doc, evalTime)) {
		if f.Status != StatusUnknown && f.Status != StatusNotApplicable {
			t.Errorf("%s returned %s with no data collected", f.ControlID, f.Status)
		}
		if f.Status == StatusUnknown && f.Reason == "" {
			t.Errorf("%s: unknown without a reason", f.ControlID)
		}
	}
}

// A category the collector tried and failed to read is also unknown, with
// the collector's error carried into the reason.
func TestCatalog_FailedReadIsUnknownWithReason(t *testing.T) {
	doc := asset.Document{SchemaVersion: "1", Fields: map[string]asset.Field{
		"security.bitlocker": {Source: "security.bitlocker", Error: "powershell: Access denied"},
	}}
	got := byID(Evaluate(Catalog(), NewFacts(doc, evalTime)))["CD-DISK-001"]
	if got.Status != StatusUnknown || !strings.Contains(got.Reason, "Access denied") {
		t.Fatalf("CD-DISK-001 = %+v, want unknown carrying the collector's error", got)
	}
}

func TestEvaluate_HardenedFixturePassesEverythingMeasurable(t *testing.T) {
	findings := Evaluate(Catalog(), loadFixture(t, "hardened.json"))
	for _, f := range findings {
		if f.Status == StatusFail {
			t.Errorf("%s failed on the hardened fixture: %s (observed %v)", f.ControlID, f.Reason, f.Observed)
		}
		if f.Status == StatusUnknown {
			t.Errorf("%s unknown on the hardened fixture: %s", f.ControlID, f.Reason)
		}
	}
	score, counts := Score(findings)
	if score == nil || *score != 100 {
		t.Fatalf("score = %v, want 100", score)
	}
	if counts.Failed != 0 || counts.Passed == 0 {
		t.Fatalf("counts = %+v", counts)
	}
}

func TestEvaluate_WeakFixtureFailsExpectedControls(t *testing.T) {
	got := byID(Evaluate(Catalog(), loadFixture(t, "weak.json")))
	wantFail := []string{
		"CD-FW-001", "CD-FW-002", "CD-AV-001", "CD-AV-002", "CD-UAC-001", "CD-UAC-002", "CD-UAC-003",
		"CD-LSA-001", "CD-LSA-002", "CD-LSA-003", "CD-LSA-004", "CD-VBS-001", "CD-VBS-002",
		"CD-RDP-001", "CD-SMB-001", "CD-SMB-002", "CD-LOG-001", "CD-ACC-001", "CD-ACC-002",
		"CD-ACC-003", "CD-ACC-004", "CD-UPD-001", "CD-BOOT-001", "CD-BOOT-002", "CD-DISK-001",
		"CD-WINRM-001", "CD-AUTO-001", "CD-NET-001",
	}
	for _, id := range wantFail {
		if f := got[id]; f.Status != StatusFail {
			t.Errorf("%s = %s (%s), want fail on the weak fixture", id, f.Status, f.Reason)
		} else if f.Reason == "" || f.Observed == nil {
			t.Errorf("%s failed without a reason and observed evidence", id)
		}
	}
	if len(wantFail) != len(Catalog()) {
		t.Errorf("weak fixture covers %d controls but the catalog has %d; extend the fixture", len(wantFail), len(Catalog()))
	}
	score, _ := Score(Evaluate(Catalog(), loadFixture(t, "weak.json")))
	if score == nil || *score != 0 {
		t.Fatalf("score = %v, want 0", score)
	}
}

func TestCredentialGuard_NotApplicableOnHome(t *testing.T) {
	doc := asset.Document{SchemaVersion: "1", Fields: map[string]asset.Field{
		"os.version":            {Source: "os.version", Value: json.RawMessage(`{"product_name":"Windows 11 Home","current_build":"26100"}`)},
		"security.device_guard": {Source: "x", Value: json.RawMessage(`{"vbs_status":2,"services_configured":[2],"services_running":[2]}`)},
	}}
	if got := byID(Evaluate(Catalog(), NewFacts(doc, evalTime)))["CD-VBS-001"]; got.Status != StatusNotApplicable {
		t.Fatalf("CD-VBS-001 on Home = %s, want not_applicable", got.Status)
	}
}

func TestSMBSigning_UnsetDependsOnBuild(t *testing.T) {
	mk := func(build string) Facts {
		return NewFacts(asset.Document{SchemaVersion: "1", Fields: map[string]asset.Field{
			"os.version":  {Source: "os.version", Value: json.RawMessage(`{"product_name":"Windows 11 Pro","current_build":"` + build + `"}`)},
			"network.smb": {Source: "x", Value: json.RawMessage(`{"` + jsonKey(keyLanmanServer) + `":{"present":true,"values":{}}}`)},
		}}, evalTime)
	}
	if got := checkSMBSigning(mk("26100")); got.Status != StatusPass {
		t.Errorf("unset on 24H2 = %s, want pass", got.Status)
	}
	if got := checkSMBSigning(mk("22631")); got.Status != StatusFail {
		t.Errorf("unset on 23H2 = %s, want fail", got.Status)
	}
}

func jsonKey(k string) string { return strings.ReplaceAll(k, `\`, `\\`) }

func TestScore_FormulaAndExclusions(t *testing.T) {
	findings := []Finding{
		{Severity: SeverityHigh, Status: StatusPass},    // 5 of 5
		{Severity: SeverityMedium, Status: StatusFail},  // 0 of 3
		{Severity: SeverityLow, Status: StatusPass},     // 1 of 1
		{Severity: SeverityHigh, Status: StatusUnknown}, // excluded
		{Severity: SeverityHigh, Status: StatusNotApplicable},
	}
	score, c := Score(findings)
	// 100 × (1 − 3/9) = 66.67
	if score == nil || *score != 66.67 {
		t.Fatalf("score = %v, want 66.67", score)
	}
	if c != (Counts{Passed: 2, Failed: 1, Unknown: 1, NotApplicable: 1}) {
		t.Fatalf("counts = %+v", c)
	}
	if s, _ := Score([]Finding{{Severity: SeverityHigh, Status: StatusUnknown}}); s != nil {
		t.Fatalf("score with nothing measurable = %v, want nil", *s)
	}
}

func TestParseRegNumber(t *testing.T) {
	for in, want := range map[string]int64{"0x1": 1, "0xff": 255, "0X2": 2, "5": 5} {
		if got, ok := parseRegNumber(in); !ok || got != want {
			t.Errorf("parseRegNumber(%q) = %d, %v", in, got, ok)
		}
	}
	if _, ok := parseRegNumber("abc"); ok {
		t.Error("parseRegNumber accepted garbage")
	}
}
