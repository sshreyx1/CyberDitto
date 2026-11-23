package collect

// All returns every collector this build knows about, in a fixed order
// (so manifest/dry-run output is stable run to run). This Phase 2 slice
// deliberately covers a subset of MASTER_PLAN §5.10's full L1 category
// list — OS version, services, firewall profiles, Defender status — not
// all of it yet; the remaining categories (audit policy, registry policy
// keys, local users/groups, installed software, BitLocker/Secure
// Boot/TPM/VBS/HVCI/Credential Guard/LSA, SChannel/RDP/WinRM/PowerShell
// logging, NICs/ports/DNS, shares, AppLocker/WDAC) are tracked in
// docs/PROGRESS.md, not forgotten — each is its own Collector added the
// same way these are.
func All() []Collector {
	reg := RegExeReader{}
	ps := SystemPowerShell{}
	return []Collector{
		OSVersionCollector{Reader: reg},
		ServicesCollector{Runner: ps},
		FirewallCollector{Runner: ps},
		DefenderCollector{Runner: ps},
	}
}
