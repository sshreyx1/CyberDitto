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
$progressFile = Join-Path $scriptDir "progress.json"

function Write-Log {
    param($Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    "$timestamp - $Message" | Out-File -Append -FilePath $logFile
}

function Update-Progress {
    param(
        [string]$Stage,
        [int]$Progress,
        [string]$Status
    )
    @{
        stage = $Stage
        progress = $Progress
        status = $Status
        phase = if ($Progress -ge 100) { "completed" } else { "scanning" }
    } | ConvertTo-Json | Set-Content -Path $progressFile -Encoding UTF8

    Write-Log "Progress: $Stage - $Progress% - $Status"
}

try {
    Update-Progress -Stage "Initializing" -Progress 0 -Status "Starting security scan..."
    Start-Sleep -Seconds 1

    # System Information Collection (10-20%)
    Update-Progress -Stage "System Info" -Progress 10 -Status "Collecting system information..."
    $systemInfo = @{
        osVersion = [string](Get-CimInstance Win32_OperatingSystem).Version
        cpuModel = [string](Get-CimInstance Win32_Processor).Name
        memory = [long](Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory
        diskSpace = [long](Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='C:'").Size
        lastBootTime = [string](Get-CimInstance Win32_OperatingSystem).LastBootUpTime
        osName = [string](Get-CimInstance Win32_OperatingSystem).Caption
    }
    Write-Log "System info collected successfully"
    Update-Progress -Stage "System Info" -Progress 20 -Status "System information collected"
    Start-Sleep -Seconds 1

    # Network Information Collection (30-40%)
    Update-Progress -Stage "Network" -Progress 30 -Status "Scanning network configuration..."
    $networkInfo = @{
        interfaces = @(Get-NetAdapter | Where-Object Status -eq 'Up' | ForEach-Object {
            $config = Get-NetIPConfiguration -InterfaceIndex $_.ifIndex
            @{
                name = $_.Name
                ipAddress = $config.IPv4Address.IPAddress
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
    Update-Progress -Stage "Network" -Progress 40 -Status "Network information collected"
    Start-Sleep -Seconds 1

    # Security Information Collection - Part 1 (50-60%)
    Update-Progress -Stage "Security" -Progress 50 -Status "Analyzing security settings..."
    $defender = Get-MpComputerStatus
    Write-Log "Windows Defender status retrieved"
    Update-Progress -Stage "Security" -Progress 60 -Status "Security settings analyzed"
    Start-Sleep -Seconds 1

    # Security Information Collection - Part 2 (70-80%)
    Update-Progress -Stage "Security" -Progress 70 -Status "Collecting security configurations..."
    $firewallRules = Get-NetFirewallRule -Enabled True
    $securityInfo = @{
        firewall = @{
            enabled = (Get-NetFirewallProfile).Enabled -contains $true
            rules = ($firewallRules | Measure-Object).Count
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
            level = @{
                ConsentPromptBehaviorAdmin = (Get-ItemProperty HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System).ConsentPromptBehaviorAdmin
            }
        }
    }
    Write-Log "Security info collected successfully"
    Update-Progress -Stage "Security" -Progress 80 -Status "Security configurations collected"
    Start-Sleep -Seconds 1

    # Monitoring Data Collection (90-95%)
    Update-Progress -Stage "Monitoring" -Progress 90 -Status "Gathering monitoring data..."
    $monitoring = @{
        runningProcesses = @(Get-Process | 
            Select-Object -First 50 ProcessName, Id, CPU, WorkingSet)
        startupPrograms = @(Get-CimInstance Win32_StartupCommand | 
            Select-Object Name, Command, Location, User)
        scheduledTasks = @(Get-ScheduledTask | Where-Object State -eq 'Ready' |
            Select-Object TaskName, State, TaskPath)
    }
    Write-Log "Monitoring data collected successfully"
    Update-Progress -Stage "Finalizing" -Progress 95 -Status "Preparing scan results..."
    Start-Sleep -Seconds 1

    # Compile and Export Results
    $securityPosture = @{
        systemInfo = $systemInfo
        networkInfo = $networkInfo
        securityInfo = $securityInfo
        monitoring = $monitoring
        scanTime = Get-Date -Format "o"
    }

    # Export to UTF-8 without BOM
    $jsonContent = $securityPosture | ConvertTo-Json -Depth 10
    [System.IO.File]::WriteAllText($jsonFile, $jsonContent, [System.Text.UTF8Encoding]::new($false))
    
    Update-Progress -Stage "Complete" -Progress 100 -Status "Scan completed successfully"
    Write-Log "Results saved successfully to: $jsonFile"
}
catch {
    $errorMessage = "Error during scan: $($_.Exception.Message)`nStack Trace: $($_.ScriptStackTrace)"
    Write-Log $errorMessage
    Update-Progress -Stage "Error" -Progress 0 -Status $errorMessage
    throw $errorMessage
}