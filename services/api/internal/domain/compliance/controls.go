package compliance

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// Registry key paths, exactly as the collector reports them
// (services/collector/internal/collect/registry_policy.go).
const (
	keyUAC          = `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`
	keyLSA          = `HKLM\SYSTEM\CurrentControlSet\Control\Lsa`
	keyTerminalSrv  = `HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server`
	keyRDPTcp       = `HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp`
	keyRDPPolicy    = `HKLM\SOFTWARE\Policies\Microsoft\Windows NT\Terminal Services`
	keyPSScriptLog  = `HKLM\SOFTWARE\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging`
	keySecureBoot   = `HKLM\SYSTEM\CurrentControlSet\Control\SecureBoot\State`
	keySMB1Client   = `HKLM\SYSTEM\CurrentControlSet\Services\mrxsmb10`
	keyLanmanServer = `HKLM\SYSTEM\CurrentControlSet\Services\LanmanServer\Parameters`
	keyWinRMPolicy  = `HKLM\SOFTWARE\Policies\Microsoft\Windows\WinRM\Service`
	keyAutorun      = `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer`
)

// Thresholds a reviewer may reasonably want to tune later; kept as named
// constants so the catalog states them instead of burying magic numbers.
const (
	maxSignatureAgeDays = 7
	maxPatchAgeDays     = 45
	maxLocalAdmins      = 2
	// Windows 11 24H2 (build 26100) made SMB signing required by default.
	buildSMBSigningDefault = 26100
)

