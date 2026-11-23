package collect

import "context"

const defenderScript = `ConvertTo-Json -Compress -InputObject (Get-MpComputerStatus | Select-Object AntivirusEnabled, RealTimeProtectionEnabled, AntispywareEnabled, IoavProtectionEnabled, NISEnabled, AntivirusSignatureAge)`

// DefenderStatus is Windows Defender's current protection state
// (docs/MASTER_PLAN.md §5.10: "Defender prefs/status"). Get-MpComputerStatus
// always returns exactly one object, so this is decoded directly rather
// than through the @(...)-array wrapping the list collectors use.
type DefenderStatus struct {
	AntivirusEnabled          bool `json:"AntivirusEnabled"`
	RealTimeProtectionEnabled bool `json:"RealTimeProtectionEnabled"`
	AntispywareEnabled        bool `json:"AntispywareEnabled"`
	IoavProtectionEnabled     bool `json:"IoavProtectionEnabled"`
	NISEnabled                bool `json:"NISEnabled"`
	AntivirusSignatureAge     int  `json:"AntivirusSignatureAge"` // days since last signature update
}

// DefenderCollector reads Windows Defender's protection toggles and
// signature freshness. On a machine without the Defender PowerShell module
// (e.g. a third-party AV has replaced it), Collect returns an error and
// the category is reported as unknown — never assumed enabled or disabled.
type DefenderCollector struct {
	Runner PowerShellRunner
}

func (DefenderCollector) ID() string { return "defender.status" }

func (DefenderCollector) Description() string {
	return "Windows Defender protection toggles (real-time, antivirus, antispyware, network inspection) and signature age, via Get-MpComputerStatus."
}

func (c DefenderCollector) Collect(ctx context.Context) (any, error) {
	var status DefenderStatus
	if err := runJSON(ctx, c.Runner, defenderScript, &status); err != nil {
		return nil, err
	}
	return status, nil
}
