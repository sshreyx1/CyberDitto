// Package remediation holds the reviewed fix templates for failed controls
// and renders them into an exportable PowerShell bundle
// (docs/MASTER_PLAN.md §5.14). Templates over free-form generation: every
// fix here is a fixed, reviewed change with a recorded rollback. Nothing
// in this package applies anything; CyberDitto never changes a production
// machine itself — a person runs the exported bundle.
package remediation

// RegSetting is one DWORD a template sets. A present value is compliant if
// it's in Accept (default: just Value) or, when Reject is set, if it's not
// in Reject — mirroring exactly what the compliance control accepts, so the
// bundle never "fixes" a value the control already passes (e.g. it must not
// downgrade RunAsPPL=2, UEFI-locked, to 1). AbsentCompliant means Windows'
// default for an unset value already satisfies the control.
type RegSetting struct {
	Path            string // PowerShell provider path, e.g. HKLM:\SYSTEM\...
	Name            string
	Value           int
	AbsentCompliant bool
	Accept          []int
	Reject          []int
}

func (r RegSetting) accepted() []int {
	if len(r.Accept) > 0 {
		return r.Accept
	}
	return []int{r.Value}
}

// Compliant reports whether a present value v satisfies the setting.
func (r RegSetting) Compliant(v int) bool {
	if len(r.Reject) > 0 {
		for _, x := range r.Reject {
			if v == x {
				return false
			}
		}
		return true
	}
	for _, x := range r.accepted() {
		if v == x {
			return true
		}
	}
	return false
}

// ScriptAction is a non-registry fix as four PowerShell scriptblock
// bodies sharing one contract: Get returns the current state (JSON-
// serialisable, saved for rollback), Test($s) says whether $s is already
// compliant, Set applies the fix, Restore($s) puts back a saved state.
type ScriptAction struct {
	Get     string
	Test    string
	Set     string
	Restore string
}

type Template struct {
	ControlID      string
	Summary        string
	Registry       []RegSetting
	Script         *ScriptAction
	RebootRequired bool
	Risk           string
}

const (
	pathUAC        = `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`
	pathLSA        = `HKLM:\SYSTEM\CurrentControlSet\Control\Lsa`
	pathRDPTcp     = `HKLM:\SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp`
	pathLanmanSrv  = `HKLM:\SYSTEM\CurrentControlSet\Services\LanmanServer\Parameters`
	pathPSScript   = `HKLM:\SOFTWARE\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging`
	pathWinRM      = `HKLM:\SOFTWARE\Policies\Microsoft\Windows\WinRM\Service`
	pathExplorerPo = `HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer`
)