// Catalog returns the native engine's controls in a stable order.
func Catalog() []Control {
	return []Control{
		{
			ID: "CD-FW-001", Title: "Windows Firewall is on for every profile", Category: "Network",
			Severity:    SeverityHigh,
			Rationale:   "A disabled profile leaves every listening service reachable from whichever network type that profile covers.",
			Remediation: "Turn the firewall on for the Domain, Private and Public profiles (Windows Security > Firewall & network protection, or Set-NetFirewallProfile -All -Enabled True).",
			Techniques:  []string{"T1562.004"},
			Reads:       []string{"firewall.profiles"},
			Check:       checkFirewallEnabled,
		},
		{
			ID: "CD-FW-002", Title: "Inbound connections are blocked by default", Category: "Network",
			Severity:    SeverityMedium,
			Rationale:   "Default-allow inbound means any new listening service is exposed without anyone deciding it should be.",
			Remediation: "Set the default inbound action to Block on every profile (Set-NetFirewallProfile -All -DefaultInboundAction Block) and allow specific services by rule.",
			Techniques:  []string{"T1133", "T1210"},
			Reads:       []string{"firewall.profiles"},
			Check:       checkFirewallInboundDefault,
		},
		{
			ID: "CD-AV-001", Title: "Defender real-time protection is on", Category: "Endpoint protection",
			Severity:    SeverityHigh,
			Rationale:   "Without real-time scanning, malicious files run before any scan sees them.",
			Remediation: "Enable real-time protection (Set-MpPreference -DisableRealtimeMonitoring $false) or confirm a third-party AV is providing it.",
			Techniques:  []string{"T1562.001", "T1204.002"},
			Reads:       []string{"defender.status"},
			Check:       checkDefenderRealtime,
		},
		{
			ID: "CD-AV-002", Title: fmt.Sprintf("Antivirus signatures are no more than %d days old", maxSignatureAgeDays), Category: "Endpoint protection",
			Severity:    SeverityMedium,
			Rationale:   "Stale signatures miss recently catalogued malware.",
			Remediation: "Update signatures (Update-MpSignature) and check why automatic updates stopped.",
			Techniques:  []string{"T1204.002"},
			Reads:       []string{"defender.status"},
			Check:       checkDefenderSignatureAge,
		},
		{
			ID: "CD-UAC-001", Title: "User Account Control is enabled", Category: "Privilege",
			Severity:    SeverityHigh,
			Rationale:   "With UAC off, every administrator process runs fully elevated and malware inherits that elevation silently.",
			Remediation: "Set EnableLUA to 1 under " + keyUAC + " (or the 'Run all administrators in Admin Approval Mode' policy) and restart.",
			Techniques:  []string{"T1548.002"},
			Reads:       []string{"security.uac"},
			Check:       checkUACEnabled,
		},
		{
			ID: "CD-UAC-002", Title: "Administrators are prompted before elevation", Category: "Privilege",
			Severity:    SeverityMedium,
			Rationale:   "'Elevate without prompting' lets any process an admin runs elevate with no user decision.",
			Remediation: "Set ConsentPromptBehaviorAdmin to 2 (prompt for consent on the secure desktop) or 5 (the Windows default).",
			Techniques:  []string{"T1548.002"},
			Reads:       []string{"security.uac"},
			Check:       checkUACAdminPrompt,
		},
		{
			ID: "CD-UAC-003", Title: "Remote logons by local admins get a filtered token", Category: "Privilege",
			Severity:    SeverityMedium,
			Rationale:   "LocalAccountTokenFilterPolicy=1 gives remote local-admin logons a full admin token, which is what makes pass-the-hash with a local account useful laterally.",
			Remediation: "Remove LocalAccountTokenFilterPolicy or set it to 0 under " + keyUAC + ".",
			Techniques:  []string{"T1550.002", "T1021.002"},
			Reads:       []string{"security.uac"},
			Check:       checkUACTokenFilter,
		},
		{
			ID: "CD-LSA-001", Title: "LSA runs as a protected process", Category: "Credential protection",
			Severity:    SeverityHigh,
			Rationale:   "Protected-process LSA blocks most tools that read credentials out of LSASS memory.",
			Remediation: "Set RunAsPPL to 1 (or 2, UEFI-locked) under " + keyLSA + " and restart; check driver compatibility first.",
			Techniques:  []string{"T1003.001"},
			Reads:       []string{"security.lsa"},
			Check:       checkLSAProtection,
		},
		{
			ID: "CD-LSA-002", Title: "LAN Manager password hashes are not stored", Category: "Credential protection",
			Severity:    SeverityMedium,
			Rationale:   "LM hashes are trivially cracked.",
			Remediation: "Set NoLMHash to 1 under " + keyLSA + " and have users change their passwords afterwards.",
			Techniques:  []string{"T1003.002"},
			Reads:       []string{"security.lsa"},
			Check:       checkNoLMHash,
		},
		{
			ID: "CD-LSA-003", Title: "Only NTLMv2 is sent and LM/NTLMv1 are refused", Category: "Credential protection",
			Severity:    SeverityMedium,
			Rationale:   "Accepting LM/NTLMv1 lets an attacker who can relay or capture authentication downgrade it to something crackable.",
			Remediation: "Set LmCompatibilityLevel to 5 under " + keyLSA + " (Group Policy: 'LAN Manager authentication level' = 'Send NTLMv2 response only. Refuse LM & NTLM').",
			Techniques:  []string{"T1557.001", "T1110.002"},
			Reads:       []string{"security.lsa"},
			Check:       checkLmCompatibility,
		},
		{
			ID: "CD-LSA-004", Title: "Anonymous enumeration of accounts is blocked", Category: "Credential protection",
			Severity:    SeverityLow,
			Rationale:   "Anonymous SAM enumeration hands an unauthenticated attacker a list of usernames to target.",
			Remediation: "Set RestrictAnonymousSAM to 1 under " + keyLSA + " (the Windows default).",
			Techniques:  []string{"T1087.001"},
			Reads:       []string{"security.lsa"},
			Check:       checkRestrictAnonymousSAM,
		},
		{
			ID: "CD-VBS-001", Title: "Credential Guard is running", Category: "Credential protection",
			Severity:    SeverityMedium,
			Rationale:   "Credential Guard isolates derived domain credentials in a VBS enclave LSASS can't hand out.",
			Remediation: "Enable Credential Guard via Group Policy (Device Guard > Turn On Virtualization Based Security) on a supported edition.",
			Techniques:  []string{"T1003.001"},
			Reads:       []string{"security.device_guard", "os.version"},
			Check:       checkCredentialGuard,
		},
		{
			ID: "CD-VBS-002", Title: "Memory integrity (HVCI) is running", Category: "Exploit protection",
			Severity:    SeverityMedium,
			Rationale:   "HVCI stops unsigned or tampered kernel code, closing off a common path from admin to kernel.",
			Remediation: "Turn on Core isolation > Memory integrity in Windows Security, after checking for incompatible drivers.",
			Techniques:  []string{"T1068", "T1014"},
			Reads:       []string{"security.device_guard"},
			Check:       checkHVCI,
		},
		{
			ID: "CD-RDP-001", Title: "Remote Desktop is off, or requires Network Level Authentication", Category: "Remote access",
			Severity:    SeverityHigh,
			Rationale:   "RDP without NLA exposes the logon screen (and pre-auth attack surface) to anyone who can reach port 3389.",
			Remediation: "Disable Remote Desktop if it isn't needed; otherwise require NLA (UserAuthentication=1) and restrict who can reach it.",
			Techniques:  []string{"T1021.001", "T1133"},
			Reads:       []string{"remote.rdp"},
			Check:       checkRDP,
		},
		{
			ID: "CD-SMB-001", Title: "SMBv1 is not in use", Category: "Network",
			Severity:    SeverityHigh,
			Rationale:   "SMBv1 is obsolete and has had wormable remote code execution flaws.",
			Remediation: "Remove the SMB 1.0/CIFS optional feature (Disable-WindowsOptionalFeature -Online -FeatureName SMB1Protocol).",
			Techniques:  []string{"T1210"},
			Reads:       []string{"network.smb"},
			Check:       checkSMBv1,
		},
		{
			ID: "CD-SMB-002", Title: "The SMB server requires signing", Category: "Network",
			Severity:    SeverityMedium,
			Rationale:   "Without required signing, SMB sessions to this machine can be relayed or tampered with in transit.",
			Remediation: "Set RequireSecuritySignature to 1 under " + keyLanmanServer + " (Set-SmbServerConfiguration -RequireSecuritySignature $true).",
			Techniques:  []string{"T1557.001"},
			Reads:       []string{"network.smb", "os.version"},
			Check:       checkSMBSigning,
		},
		{
			ID: "CD-LOG-001", Title: "PowerShell script block logging is on", Category: "Logging",
			Severity:    SeverityMedium,
			Rationale:   "Script block logging records the PowerShell actually executed, including deobfuscated code — the main evidence source for PowerShell-based attacks.",
			Remediation: "Enable 'Turn on PowerShell Script Block Logging' (EnableScriptBlockLogging=1 under " + keyPSScriptLog + ").",
			Techniques:  []string{"T1059.001"},
			Reads:       []string{"logging.powershell"},
			Check:       checkScriptBlockLogging,
		},
		{
			ID: "CD-ACC-001", Title: "The Guest account is disabled", Category: "Accounts",
			Severity:    SeverityHigh,
			Rationale:   "An enabled Guest account is an unauthenticated foothold.",
			Remediation: "Disable-LocalUser -Name Guest.",
			Techniques:  []string{"T1078.001"},
			Reads:       []string{"accounts.local_users"},
			Check:       checkGuestDisabled,
		},
		{
			ID: "CD-ACC-002", Title: "The built-in Administrator account is disabled", Category: "Accounts",
			Severity:    SeverityMedium,
			Rationale:   "The built-in Administrator has a well-known RID and is exempt from lockout by default, which makes it the first account guessed.",
			Remediation: "Use named admin accounts and disable the built-in one (Disable-LocalUser -SID <machine SID>-500).",
			Techniques:  []string{"T1078.003", "T1110"},
			Reads:       []string{"accounts.local_users"},
			Check:       checkBuiltinAdminDisabled,
		},
		{
			ID: "CD-ACC-003", Title: "Every enabled local account requires a password", Category: "Accounts",
			Severity:    SeverityHigh,
			Rationale:   "An account flagged 'password not required' can be given a blank password.",
			Remediation: "Set-LocalUser -Name <name> -PasswordRequired $true for each listed account and set a password.",
			Techniques:  []string{"T1078.003"},
			Reads:       []string{"accounts.local_users"},
			Check:       checkPasswordsRequired,
		},
		{
			ID: "CD-ACC-004", Title: fmt.Sprintf("No more than %d accounts are local administrators", maxLocalAdmins), Category: "Accounts",
			Severity:    SeverityLow,
			Rationale:   "Every extra admin account is another credential that yields full control if stolen.",
			Remediation: "Remove day-to-day accounts from the local Administrators group; use a separate admin account when elevation is needed.",
			Techniques:  []string{"T1078.003"},
			Reads:       []string{"accounts.local_admins"},
			Check:       checkLocalAdminCount,
		},
		{
			ID: "CD-UPD-001", Title: fmt.Sprintf("An update was installed in the last %d days", maxPatchAgeDays), Category: "Patching",
			Severity:    SeverityMedium,
			Rationale:   "Monthly cumulative updates close publicly known vulnerabilities; a machine that has stopped taking them accumulates exploitable ones.",
			Remediation: "Run Windows Update and find out why updates stopped (paused, failing, or blocked by policy).",
			Techniques:  []string{"T1068", "T1203"},
			Reads:       []string{"updates.hotfixes"},
			Check:       checkPatchRecency,
		},
		{
			ID: "CD-BOOT-001", Title: "Secure Boot is on", Category: "Boot integrity",
			Severity:    SeverityMedium,
			Rationale:   "Secure Boot stops unsigned bootloaders and bootkits from loading before Windows.",
			Remediation: "Enable Secure Boot in UEFI firmware settings (requires UEFI mode, not legacy BIOS).",
			Techniques:  []string{"T1542.003"},
			Reads:       []string{"boot.secure_boot"},
			Check:       checkSecureBoot,
		},
		{
			ID: "CD-BOOT-002", Title: "A TPM 2.0 is present and active", Category: "Boot integrity",
			Severity:    SeverityLow,
			Rationale:   "BitLocker key protection, measured boot and several VBS features depend on a working TPM 2.0.",
			Remediation: "Enable the TPM (often 'PTT' or 'fTPM') in firmware settings.",
			Techniques:  []string{"T1542"},
			Reads:       []string{"security.tpm"},
			Check:       checkTPM,
		},
		{
			ID: "CD-DISK-001", Title: "The OS volume is BitLocker-protected", Category: "Data protection",
			Severity:    SeverityHigh,
			Rationale:   "Without disk encryption, a lost or stolen device gives up every file and cached credential on it.",
			Remediation: "Turn on BitLocker (or Device Encryption on Home editions) for the OS drive and escrow the recovery key.",
			Techniques:  []string{"T1005"},
			Reads:       []string{"security.bitlocker"},
			Check:       checkBitLocker,
		},
		{
			ID: "CD-WINRM-001", Title: "WinRM does not allow Basic auth or unencrypted traffic", Category: "Remote access",
			Severity:    SeverityMedium,
			Rationale:   "Basic auth over unencrypted WinRM sends reusable credentials in clear text.",
			Remediation: "Set AllowBasic and AllowUnencryptedTraffic to 0 (or remove them) under " + keyWinRMPolicy + ".",
			Techniques:  []string{"T1021.006", "T1040"},
			Reads:       []string{"remote.winrm"},
			Check:       checkWinRM,
		},
		{
			ID: "CD-AUTO-001", Title: "AutoRun is disabled for all drives", Category: "Removable media",
			Severity:    SeverityLow,
			Rationale:   "AutoRun lets inserted media start code with no user action.",
			Remediation: "Set NoDriveTypeAutoRun to 0xFF under " + keyAutorun + " ('Turn off Autoplay' for all drives).",
			Techniques:  []string{"T1091"},
			Reads:       []string{"security.autorun"},
			Check:       checkAutorun,
		},
		{
			ID: "CD-NET-001", Title: "No legacy clear-text remote-access services are listening", Category: "Network",
			Severity:    SeverityMedium,
			Rationale:   "Telnet, FTP, rsh/rlogin and VNC send credentials in clear text or weakly protected, and are frequent initial-access targets.",
			Remediation: "Stop and remove the listed service; use SSH, SFTP or RDP with NLA instead.",
			Techniques:  []string{"T1021", "T1040"},
			Reads:       []string{"network.listening_ports"},
			Check:       checkLegacyListeners,
		},
	}
}

