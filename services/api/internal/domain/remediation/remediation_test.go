package remediation

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
	"github.com/sshreyx1/cyberditto/services/api/internal/domain/compliance"
)

func controlsByID() map[string]compliance.Control {
	out := map[string]compliance.Control{}
	for _, c := range compliance.Catalog() {
		out[c.ID] = c
	}
	return out
}

func TestTemplates_TargetRealControls(t *testing.T) {
	controls := controlsByID()
	for id, tpl := range Templates() {
		if _, ok := controls[id]; !ok {
			t.Errorf("template %s has no matching compliance control", id)
		}
		if (len(tpl.Registry) == 0) == (tpl.Script == nil) {
			t.Errorf("template %s must have exactly one of Registry or Script", id)
		}
		if tpl.Script != nil && (tpl.Script.Get == "" || tpl.Script.Test == "" || tpl.Script.Set == "" || tpl.Script.Restore == "") {
			t.Errorf("template %s script action is missing a part (every fix needs a rollback)", id)
		}
	}
}

// pathCategory maps a template's registry path to the collector category
// and key spelling the compliance check reads it from.
var pathCategory = map[string]string{
	pathUAC:        "security.uac",
	pathLSA:        "security.lsa",
	pathRDPTcp:     "remote.rdp",
	pathLanmanSrv:  "network.smb",
	pathPSScript:   "logging.powershell",
	pathWinRM:      "remote.winrm",
	pathExplorerPo: "security.autorun",
}

// factsWithRegistry builds a snapshot containing only the given registry
// values (nil value = value absent but key present).
type regValue struct {
	r RegSetting
	v *int
}

func factsWithRegistry(t *testing.T, controlID string, values []regValue) compliance.Facts {
	t.Helper()
	fields := map[string]asset.Field{
		"os.version": {Source: "os.version", Value: json.RawMessage(`{"product_name":"Windows 11 Pro","current_build":"22631"}`)},
	}
	cats := map[string]map[string]any{}
	for _, rv := range values {
		r, v := rv.r, rv.v
		cat := pathCategory[r.Path]
		if cat == "" {
			t.Fatalf("%s: no category mapping for %s", controlID, r.Path)
		}
		key := strings.Replace(r.Path, "HKLM:", "HKLM", 1)
		if cats[cat] == nil {
			cats[cat] = map[string]any{}
		}
		entry, _ := cats[cat][key].(map[string]any)
		if entry == nil {
			entry = map[string]any{"present": true, "values": map[string]string{}}
			cats[cat][key] = entry
		}
		if v != nil {
			entry["values"].(map[string]string)[r.Name] = fmt.Sprintf("0x%x", *v)
		}
	}
	// RDP: Remote Desktop enabled, so the check depends on the NLA value
	// rather than passing because RDP is off.
	if controlID == "CD-RDP-001" {
		cats["remote.rdp"][`HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server`] = map[string]any{"present": true, "values": map[string]string{"fDenyTSConnections": "0x0"}}
	}
	for cat, keys := range cats {
		raw, _ := json.Marshal(keys)
		fields[cat] = asset.Field{Source: cat, Value: raw}
	}
	return compliance.NewFacts(asset.Document{SchemaVersion: "1", Fields: fields}, time.Now())
}

// TestRegistryTemplates_FixMakesTheControlPass: setting exactly the values
// a template sets must turn its control to pass.
func TestRegistryTemplates_FixMakesTheControlPass(t *testing.T) {
	controls := controlsByID()
	for id, tpl := range Templates() {
		if len(tpl.Registry) == 0 {
			continue
		}
		var values []regValue
		for _, r := range tpl.Registry {
			v := r.Value
			values = append(values, regValue{r, &v})
		}
		if got := controls[id].Check(factsWithRegistry(t, id, values)); got.Status != compliance.StatusPass {
			t.Errorf("%s: after applying the template's values the control is %s (%s)", id, got.Status, got.Reason)
		}
	}
}

// TestRegistryTemplates_AgreeWithControlOnEveryValue is the stronger
// contract: for each setting and a range of values, the bundle's
// "already compliant" decision must equal the control's verdict. This is
// what stops the bundle from rewriting a value the assessment passes
// (it caught RunAsPPL = 2 being "fixed" down to 1).
func TestRegistryTemplates_AgreeWithControlOnEveryValue(t *testing.T) {
	controls := controlsByID()
	for id, tpl := range Templates() {
		for _, r := range tpl.Registry {
			probes := []int{0, 1, 2, 3, 4, 5, r.Value}
			for _, v := range probes {
				var values []regValue
				for _, other := range tpl.Registry {
					switch {
					case other.Name == r.Name:
						vv := v
						values = append(values, regValue{other, &vv})
					case !other.AbsentCompliant:
						ov := other.Value
						values = append(values, regValue{other, &ov})
					default:
						values = append(values, regValue{other, nil})
					}
				}
				controlPasses := controls[id].Check(factsWithRegistry(t, id, values)).Status == compliance.StatusPass
				if controlPasses != r.Compliant(v) {
					t.Errorf("%s %s=%d: control passes=%v but bundle compliant=%v", id, r.Name, v, controlPasses, r.Compliant(v))
				}
			}
			// And absence must agree too.
			var values []regValue
			for _, other := range tpl.Registry {
				if other.Name != r.Name && !other.AbsentCompliant {
					ov := other.Value
					values = append(values, regValue{other, &ov})
				} else {
					values = append(values, regValue{other, nil})
				}
			}
			controlPasses := controls[id].Check(factsWithRegistry(t, id, values)).Status == compliance.StatusPass
			if controlPasses != r.AbsentCompliant {
				t.Errorf("%s %s absent: control passes=%v but AbsentCompliant=%v", id, r.Name, controlPasses, r.AbsentCompliant)
			}
		}
	}
}

