package collect

import (
	"context"
	"encoding/json"
)

// psJSONCollector runs one fixed PowerShell script and decodes its JSON
// output into T. Every script is a compile-time constant (see
// SystemPowerShell for why that matters) and is prefixed with
// $ErrorActionPreference='Stop' so any cmdlet failure — most commonly
// "Access denied" for categories that need elevation — becomes a non-zero
// exit and the category is reported as unknown, never as an empty list
// that would read like "nothing found".
type psJSONCollector[T any] struct {
	Runner     PowerShellRunner
	CategoryID string
	Summary    string
	Script     string
}

func (c psJSONCollector[T]) ID() string          { return c.CategoryID }
func (c psJSONCollector[T]) Description() string { return c.Summary }

func (c psJSONCollector[T]) Collect(ctx context.Context) (any, error) {
	var v T
	if err := runJSON(ctx, c.Runner, "$ErrorActionPreference='Stop'; "+c.Script, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// LocalUser is a local account's name and security-relevant flags. RID is
// the last SID component (500 = built-in Administrator, 501 = Guest,
// >= 1000 = accounts someone created) — enough to identify built-ins
// without collecting the machine SID. No hashes, no last-logon times.
type LocalUser struct {
	Name                 string `json:"name"`
	Enabled              bool   `json:"enabled"`
	RID                  int    `json:"rid"`
	PasswordRequired     bool   `json:"password_required"`
	PasswordNeverExpires bool   `json:"password_never_expires"`
}

// LocalAdmins counts members of the local Administrators group by where
// the account comes from. Counts only: who the admins are is not needed to
// judge "too many admins", so it isn't collected.
type LocalAdmins struct {
	Total            int `json:"total"`
	Local            int `json:"local"`
	ActiveDirectory  int `json:"active_directory"`
	AzureAD          int `json:"azure_ad"`
	MicrosoftAccount int `json:"microsoft_account"`
	Unknown          int `json:"unknown"`
}

type Hotfix struct {
	ID          string `json:"id"`
	InstalledOn string `json:"installed_on,omitempty"` // yyyy-mm-dd; empty when Windows doesn't record it
}

// intList accepts both a JSON array and a bare number: CIM array
// properties with one element don't always serialise as arrays.
type intList []int

func (l *intList) UnmarshalJSON(b []byte) error {
	var many []int
	if err := json.Unmarshal(b, &many); err == nil {
		*l = many
		return nil
	}
	var one int
	if err := json.Unmarshal(b, &one); err != nil {
		if string(b) == "null" {
			*l = nil
			return nil
		}
		return err
	}
	*l = intList{one}
	return nil
}

// DeviceGuard is Win32_DeviceGuard's state. Service codes: 1 = Credential
// Guard, 2 = HVCI (memory integrity), 3 = System Guard Secure Launch,
// 4 = SMM firmware measurement. VBS status: 0 off, 1 configured but not
// running, 2 running.
type DeviceGuard struct {
	VirtualizationBasedSecurityStatus int     `json:"vbs_status"`
	SecurityServicesConfigured        intList `json:"services_configured"`
	SecurityServicesRunning           intList `json:"services_running"`
}

type BitLockerVolume struct {
	MountPoint       string `json:"mount_point"`
	VolumeType       string `json:"volume_type"`
	ProtectionStatus string `json:"protection_status"`
	EncryptionMethod string `json:"encryption_method"`
	VolumeStatus     string `json:"volume_status"`
}

type TPM struct {
	Present     bool   `json:"present"`
	Enabled     bool   `json:"enabled"`
	Activated   bool   `json:"activated"`
	SpecVersion string `json:"spec_version"`
}

type ListeningPort struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Process string `json:"process,omitempty"`
}

type InstalledSoftware struct {
	Name      string `json:"name"`
	Version   string `json:"version,omitempty"`
	Publisher string `json:"publisher,omitempty"`
}

func powershellCategoryCollectors(ps PowerShellRunner) []Collector {
	return []Collector{
		psJSONCollector[[]LocalUser]{
			Runner: ps, CategoryID: "accounts.local_users",
			Summary: "Local user account names with enabled/password flags and RID (no hashes, no SIDs, no logon history).",
			Script:  `ConvertTo-Json -Compress -InputObject @(Get-LocalUser | ForEach-Object { [pscustomobject]@{ name=$_.Name; enabled=[bool]$_.Enabled; rid=[int]$_.SID.Value.Split('-')[-1]; password_required=[bool]$_.PasswordRequired; password_never_expires=(-not $_.PasswordExpires) } })`,
		},
		psJSONCollector[LocalAdmins]{
			Runner: ps, CategoryID: "accounts.local_admins",
			Summary: "How many accounts are in the local Administrators group, by account source (counts only, no names).",
			Script:  `$m = @(Get-LocalGroupMember -SID 'S-1-5-32-544'); ConvertTo-Json -Compress -InputObject ([pscustomobject]@{ total=$m.Count; local=@($m | Where-Object { [int]$_.PrincipalSource -eq 1 }).Count; active_directory=@($m | Where-Object { [int]$_.PrincipalSource -eq 2 }).Count; azure_ad=@($m | Where-Object { [int]$_.PrincipalSource -eq 3 }).Count; microsoft_account=@($m | Where-Object { [int]$_.PrincipalSource -eq 4 }).Count; unknown=@($m | Where-Object { [int]$_.PrincipalSource -eq 0 }).Count })`,
		},
		psJSONCollector[[]Hotfix]{
			Runner: ps, CategoryID: "updates.hotfixes",
			Summary: "Installed Windows update KB numbers and install dates (via Get-HotFix).",
			Script:  `ConvertTo-Json -Compress -InputObject @(Get-HotFix | ForEach-Object { [pscustomobject]@{ id=$_.HotFixID; installed_on=$(if ($_.InstalledOn) { $_.InstalledOn.ToString('yyyy-MM-dd') } else { '' }) } })`,
		},
		psJSONCollector[DeviceGuard]{
			Runner: ps, CategoryID: "security.device_guard",
			Summary: "Virtualization-based security status and which VBS services (Credential Guard, HVCI, ...) are configured and running.",
			Script:  `$d = Get-CimInstance -Namespace 'root\Microsoft\Windows\DeviceGuard' -ClassName Win32_DeviceGuard; ConvertTo-Json -Compress -InputObject ([pscustomobject]@{ vbs_status=[int]$d.VirtualizationBasedSecurityStatus; services_configured=@($d.SecurityServicesConfigured | ForEach-Object { [int]$_ }); services_running=@($d.SecurityServicesRunning | ForEach-Object { [int]$_ }) })`,
		},
		psJSONCollector[[]BitLockerVolume]{
			Runner: ps, CategoryID: "security.bitlocker",
			Summary: "BitLocker protection status and encryption method per volume (requires an elevated collector; otherwise reported as unknown).",
			Script:  `ConvertTo-Json -Compress -InputObject @(Get-BitLockerVolume | ForEach-Object { [pscustomobject]@{ mount_point=[string]$_.MountPoint; volume_type=[string]$_.VolumeType; protection_status=[string]$_.ProtectionStatus; encryption_method=[string]$_.EncryptionMethod; volume_status=[string]$_.VolumeStatus } })`,
		},
		psJSONCollector[TPM]{
			Runner: ps, CategoryID: "security.tpm",
			Summary: "Whether a TPM is present, enabled, and activated, and its spec version (requires an elevated collector; otherwise reported as unknown).",
			Script:  `$t = Get-CimInstance -Namespace 'root\cimv2\Security\MicrosoftTpm' -ClassName Win32_Tpm; ConvertTo-Json -Compress -InputObject ([pscustomobject]@{ present=($null -ne $t); enabled=[bool]$t.IsEnabled_InitialValue; activated=[bool]$t.IsActivated_InitialValue; spec_version=[string]$t.SpecVersion })`,
		},
		psJSONCollector[[]ListeningPort]{
			Runner: ps, CategoryID: "network.listening_ports",
			Summary: "TCP ports in the listening state, with the local address and owning process name (no connection or traffic data).",
			Script:  `$p = @{}; Get-Process | ForEach-Object { $p[[int]$_.Id] = $_.ProcessName }; ConvertTo-Json -Compress -InputObject @(Get-NetTCPConnection -State Listen | Sort-Object LocalAddress, LocalPort -Unique | ForEach-Object { [pscustomobject]@{ address=[string]$_.LocalAddress; port=[int]$_.LocalPort; process=[string]$p[[int]$_.OwningProcess] } })`,
		},
		psJSONCollector[[]InstalledSoftware]{
			Runner: ps, CategoryID: "software.installed",
			Summary: "Machine-wide installed programs' display name, version, and publisher (from the Uninstall registry keys).",
			Script:  `ConvertTo-Json -Compress -InputObject @(Get-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\*', 'HKLM:\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*' -ErrorAction SilentlyContinue | Where-Object { $_.DisplayName } | Sort-Object DisplayName -Unique | ForEach-Object { [pscustomobject]@{ name=[string]$_.DisplayName; version=[string]$_.DisplayVersion; publisher=[string]$_.Publisher } })`,
		},
	}
}