// --- firewall ---

type firewallProfile struct {
	Name                  string `json:"Name"`
	Enabled               int    `json:"Enabled"`
	DefaultInboundAction  int    `json:"DefaultInboundAction"`
	DefaultOutboundAction int    `json:"DefaultOutboundAction"`
}

// NetSecurity's GpoBoolean: 0 False, 1 True, 2 NotConfigured (falls back
// to the local default, which is on). Action: 0 NotConfigured (inbound
// default is Block), 2 Allow, 4 Block.
const (
	gpoFalse    = 0
	actionAllow = 2
)

func checkFirewallEnabled(f Facts) Result {
	var profiles []firewallProfile
	if ok, why := f.decode("firewall.profiles", &profiles); !ok {
		return unknown(why)
	}
	if len(profiles) == 0 {
		return unknown("no firewall profiles were reported")
	}
	var off []string
	observed := map[string]int{}
	for _, p := range profiles {
		observed[p.Name] = p.Enabled
		if p.Enabled == gpoFalse {
			off = append(off, p.Name)
		}
	}
	const want = "every profile enabled (1) or not configured (2, which defaults to on)"
	if len(off) > 0 {
		return fail(observed, want, "off for: "+strings.Join(off, ", "))
	}
	return pass(observed, want)
}

func checkFirewallInboundDefault(f Facts) Result {
	var profiles []firewallProfile
	if ok, why := f.decode("firewall.profiles", &profiles); !ok {
		return unknown(why)
	}
	if len(profiles) == 0 {
		return unknown("no firewall profiles were reported")
	}
	var allow []string
	observed := map[string]int{}
	for _, p := range profiles {
		observed[p.Name] = p.DefaultInboundAction
		if p.DefaultInboundAction == actionAllow {
			allow = append(allow, p.Name)
		}
	}
	const want = "default inbound action Block (4) or not configured (0, which defaults to Block) on every profile"
	if len(allow) > 0 {
		return fail(observed, want, "default-allow inbound on: "+strings.Join(allow, ", "))
	}
	return pass(observed, want)
}

