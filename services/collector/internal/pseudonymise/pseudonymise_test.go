package pseudonymise

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplace_StablePerKindAndValue(t *testing.T) {
	m := NewMapper()
	a, _ := m.Replace("HOST", "DESKTOP-ABC")
	b, _ := m.Replace("HOST", "DESKTOP-ABC")
	if a != b || !strings.HasPrefix(a, "HOST-") {
		t.Fatalf("Replace not stable: %q vs %q", a, b)
	}
	u, _ := m.Replace("USER", "DESKTOP-ABC")
	if u == a || !strings.HasPrefix(u, "USER-") {
		t.Fatalf("same value under a different kind must get its own placeholder; got %q and %q", a, u)
	}
	other, _ := m.Replace("HOST", "other")
	if other == a {
		t.Fatal("different values got the same placeholder")
	}
}

func TestReplace_EmptyPassesThroughAndBadKindRejected(t *testing.T) {
	m := NewMapper()
	if got, err := m.Replace("HOST", ""); err != nil || got != "" {
		t.Fatalf("Replace(empty) = %q, %v", got, err)
	}
	if _, err := m.Replace("host-x", "v"); err == nil {
		t.Fatal("a kind containing '-' or lowercase must be rejected")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "map.json")
	m := NewMapper()
	host, _ := m.Replace("HOST", "my-host")
	user, _ := m.Replace("USER", "alice")
	if err := m.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm&0o077 != 0 {
		t.Errorf("map file permissions = %v, want owner-only", perm)
	}

	re, err := LoadMapper(path)
	if err != nil {
		t.Fatalf("LoadMapper: %v", err)
	}
	if got, _ := re.Replace("HOST", "my-host"); got != host {
		t.Errorf("host placeholder after reload = %q, want %q", got, host)
	}
	if got, _ := re.Replace("USER", "alice"); got != user {
		t.Errorf("user placeholder after reload = %q, want %q", got, user)
	}
}

func TestLoadMapper_MissingFileAndMalformed(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadMapper(filepath.Join(dir, "none.json")); err != nil {
		t.Fatalf("missing file should start empty: %v", err)
	}
	bad := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(bad, []byte(`{"mapping":{"nodash":"x"}}`), 0o600)
	if _, err := LoadMapper(bad); err == nil {
		t.Fatal("malformed placeholder in map must be rejected")
	}
}
