# First check for admin privileges
$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
$isAdmin = $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if (-not $isAdmin) {
    Write-Host "Requesting administrative privileges..."
    try {
        Start-Process powershell -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File `"$PSCommandPath`"" -Verb RunAs -Wait
        exit
    }
    catch {
        Write-Error "This script requires administrative privileges to run. Please run as administrator."
        exit 1
    }
}

# Force UTF-8 without BOM encoding
$PSDefaultParameterValues['Out-File:Encoding'] = 'utf8'
$PSDefaultParameterValues['*:Encoding'] = 'utf8'
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$ErrorActionPreference = "Continue"
$ProgressPreference = "SilentlyContinue"

# Get the current script directory for output
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$jsonFile = Join-Path $scriptDir "security_posture.json"

Write-Host "Script directory: $scriptDir"
Write-Host "JSON output file: $jsonFile"

function Get-SystemSecurityInfo {
    @{
        securityCenter = Get-CimInstance -Namespace root/SecurityCenter2 -ClassName SecurityCenter2
        bitLocker = Get-BitLockerVolume -ErrorAction SilentlyContinue
        secureBootStatus = Confirm-SecureBootUEFI
        tpm = Get-Tpm
        uacConfig = @{
            enabled = (Get-ItemProperty HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System).EnableLUA
            consentPromptBehavior = (Get-ItemProperty HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System).ConsentPromptBehaviorAdmin
            secureDesktop = (Get-ItemProperty HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System).PromptOnSecureDesktop
        }
        localUsers = Get-LocalUser | Select-Object Name, Enabled, PasswordRequired, PasswordLastSet, LastLogon
        localGroups = Get-LocalGroupMember -Group "Administrators"
        passwordPolicy = net accounts | Select-String "Password" | ForEach-Object { $_.ToString().Trim() }
        credentialGuard = Get-CimInstance -ClassName Win32_DeviceGuard -Namespace root\Microsoft\Windows\DeviceGuard
    }
}

function Get-DetailedNetworkSecurity {
    @{
        firewallRules = Get-NetFirewallRule -Enabled True | Select-Object Name, DisplayName, Direction, Action, Profile
        networkShares = Get-SmbShare | Select-Object Name, Path, Description
        smbConfig = Get-SmbServerConfiguration | Select-Object EnableSMB1Protocol, EnableSMB2Protocol
        tlsSettings = Get-TlsCipherSuite
        networkAdapters = Get-NetAdapter | Select-Object Name, InterfaceDescription, Status, MacAddress, LinkSpeed
        ipsecPolicies = Get-NetIPsecRule
        rdpConfig = Get-ItemProperty "HKLM:\SYSTEM\CurrentControllerSet\Control\Terminal Server" -Name "fDenyTSConnections" -ErrorAction SilentlyContinue
        wifiSettings = netsh wlan show interfaces
        proxySettings = Get-ItemProperty "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings"
    }
}

function Get-EndpointSecurityInfo {
    @{
        antivirusProducts = Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntiVirusProduct
        defenderConfig = @{
            realTimeProtection = Get-MpPreference | Select-Object DisableRealtimeMonitoring
            cloudProtection = Get-MpPreference | Select-Object MAPSReporting
            networkInspection = Get-MpPreference | Select-Object DisableIOAVProtection
            controlledFolderAccess = Get-MpPreference | Select-Object EnableControlledFolderAccess
            exploitProtection = Get-ProcessMitigation -System
        }
        applicationGuard = Get-WindowsOptionalFeature -Online -FeatureName Windows-Defender-ApplicationGuard
    }
}

function Get-SoftwareSecurityInfo {
    @{
        installedSoftware = Get-WmiObject -Class Win32_Product | Select-Object Name, Version, Vendor
        runningServices = Get-Service | Where-Object Status -eq 'Running' | Select-Object Name, DisplayName, StartType
        startupPrograms = Get-CimInstance Win32_StartupCommand
        powershellPolicy = Get-ExecutionPolicy -List
        dotNetVersions = Get-ChildItem 'HKLM:\SOFTWARE\Microsoft\NET Framework Setup\NDP' -Recurse | Get-ItemProperty -Name Version -ErrorAction SilentlyContinue
        codeIntegrityPolicies = Get-CimInstance -ClassName Win32_DeviceGuard -Namespace root\Microsoft\Windows\DeviceGuard
    }
}

function Get-UpdateManagementInfo {
    @{
        windowsUpdate = @{
            settings = Get-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update'
            history = Get-HotFix | Sort-Object -Property InstalledOn -Descending
            pendingReboot = Test-Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\RebootRequired'
        }
        osVersion = [System.Environment]::OSVersion.Version
        osFullName = (Get-WmiObject -class Win32_OperatingSystem).Caption
        lastBootTime = (Get-WmiObject -class Win32_OperatingSystem).LastBootUpTime
    }
}

function Get-LoggingAndAuditInfo {
    @{
        eventLogSettings = Get-WinEvent -ListLog * | Where-Object IsEnabled -eq $true | Select-Object LogName, LogMode, MaximumSizeInBytes
        auditPolicies = auditpol /get /category:*
        powershellLogging = Get-ItemProperty HKLM:\SOFTWARE\Wow6432Node\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging -ErrorAction SilentlyContinue
    }
}

function Get-AccessControlInfo {
    @{
        criticalFolderPermissions = Get-Acl C:\Windows | Select-Object -ExpandProperty Access
        registryPermissions = Get-Acl HKLM:\SOFTWARE | Select-Object -ExpandProperty Access
        scheduledTasks = Get-ScheduledTask | Where-Object State -eq 'Ready' | Select-Object TaskName, State, TaskPath
        remoteAccess = Get-ItemProperty -Path 'HKLM:\System\CurrentControlSet\Control\Terminal Server' -Name 'fDenyTSConnections'
    }
}

function Get-MemoryProtectionInfo {
    @{
        dep = Get-CimInstance -ClassName Win32_OperatingSystem | Select-Object DataExecutionPrevention_Available
        aslr = Get-ProcessMitigation -System
        memoryIntegrity = Get-CimInstance -Namespace root\Microsoft\Windows\DeviceGuard -ClassName Win32_DeviceGuard
    }
}

# Create comprehensive security configuration object
$securityPosture = @{
    systemSecurity = Get-SystemSecurityInfo
    networkSecurity = Get-DetailedNetworkSecurity
    endpointSecurity = Get-EndpointSecurityInfo
    softwareSecurity = Get-SoftwareSecurityInfo
    updateManagement = Get-UpdateManagementInfo
    loggingAndAudit = Get-LoggingAndAuditInfo
    accessControl = Get-AccessControlInfo
    memoryProtection = Get-MemoryProtectionInfo
    systemInfo = @{
        osVersion = [string](Get-CimInstance Win32_OperatingSystem).Version
        cpuModel = [string](Get-CimInstance Win32_Processor).Name
        memory = [long](Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory
        diskSpace = [long](Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='C:'").Size
    }
}

try {
    # Create directory if it doesn't exist
    $directory = Split-Path -Parent $jsonFile
    if (-not (Test-Path -Path $directory)) {
        New-Item -ItemType Directory -Path $directory -Force | Out-Null
    }

    # Export to UTF-8 without BOM
    $jsonContent = $securityPosture | ConvertTo-Json -Depth 10
    [System.IO.File]::WriteAllText($jsonFile, $jsonContent, [System.Text.UTF8Encoding]::new($false))

    Write-Host "Security posture data collected and saved to: $jsonFile"
} catch {
    Write-Error "Failed to save security configuration: $_"
    exit 1
}