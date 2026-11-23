package collect

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mapRegistry struct {
	keys map[string]map[string]string // key -> values; absent key -> ErrKeyNotFound
	fail map[string]error
}

func (m mapRegistry) ReadKey(ctx context.Context, key string) (map[string]string, error) {
	if err, ok := m.fail[key]; ok {
		return nil, err
	}
	v, ok := m.keys[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return v, nil
}

func TestRegistryPolicyCollector_PresentAbsentAndCaseInsensitive(t *testing.T) {
	c := RegistryPolicyCollector{
		Reader: mapRegistry{keys: map[string]map[string]string{
			keyLSA: {"restrictanonymous": "0x0", "RunAsPPL": "0x2", "Unrelated": "x"},
		}},
		CategoryID: "security.lsa",
		Keys: []RegistryKeySpec{
			{Key: keyLSA, Values: []string{"RestrictAnonymous", "RunAsPPL", "NoLMHash"}},
			{Key: keyPSScriptLog, Values: []string{"EnableScriptBlockLogging"}},
		},
	}
	got, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	res := got.(map[string]RegistryKeyResult)

	lsa := res[keyLSA]
	if !lsa.Present || lsa.Values["RestrictAnonymous"] != "0x0" || lsa.Values["RunAsPPL"] != "0x2" {
		t.Errorf("LSA result = %+v", lsa)
	}
	if _, set := lsa.Values["NoLMHash"]; set {
		t.Error("an unset value must be absent, not defaulted")
	}
	if _, leaked := lsa.Values["Unrelated"]; leaked {
		t.Error("values that weren't requested must not be reported")
	}
	if ps := res[keyPSScriptLog]; ps.Present || len(ps.Values) != 0 {
		t.Errorf("missing key should be Present=false with no values, got %+v", ps)
	}
}

func TestRegistryPolicyCollector_ReadFailureMakesCategoryUnknown(t *testing.T) {
	c := RegistryPolicyCollector{
		Reader:     mapRegistry{fail: map[string]error{keyLSA: errors.New("access denied")}},
		CategoryID: "security.lsa",
		Keys:       []RegistryKeySpec{{Key: keyLSA, Values: []string{"RunAsPPL"}}},
	}
	if _, err := c.Collect(context.Background()); err == nil {
		t.Fatal("an unreadable key must fail the category, not report it as not configured")
	}
}

func TestPSJSONCollector_DecodesAndSetsStrictErrors(t *testing.T) {
	ps := &fakePowerShell{output: []byte(`{"vbs_status":2,"services_configured":2,"services_running":[2,3]}`)}
	c := psJSONCollector[DeviceGuard]{Runner: ps, CategoryID: "security.device_guard", Script: "x"}
	got, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	dg := got.(DeviceGuard)
	if dg.VirtualizationBasedSecurityStatus != 2 || len(dg.SecurityServicesConfigured) != 1 || len(dg.SecurityServicesRunning) != 2 {
		t.Errorf("DeviceGuard = %+v (a bare number must decode as a one-element list)", dg)
	}
	if !strings.HasPrefix(ps.script, "$ErrorActionPreference='Stop'") {
		t.Errorf("script must run with terminating errors so failures surface as unknown; got %q", ps.script)
	}
}

func TestAll_IDsUniqueAndDescribed(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range All() {
		if seen[c.ID()] {
			t.Errorf("duplicate category id %q", c.ID())
		}
		seen[c.ID()] = true
		if c.Description() == "" {
			t.Errorf("category %q has no manifest description", c.ID())
		}
	}
}
