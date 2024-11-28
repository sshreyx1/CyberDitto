# First check if running as administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

# If not admin, try to elevate
if (-not $isAdmin) {
    Write-Host "Script requires administrative privileges. Requesting elevation..."
    try {
        # Create new process with elevated privileges
        $argList = "-File `"$($MyInvocation.MyCommand.Path)`""
        Start-Process powershell -ArgumentList $argList -Verb RunAs -Wait
    }
    catch {
        Write-Error "Failed to get administrative privileges. Script cannot continue."
        Write-Error "Please run this script as an administrator."
        exit 1
    }
    exit
}

Write-Host "Running with administrative privileges. Starting security scan..."

# Set encoding and preferences
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)
$ErrorActionPreference = "Stop" # Changed to Stop so we fail fast on errors
$ProgressPreference = "SilentlyContinue"

try {
    # Create the security configuration object
    $securityPosture = @{
        scanInfo = @{
            timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
            computerName = $env:COMPUTERNAME
            userName = [System.Security.Principal.WindowsIdentity]::GetCurrent().Name
        }
        
        systemInfo = @{
            osVersion = (Get-CimInstance Win32_OperatingSystem).Version
            osName = (Get-CimInstance Win32_OperatingSystem).Caption
            cpuModel = (Get-CimInstance Win32_Processor).Name
            memory = (Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory
            diskSpace = (Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='C:'").Size
            lastBootTime = (Get-CimInstance Win32_OperatingSystem).LastBootUpTime
        }
        
        networkInfo = @{
            interfaces = @(Get-NetAdapter | Where-Object Status -eq 'Up' | ForEach-Object {
                $config = Get-NetIPConfiguration -InterfaceIndex $_.ifIndex
                @{
                    name = $_.Name
                    ipAddress = $config.IPv4Address.IPAddress
                    macAddress = $_.MacAddress
                    linkSpeed = $_.LinkSpeed
                }
            })
            openPorts = @(Get-NetTCPConnection -State Listen | Select-Object LocalPort, RemoteAddress, State, OwningProcess)
            dnsServers = @(Get-DnsClientServerAddress | Select-Object -ExpandProperty ServerAddresses)
            services = @(Get-Service | Where-Object Status -eq 'Running' | Select-Object Name, DisplayName, StartType)
        }
        
        securityInfo = @{
            firewall = @{
                profiles = @(Get-NetFirewallProfile | Select-Object Name, Enabled)
                rules = @(Get-NetFirewallRule | Where-Object Enabled -eq $true | 
                    Select-Object DisplayName, Direction, Action, Profile)
            }
            antivirus = @{
                windowsDefender = Get-MpComputerStatus | Select-Object -Property @(
                    'AntivirusEnabled',
                    'RealTimeProtectionEnabled',
                    'IoavProtectionEnabled',
                    'AntispywareEnabled',
                    'BehaviorMonitorEnabled'
                )
                thirdParty = @(Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntivirusProduct | 
                    Select-Object displayName, productState)
            }
            updates = @{
                lastInstalled = (Get-HotFix | Sort-Object -Property InstalledOn -Descending | 
                    Select-Object -First 5 | Select-Object HotFixID, Description, InstalledOn)
                pendingReboot = Test-Path "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\WindowsUpdate\Auto Update\RebootRequired"
            }
            bitLocker = @(Get-BitLockerVolume | Select-Object MountPoint, VolumeStatus, EncryptionMethod, ProtectionStatus)
            uacSettings = Get-ItemProperty HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System | 
                Select-Object EnableLUA, ConsentPromptBehaviorAdmin, PromptOnSecureDesktop
        }
        
        vulnerabilityInfo = @{
            smbv1Enabled = (Get-WindowsOptionalFeature -Online -FeatureName SMB1Protocol).State -eq 'Enabled'
            powerShellV2Enabled = (Get-WindowsOptionalFeature -Online -FeatureName MicrosoftWindowsPowerShellV2Root).State -eq 'Enabled'
            rdpEnabled = (Get-ItemProperty -Path 'HKLM:\System\CurrentControlSet\Control\Terminal Server' -Name "fDenyTSConnections").fDenyTSConnections -eq 0
            autoPlayEnabled = (Get-ItemProperty -Path 'HKLM:\Software\Microsoft\Windows\CurrentVersion\Policies\Explorer' -Name "NoDriveTypeAutoRun").NoDriveTypeAutoRun -eq 0
        }
    }

    # Export the results
    $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
    $jsonFile = Join-Path $scriptDir "security_posture.json"
    
    Write-Host "Exporting security posture to $jsonFile"
    $jsonContent = $securityPosture | ConvertTo-Json -Depth 10
    [System.IO.File]::WriteAllText($jsonFile, $jsonContent, [System.Text.UTF8Encoding]::new($false))
    
    Write-Host "Security configuration collected successfully and saved to: $jsonFile"
}
catch {
    Write-Error "Critical error during security scan: $_"
    exit 1
}