// --- Defender ---

type defenderStatus struct {
	AntivirusEnabled          bool `json:"AntivirusEnabled"`
	RealTimeProtectionEnabled bool `json:"RealTimeProtectionEnabled"`
	AntivirusSignatureAge     int  `json:"AntivirusSignatureAge"`
}

func checkDefenderRealtime(f Facts) Result {
	var d defenderStatus
	if ok, why := f.decode("defender.status", &d); !ok {
		return unknown(why)
	}
	obs := map[string]bool{"AntivirusEnabled": d.AntivirusEnabled, "RealTimeProtectionEnabled": d.RealTimeProtectionEnabled}
	const want = "AntivirusEnabled and RealTimeProtectionEnabled both true"
	if !d.AntivirusEnabled || !d.RealTimeProtectionEnabled {
		return fail(obs, want, "Defender antivirus or real-time protection is off")
	}
	return pass(obs, want)
}

func checkDefenderSignatureAge(f Facts) Result {
	var d defenderStatus
	if ok, why := f.decode("defender.status", &d); !ok {
		return unknown(why)
	}
	obs := map[string]int{"AntivirusSignatureAge": d.AntivirusSignatureAge}
	want := fmt.Sprintf("AntivirusSignatureAge <= %d days", maxSignatureAgeDays)
	if d.AntivirusSignatureAge > maxSignatureAgeDays {
		return fail(obs, want, fmt.Sprintf("signatures are %d days old", d.AntivirusSignatureAge))
	}
	return pass(obs, want)
}

