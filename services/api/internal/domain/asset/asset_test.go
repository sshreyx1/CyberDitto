package asset

import (
	"errors"
	"strings"
	"testing"
)

const validDoc = `{
  "schema_version": "1",
  "hostname": "HOST-59c4a913",
  "collected_at": "2026-11-23T06:00:00Z",
  "fields": {
    "os.version": {"source": "os.version", "collected_at": "2026-11-23T06:00:00Z", "value": {"product_name": "Windows 11 Pro"}},
    "defender.status": {"source": "defender.status", "collected_at": "2026-11-23T06:00:00Z", "error": "module not found"}
  }
}`

func TestParseDocument_Valid(t *testing.T) {
	doc, err := ParseDocument([]byte(validDoc))
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	if !doc.Fields["os.version"].Known() {
		t.Error("os.version should be known")
	}
	if doc.Fields["defender.status"].Known() {
		t.Error("defender.status carries an error and must not be known")
	}
}

func TestParseDocument_Rejects(t *testing.T) {
	cases := map[string]string{
		"empty":           ``,
		"not json":        `not json`,
		"wrong version":   strings.Replace(validDoc, `"schema_version": "1"`, `"schema_version": "2"`, 1),
		"unknown top key": strings.Replace(validDoc, `"hostname"`, `"extra": 1, "hostname"`, 1),
		"no fields":       `{"schema_version":"1","hostname":"h","collected_at":"2026-11-23T06:00:00Z","fields":{}}`,
		"missing time":    `{"schema_version":"1","hostname":"h","fields":{"a":{"source":"a","collected_at":"2026-11-23T06:00:00Z","value":1}}}`,
		"value and error": `{"schema_version":"1","hostname":"h","collected_at":"2026-11-23T06:00:00Z","fields":{"a":{"source":"a","collected_at":"2026-11-23T06:00:00Z","value":1,"error":"x"}}}`,
		"neither":         `{"schema_version":"1","hostname":"h","collected_at":"2026-11-23T06:00:00Z","fields":{"a":{"source":"a","collected_at":"2026-11-23T06:00:00Z"}}}`,
		"null value":      `{"schema_version":"1","hostname":"h","collected_at":"2026-11-23T06:00:00Z","fields":{"a":{"source":"a","collected_at":"2026-11-23T06:00:00Z","value":null}}}`,
		"bad key":         `{"schema_version":"1","hostname":"h","collected_at":"2026-11-23T06:00:00Z","fields":{"Bad Key":{"source":"a","collected_at":"2026-11-23T06:00:00Z","value":1}}}`,
		"no source":       `{"schema_version":"1","hostname":"h","collected_at":"2026-11-23T06:00:00Z","fields":{"a":{"collected_at":"2026-11-23T06:00:00Z","value":1}}}`,
		"trailing data":   validDoc + `{}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseDocument([]byte(body)); !errors.Is(err, ErrInvalidDocument) {
				t.Fatalf("ParseDocument error = %v, want ErrInvalidDocument", err)
			}
		})
	}
}

func TestParseDocument_TooLarge(t *testing.T) {
	big := make([]byte, MaxDocumentBytes+1)
	if _, err := ParseDocument(big); !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("oversize document error = %v, want ErrInvalidDocument", err)
	}
}

func TestValidateName(t *testing.T) {
	if got, err := ValidateName("  laptop-01  "); err != nil || got != "laptop-01" {
		t.Fatalf("ValidateName = %q, %v", got, err)
	}
	if _, err := ValidateName("   "); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("blank name error = %v", err)
	}
	if _, err := ValidateName(strings.Repeat("a", MaxNameLen+1)); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("long name error = %v", err)
	}
}
