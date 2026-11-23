package collect

import (
	"context"
	"errors"
	"strings"
)

// RegistryKeySpec names one key and the value names to report from it.
type RegistryKeySpec struct {
	Key    string
	Values []string
}

// RegistryKeyResult is what was found under one key. Present=false means
// the key doesn't exist (policy not configured); a requested value missing
// from Values means that value isn't set. Both are real readings, distinct
// from the whole category failing.
type RegistryKeyResult struct {
	Present bool              `json:"present"`
	Values  map[string]string `json:"values"`
}

// RegistryPolicyCollector reports specific values from specific registry
// keys — the shape almost every Windows hardening setting takes (UAC, LSA,
// RDP, PowerShell logging, Secure Boot state, ...). Output is keyed by the
// full key path. Value names are matched case-insensitively (the registry
// itself is case-insensitive, and real machines mix casings, e.g.
// "restrictanonymous") and reported under the spelling requested here.
type RegistryPolicyCollector struct {
	Reader     RegistryReader
	CategoryID string
	Summary    string
	Keys       []RegistryKeySpec
}

func (c RegistryPolicyCollector) ID() string          { return c.CategoryID }
func (c RegistryPolicyCollector) Description() string { return c.Summary }

func (c RegistryPolicyCollector) Collect(ctx context.Context) (any, error) {
	out := make(map[string]RegistryKeyResult, len(c.Keys))
	for _, spec := range c.Keys {
		raw, err := c.Reader.ReadKey(ctx, spec.Key)
		if errors.Is(err, ErrKeyNotFound) {
			out[spec.Key] = RegistryKeyResult{Present: false, Values: map[string]string{}}
			continue
		}
		if err != nil {
			// One unreadable key makes the whole category unknown: a
			// partial result would look like "not configured" for the keys
			// we couldn't read, which is exactly the false-negative the
			// collector exists to avoid.
			return nil, err
		}
		lower := make(map[string]string, len(raw))
		for name, v := range raw {
			lower[strings.ToLower(name)] = v
		}
		values := map[string]string{}
		for _, name := range spec.Values {
			if v, ok := lower[strings.ToLower(name)]; ok {
				values[name] = v
			}
		}
		out[spec.Key] = RegistryKeyResult{Present: true, Values: values}
	}
	return out, nil
}

// Registry key paths used by the policy collectors below. Exported names
// aren't needed; they're grouped here so the full read set is auditable in
// one place.
const (
	keyUAC          = `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`
	keyLSA          = `HKLM\SYSTEM\CurrentControlSet\Control\Lsa`
	keyTerminalSrv  = `HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server`
	keyRDPTcp       = `HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp`
	keyRDPPolicy    = `HKLM\SOFTWARE\Policies\Microsoft\Windows NT\Terminal Services`
	keyPSScriptLog  = `HKLM\SOFTWARE\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging`
	keyPSModuleLog  = `HKLM\SOFTWARE\Policies\Microsoft\Windows\PowerShell\ModuleLogging`
	keyPSTranscript = `HKLM\SOFTWARE\Policies\Microsoft\Windows\PowerShell\Transcription`
	keySecureBoot   = `HKLM\SYSTEM\CurrentControlSet\Control\SecureBoot\State`
	keySMB1Client   = `HKLM\SYSTEM\CurrentControlSet\Services\mrxsmb10`
	keyLanmanServer = `HKLM\SYSTEM\CurrentControlSet\Services\LanmanServer\Parameters`
	keyWinRMPolicy  = `HKLM\SOFTWARE\Policies\Microsoft\Windows\WinRM\Service`
	keyAutorun      = `HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\Explorer`
)

func registryPolicyCollectors(reg RegistryReader) []Collector {
	return []Collector{
		RegistryPolicyCollector{
			Reader: reg, CategoryID: "security.uac",
			Summary: "User Account Control policy values (EnableLUA, admin/standard-user consent prompt behaviour, secure desktop).",
			Keys:    []RegistryKeySpec{{Key: keyUAC, Values: []string{"EnableLUA", "ConsentPromptBehaviorAdmin", "ConsentPromptBehaviorUser", "PromptOnSecureDesktop", "FilterAdministratorToken", "LocalAccountTokenFilterPolicy"}}},
		},
		RegistryPolicyCollector{
			Reader: reg, CategoryID: "security.lsa",
			Summary: "LSA protection and authentication-hardening values (RunAsPPL, LM hash storage, LAN Manager auth level, anonymous access restrictions).",
			Keys:    []RegistryKeySpec{{Key: keyLSA, Values: []string{"RunAsPPL", "NoLMHash", "LmCompatibilityLevel", "RestrictAnonymous", "RestrictAnonymousSAM", "EveryoneIncludesAnonymous", "DisableDomainCreds", "LsaCfgFlags"}}},
		},
		RegistryPolicyCollector{
			Reader: reg, CategoryID: "remote.rdp",
			Summary: "Remote Desktop enablement, Network Level Authentication, and RDP security-layer settings (local config and Group Policy).",
			Keys: []RegistryKeySpec{
				{Key: keyTerminalSrv, Values: []string{"fDenyTSConnections"}},
				{Key: keyRDPTcp, Values: []string{"UserAuthentication", "SecurityLayer", "MinEncryptionLevel", "PortNumber"}},
				{Key: keyRDPPolicy, Values: []string{"fDenyTSConnections", "UserAuthentication", "SecurityLayer", "MinEncryptionLevel", "fDisableCdm"}},
			},
		},
		RegistryPolicyCollector{
			Reader: reg, CategoryID: "logging.powershell",
			Summary: "PowerShell script-block logging, module logging, and transcription policy values.",
			Keys: []RegistryKeySpec{
				{Key: keyPSScriptLog, Values: []string{"EnableScriptBlockLogging", "EnableScriptBlockInvocationLogging"}},
				{Key: keyPSModuleLog, Values: []string{"EnableModuleLogging"}},
				{Key: keyPSTranscript, Values: []string{"EnableTranscripting", "EnableInvocationHeader"}},
			},
		},
		RegistryPolicyCollector{
			Reader: reg, CategoryID: "boot.secure_boot",
			Summary: "Whether UEFI Secure Boot is reported as enabled (registry state value; no firmware access).",
			Keys:    []RegistryKeySpec{{Key: keySecureBoot, Values: []string{"UEFISecureBootEnabled"}}},
		},
		RegistryPolicyCollector{
			Reader: reg, CategoryID: "network.smb",
			Summary: "SMBv1 client driver start type and SMB server signing/SMBv1 values.",
			Keys: []RegistryKeySpec{
				{Key: keySMB1Client, Values: []string{"Start"}},
				{Key: keyLanmanServer, Values: []string{"SMB1", "RequireSecuritySignature", "EnableSecuritySignature", "RestrictNullSessAccess", "NullSessionShares"}},
			},
		},
		RegistryPolicyCollector{
			Reader: reg, CategoryID: "remote.winrm",
			Summary: "WinRM service Group Policy values (basic auth, unencrypted traffic, remote management enablement).",
			Keys:    []RegistryKeySpec{{Key: keyWinRMPolicy, Values: []string{"AllowBasic", "AllowUnencryptedTraffic", "AllowAutoConfig", "DisableRunAs"}}},
		},
		RegistryPolicyCollector{
			Reader: reg, CategoryID: "security.autorun",
			Summary: "AutoRun/AutoPlay policy values (NoDriveTypeAutoRun, NoAutorun).",
			Keys:    []RegistryKeySpec{{Key: keyAutorun, Values: []string{"NoDriveTypeAutoRun", "NoAutorun"}}},
		},
	}
}
