package emulation

import (
	"strings"
	"testing"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/twin"
)

func TestDenied_Techniques(t *testing.T) {
	for _, id := range []string{"T1485", "T1486", "T1490", "T1561.001", "T1529", "T1046", "T1070.001"} {
		if denied, _ := Denied(Test{TechniqueID: id}); !denied {
			t.Errorf("technique %s should be denied", id)
		}
	}
	for _, id := range []string{"T1003.001", "T1070.004", "T1059.001", "T14850"} {
		if denied, why := Denied(Test{TechniqueID: id}); denied {
			t.Errorf("technique %s wrongly denied: %s", id, why)
		}
	}
}

func TestDenied_Commands(t *testing.T) {
	deny := []string{
		"vssadmin delete shadows /all /quiet",
		"VSSADMIN.EXE Delete Shadows /For=C:",
		"wmic shadowcopy delete",
		"bcdedit /set {default} recoveryenabled No",
		"wbadmin delete catalog -quiet",
		"cipher /w:C:",
		"format c: /q",
		"shutdown /r /t 0",
		"Restart-Computer -Force",
		"wevtutil cl Security",
		"Clear-EventLog -LogName System",
		`del /f /s /q C:\*`,
		`rd /s /q C:\Windows`,
		`Remove-Item -Recurse -Force C:\Windows\System32\*`,
		`Remove-Item -Recurse -Force "C:\"`,
		"Invoke-WebRequest -Uri x -OutFile y",
		"(New-Object Net.WebClient).DownloadFile('a','b')",
		"curl.exe -o x y",
		"bitsadmin /transfer x",
		"iex (iwr http://example.invalid/x.ps1)",
	}
	for _, cmd := range deny {
		if denied, _ := Denied(Test{TechniqueID: "T1059.001", Command: cmd}); !denied {
			t.Errorf("command should be denied: %q", cmd)
		}
		if denied, _ := Denied(Test{TechniqueID: "T1059.001", Cleanup: cmd}); !denied {
			t.Errorf("cleanup should be denied: %q", cmd)
		}
	}
	allow := []string{
		`Remove-Item -path C:\Windows\Temp\art-marker.txt -Force -ErrorAction Ignore`,
		`del %temp%\sam`,
		`reg save HKLM\sam %temp%\sam`,
		"net user",
		"schtasks /delete /tn spawn /f",
		"format-list",
	}
	for _, cmd := range allow {
		if denied, why := Denied(Test{TechniqueID: "T1059.001", Command: cmd}); denied {
			t.Errorf("command wrongly denied (%s): %q", why, cmd)
		}
	}
}

func TestDenied_InputDefaults(t *testing.T) {
	tc := Test{TechniqueID: "T1059.001", Command: "#{cmd}", Inputs: map[string]Input{"cmd": {Default: "vssadmin delete shadows /all"}}}
	if denied, _ := Denied(tc); !denied {
		t.Error("a destructive input default must be denied even if the command template is innocuous")
	}
}

func goodInput() PreflightInput {
	return PreflightInput{
		TwinState:       twin.StateReady,
		Isolation:       IsolationReport{Verified: true, Detail: "ok"},
		CatalogVerified: true,
		Tests:           []Test{{GUID: "g1", TechniqueID: "T1087.001", Command: "net user", RiskTier: TierBenign, Origin: OriginCatalog}},
		Approver:        true,
		Acknowledged:    true,
	}
}

func failedChecks(checks []Check) []string {
	var out []string
	for _, c := range checks {
		if !c.Passed {
			out = append(out, c.ID)
		}
	}
	return out
}

func TestPreflight_AllChecksPass(t *testing.T) {
	checks, ok := Preflight(goodInput())
	if !ok {
		t.Fatalf("preflight failed: %v", failedChecks(checks))
	}
}

func TestPreflight_EachViolationBlocks(t *testing.T) {
	cases := map[string]func(*PreflightInput){
		"twin.ready":            func(in *PreflightInput) { in.TwinState = twin.StateImaging },
		"lab.isolated":          func(in *PreflightInput) { in.Isolation.Verified = false },
		"catalog.verified":      func(in *PreflightInput) { in.CatalogVerified = false },
		"tests.catalog_only":    func(in *PreflightInput) { in.Tests[0].Origin = OriginDraft },
		"tests.not_destructive": func(in *PreflightInput) { in.Tests[0].RiskTier = TierDestructive },
		"run.authorised":        func(in *PreflightInput) { in.Acknowledged = false },
	}
	for want, mutate := range cases {
		in := goodInput()
		in.Tests = append([]Test(nil), in.Tests...)
		mutate(&in)
		checks, ok := Preflight(in)
		if ok {
			t.Errorf("%s violation did not block the run", want)
		}
		if f := failedChecks(checks); len(f) != 1 || f[0] != want {
			t.Errorf("%s violation: failed checks = %v", want, f)
		}
	}
}

func TestPreflight_DenylistedCommandBlocksEvenWhenTierSaysIntrusive(t *testing.T) {
	in := goodInput()
	in.Tests = []Test{{GUID: "g", TechniqueID: "T1059.001", Command: "vssadmin delete shadows /all", RiskTier: TierIntrusive, Origin: OriginCatalog}}
	checks, ok := Preflight(in)
	if ok || !strings.Contains(strings.Join(failedChecks(checks), ","), "tests.not_destructive") {
		t.Fatalf("denylisted command passed preflight: %v", failedChecks(checks))
	}
}

func TestPreflight_Limits(t *testing.T) {
	in := goodInput()
	in.Tests = nil
	if _, ok := Preflight(in); ok {
		t.Error("a run with no tests must not pass")
	}
	in = goodInput()
	for i := 0; i < MaxTestsPerRun; i++ {
		in.Tests = append(in.Tests, in.Tests[0])
	}
	if _, ok := Preflight(in); ok {
		t.Errorf("a run with %d tests must not pass (limit %d)", len(in.Tests), MaxTestsPerRun)
	}
}

func TestPreflight_UnknownTestBlocks(t *testing.T) {
	in := goodInput()
	in.UnknownTests = []string{"not-a-real-guid"}
	checks, ok := Preflight(in)
	if ok || failedChecks(checks)[0] != "tests.catalog_only" {
		t.Fatalf("an unknown test GUID passed preflight: %v", failedChecks(checks))
	}
}

func TestPreflight_NonApproverBlocks(t *testing.T) {
	in := goodInput()
	in.Approver = false
	if _, ok := Preflight(in); ok {
		t.Fatal("a non-approver passed preflight")
	}
}
