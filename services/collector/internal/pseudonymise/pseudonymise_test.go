package pseudonymise

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplace_SameValueGetsSamePlaceholder(t *testing.T) {
	m := NewMapper("HOST")

	first, err := m.Replace("DESKTOP-ABC123")
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	second, err := m.Replace("DESKTOP-ABC123")
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	if first != second {
		t.Errorf("Replace returned %q then %q for the same input", first, second)
	}
	if !strings.HasPrefix(first, "HOST-") {
		t.Errorf("placeholder %q does not carry the HOST- prefix", first)
	}
}

func TestReplace_DifferentValuesGetDifferentPlaceholders(t *testing.T) {
	m := NewMapper("HOST")

	a, _ := m.Replace("host-a")
	b, _ := m.Replace("host-b")
	if a == b {
		t.Errorf("two different values got the same placeholder %q", a)
	}
}

func TestReplace_EmptyValuePassesThrough(t *testing.T) {
	m := NewMapper("HOST")
	got, err := m.Replace("")
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	if got != "" {
		t.Errorf("Replace(\"\") = %q, want empty", got)
	}
}

func TestSaveAndLoad_RoundTripsPlaceholders(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "map.json")

	m := NewMapper("HOST")
	placeholder, err := m.Replace("my-real-hostname")
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	if err := m.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat map file: %v", err)
	}
	if perm := info.Mode().Perm(); perm&0o077 != 0 {
		t.Errorf("map file permissions = %v, want no access for group/other", perm)
	}

	reloaded, err := LoadMapper(path)
	if err != nil {
		t.Fatalf("LoadMapper: %v", err)
	}
	again, err := reloaded.Replace("my-real-hostname")
	if err != nil {
		t.Fatalf("Replace after reload: %v", err)
	}
	if again != placeholder {
		t.Errorf("after reload, Replace gave %q, want the original placeholder %q", again, placeholder)
	}
}

func TestLoadMapper_MissingFileStartsEmpty(t *testing.T) {
	m, err := LoadMapper(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err != nil {
		t.Fatalf("LoadMapper on a missing file should not error: %v", err)
	}
	if _, err := m.Replace("anything"); err != nil {
		t.Fatalf("Replace on a freshly-created mapper: %v", err)
	}
}
