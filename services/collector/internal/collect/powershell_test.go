package collect

import (
	"context"
	"testing"
)

type fakePowerShell struct {
	script string // records the last script it was asked to run, for assertions
	output []byte
	err    error
}

func (f *fakePowerShell) Run(ctx context.Context, script string) ([]byte, error) {
	f.script = script
	return f.output, f.err
}

func TestServicesCollector_Collect(t *testing.T) {
	ps := &fakePowerShell{output: []byte(`[{"Name":"Spooler","DisplayName":"Print Spooler","Status":4,"StartType":2}]`)}
	c := ServicesCollector{Runner: ps}

	got, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	services, ok := got.([]Service)
	if !ok || len(services) != 1 {
		t.Fatalf("Collect() = %#v, want one Service", got)
	}
	if services[0].Name != "Spooler" || services[0].Status != 4 {
		t.Errorf("services[0] = %+v, want Name=Spooler Status=4", services[0])
	}
	if ps.script == "" {
		t.Error("ServicesCollector never ran a script")
	}
}

func TestServicesCollector_Collect_EmptyResult(t *testing.T) {
	ps := &fakePowerShell{output: []byte(`[]`)}
	c := ServicesCollector{Runner: ps}

	got, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if services := got.([]Service); len(services) != 0 {
		t.Errorf("got %d services, want 0", len(services))
	}
}

func TestFirewallCollector_Collect(t *testing.T) {
	ps := &fakePowerShell{output: []byte(`[{"Name":"Domain","Enabled":1,"DefaultInboundAction":1,"DefaultOutboundAction":2},{"Name":"Public","Enabled":1,"DefaultInboundAction":1,"DefaultOutboundAction":2}]`)}
	c := FirewallCollector{Runner: ps}

	got, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	profiles := got.([]FirewallProfile)
	if len(profiles) != 2 || profiles[0].Name != "Domain" {
		t.Errorf("Collect() = %+v, want 2 profiles starting with Domain", profiles)
	}
}

func TestDefenderCollector_Collect(t *testing.T) {
	ps := &fakePowerShell{output: []byte(`{"AntivirusEnabled":true,"RealTimeProtectionEnabled":true,"AntispywareEnabled":true,"IoavProtectionEnabled":true,"NISEnabled":true,"AntivirusSignatureAge":1}`)}
	c := DefenderCollector{Runner: ps}

	got, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	status := got.(DefenderStatus)
	if !status.AntivirusEnabled || status.AntivirusSignatureAge != 1 {
		t.Errorf("Collect() = %+v", status)
	}
}

func TestDefenderCollector_Collect_PowerShellFailurePropagates(t *testing.T) {
	ps := &fakePowerShell{err: errFakeRegistry}
	c := DefenderCollector{Runner: ps}

	if _, err := c.Collect(context.Background()); err == nil {
		t.Fatal("expected an error when the PowerShell call itself fails")
	}
}