// Templates returns every automated fix, keyed by control ID. A failed
// control with no entry here is exported as a manual step.
func Templates() map[string]Template {
	list := []Template{
		{ControlID: "CD-UAC-001", Summary: "Turn User Account Control on (EnableLUA = 1).",
			Registry: []RegSetting{{Path: pathUAC, Name: "EnableLUA", Value: 1, AbsentCompliant: true}}, RebootRequired: true},
		{ControlID: "CD-UAC-002", Summary: "Prompt administrators before elevation (ConsentPromptBehaviorAdmin = 5, the Windows default, if it is 0).",
			Registry: []RegSetting{{Path: pathUAC, Name: "ConsentPromptBehaviorAdmin", Value: 5, AbsentCompliant: true, Reject: []int{0}}}},
		{ControlID: "CD-UAC-003", Summary: "Filter remote local-administrator tokens (LocalAccountTokenFilterPolicy = 0).",
			Registry: []RegSetting{{Path: pathUAC, Name: "LocalAccountTokenFilterPolicy", Value: 0, AbsentCompliant: true, Reject: []int{1}}},
			Risk:     "Tools that rely on remote admin access with a local account (some management agents) stop working; use domain accounts or LAPS-managed accounts for them."},
		{ControlID: "CD-LSA-001", Summary: "Run LSA as a protected process (RunAsPPL = 1; an existing 2, UEFI-locked, is kept).",
			Registry: []RegSetting{{Path: pathLSA, Name: "RunAsPPL", Value: 1, Accept: []int{1, 2}}}, RebootRequired: true,
			Risk: "Unsigned LSA plug-ins and some smart-card or password-filter drivers fail to load; check the CodeIntegrity event log for blocked modules after the reboot."},
		{ControlID: "CD-LSA-002", Summary: "Stop storing LAN Manager password hashes (NoLMHash = 1).",
			Registry: []RegSetting{{Path: pathLSA, Name: "NoLMHash", Value: 1, AbsentCompliant: true}},
			Risk:     "Existing LM hashes are only cleared when each password is next changed."},
		{ControlID: "CD-LSA-003", Summary: "Send NTLMv2 only and refuse LM and NTLMv1 (LmCompatibilityLevel = 5).",
			Registry: []RegSetting{{Path: pathLSA, Name: "LmCompatibilityLevel", Value: 5}},
			Risk:     "Very old clients or appliances that only speak LM/NTLMv1 can no longer authenticate to this machine."},
		{ControlID: "CD-LSA-004", Summary: "Block anonymous enumeration of local accounts (RestrictAnonymousSAM = 1).",
			Registry: []RegSetting{{Path: pathLSA, Name: "RestrictAnonymousSAM", Value: 1, AbsentCompliant: true}}},
		{ControlID: "CD-RDP-001", Summary: "Require Network Level Authentication for Remote Desktop (UserAuthentication = 1).",
			Registry: []RegSetting{{Path: pathRDPTcp, Name: "UserAuthentication", Value: 1, AbsentCompliant: true}},
			Risk:     "If Group Policy sets this, the policy wins; change it there instead. Clients that don't support NLA can no longer connect."},
		{ControlID: "CD-SMB-002", Summary: "Require SMB signing on the SMB server (RequireSecuritySignature = 1).",
			Registry: []RegSetting{{Path: pathLanmanSrv, Name: "RequireSecuritySignature", Value: 1}},
			Risk:     "SMB clients that can't sign (very old devices, some NAS or printers) can no longer connect to shares on this machine."},
		{ControlID: "CD-LOG-001", Summary: "Turn on PowerShell script block logging (EnableScriptBlockLogging = 1).",
			Registry: []RegSetting{{Path: pathPSScript, Name: "EnableScriptBlockLogging", Value: 1}},
			Risk:     "Increases event log volume (Microsoft-Windows-PowerShell/Operational); size the log accordingly."},
		{ControlID: "CD-WINRM-001", Summary: "Disallow WinRM Basic authentication and unencrypted traffic.",
			Registry: []RegSetting{
				{Path: pathWinRM, Name: "AllowBasic", Value: 0, AbsentCompliant: true, Reject: []int{1}},
				{Path: pathWinRM, Name: "AllowUnencryptedTraffic", Value: 0, AbsentCompliant: true, Reject: []int{1}},
			}},
		{ControlID: "CD-AUTO-001", Summary: "Turn off AutoRun for every drive type (NoDriveTypeAutoRun = 0xFF).",
			Registry: []RegSetting{{Path: pathExplorerPo, Name: "NoDriveTypeAutoRun", Value: 0xFF}}},
		{ControlID: "CD-FW-001", Summary: "Turn Windows Firewall on for every profile.", Script: &ScriptAction{
			Get:     `@(Get-NetFirewallProfile | ForEach-Object { [pscustomobject]@{ Name = $_.Name; Enabled = $_.Enabled.ToString() } })`,
			Test:    `param($s) -not ($s | Where-Object { $_.Enabled -eq 'False' })`,
			Set:     `Set-NetFirewallProfile -All -Enabled True`,
			Restore: `param($s) foreach ($p in $s) { Set-NetFirewallProfile -Name $p.Name -Enabled $p.Enabled }`,
		}},
		{ControlID: "CD-FW-002", Summary: "Block inbound connections by default on every profile.", Script: &ScriptAction{
			Get:     `@(Get-NetFirewallProfile | ForEach-Object { [pscustomobject]@{ Name = $_.Name; DefaultInboundAction = $_.DefaultInboundAction.ToString() } })`,
			Test:    `param($s) -not ($s | Where-Object { $_.DefaultInboundAction -eq 'Allow' })`,
			Set:     `Set-NetFirewallProfile -All -DefaultInboundAction Block`,
			Restore: `param($s) foreach ($p in $s) { Set-NetFirewallProfile -Name $p.Name -DefaultInboundAction $p.DefaultInboundAction }`,
		}, Risk: "Inbound services that relied on default-allow need explicit allow rules."},
		{ControlID: "CD-AV-001", Summary: "Turn Defender real-time protection on.", Script: &ScriptAction{
			Get:     `[bool](Get-MpPreference).DisableRealtimeMonitoring`,
			Test:    `param($s) -not $s`,
			Set:     `Set-MpPreference -DisableRealtimeMonitoring $false`,
			Restore: `param($s) Set-MpPreference -DisableRealtimeMonitoring ([bool]$s)`,
		}, Risk: "Fails if another antivirus product is registered as primary, or Tamper Protection blocks the change; turn it on in Windows Security instead."},
		{ControlID: "CD-ACC-001", Summary: "Disable the built-in Guest account.", Script: &ScriptAction{
			Get:     `[bool](Get-LocalUser | Where-Object { $_.SID.Value -like '*-501' }).Enabled`,
			Test:    `param($s) -not $s`,
			Set:     `Get-LocalUser | Where-Object { $_.SID.Value -like '*-501' } | Disable-LocalUser`,
			Restore: `param($s) if ($s) { Get-LocalUser | Where-Object { $_.SID.Value -like '*-501' } | Enable-LocalUser }`,
		}},
		{ControlID: "CD-ACC-002", Summary: "Disable the built-in Administrator account, only if another enabled administrator exists.", Script: &ScriptAction{
			Get:  `[bool](Get-LocalUser | Where-Object { $_.SID.Value -like '*-500' }).Enabled`,
			Test: `param($s) -not $s`,
			Set: `$others = @(Get-LocalGroupMember -SID 'S-1-5-32-544' | Where-Object { $_.SID.Value -notlike '*-500' -and $_.ObjectClass -eq 'User' } | Where-Object {
    $u = Get-LocalUser -SID $_.SID -ErrorAction SilentlyContinue
    ($null -eq $u) -or $u.Enabled
})
if ($others.Count -eq 0) { throw 'refusing: disabling the built-in Administrator would leave no enabled administrator account' }
Get-LocalUser | Where-Object { $_.SID.Value -like '*-500' } | Disable-LocalUser`,
			Restore: `param($s) if ($s) { Get-LocalUser | Where-Object { $_.SID.Value -like '*-500' } | Enable-LocalUser }`,
		}, Risk: "Make sure you can sign in with another administrator account before applying."},
	}
	out := make(map[string]Template, len(list))
	for _, t := range list {
		out[t.ControlID] = t
	}
	return out
}
