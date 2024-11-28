# Force UTF-8 without BOM encoding
$PSDefaultParameterValues['Out-File:Encoding'] = 'utf8'
$PSDefaultParameterValues['*:Encoding'] = 'utf8'
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$ErrorActionPreference = "SilentlyContinue"
$ProgressPreference = "SilentlyContinue"

# Get the current script directory for output
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$jsonFile = Join-Path $scriptDir "security_posture.json"

function Get-AntivirusProducts {
    try {
        $products = @()
        
        # Try SecurityCenter2 first
        try {
            $products += Get-CimInstance -Namespace root/SecurityCenter2 -ClassName AntiVirusProduct | 
                ForEach-Object { $_.displayName }
        } catch {}
        
        # Check Windows Defender status as fallback
        if ($products.Count -eq 0) {
            $defender = Get-MpComputerStatus
            if ($defender.AntivirusEnabled) {
                $products += "Windows Defender"
            }
        }
        
        return $products
    } catch {
        return @("Windows Defender")
    }
}

function Get-SecuritySettings {
    try {
        $defender = Get-MpComputerStatus
        @{
            realTimeProtection = $defender.RealTimeProtectionEnabled
            antispyware = $defender.AntispywareEnabled
            behaviorMonitor = $defender.BehaviorMonitorEnabled
            ioavProtection = $defender.IoavProtectionEnabled
            networkProtection = $defender.IoavProtectionEnabled
        }
    } catch {
        @{
            realTimeProtection = $false
            antispyware = $false
            behaviorMonitor = $false
            ioavProtection = $false
            networkProtection = $false
        }
    }
}

# Create the primary data structure
$securityPosture = @{
    systemInfo = @{
        osVersion = [string](Get-CimInstance Win32_OperatingSystem).Version
        cpuModel = [string](Get-CimInstance Win32_Processor).Name
        memory = [long](Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory
        diskSpace = [long](Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='C:'").Size
        lastBootTime = [string](Get-CimInstance Win32_OperatingSystem).LastBootUpTime
        osName = [string](Get-CimInstance Win32_OperatingSystem).Caption
    }
    
    networkInfo = @{
        interfaces = @(Get-NetAdapter | Where-Object Status -eq 'Up' | ForEach-Object {
            $config = Get-NetIPConfiguration -InterfaceIndex $_.ifIndex
            @{
                name = $_.Name
                ipAddress = ($config.IPv4Address.IPAddress)
                macAddress = $_.MacAddress
                linkSpeed = $_.LinkSpeed
                dnsServers = @($config.DNSServer.ServerAddresses)
            }
        })
        openPorts = @(Get-NetTCPConnection -State Listen | 
            Select-Object -Property LocalPort, RemoteAddress, State, OwningProcess |
            Sort-Object LocalPort -Unique)
        dnsServers = @(Get-DnsClientServerAddress | Select-Object -ExpandProperty ServerAddresses)
        services = @(Get-Service | Where-Object Status -eq 'Running' | Select-Object -ExpandProperty Name)
    }
    
    securityInfo = @{
        firewall = @{
            enabled = (Get-NetFirewallProfile).Enabled -contains $true
            rules = (Get-NetFirewallRule | Where-Object Enabled -eq $true | Measure-Object).Count
            inboundRules = @(Get-NetFirewallRule -Direction Inbound -Enabled True |
                Select-Object -First 100 DisplayName, Action, Profile)
            outboundRules = @(Get-NetFirewallRule -Direction Outbound -Enabled True |
                Select-Object -First 100 DisplayName, Action, Profile)
        }
        antivirus = @(Get-AntivirusProducts)
        defenderStatus = Get-SecuritySettings
        updates = @(Get-HotFix | Sort-Object -Property InstalledOn -Descending | 
            Select-Object -First 10 HotFixID, InstalledOn)
        uac = @{
            enabled = (Get-ItemProperty HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System).EnableLUA -eq 1
            level = Get-ItemProperty HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System -Name ConsentPromptBehaviorAdmin
        }
    }

    monitoring = @{
        runningProcesses = @(Get-Process | Select-Object -First 50 ProcessName, Id, CPU, WorkingSet)
        startupPrograms = @(Get-CimInstance Win32_StartupCommand | 
            Select-Object Name, Command, Location, User)
        scheduledTasks = @(Get-ScheduledTask | Where-Object State -eq 'Ready' |
            Select-Object TaskName, State, TaskPath)
    }
}

try {
    # Export to UTF-8 without BOM
    $jsonContent = $securityPosture | ConvertTo-Json -Depth 10 -Compress
    [System.IO.File]::WriteAllText($jsonFile, $jsonContent, [System.Text.UTF8Encoding]::new($false))
} catch {
    # Write error to separate file for debugging
    $errorFile = Join-Path $scriptDir "scan_error.log"
    [System.IO.File]::WriteAllText($errorFile, $_.Exception.Message, [System.Text.UTF8Encoding]::new($false))
    throw
}