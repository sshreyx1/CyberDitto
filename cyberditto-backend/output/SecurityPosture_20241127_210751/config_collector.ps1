# Force UTF-8 without BOM encoding
$PSDefaultParameterValues['Out-File:Encoding'] = 'utf8'
$PSDefaultParameterValues['*:Encoding'] = 'utf8'
$OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# Get the current script directory for output
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$jsonFile = Join-Path $scriptDir "security_posture.json"
$logFile = Join-Path $scriptDir "scan.log"

function Write-Log {
    param($Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    "$timestamp - $Message" | Out-File -Append -FilePath $logFile
}

function Update-Progress {
    param($Stage, $Progress, $Status)
    $progressFile = Join-Path $scriptDir "progress.json"
    @{
        stage = $Stage
        progress = $Progress
        status = $Status
    } | ConvertTo-Json | Out-File $progressFile
    Write-Log "Progress: $Stage - $Progress% - $Status"
}

Write-Log "Starting security scan..."
Update-Progress "Initializing" 0 "Starting security scan..."

try {
    Update-Progress "System Info" 10 "Collecting system information..."
    $systemInfo = @{
        osVersion = [string](Get-CimInstance Win32_OperatingSystem).Version
        cpuModel = [string](Get-CimInstance Win32_Processor).Name
        memory = [long](Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory
        diskSpace = [long](Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='C:'").Size
        lastBootTime = [string](Get-CimInstance Win32_OperatingSystem).LastBootUpTime
        osName = [string](Get-CimInstance Win32_OperatingSystem).Caption
    }
    Write-Log "System info collected successfully"

    Update-Progress "Network" 30 "Scanning network configuration..."
    $networkInfo = @{
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
    Write-Log "Network info collected successfully"

    Update-Progress "Security" 50 "Analyzing security settings..."
    # Get Windows Defender status
    $defender = Get-MpComputerStatus
    Write-Log "Windows Defender status retrieved"

    # Get firewall rules
    $firewallRules = Get-NetFirewallRule -Enabled True
    Write-Log "Firewall rules retrieved: $($firewallRules.Count) rules found"

    Update-Progress "Security" 70 "Collecting security configurations..."
    $securityInfo = @{
        firewall = @{
            enabled = (Get-NetFirewallProfile).Enabled -contains $true
            rules = $firewallRules.Count
            inboundRules = @($firewallRules | Where-Object Direction -eq "Inbound" |
                Select-Object -First 100 DisplayName, Action, Profile)
            outboundRules = @($firewallRules | Where-Object Direction -eq "Outbound" |
                Select-Object -First 100 DisplayName, Action, Profile)
        }
        antivirus = @{
            windowsDefender = @{
                enabled = $defender.AntivirusEnabled
                realTimeProtection = $defender.RealTimeProtectionEnabled
                antispyware = $defender.AntispywareEnabled
                behaviorMonitor = $defender.BehaviorMonitorEnabled
                ioavProtection = $defender.IoavProtectionEnabled
                networkProtection = $defender.IoavProtectionEnabled
            }
        }
        updates = @(Get-HotFix | Sort-Object -Property InstalledOn -Descending | 
            Select-Object -First 10 HotFixID, InstalledOn)
        uac = @{
            enabled = (Get-ItemProperty HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System).EnableLUA -eq 1
            level = Get-ItemProperty HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System -Name ConsentPromptBehaviorAdmin
        }
    }
    Write-Log "Security info collected successfully"

    Update-Progress "Monitoring" 90 "Gathering monitoring data..."
    $monitoring = @{
        runningProcesses = @(Get-Process | Select-Object -First 50 ProcessName, Id, CPU, WorkingSet)
        startupPrograms = @(Get-CimInstance Win32_StartupCommand | 
            Select-Object Name, Command, Location, User)
        scheduledTasks = @(Get-ScheduledTask | Where-Object State -eq 'Ready' |
            Select-Object TaskName, State, TaskPath)
    }
    Write-Log "Monitoring data collected successfully"

    Update-Progress "Finalizing" 95 "Preparing scan results..."
    $securityPosture = @{
        systemInfo = $systemInfo
        networkInfo = $networkInfo
        securityInfo = $securityInfo
        monitoring = $monitoring
    }

    # Export to UTF-8 without BOM
    $jsonContent = $securityPosture | ConvertTo-Json -Depth 10 -Compress
    [System.IO.File]::WriteAllText($jsonFile, $jsonContent, [System.Text.UTF8Encoding]::new($false))
    
    Update-Progress "Complete" 100 "Scan completed successfully"
    Write-Log "Scan completed successfully. Results saved to: $jsonFile"
}
catch {
    $errorMessage = "Error during scan: $($_.Exception.Message)`nStack Trace: $($_.ScriptStackTrace)"
    Write-Log $errorMessage
    Update-Progress "Error" 0 $errorMessage
    throw $errorMessage
}

Update-Progress "Complete" 100 "Scan completed successfully"
Write-Log "Scan completed successfully"

# Export the final results
try {
    $jsonContent = $securityPosture | ConvertTo-Json -Depth 10 -Compress
    [System.IO.File]::WriteAllText($jsonFile, $jsonContent, [System.Text.UTF8Encoding]::new($false))
    
    # Write final progress
    @{
        stage = "Complete"
        progress = 100
        status = "Scan completed successfully"
    } | ConvertTo-Json | Out-File $progressFile

    Write-Log "Results saved successfully to: $jsonFile"
} catch {
    $errorMessage = "Error during results export: $($_.Exception.Message)"
    Write-Log $errorMessage
    Update-Progress "Error" 0 $errorMessage
    throw $errorMessage
}