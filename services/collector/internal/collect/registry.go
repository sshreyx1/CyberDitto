package collect

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// RegistryReader abstracts a single registry key read so OSVersionCollector
// is testable with fixed data instead of a live Windows registry
// (docs/MASTER_PLAN.md §5.10: "golden-file tests with recorded outputs").
type RegistryReader interface {
	// ReadKey returns every value name -> string value under key, exactly
	// as reg.exe reports them (a REG_DWORD like UBR comes back as a hex
	// string such as "0x26f6" — this package never reinterprets registry
	// types, it only reads and reports them).
	ReadKey(ctx context.Context, key string) (map[string]string, error)
}

// RegExeReader reads the registry by shelling out to reg.exe with a fixed
// argument list (docs/MASTER_PLAN.md §6.3: no shell-string building — key
// is a compile-time constant here, never user input, and is passed as its
// own argv element, never concatenated into a shell command string).
type RegExeReader struct{}

func (RegExeReader) ReadKey(ctx context.Context, key string) (map[string]string, error) {
	// "reg.exe", not "reg": Go's exec.LookPath resolves the bare name via
	// PATHEXT on native Windows, but a WSL2-interop invocation of this
	// same binary needs the extension spelled out to find it at all.
	cmd := exec.CommandContext(ctx, "reg.exe", "query", key)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("reg query %q: %w", key, err)
	}
	return parseRegQueryOutput(out), nil
}

// regLinePattern matches one value line from `reg query`, e.g.:
//
//	ProductName    REG_SZ    Windows 11 Pro
//	EditionSubManufacturer    REG_SZ
//
// The value itself may be empty (as above) or contain internal single
// spaces (e.g. "Windows 11 Pro") — only the name/type separators are runs
// of whitespace wide enough to not appear inside a value in practice, so
// the value is captured as "everything after the type token", not
// re-split on whitespace.
var regLinePattern = regexp.MustCompile(`^\s{2,}(\S+)\s+(REG_\S+)\s*(.*?)\s*$`)

func parseRegQueryOutput(out []byte) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimRight(line, "\r")
		m := regLinePattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		values[m[1]] = m[3]
	}
	return values
}

const currentVersionKey = `HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`

// OSVersion is the OS version collector's reported shape — deliberately
// just the raw registry strings/hex values, not a parsed/normalised
// struct: normalising (e.g. hex UBR to int) is a presentation concern for
// whatever reads the snapshot later, not something this collector should
// get to interpret and risk getting wrong.
type OSVersion struct {
	ProductName    string `json:"product_name,omitempty"`
	DisplayVersion string `json:"display_version,omitempty"`
	CurrentBuild   string `json:"current_build,omitempty"`
	UBR            string `json:"ubr,omitempty"`
}

// OSVersionCollector reads OS edition/build/UBR
// (docs/MASTER_PLAN.md §5.10's first L1 item) from the registry.
type OSVersionCollector struct {
	Reader RegistryReader
}

func (OSVersionCollector) ID() string { return "os.version" }

func (OSVersionCollector) Description() string {
	return "Windows edition, display version, build number, and update build revision (registry read only; no telemetry, no network call)."
}

func (c OSVersionCollector) Collect(ctx context.Context) (any, error) {
	values, err := c.Reader.ReadKey(ctx, currentVersionKey)
	if err != nil {
		return nil, err
	}
	return OSVersion{
		ProductName:    values["ProductName"],
		DisplayVersion: values["DisplayVersion"],
		CurrentBuild:   values["CurrentBuild"],
		UBR:            values["UBR"],
	}, nil
}
