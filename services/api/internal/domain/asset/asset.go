// Package asset holds assets (real machines a workspace tracks) and their
// immutable configuration snapshots, plus validation of the collector's
// snapshot document format. Pure: no I/O (docs/MASTER_PLAN.md §5.1, §5.6).
package asset

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Asset struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// MaxNameLen bounds asset names; they're display labels, not documents.
const MaxNameLen = 200

var ErrInvalidName = errors.New("asset: name must be 1-200 characters")

// ValidateName trims and checks an asset name.
func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > MaxNameLen {
		return "", ErrInvalidName
	}
	return name, nil
}

// Snapshot is one immutable collector upload for an asset. Data is the
// exact bytes that were uploaded and validated; SHA256 is computed over
// those bytes, so a snapshot is content-addressed by what was received.
type Snapshot struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	AssetID     uuid.UUID
	Version     int
	SHA256      string
	Data        json.RawMessage // nil in list/summary results
	CreatedAt   time.Time
}

// Document mirrors the collector's output format, schema version 1
// (services/collector/internal/snapshot). The two modules are deliberately
// separate deployables, so the shape is restated here rather than imported;
// SupportedSchemaVersion is what keeps them honest with each other.
type Document struct {
	SchemaVersion string           `json:"schema_version"`
	Hostname      string           `json:"hostname"`
	CollectedAt   time.Time        `json:"collected_at"`
	Fields        map[string]Field `json:"fields"`
}

// Field is one collected category. Exactly one of Value or Error is set:
// a category the collector couldn't read is "unknown", never a default.
type Field struct {
	Source      string          `json:"source"`
	CollectedAt time.Time       `json:"collected_at"`
	Value       json.RawMessage `json:"value,omitempty"`
	Error       string          `json:"error,omitempty"`
}

// Known reports whether the collector actually read this category.
func (f Field) Known() bool { return f.Error == "" && len(f.Value) > 0 }

const SupportedSchemaVersion = "1"

// MaxDocumentBytes caps an upload. A real Windows host's snapshot with
// every L1 category is in the hundreds of KB; 5 MiB is generous headroom
// without letting an upload be used to exhaust memory.
const MaxDocumentBytes = 5 << 20

var fieldKeyPattern = regexp.MustCompile(`^[a-z0-9_]+(\.[a-z0-9_]+)*$`)

// ErrInvalidDocument wraps every validation failure so callers can map it
// to a 400 without string-matching.
var ErrInvalidDocument = errors.New("asset: invalid snapshot document")

// ParseDocument strictly decodes and validates a collector snapshot. It
// rejects unknown top-level/field keys, unsupported schema versions, and
// fields that claim both a value and an error (or neither).
func ParseDocument(data []byte) (Document, error) {
	if len(data) == 0 {
		return Document{}, fmt.Errorf("%w: empty body", ErrInvalidDocument)
	}
	if len(data) > MaxDocumentBytes {
		return Document{}, fmt.Errorf("%w: larger than %d bytes", ErrInvalidDocument, MaxDocumentBytes)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var doc Document
	if err := dec.Decode(&doc); err != nil {
		return Document{}, fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}
	if dec.More() {
		return Document{}, fmt.Errorf("%w: trailing data after the JSON document", ErrInvalidDocument)
	}

	if doc.SchemaVersion != SupportedSchemaVersion {
		return Document{}, fmt.Errorf("%w: schema_version %q is not supported (want %q)", ErrInvalidDocument, doc.SchemaVersion, SupportedSchemaVersion)
	}
	if doc.CollectedAt.IsZero() {
		return Document{}, fmt.Errorf("%w: collected_at is missing", ErrInvalidDocument)
	}
	if len(doc.Fields) == 0 {
		return Document{}, fmt.Errorf("%w: no fields", ErrInvalidDocument)
	}
	for key, f := range doc.Fields {
		if !fieldKeyPattern.MatchString(key) {
			return Document{}, fmt.Errorf("%w: field key %q is not a dotted lowercase identifier", ErrInvalidDocument, key)
		}
		if f.Source == "" {
			return Document{}, fmt.Errorf("%w: field %q has no source", ErrInvalidDocument, key)
		}
		hasValue := len(f.Value) > 0 && string(f.Value) != "null"
		hasError := f.Error != ""
		if hasValue == hasError {
			return Document{}, fmt.Errorf("%w: field %q must have exactly one of value or error", ErrInvalidDocument, key)
		}
	}
	return doc, nil
}
