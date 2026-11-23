// Package pseudonymise replaces specific sensitive strings (hostnames,
// usernames, MACs, IPs) with stable placeholders before a snapshot leaves
// the machine, keeping the real values only in a local map file that is
// never uploaded or embedded in the snapshot itself
// (docs/MASTER_PLAN.md §5.10, §6.5: "pseudonymisation... with a local
// reversible map that never leaves the machine").
package pseudonymise

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

// Mapper holds the reversible value<->placeholder map for one machine. It
// is not safe for concurrent use — the collector is a single-threaded CLI
// run, not a long-lived service.
type Mapper struct {
	// prefix labels placeholders by what they replace (e.g. "HOST", so a
	// reader of the snapshot can tell a pseudonymised hostname from a
	// pseudonymised username without needing the map itself).
	prefix       string
	valueToPlace map[string]string
	placeToValue map[string]string
}

func NewMapper(prefix string) *Mapper {
	return &Mapper{
		prefix:       prefix,
		valueToPlace: map[string]string{},
		placeToValue: map[string]string{},
	}
}

// Replace returns value's placeholder, minting a new random one and
// recording the mapping the first time a given value is seen, so the same
// real value always maps to the same placeholder within one run (and
// across runs, once Load has restored a prior map).
func (m *Mapper) Replace(value string) (string, error) {
	if value == "" {
		return value, nil
	}
	if placeholder, ok := m.valueToPlace[value]; ok {
		return placeholder, nil
	}

	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("pseudonymise: generate placeholder: %w", err)
	}
	placeholder := fmt.Sprintf("%s-%s", m.prefix, hex.EncodeToString(buf[:]))

	m.valueToPlace[value] = placeholder
	m.placeToValue[placeholder] = value
	return placeholder, nil
}

// mapFile is the on-disk shape: a flat placeholder -> real value map. The
// reverse (value -> placeholder) index is rebuilt from it on Load, so the
// file itself never needs to store both directions.
type mapFile struct {
	Prefix  string            `json:"prefix"`
	Mapping map[string]string `json:"mapping"` // placeholder -> real value
}

// Save writes the current map to path as JSON, creating or overwriting it.
// This file contains real, sensitive values (the whole point of it existing
// is to let an operator reverse a placeholder locally later) and must
// never be uploaded alongside the snapshot it corresponds to.
func (m *Mapper) Save(path string) error {
	data, err := json.MarshalIndent(mapFile{Prefix: m.prefix, Mapping: m.placeToValue}, "", "  ")
	if err != nil {
		return fmt.Errorf("pseudonymise: encode map: %w", err)
	}
	// 0o600: this file holds real hostnames/usernames; only the owner
	// should be able to read it.
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("pseudonymise: write map: %w", err)
	}
	return nil
}

// LoadMapper reads a previously saved map back, so a second collector run
// on the same machine reuses the same placeholders instead of minting new
// ones for values it has already pseudonymised.
func LoadMapper(path string) (*Mapper, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return NewMapper("HOST"), nil
	}
	if err != nil {
		return nil, fmt.Errorf("pseudonymise: read map: %w", err)
	}

	var file mapFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("pseudonymise: decode map: %w", err)
	}

	m := NewMapper(file.Prefix)
	for placeholder, value := range file.Mapping {
		m.placeToValue[placeholder] = value
		m.valueToPlace[value] = placeholder
	}
	return m, nil
}
