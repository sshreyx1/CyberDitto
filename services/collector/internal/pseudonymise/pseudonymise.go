// Package pseudonymise replaces sensitive strings (hostnames, usernames,
// ...) with stable placeholders before a snapshot leaves the machine,
// keeping the real values only in a local map file that is never uploaded
// or embedded in the snapshot (docs/MASTER_PLAN.md §5.10, §6.5).
package pseudonymise

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// kindPattern: placeholders look like "HOST-1a2b3c4d", so a kind must be
// letters only (the map file recovers each entry's kind from the text
// before the first "-").
var kindPattern = regexp.MustCompile(`^[A-Z]+$`)

// Mapper holds the reversible value<->placeholder map for one machine.
// Not safe for concurrent use; the collector is a single-run CLI.
type Mapper struct {
	valueToPlace map[string]string // kind + "\x00" + value -> placeholder
	placeToValue map[string]string // placeholder -> value
}

func NewMapper() *Mapper {
	return &Mapper{valueToPlace: map[string]string{}, placeToValue: map[string]string{}}
}

// Replace returns value's placeholder for kind (e.g. "HOST", "USER"),
// minting a random one the first time a value is seen, so the same value
// always maps to the same placeholder — within one run, and across runs
// once LoadMapper has restored an earlier map.
func (m *Mapper) Replace(kind, value string) (string, error) {
	if !kindPattern.MatchString(kind) {
		return "", fmt.Errorf("pseudonymise: invalid kind %q", kind)
	}
	if value == "" {
		return value, nil
	}
	key := kind + "\x00" + value
	if p, ok := m.valueToPlace[key]; ok {
		return p, nil
	}
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("pseudonymise: generate placeholder: %w", err)
	}
	p := kind + "-" + hex.EncodeToString(buf[:])
	m.valueToPlace[key] = p
	m.placeToValue[p] = value
	return p, nil
}

type mapFile struct {
	Mapping map[string]string `json:"mapping"` // placeholder -> real value
}

// Save writes the map to path (0600). It contains the real values — that
// is its whole purpose — and must never be uploaded with the snapshot.
func (m *Mapper) Save(path string) error {
	data, err := json.MarshalIndent(mapFile{Mapping: m.placeToValue}, "", "  ")
	if err != nil {
		return fmt.Errorf("pseudonymise: encode map: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("pseudonymise: write map: %w", err)
	}
	return nil
}

// LoadMapper restores a saved map, or returns an empty one if path doesn't
// exist yet.
func LoadMapper(path string) (*Mapper, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return NewMapper(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("pseudonymise: read map: %w", err)
	}
	var file mapFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("pseudonymise: decode map: %w", err)
	}
	m := NewMapper()
	for p, value := range file.Mapping {
		kind, _, ok := strings.Cut(p, "-")
		if !ok || !kindPattern.MatchString(kind) {
			return nil, fmt.Errorf("pseudonymise: malformed placeholder %q in map", p)
		}
		m.placeToValue[p] = value
		m.valueToPlace[kind+"\x00"+value] = p
	}
	return m, nil
}
