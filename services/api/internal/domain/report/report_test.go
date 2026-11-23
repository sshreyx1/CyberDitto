package report

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/correlation"
)

func input(t *testing.T) Input {
	t.Helper()
	data, err := os.ReadFile("../compliance/testdata/weak.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, _ := asset.ParseDocument(data)
	findings := compliance.Evaluate(compliance.Catalog(), compliance.NewFacts(doc, time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)))
	findings[0].Status = compliance.StatusUnknown
	findings[0].Reason = "collector could not read it"
	score, counts := compliance.Score(findings)
	a := compliance.Assessment{ID: uuid.New(), SnapshotID: uuid.New(), Engine: "native", EngineVersion: compliance.EngineVersion, Score: score, Counts: counts, Findings: findings}
	return Input{AssetName: "laptop", Assessment: a, Gaps: correlation.Correlate(findings, nil), GeneratedAt: time.Now()}
}

func roundTrip(t *testing.T, v any) map[string]any {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestSARIF_Shape(t *testing.T) {
	in := input(t)
	s := roundTrip(t, SARIF(in, compliance.Catalog()))
	if s["version"] != "2.1.0" {
		t.Fatalf("version = %v", s["version"])
	}
	run := s["runs"].([]any)[0].(map[string]any)
	rules := run["tool"].(map[string]any)["driver"].(map[string]any)["rules"].([]any)
	if len(rules) != len(compliance.Catalog()) {
		t.Errorf("rules = %d, want one per control", len(rules))
	}
	results := run["results"].([]any)
	if len(results) != in.Assessment.Counts.Failed+in.Assessment.Counts.Unknown {
		t.Errorf("results = %d, want failed+unknown = %d", len(results), in.Assessment.Counts.Failed+in.Assessment.Counts.Unknown)
	}
	levels := map[string]int{}
	for _, r := range results {
		rm := r.(map[string]any)
		levels[rm["level"].(string)]++
		ri := int(rm["ruleIndex"].(float64))
		if rules[ri].(map[string]any)["id"] != rm["ruleId"] {
			t.Errorf("ruleIndex %d does not point at rule %v", ri, rm["ruleId"])
		}
	}
	if levels["none"] != 1 || levels["error"] == 0 || levels["warning"] == 0 {
		t.Errorf("levels = %v (unknown must be 'none', high 'error', medium 'warning')", levels)
	}
}

func TestNavigator_Shape(t *testing.T) {
	in := input(t)
	n := roundTrip(t, Navigator(in))
	if n["domain"] != "enterprise-attack" {
		t.Fatalf("domain = %v", n["domain"])
	}
	techs := n["techniques"].([]any)
	if len(techs) == 0 {
		t.Fatal("no techniques in layer")
	}
	seen := map[string]bool{}
	for _, tt := range techs {
		tm := tt.(map[string]any)
		id := tm["techniqueID"].(string)
		if seen[id] {
			t.Errorf("technique %s listed twice; the layer must aggregate per technique", id)
		}
		seen[id] = true
		if tm["score"].(float64) <= 0 {
			t.Errorf("%s has non-positive score", id)
		}
	}
}

func TestJSON_IncludesEverything(t *testing.T) {
	in := input(t)
	j := roundTrip(t, JSON(in))
	if len(j["findings"].([]any)) != len(in.Assessment.Findings) || len(j["gaps"].([]any)) != len(in.Gaps) {
		t.Fatalf("findings/gaps missing from the JSON report")
	}
}

func TestParseFormat(t *testing.T) {
	for _, s := range []string{"json", "sarif", "navigator"} {
		if _, ok := ParseFormat(s); !ok {
			t.Errorf("%s rejected", s)
		}
	}
	if _, ok := ParseFormat("pdf"); ok {
		t.Error("pdf accepted but not implemented")
	}
}
