# resources/scripts/prepare_vm.ps1

# Error handling
$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

# Logging function
function Write-Log {
    param($Message)
    $logMessage = "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss'): $Message"
    Write-Host $logMessage
    Add-Content -Path "C:\AtomicRedTeam\setup_log.txt" -Value $logMessage
}

Write-Log "Starting VM preparation..."

# Create necessary directories
Write-Log "Creating directories..."
$directories = @(
    "C:\AtomicRedTeam",
    "C:\AtomicRedTeam\atomics",
    "C:\AtomicRedTeam\invoke-atomicredteam",
    "C:\AtomicResults",
    "C:\AtomicScripts"
)

foreach ($dir in $directories) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force
        Write-Log "Created directory: $dir"
    }
}

# Download and install Atomic Red Team
Write-Log "Downloading Atomic Red Team..."
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

try {
    # Download and install the framework
    Invoke-WebRequest `
        -Uri "https://raw.githubusercontent.com/redcanaryco/invoke-atomicredteam/master/install-atomicredteam.ps1" `
        -OutFile "C:\AtomicRedTeam\install-atomicredteam.ps1"

    Write-Log "Installing Atomic Red Team..."
    Set-ExecutionPolicy Bypass -Scope Process -Force
    & "C:\AtomicRedTeam\install-atomicredteam.ps1" -Branch "master" -InstallPath "C:\AtomicRedTeam"

    # Clone the atomics repository
    Write-Log "Cloning Atomic Red Team repository..."
    Push-Location "C:\AtomicRedTeam"
    & git clone https://github.com/redcanaryco/atomic-red-team.git atomics
    Pop-Location
} catch {
    Write-Log "Error during Atomic Red Team installation: $_"
    throw
}

# Configure PowerShell remoting
Write-Log "Configuring PowerShell remoting..."
try {
    Enable-PSRemoting -Force
    Set-Item WSMan:\localhost\Client\TrustedHosts -Value "*" -Force
    
    # Create dedicated service account
    $password = ConvertTo-SecureString "AtomicTest123!" -AsPlainText -Force
    $username = "AtomicTester"
    
    if (-not (Get-LocalUser -Name $username -ErrorAction SilentlyContinue)) {
        New-LocalUser -Name $username -Password $password -FullName "Atomic Test Account" -Description "Account for running Atomic Red Team tests"
        Add-LocalGroupMember -Group "Administrators" -Member $username
        Write-Log "Created service account: $username"
    }
} catch {
    Write-Log "Error during PowerShell remoting setup: $_"
    throw
}

# Install additional PowerShell modules
Write-Log "Installing required PowerShell modules..."
try {
    Install-PackageProvider -Name NuGet -MinimumVersion 2.8.5.201 -Force
    Install-Module -Name PowerShellGet -Force -AllowClobber
    Install-Module -Name PSWSMan -Force
} catch {
    Write-Log "Error during module installation: $_"
    throw
}

# Create test execution directory structure
Write-Log "Setting up test execution environment..."
$testDirs = @(
    "C:\AtomicScripts\Tactics",
    "C:\AtomicScripts\Techniques",
    "C:\AtomicScripts\Custom",
    "C:\AtomicResults\Logs",
    "C:\AtomicResults\Reports"
)

foreach ($dir in $testDirs) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force
        Write-Log "Created test directory: $dir"
    }
}

# Configure execution policy and system settings
Write-Log "Configuring system settings..."
try {
    Set-ExecutionPolicy Bypass -Scope LocalMachine -Force
    Set-MpPreference -DisableRealtimeMonitoring $true -Force
    reg add "HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System" /v EnableLUA /t REG_DWORD /d 0 /f
} catch {
    Write-Log "Error during system configuration: $_"
    throw
}

# Verify installation
Write-Log "Verifying installation..."
$verificationChecks = @{
    "Atomic Red Team Installation" = Test-Path "C:\AtomicRedTeam\invoke-atomicredteam\Invoke-AtomicRedTeam.psd1"
    "Atomics Repository" = Test-Path "C:\AtomicRedTeam\atomics"
    "Results Directory" = Test-Path "C:\AtomicResults"
    "Scripts Directory" = Test-Path "C:\AtomicScripts"
    "Service Account" = Get-LocalUser -Name "AtomicTester" -ErrorAction SilentlyContinue
}

$verification = $true
foreach ($check in $verificationChecks.GetEnumerator()) {
    if (-not $check.Value) {
        Write-Log "Verification failed: $($check.Name)"
        $verification = $false
    }
}

if ($verification) {
    Write-Log "VM preparation completed successfully"
} else {
    Write-Log "VM preparation completed with errors"
    throw "Installation verification failed"
}

# Create verification file for backend to check
$verificationFile = "C:\AtomicRedTeam\setup_complete"
New-Item -ItemType File -Path $verificationFile -Force
Add-Content -Path $verificationFile -Value (Get-Date -Format "yyyy-MM-dd HH:mm:ss")

Write-Log "Setup completion marker created at: $verificationFile"