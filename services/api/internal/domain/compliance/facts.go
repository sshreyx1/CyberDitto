package compliance

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/asset"
)

// Facts is typed, read-only access to one collector snapshot for controls.
// Every accessor reports whether the data was actually collected, and if
// not, why — that reason becomes the control's "unknown" explanation.
type Facts struct {
	doc asset.Document
	now time.Time
}

func NewFacts(doc asset.Document, now time.Time) Facts {
	return Facts{doc: doc, now: now}
}

// Now is the evaluation time (injected so results are reproducible in
// tests; age-based controls like update recency use it).
func (f Facts) Now() time.Time { return f.now }

// decode loads category into v. ok=false means unknown, with the reason.
func (f Facts) decode(category string, v any) (ok bool, reason string) {
	field, present := f.doc.Fields[category]
	if !present {
		return false, fmt.Sprintf("the snapshot does not include %q (collected by an older collector, or the category was skipped)", category)
	}
	if !field.Known() {
		return false, fmt.Sprintf("the collector could not read %q: %s", category, field.Error)
	}
	if err := json.Unmarshal(field.Value, v); err != nil {
		return false, fmt.Sprintf("%q is not in the expected format: %v", category, err)
	}
	return true, ""
}

// RegKey is one key from a registry-policy category.
type RegKey struct {
	Present bool              `json:"present"`
	Values  map[string]string `json:"values"`
}

type registry map[string]RegKey

func (f Facts) registry(category string) (registry, bool, string) {
	var r registry
	ok, reason := f.decode(category, &r)
	return r, ok, reason
}

// dword returns a numeric registry value. set=false means the value (or
// its whole key) isn't configured, so Windows' default applies.
func (r registry) dword(key, name string) (value int64, set bool) {
	k, ok := r[key]
	if !ok || !k.Present {
		return 0, false
	}
	raw, ok := k.Values[name]
	if !ok {
		return 0, false
	}
	n, ok := parseRegNumber(raw)
	return n, ok
}

// parseRegNumber accepts reg.exe's REG_DWORD rendering ("0x1") and plain
// decimals.
func parseRegNumber(s string) (int64, bool) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(strings.ToLower(s), "0x") {
		n, err := strconv.ParseInt(s[2:], 16, 64)
		return n, err == nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	return n, err == nil
}

// effective is a value's effective setting: the explicit value if set,
// otherwise Windows' documented default.
type effective struct {
	Value    int64 `json:"value"`
	Explicit bool  `json:"explicit"`
}

func (r registry) effectiveDword(key, name string, windowsDefault int64) effective {
	if v, set := r.dword(key, name); set {
		return effective{Value: v, Explicit: true}
	}
	return effective{Value: windowsDefault, Explicit: false}
}

// OSVersion mirrors the collector's os.version category.
type OSVersion struct {
	ProductName    string `json:"product_name"`
	DisplayVersion string `json:"display_version"`
	CurrentBuild   string `json:"current_build"`
	UBR            string `json:"ubr"`
}

func (f Facts) osVersion() (OSVersion, bool) {
	var v OSVersion
	ok, _ := f.decode("os.version", &v)
	return v, ok
}

// build returns the OS build number if known.
func (f Facts) build() (int, bool) {
	v, ok := f.osVersion()
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(v.CurrentBuild)
	return n, err == nil
}

// isHomeEdition reports whether the OS is a Home edition, which lacks some
// enterprise features (Credential Guard, Group Policy-managed settings).
func (f Facts) isHomeEdition() (home bool, known bool) {
	v, ok := f.osVersion()
	if !ok || v.ProductName == "" {
		return false, false
	}
	return strings.Contains(strings.ToLower(v.ProductName), "home"), true
}