func weakFindings(t *testing.T) []compliance.Finding {
	t.Helper()
	data, err := os.ReadFile("../compliance/testdata/weak.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := asset.ParseDocument(data)
	if err != nil {
		t.Fatal(err)
	}
	return compliance.Evaluate(compliance.Catalog(), compliance.NewFacts(doc, time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)))
}

func TestRender_SplitsAutomatedAndManual(t *testing.T) {
	findings := weakFindings(t)
	b := Render(Meta{AssessmentID: "0123456789abcdef", SnapshotID: "s", EngineVersion: compliance.EngineVersion, GeneratedAt: time.Now()}, findings)
	if len(b.Automated) != len(Templates()) {
		t.Errorf("automated = %d, want every template (%d) since every control fails", len(b.Automated), len(Templates()))
	}
	if len(b.Automated)+len(b.Manual) != len(compliance.Catalog()) {
		t.Errorf("automated %d + manual %d != %d failed controls", len(b.Automated), len(b.Manual), len(compliance.Catalog()))
	}
	if b.Filename != "cyberditto-remediation-01234567.ps1" {
		t.Errorf("filename = %s", b.Filename)
	}
	for _, want := range []string{"[switch]$Apply", "-Rollback", "Dry run", "Saved the previous values", "Test-IsAdmin"} {
		if !strings.Contains(b.Script, want) {
			t.Errorf("script is missing %q", want)
		}
	}
}

func TestRender_IgnoresNonFailures(t *testing.T) {
	findings := []compliance.Finding{
		{ControlID: "CD-LSA-002", Status: compliance.StatusPass},
		{ControlID: "CD-LSA-003", Status: compliance.StatusUnknown},
		{ControlID: "CD-VBS-001", Status: compliance.StatusNotApplicable},
	}
	b := Render(Meta{AssessmentID: "a"}, findings)
	if len(b.Automated) != 0 || len(b.Manual) != 0 {
		t.Fatalf("non-failing controls were included: %+v %+v", b.Automated, b.Manual)
	}
}

func TestPSQuote(t *testing.T) {
	if got := psQuote("it's"); got != "'it''s'" {
		t.Errorf("psQuote = %s", got)
	}
}

// TestRender_ParsesAsPowerShell runs PowerShell's own parser over generated
// bundles (it does not execute them). Skipped when no PowerShell is
// installed; CI's Linux runners have pwsh.
func TestRender_ParsesAsPowerShell(t *testing.T) {
	exe := ""
	for _, c := range []string{"pwsh", "powershell.exe"} {
		if p, err := exec.LookPath(c); err == nil {
			exe = p
			break
		}
	}
	if exe == "" {
		t.Skip("no PowerShell available to parse the bundle")
	}
	bundles := map[string]Bundle{
		"all failing": Render(Meta{AssessmentID: "aaaaaaaa-bbbb"}, weakFindings(t)),
		"empty":       Render(Meta{AssessmentID: "bbbbbbbb-cccc"}, nil),
		"manual only": Render(Meta{AssessmentID: "cccccccc-dddd"}, []compliance.Finding{{ControlID: "CD-UPD-001", Title: "it's old", Status: compliance.StatusFail, Remediation: "Run Windows Update's \"check\""}}),
	}
	for name, b := range bundles {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "bundle.ps1")
			if err := os.WriteFile(path, []byte(b.Script), 0o644); err != nil {
				t.Fatal(err)
			}
			psPath := path
			if runtime.GOOS == "linux" && strings.HasSuffix(exe, ".exe") {
				out, err := exec.Command("wslpath", "-w", path).Output()
				if err != nil {
					t.Skipf("powershell.exe via WSL interop but wslpath failed: %v", err)
				}
				psPath = strings.TrimSpace(string(out))
			}
			check := fmt.Sprintf(`$e = $null; [void][System.Management.Automation.Language.Parser]::ParseFile('%s', [ref]$null, [ref]$e); if ($e -and $e.Count -gt 0) { $e | ForEach-Object { $_.ToString() }; exit 1 }`,
				strings.ReplaceAll(psPath, "'", "''"))
			out, err := exec.Command(exe, "-NoProfile", "-NonInteractive", "-Command", check).CombinedOutput()
			if err != nil {
				t.Fatalf("generated bundle does not parse:\n%s", out)
			}
		})
	}
}
