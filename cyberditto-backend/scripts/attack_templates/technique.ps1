# scripts/attack_templates/technique.ps1
param(
    [Parameter(Mandatory=$true)]
    [string]$TechniqueID,
    [string]$TestNumbers = "all",
    [string]$OutputPath = "C:\AtomicResults"
)

$ErrorActionPreference = "Continue"
$ProgressPreference = "SilentlyContinue"

function Write-Log {
    param($Message)
    $logMessage = "$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss'): $Message"
    Write-Host $logMessage
    Add-Content -Path "C:\AtomicResults\execution.log" -Value $logMessage
}

# Import Atomic Red Team module
Write-Log "Importing Atomic Red Team module..."
Import-Module "C:\AtomicRedTeam\invoke-atomicredteam\Invoke-AtomicRedTeam.psd1" -Force
$PSDefaultParameterValues = @{"Invoke-AtomicTest:PathToAtomicsFolder"="C:\AtomicRedTeam\atomics"}

# Load test data
$templateData = Import-Csv "C:\AtomicRedTeam\windows-index.csv"

# Create results directory
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$resultsDir = Join-Path $OutputPath "TechniqueTest_${TechniqueID}_${timestamp}"
New-Item -ItemType Directory -Path $resultsDir -Force | Out-Null

Write-Log "Starting technique execution for: $TechniqueID"
$results = @()

# Filter tests for the specific technique
$techniqueTests = $templateData | Where-Object { $_.Technique -eq $TechniqueID }

if (-not $techniqueTests) {
    Write-Log "Error: No tests found for technique $TechniqueID"
    throw "Technique not found"
}

# Determine which tests to run
$testsToRun = @()
if ($TestNumbers -eq "all") {
    $testsToRun = $techniqueTests
} else {
    $testNums = $TestNumbers -split ',' | ForEach-Object { $_.Trim() }
    $testsToRun = $techniqueTests | Where-Object { $testNums -contains $_.TestNumber }
}

Write-Log "Found $($testsToRun.Count) tests to execute"

foreach ($test in $testsToRun) {
    Write-Log "Executing Test $($test.TestNumber): $($test.TestName)"
    
    # Create test-specific log file
    $testLog = Join-Path $resultsDir "test_$($test.TestNumber)_log.txt"
    
    try {
        # Run prerequisites
        Write-Log "Running prerequisites for test $($test.TestNumber)..."
        $prereqOutput = Invoke-AtomicTest $test.Technique -TestNumbers $test.TestNumber -GetPrereqs -ErrorAction Stop *>&1 | Out-String
        Add-Content -Path $testLog -Value "Prerequisites Output:`n$prereqOutput`n"
        
        # Execute test
        Write-Log "Executing test..."
        $testOutput = Invoke-AtomicTest $test.Technique -TestNumbers $test.TestNumber -ErrorAction Stop *>&1 | Out-String
        Add-Content -Path $testLog -Value "Test Output:`n$testOutput`n"
        
        # Run cleanup
        Write-Log "Running cleanup..."
        $cleanupOutput = Invoke-AtomicTest $test.Technique -TestNumbers $test.TestNumber -Cleanup -ErrorAction Stop *>&1 | Out-String
        Add-Content -Path $testLog -Value "Cleanup Output:`n$cleanupOutput`n"
        
        $status = "Pass"
        $errorMsg = ""
    }
    catch {
        $status = "Fail"
        $errorMsg = $_.Exception.Message
        Write-Log "Error during test execution: $errorMsg"
        Add-Content -Path $testLog -Value "Error: $errorMsg"
        
        # Attempt cleanup after error
        try {
            Invoke-AtomicTest $test.Technique -TestNumbers $test.TestNumber -Cleanup -ErrorAction Stop
        }
        catch {
            Add-Content -Path $testLog -Value "Cleanup after error failed: $($_.Exception.Message)"
        }
    }
    
    $results += [PSCustomObject]@{
        'Timestamp' = (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
        'Status' = $status
        'Tactic' = $test.Tactic
        'Technique' = $test.Technique
        'TechniqueName' = $test.'Technique Name'
        'TestNumber' = $test.TestNumber
        'TestName' = $test.TestName
        'TestGUID' = $test.'Test GUID'
        'ExecutorName' = $test.'Executor Name'
        'ErrorMessage' = $errorMsg
        'LogFile' = $testLog
    }
}

# Export results
$resultsPath = Join-Path $resultsDir "atomic_results.csv"
$results | Export-Csv -Path $resultsPath -NoTypeInformation

Write-Log "Technique execution completed. Results saved to: $resultsPath"
return $resultsPath