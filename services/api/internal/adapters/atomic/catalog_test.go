package atomic

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain/emulation"
)

func TestEmbeddedCatalog_IsVerified(t *testing.T) {
	c := Load()
	if ok, detail := c.Verified(); !ok {
		t.Fatalf("embedded catalog failed verification: %s (if you re-ran catalog-import, update CatalogSHA256)", detail)
	}
	if c.Commit() != PinnedCommit || len(c.Tests()) == 0 || len(c.Scenarios()) == 0 {
		t.Fatalf("catalog commit %s, %d tests, %d scenarios", c.Commit(), len(c.Tests()), len(c.Scenarios()))
	}
	for _, s := range c.Scenarios() {
		for _, g := range s.TestGUIDs {
			if _, ok := c.Test(g); !ok {
				t.Errorf("scenario %s references unknown test %s", s.ID, g)
			}
		}
	}
}

func hashOf(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// §6.4 invariant 7: a tampered catalog is detected, even a one-byte change.
func TestTamperedCatalog_FailsVerification(t *testing.T) {
	tampered := bytes.Replace(embedded, []byte(`"risk_tier": "benign"`), []byte(`"risk_tier": "benign" `), 1)
	if bytes.Equal(tampered, embedded) {
		t.Fatal("test setup: tampering did not change the bytes")
	}
	if ok, _ := load(tampered, CatalogSHA256, PinnedCommit).Verified(); ok {
		t.Fatal("a modified catalog passed verification")
	}
}

func TestWrongPin_FailsVerification(t *testing.T) {
	if ok, detail := load(embedded, CatalogSHA256, strings.Repeat("0", 40)).Verified(); ok || !strings.Contains(detail, "pinned") {
		t.Fatalf("catalog with a different pin verified: %s", detail)
	}
}

// Even a catalog whose hash matches is rejected if it contains a test the
// current denylist forbids (e.g. the denylist was tightened after import).
func TestCatalogWithDeniedTest_FailsVerification(t *testing.T) {
	bad := bytes.Replace(embedded, []byte(`"command": "`), []byte(`"command": "vssadmin delete shadows /all /quiet & `), 1)
	if bytes.Equal(bad, embedded) {
		t.Fatal("test setup: no command field found to tamper with")
	}
	c := load(bad, hashOf(bad), PinnedCommit)
	if ok, detail := c.Verified(); ok || !strings.Contains(detail, "denylisted") {
		t.Fatalf("catalog containing a destructive command verified: %s", detail)
	}
}

func TestCatalog_OnlySchedulableTiers(t *testing.T) {
	for _, tc := range Load().Tests() {
		if tc.RiskTier != emulation.TierBenign && tc.RiskTier != emulation.TierIntrusive {
			t.Errorf("%s has tier %s", tc.GUID, tc.RiskTier)
		}
		if tc.Origin != emulation.OriginCatalog {
			t.Errorf("%s has origin %s", tc.GUID, tc.Origin)
		}
	}
}
