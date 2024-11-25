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

# Create the primary data structure
$securityPosture = @{
    systemInfo = @{
        osVersion = [string](Get-CimInstance Win32_OperatingSystem).Version
        cpuModel = [string](Get-CimInstance Win32_Processor).Name
        memory = [long](Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory
        diskSpace = [long](Get-CimInstance Win32_LogicalDisk -Filter "DeviceID='C:'").Size
    }
    networkInfo = @{
        interfaces = @(Get-NetAdapter | Where-Object Status -eq 'Up' | ForEach-Object {
            $config = Get-NetIPConfiguration -InterfaceIndex $_.ifIndex
            @{
                name = $_.Name
                ipAddress = ($config.IPv4Address.IPAddress)
                macAddress = $_.MacAddress
            }
        })
        openPorts = @(Get-NetTCPConnection -State Listen | Select-Object -ExpandProperty LocalPort)
        dnsServers = @(Get-DnsClientServerAddress | Select-Object -ExpandProperty ServerAddresses)
        services = @(Get-Service | Where-Object Status -eq 'Running' | Select-Object -ExpandProperty Name)
    }
    securityInfo = @{
        firewall = @{
            enabled = $true
            rules = (Get-NetFirewallRule | Measure-Object).Count
        }
        antivirus = @(Get-MpComputerStatus | ForEach-Object { "Windows Defender" })
        updates = @((Get-HotFix | Sort-Object -Property InstalledOn -Descending | Select-Object -First 5).HotFixID)
        uac = @{
            enabled = $true
        }
    }
}

# Export to UTF-8 without BOM
$jsonContent = $securityPosture | ConvertTo-Json -Depth 10
[System.IO.File]::WriteAllText($jsonFile, $jsonContent, [System.Text.UTF8Encoding]::new($false))

Write-Host "Security posture data collected and saved to: $jsonFile"