// --- UAC ---

func checkUACEnabled(f Facts) Result {
	r, ok, why := f.registry("security.uac")
	if !ok {
		return unknown(why)
	}
	v := r.effectiveDword(keyUAC, "EnableLUA", 1)
	const want = "EnableLUA = 1 (Windows default when unset)"
	if v.Value != 1 {
		return fail(v, want, "UAC is turned off")
	}
	return pass(v, want)
}

func checkUACAdminPrompt(f Facts) Result {
	r, ok, why := f.registry("security.uac")
	if !ok {
		return unknown(why)
	}
	v := r.effectiveDword(keyUAC, "ConsentPromptBehaviorAdmin", 5)
	const want = "ConsentPromptBehaviorAdmin != 0 (Windows default 5 when unset)"
	if v.Value == 0 {
		return fail(v, want, "administrators are elevated without any prompt")
	}
	return pass(v, want)
}

func checkUACTokenFilter(f Facts) Result {
	r, ok, why := f.registry("security.uac")
	if !ok {
		return unknown(why)
	}
	v := r.effectiveDword(keyUAC, "LocalAccountTokenFilterPolicy", 0)
	const want = "LocalAccountTokenFilterPolicy != 1 (Windows default 0 when unset)"
	if v.Value == 1 {
		return fail(v, want, "remote local-admin logons receive an unfiltered admin token")
	}
	return pass(v, want)
}

// --- LSA ---

func checkLSAProtection(f Facts) Result {
	r, ok, why := f.registry("security.lsa")
	if !ok {
		return unknown(why)
	}
	v, set := r.dword(keyLSA, "RunAsPPL")
	const want = "RunAsPPL = 1 or 2"
	if !set {
		return fail(effective{Explicit: false}, want, "RunAsPPL is not set; LSA protection is only certain when it is explicitly configured")
	}
	if v != 1 && v != 2 {
		return fail(effective{Value: v, Explicit: true}, want, "LSA protection is configured off")
	}
	return pass(effective{Value: v, Explicit: true}, want)
}

func checkNoLMHash(f Facts) Result {
	r, ok, why := f.registry("security.lsa")
	if !ok {
		return unknown(why)
	}
	v := r.effectiveDword(keyLSA, "NoLMHash", 1)
	const want = "NoLMHash = 1 (Windows default when unset)"
	if v.Value != 1 {
		return fail(v, want, "LM hashes are stored for new passwords")
	}
	return pass(v, want)
}

func checkLmCompatibility(f Facts) Result {
	r, ok, why := f.registry("security.lsa")
	if !ok {
		return unknown(why)
	}
	v := r.effectiveDword(keyLSA, "LmCompatibilityLevel", 3)
	const want = "LmCompatibilityLevel = 5 (Windows defaults to 3 when unset: sends NTLMv2 but still accepts LM/NTLMv1)"
	if v.Value < 5 {
		return fail(v, want, fmt.Sprintf("level %d still accepts weaker LM/NTLM responses", v.Value))
	}
	return pass(v, want)
}

