package collect

// All returns every collector this build knows about, in a fixed order (so
// manifest/dry-run output is stable run to run). Still not every category
// MASTER_PLAN §5.10 lists — see docs/PROGRESS.md for what remains (audit
// policy, scheduled tasks, shares, AppLocker/WDAC, firewall rules, ...).
func All() []Collector {
	reg := RegExeReader{}
	ps := SystemPowerShell{}
	collectors := []Collector{
		OSVersionCollector{Reader: reg},
		ServicesCollector{Runner: ps},
		FirewallCollector{Runner: ps},
		DefenderCollector{Runner: ps},
	}
	collectors = append(collectors, registryPolicyCollectors(reg)...)
	collectors = append(collectors, powershellCategoryCollectors(ps)...)
	return collectors
}
