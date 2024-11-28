# Security scan script with privilege elevation
param([switch]$Elevated)

function Test-Admin {
    $currentUser = New-Object Security.Principal.WindowsPrincipal $([Security.Principal.WindowsIdentity]::GetCurrent())
    $currentUser.IsInRole([Security.Principal.WindowsBuiltinRole]::Administrator)
}

# Self-elevate the script if required
if ((Test-Admin) -eq $false)  {
    if ($elevated) {
        Write-Warning "Failed to elevate privileges. Some security information may be limited."
    } else {
        Start-Process powershell.exe -Verb RunAs -ArgumentList ('-noprofile -file "{0}" -elevated' -f ($myinvocation.MyCommand.Definition))
        exit
    }
}

Write-Host "Running with administrative privileges..."

# Set encoding and preferences
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
$ErrorActionPreference = "Continue"
$ProgressPreference = "SilentlyContinue"

# Create the security configuration object
$securityPosture = @{
    systemInfo = @{
        osVersion = (Get-CimInstance Win32_OperatingSystem).Version
        cpuModel = (Get-CimInstance Win32_Processor).Name
        memory = (Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory
        diskSpace = (Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='C:'").Size
    }
    networkInfo = @{
        interfaces = @()
        openPorts = @()
        dnsServers = @()
        services = @()
    }
    securityInfo = @{
        firewall = @{
            enabled = $false
            rules = 0
        }
        antivirus = @()
        updates = @()
        uac = @{
            enabled = $false
        }
    }
}

# Collect network information
try {
    $securityPosture.networkInfo.interfaces = @(Get-NetAdapter | Where-Object Status -eq 'Up' | ForEach-Object {
        $config = Get-NetIPConfiguration -InterfaceIndex $_.ifIndex
        @{
            name = $_.Name
            ipAddress = $config.IPv4Address.IPAddress
            macAddress = $_.MacAddress
        }
    })
    $securityPosture.networkInfo.openPorts = @(Get-NetTCPConnection -State Listen).LocalPort
    $securityPosture.networkInfo.dnsServers = @(Get-DnsClientServerAddress).ServerAddresses
    $securityPosture.networkInfo.services = @(Get-Service | Where-Object Status -eq 'Running').Name
} catch {
    Write-Warning "Error collecting network information: $_"
}

# Collect security information
try {
    # Firewall status
    $firewallProfiles = Get-NetFirewallProfile -ErrorAction SilentlyContinue
    $securityPosture.securityInfo.firewall.enabled = $firewallProfiles | Where-Object Enabled -eq $true | Measure-Object | Select-Object -ExpandProperty Count
    $securityPosture.securityInfo.firewall.rules = (Get-NetFirewallRule | Measure-Object).Count

    # Antivirus
    $securityPosture.securityInfo.antivirus = @(Get-MpComputerStatus | ForEach-Object { "Windows Defender" })

    # Windows Updates
    $securityPosture.securityInfo.updates = @(Get-HotFix | Sort-Object -Property InstalledOn -Descending | Select-Object -First 5 -ExpandProperty HotFixID)

    # UAC Status
    $securityPosture.securityInfo.uac.enabled = (Get-ItemProperty HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System).EnableLUA -eq 1
} catch {
    Write-Warning "Error collecting security information: $_"
}

# Export the results
try {
    $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    $jsonFile = Join-Path $scriptDir "security_posture.json"
    
    $jsonContent = $securityPosture | ConvertTo-Json -Depth 10
    [System.IO.File]::WriteAllText($jsonFile, $jsonContent, [System.Text.UTF8Encoding]::new($false))
    
    Write-Host "Security configuration collected successfully and saved to: $jsonFile"
} catch {
    Write-Error "Failed to save security configuration: $_"
    exit 1
}