package collect

import (
	"context"
	"testing"
)

// Captured verbatim from `reg query "HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion"`
// on a real Windows 11 host, including an empty-valued entry
// (EditionSubManufacturer) and a REG_DWORD hex value (UBR would look the
// same shape) — the parser must handle both without special-casing them.
const regQueryFixture = "HKEY_LOCAL_MACHINE\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\r\n" +
	"    ProductName    REG_SZ    Windows 11 Pro\r\n" +
	"    CurrentBuild    REG_SZ    26200\r\n" +
	"    UBR    REG_DWORD    0x26f6\r\n" +
	"    DisplayVersion    REG_SZ    25H2\r\n" +
	"    EditionSubManufacturer    REG_SZ    \r\n" +
	"\r\n"

func TestParseRegQueryOutput(t *testing.T) {
	values := parseRegQueryOutput([]byte(regQueryFixture))

	want := map[string]string{
		"ProductName":            "Windows 11 Pro",
		"CurrentBuild":           "26200",
		"UBR":                    "0x26f6",
		"DisplayVersion":         "25H2",
		"EditionSubManufacturer": "",
	}
	for k, v := range want {
		if got := values[k]; got != v {
			t.Errorf("values[%q] = %q, want %q", k, got, v)
		}
	}
	if _, ok := values["HKEY_LOCAL_MACHINE"]; ok {
		t.Error("the header line must not be parsed as a value")
	}
}

type fakeRegistryReader struct {
	values map[string]string
	err    error
}

func (f fakeRegistryReader) ReadKey(ctx context.Context, key string) (map[string]string, error) {
	return f.values, f.err
}

func TestOSVersionCollector_Collect(t *testing.T) {
	c := OSVersionCollector{Reader: fakeRegistryReader{values: map[string]string{
		"ProductName":    "Windows 11 Pro",
		"DisplayVersion": "25H2",
		"CurrentBuild":   "26200",
		"UBR":            "0x26f6",
	}}}

	got, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	want := OSVersion{ProductName: "Windows 11 Pro", DisplayVersion: "25H2", CurrentBuild: "26200", UBR: "0x26f6"}
	if got != want {
		t.Errorf("Collect() = %+v, want %+v", got, want)
	}
}

func TestOSVersionCollector_Collect_ReaderErrorPropagates(t *testing.T) {
	wantErr := errFakeRegistry
	c := OSVersionCollector{Reader: fakeRegistryReader{err: wantErr}}

	_, err := c.Collect(context.Background())
	if err != wantErr {
		t.Fatalf("Collect() error = %v, want %v", err, wantErr)
	}
}

var errFakeRegistry = fakeError("registry unavailable")

type fakeError string

func (e fakeError) Error() string { return string(e) }
