// Package atomic is the AttackCatalog adapter for the pinned Atomic Red
// Team build (docs/adr/0006). The catalog is embedded at compile time and
// verified on load: its SHA-256 must equal CatalogSHA256 and its commit
// must equal PinnedCommit. Changing either is a code change (reviewed),
// and editing catalog.json without updating the hash makes the catalog
// unverified, which preflight refuses to run against.
package atomic

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
)

// Produced by `go run ./cmd/catalog-import` from curation.json.
const (
	PinnedCommit  = "a0e3f9bc69fed2fd711bd1f1aa87a40c33482fe6"
	CatalogSHA256 = "5987b61861afb9da9cda206e391eb656b1ec4620eee0b3581bb4066f01672a93"
)

//go:embed catalog.json
var embedded []byte

type file struct {
	Source     string               `json:"source"`
	Repository string               `json:"repository"`
	Commit     string               `json:"commit"`
	CommitDate string               `json:"commit_date"`
	License    string               `json:"license"`
	Files      map[string]string    `json:"files"`
	Tests      []emulation.Test     `json:"tests"`
	Scenarios  []emulation.Scenario `json:"scenarios"`
}

// Catalog is the loaded, verified-or-not catalog. A catalog that fails
// verification still loads (so the console can show what's wrong), but
// Verified reports false and preflight blocks every run.
type Catalog struct {
	f         file
	verifyErr error
	tests     map[string]emulation.Test
	scenarios map[string]emulation.Scenario
}

// Load parses and verifies the embedded catalog.
func Load() *Catalog { return load(embedded, CatalogSHA256, PinnedCommit) }

func load(data []byte, wantSHA, wantCommit string) *Catalog {
	c := &Catalog{tests: map[string]emulation.Test{}, scenarios: map[string]emulation.Scenario{}}
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if got != wantSHA {
		c.verifyErr = fmt.Errorf("catalog hash %s does not match the pinned %s", got, wantSHA)
	}
	if err := json.Unmarshal(data, &c.f); err != nil {
		c.verifyErr = fmt.Errorf("catalog does not parse: %w", err)
		return c
	}
	if c.verifyErr == nil && c.f.Commit != wantCommit {
		c.verifyErr = fmt.Errorf("catalog commit %s is not the pinned %s", c.f.Commit, wantCommit)
	}
	for _, t := range c.f.Tests {
		// Defence in depth: the importer already refused these, but a
		// catalog is only trusted if it still passes the rules today.
		if c.verifyErr == nil {
			if t.Origin != emulation.OriginCatalog {
				c.verifyErr = fmt.Errorf("test %s has origin %q", t.GUID, t.Origin)
			} else if denied, why := emulation.Denied(t); denied {
				c.verifyErr = fmt.Errorf("test %s is denylisted: %s", t.GUID, why)
			} else if t.RiskTier != emulation.TierBenign && t.RiskTier != emulation.TierIntrusive {
				c.verifyErr = fmt.Errorf("test %s has non-schedulable tier %q", t.GUID, t.RiskTier)
			}
		}
		c.tests[t.GUID] = t
	}
	for _, s := range c.f.Scenarios {
		c.scenarios[s.ID] = s
	}
	return c
}

func (c *Catalog) Source() string     { return c.f.Source }
func (c *Catalog) Repository() string { return c.f.Repository }
func (c *Catalog) Commit() string     { return c.f.Commit }
func (c *Catalog) CommitDate() string { return c.f.CommitDate }
func (c *Catalog) License() string    { return c.f.License }

// Verified reports whether the catalog is the pinned, untampered build.
func (c *Catalog) Verified() (bool, string) {
	if c.verifyErr != nil {
		return false, c.verifyErr.Error()
	}
	return true, fmt.Sprintf("%s@%s, sha256 %s…, %d tests", c.f.Repository, c.f.Commit[:12], CatalogSHA256[:12], len(c.f.Tests))
}

func (c *Catalog) Tests() []emulation.Test { return c.f.Tests }

func (c *Catalog) Test(guid string) (emulation.Test, bool) {
	t, ok := c.tests[guid]
	return t, ok
}

func (c *Catalog) Scenarios() []emulation.Scenario { return c.f.Scenarios }

func (c *Catalog) Scenario(id string) (emulation.Scenario, bool) {
	s, ok := c.scenarios[id]
	return s, ok
}