func checkRestrictAnonymousSAM(f Facts) Result {
	r, ok, why := f.registry("security.lsa")
	if !ok {
		return unknown(why)
	}
	v := r.effectiveDword(keyLSA, "RestrictAnonymousSAM", 1)
	const want = "RestrictAnonymousSAM = 1 (Windows default when unset)"
	if v.Value != 1 {
		return fail(v, want, "anonymous users can enumerate local accounts")
	}
	return pass(v, want)
}

// --- VBS ---

type deviceGuard struct {
	VBSStatus          int   `json:"vbs_status"`
	ServicesConfigured []int `json:"services_configured"`
	ServicesRunning    []int `json:"services_running"`
}

// Win32_DeviceGuard service codes.
const (
	dgCredentialGuard = 1
	dgHVCI            = 2
)

func checkCredentialGuard(f Facts) Result {
	if home, known := f.isHomeEdition(); known && home {
		return notApplicable("Credential Guard is not available on Home editions")
	}
	var d deviceGuard
	if ok, why := f.decode("security.device_guard", &d); !ok {
		return unknown(why)
	}
	const want = "Credential Guard (service 1) in services_running"
	if !slices.Contains(d.ServicesRunning, dgCredentialGuard) {
		return fail(d, want, "Credential Guard is not running")
	}
	return pass(d, want)
}

func checkHVCI(f Facts) Result {
	var d deviceGuard
	if ok, why := f.decode("security.device_guard", &d); !ok {
		return unknown(why)
	}
	const want = "memory integrity / HVCI (service 2) in services_running"
	if !slices.Contains(d.ServicesRunning, dgHVCI) {
		return fail(d, want, "memory integrity is not running")
	}
	return pass(d, want)
}

// --- RDP ---

func checkRDP(f Facts) Result {
	r, ok, why := f.registry("remote.rdp")
	if !ok {
		return unknown(why)
	}
	// Group Policy, when set, overrides the local setting.
	deny := r.effectiveDword(keyTerminalSrv, "fDenyTSConnections", 1)
	if v, set := r.dword(keyRDPPolicy, "fDenyTSConnections"); set {
		deny = effective{Value: v, Explicit: true}
	}
	nla := r.effectiveDword(keyRDPTcp, "UserAuthentication", 1)
	if v, set := r.dword(keyRDPPolicy, "UserAuthentication"); set {
		nla = effective{Value: v, Explicit: true}
	}
	obs := map[string]effective{"fDenyTSConnections": deny, "UserAuthentication": nla}
	const want = "Remote Desktop disabled (fDenyTSConnections = 1), or enabled with NLA (UserAuthentication = 1)"
	if deny.Value == 1 {
		return pass(obs, want)
	}
	if nla.Value != 1 {
		return fail(obs, want, "Remote Desktop is enabled without Network Level Authentication")
	}
	return pass(obs, want)
}

// --- SMB ---

// Service start type 4 = disabled.
const serviceDisabled = 4

func checkSMBv1(f Facts) Result {
	r, ok, why := f.registry("network.smb")
	if !ok {
		return unknown(why)
	}
	client := r[keySMB1Client]
	clientStart, clientSet := r.dword(keySMB1Client, "Start")
	server, serverSet := r.dword(keyLanmanServer, "SMB1")
	obs := map[string]any{
		"smb1_client_driver_installed": client.Present,
		"smb1_client_start":            clientStart,
		"smb1_server_value":            server,
		"smb1_server_value_set":        serverSet,
	}
	const want = "SMBv1 client driver absent or disabled (Start = 4), and SMB1 not explicitly enabled on the server"
	var problems []string
	if client.Present && clientSet && clientStart != serviceDisabled {
		problems = append(problems, "the SMBv1 client driver is installed and enabled")
	}
	if serverSet && server == 1 {
		problems = append(problems, "the SMB server has SMB1 explicitly enabled")
	}
	if len(problems) > 0 {
		return fail(obs, want, strings.Join(problems, "; "))
	}
	return pass(obs, want)
}

func checkSMBSigning(f Facts) Result {
	r, ok, why := f.registry("network.smb")
	if !ok {
		return unknown(why)
	}
	if v, set := r.dword(keyLanmanServer, "RequireSecuritySignature"); set {
		e := effective{Value: v, Explicit: true}
		const want = "RequireSecuritySignature = 1"
		if v != 1 {
			return fail(e, want, "the SMB server does not require signing")
		}
		return pass(e, want)
	}
	build, known := f.build()
	if !known {
		return unknown("RequireSecuritySignature is unset and the OS build (which decides the default) is unknown")
	}
	obs := map[string]any{"RequireSecuritySignature": "unset", "build": build}
	want := fmt.Sprintf("RequireSecuritySignature = 1 (unset defaults to required from build %d)", buildSMBSigningDefault)
	if build >= buildSMBSigningDefault {
		return pass(obs, want)
	}
	return fail(obs, want, fmt.Sprintf("unset on build %d, where the default is not to require signing", build))
}

