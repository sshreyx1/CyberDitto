package collect

import "context"

const firewallScript = `ConvertTo-Json -Compress -InputObject @(Get-NetFirewallProfile | Select-Object Name, Enabled, DefaultInboundAction, DefaultOutboundAction)`

// FirewallProfile is one Windows Firewall profile's state
// (docs/MASTER_PLAN.md §5.10: "firewall profiles + rules" — rule-level
// detail is tracked as a follow-up, see docs/PROGRESS.md; profile state is
// the higher-value signal and lands first). This is also the exact
// category v1 hardcoded as `firewall.enabled = true` regardless of the
// real value (MASTER_PLAN §2.3 #3) — every field here is a real read.
type FirewallProfile struct {
	Name                  string `json:"Name"`
	Enabled               int    `json:"Enabled"` // 1 = enabled, 0 = disabled (NetSecurity's own on/off enum)
	DefaultInboundAction  int    `json:"DefaultInboundAction"`
	DefaultOutboundAction int    `json:"DefaultOutboundAction"`
}

// FirewallCollector reads the Domain/Private/Public firewall profiles'
// enabled state and default actions.
type FirewallCollector struct {
	Runner PowerShellRunner
}

func (FirewallCollector) ID() string { return "firewall.profiles" }

func (FirewallCollector) Description() string {
	return "Windows Firewall profile (Domain/Private/Public) enabled state and default inbound/outbound actions (via Get-NetFirewallProfile)."
}

func (c FirewallCollector) Collect(ctx context.Context) (any, error) {
	var profiles []FirewallProfile
	if err := runJSON(ctx, c.Runner, firewallScript, &profiles); err != nil {
		return nil, err
	}
	return profiles, nil
}