// --- logging ---

func checkScriptBlockLogging(f Facts) Result {
	r, ok, why := f.registry("logging.powershell")
	if !ok {
		return unknown(why)
	}
	v := r.effectiveDword(keyPSScriptLog, "EnableScriptBlockLogging", 0)
	const want = "EnableScriptBlockLogging = 1"
	if v.Value != 1 {
		return fail(v, want, "script block logging is not enabled")
	}
	return pass(v, want)
}

// --- accounts ---

type localUser struct {
	Name             string `json:"name"`
	Enabled          bool   `json:"enabled"`
	RID              int    `json:"rid"`
	PasswordRequired bool   `json:"password_required"`
}

const (
	ridAdministrator = 500
	ridGuest         = 501
)

func builtinDisabled(f Facts, rid int, label string) Result {
	var users []localUser
	if ok, why := f.decode("accounts.local_users", &users); !ok {
		return unknown(why)
	}
	for _, u := range users {
		if u.RID == rid {
			obs := map[string]any{"name": u.Name, "enabled": u.Enabled}
			want := "the built-in " + label + " account (RID " + fmt.Sprint(rid) + ") disabled"
			if u.Enabled {
				return fail(obs, want, "the account is enabled")
			}
			return pass(obs, want)
		}
	}
	return pass(map[string]any{"present": false}, "the built-in "+label+" account disabled or absent")
}

func checkGuestDisabled(f Facts) Result { return builtinDisabled(f, ridGuest, "Guest") }
func checkBuiltinAdminDisabled(f Facts) Result {
	return builtinDisabled(f, ridAdministrator, "Administrator")
}

func checkPasswordsRequired(f Facts) Result {
	var users []localUser
	if ok, why := f.decode("accounts.local_users", &users); !ok {
		return unknown(why)
	}
	var noPw []string
	for _, u := range users {
		if u.Enabled && !u.PasswordRequired {
			noPw = append(noPw, u.Name)
		}
	}
	obs := map[string]any{"enabled_accounts_not_requiring_password": noPw}
	const want = "every enabled account has password_required = true"
	if len(noPw) > 0 {
		return fail(obs, want, "enabled accounts with the 'password not required' flag: "+strings.Join(noPw, ", ")+
			" (accounts signed in with a Microsoft account often carry this flag; it still permits a blank local password)")
	}
	return pass(obs, want)
}

type localAdmins struct {
	Total int `json:"total"`
}

func checkLocalAdminCount(f Facts) Result {
	var a localAdmins
	if ok, why := f.decode("accounts.local_admins", &a); !ok {
		return unknown(why)
	}
	want := fmt.Sprintf("<= %d members of the local Administrators group", maxLocalAdmins)
	if a.Total > maxLocalAdmins {
		return fail(a, want, fmt.Sprintf("%d accounts are local administrators", a.Total))
	}
	return pass(a, want)
}

// --- patching ---

type hotfix struct {
	ID          string `json:"id"`
	InstalledOn string `json:"installed_on"`
}

func checkPatchRecency(f Facts) Result {
	var fixes []hotfix
	if ok, why := f.decode("updates.hotfixes", &fixes); !ok {
		return unknown(why)
	}
	var latest time.Time
	var latestID string
	for _, h := range fixes {
		t, err := time.Parse("2006-01-02", h.InstalledOn)
		if err != nil {
			continue
		}
		if t.After(latest) {
			latest, latestID = t, h.ID
		}
	}
	if latest.IsZero() {
		return unknown("no installed update has a recorded install date")
	}
	age := int(f.Now().Sub(latest).Hours() / 24)
	obs := map[string]any{"latest_update": latestID, "installed_on": latest.Format("2006-01-02"), "age_days": age}
	want := fmt.Sprintf("latest update installed within %d days", maxPatchAgeDays)
	if age > maxPatchAgeDays {
		return fail(obs, want, fmt.Sprintf("the newest update is %d days old", age))
	}
	return pass(obs, want)
}

// --- boot / disk ---

func checkSecureBoot(f Facts) Result {
	r, ok, why := f.registry("boot.secure_boot")
	if !ok {
		return unknown(why)
	}
	v, set := r.dword(keySecureBoot, "UEFISecureBootEnabled")
	obs := effective{Value: v, Explicit: set}
	const want = "UEFISecureBootEnabled = 1"
	if !set {
		return fail(obs, want, "no Secure Boot state is recorded (legacy BIOS boot, or Secure Boot unsupported)")
	}
	if v != 1 {
		return fail(obs, want, "Secure Boot is off")
	}
	return pass(obs, want)
}

type tpm struct {
	Present     bool   `json:"present"`
	Enabled     bool   `json:"enabled"`
	Activated   bool   `json:"activated"`
	SpecVersion string `json:"spec_version"`
}

func checkTPM(f Facts) Result {
	var t tpm
	if ok, why := f.decode("security.tpm", &t); !ok {
		return unknown(why)
	}
	const want = "TPM present, enabled, activated, spec version 2.0"
	if !t.Present || !t.Enabled || !t.Activated || !strings.HasPrefix(strings.TrimSpace(t.SpecVersion), "2.0") {
		return fail(t, want, "no active TPM 2.0")
	}
	return pass(t, want)
}

type bitLockerVolume struct {
	MountPoint       string `json:"mount_point"`
	VolumeType       string `json:"volume_type"`
	ProtectionStatus string `json:"protection_status"`
}

func checkBitLocker(f Facts) Result {
	var vols []bitLockerVolume
	if ok, why := f.decode("security.bitlocker", &vols); !ok {
		return unknown(why)
	}
	const want = "the OperatingSystem volume has protection_status On"
	for _, v := range vols {
		if v.VolumeType == "OperatingSystem" {
			if v.ProtectionStatus != "On" {
				return fail(v, want, "the OS volume is not protected")
			}
			return pass(v, want)
		}
	}
	return unknown("no OperatingSystem volume was reported")
}

// --- WinRM ---

func checkWinRM(f Facts) Result {
	r, ok, why := f.registry("remote.winrm")
	if !ok {
		return unknown(why)
	}
	basic := r.effectiveDword(keyWinRMPolicy, "AllowBasic", 0)
	unenc := r.effectiveDword(keyWinRMPolicy, "AllowUnencryptedTraffic", 0)
	obs := map[string]effective{"AllowBasic": basic, "AllowUnencryptedTraffic": unenc}
	const want = "AllowBasic and AllowUnencryptedTraffic not 1 (both default to disallowed when unset)"
	var problems []string
	if basic.Value == 1 {
		problems = append(problems, "Basic authentication is allowed")
	}
	if unenc.Value == 1 {
		problems = append(problems, "unencrypted traffic is allowed")
	}
	if len(problems) > 0 {
		return fail(obs, want, strings.Join(problems, "; "))
	}
	return pass(obs, want)
}

// --- AutoRun ---

func checkAutorun(f Facts) Result {
	r, ok, why := f.registry("security.autorun")
	if !ok {
		return unknown(why)
	}
	// Windows' default NoDriveTypeAutoRun is 0x91 (some drive types still
	// autorun).
	drives := r.effectiveDword(keyAutorun, "NoDriveTypeAutoRun", 0x91)
	noAutorun := r.effectiveDword(keyAutorun, "NoAutorun", 0)
	obs := map[string]effective{"NoDriveTypeAutoRun": drives, "NoAutorun": noAutorun}
	const want = "NoDriveTypeAutoRun = 0xFF, or NoAutorun = 1"
	if drives.Value == 0xFF || noAutorun.Value == 1 {
		return pass(obs, want)
	}
	return fail(obs, want, "AutoRun is still enabled for some drive types")
}

// --- listening services ---

type listeningPort struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Process string `json:"process"`
}

var legacyPorts = map[int]string{
	21: "FTP", 23: "Telnet", 512: "rexec", 513: "rlogin", 514: "rsh", 5800: "VNC (HTTP)", 5900: "VNC",
}

func isLoopback(addr string) bool {
	return addr == "127.0.0.1" || addr == "::1" || strings.HasPrefix(addr, "127.")
}

func checkLegacyListeners(f Facts) Result {
	var ports []listeningPort
	if ok, why := f.decode("network.listening_ports", &ports); !ok {
		return unknown(why)
	}
	var found []listeningPort
	for _, p := range ports {
		if _, legacy := legacyPorts[p.Port]; legacy && !isLoopback(p.Address) {
			found = append(found, p)
		}
	}
	obs := map[string]any{"legacy_listeners": found}
	const want = "none of TCP 21, 23, 512-514, 5800, 5900 listening on a non-loopback address"
	if len(found) > 0 {
		names := make([]string, 0, len(found))
		for _, p := range found {
			names = append(names, fmt.Sprintf("%s on %s:%d (%s)", legacyPorts[p.Port], p.Address, p.Port, p.Process))
		}
		return fail(obs, want, strings.Join(names, "; "))
	}
	return pass(obs, want)
}
