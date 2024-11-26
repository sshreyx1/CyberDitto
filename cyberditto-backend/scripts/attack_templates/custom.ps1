# scripts/attack_templates/custom.ps1
param(
    [Parameter(Mandatory=$true)]
    [string]$Techniques,
    [string]$OutputPath = "C:\AtomicResults"
)

$ErrorActionPreference = "Continue"
$ProgressPreference = "SilentlyContinue"

function Write-Log {
    param(
        [string]$Message,
        [string]$Level = "Info"
    )
    $timestamp = Get-Date -Format 'yyyy-MM-dd HH:mm:ss'
    $logMessage = "[$timestamp] [$Level] $Message"
    Write-Host $logMessage
    Add-Content -Path "C:\AtomicResults\execution.log" -Value $logMessage
}

function Write-ErrorLog {
    param(
        [string]$TechniqueID,
        [string]$TestNumber,
        [string]$ErrorMessage,
        [string]$ErrorDetails
    )
    $errorLog = @{
        Timestamp = Get-Date -Format 'yyyy-MM-dd HH:mm:ss'
        TechniqueID = $TechniqueID
        TestNumber = $TestNumber
        ErrorMessage = $ErrorMessage
        ErrorDetails = $ErrorDetails
    }
    
    $errorLogPath = Join-Path $script:ResultsDir "error_log.json"
    $errorLog | ConvertTo-Json | Add-Content -Path $errorLogPath
    Write-Log -Level "Error" -Message "Technique: $TechniqueID, Test: $TestNumber - $ErrorMessage"
}

function Execute-Technique {
    param(
        [string]$TechniqueID,
        [array]$TestData,
        [string]$ResultsDir
    )
    
    $techniqueTests = $TestData | Where-Object { $_.Technique -eq $TechniqueID }
    if (-not $techniqueTests) {
        Write-ErrorLog -TechniqueID $TechniqueID -TestNumber "N/A" `
                      -ErrorMessage "No tests found for technique" `
                      -ErrorDetails "Technique ID may be invalid or no tests are available"
        return @()
    }

    $results = @()
    
    foreach ($test in $techniqueTests) {
        Write-Log "Executing Test $($test.TestNumber) for Technique $TechniqueID"
        
        $testLog = Join-Path $ResultsDir "technique_${TechniqueID}_test_$($test.TestNumber)_log.txt"
        
        try {
            # Prerequisites
            Write-Log "Running prerequisites for $TechniqueID - Test $($test.TestNumber)"
            $prereqOutput = Invoke-AtomicTest $TechniqueID -TestNumbers $test.TestNumber -GetPrereqs -ErrorAction Stop *>&1 | Out-String
            Add-Content -Path $testLog -Value "Prerequisites Output:`n$prereqOutput`n"
            
            # Test execution
            Write-Log "Executing test $TechniqueID - Test $($test.TestNumber)"
            $testOutput = Invoke-AtomicTest $TechniqueID -TestNumbers $test.TestNumber -ErrorAction Stop *>&1 | Out-String
            Add-Content -Path $testLog -Value "Test Output:`n$testOutput`n"
            
            # Cleanup
            Write-Log "Running cleanup for $TechniqueID - Test $($test.TestNumber)"
            $cleanupOutput = Invoke-AtomicTest $TechniqueID -TestNumbers $test.TestNumber -Cleanup -ErrorAction Stop *>&1 | Out-String
            Add-Content -Path $testLog -Value "Cleanup Output:`n$cleanupOutput`n"
            
            $status = "Pass"
            $errorMsg = ""
        }
        catch {
            $status = "Fail"
            $errorMsg = $_.Exception.Message
            Write-ErrorLog -TechniqueID $TechniqueID -TestNumber $test.TestNumber `
                          -ErrorMessage $_.Exception.Message `
                          -ErrorDetails $_.Exception.StackTrace
            
            # Attempt cleanup after error
            try {
                Write-Log "Attempting cleanup after error for $TechniqueID - Test $($test.TestNumber)"
                Invoke-AtomicTest $TechniqueID -TestNumbers $test.TestNumber -Cleanup -ErrorAction Stop
            }
            catch {
                Write-ErrorLog -TechniqueID $TechniqueID -TestNumber $test.TestNumber `
                              -ErrorMessage "Cleanup failed after error" `
                              -ErrorDetails $_.Exception.Message
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
            'ExecutionOrder' = $script:executionOrder
        }
    }
    
    return $results
}

# Import Atomic Red Team module
Write-Log "Importing Atomic Red Team module..."
Import-Module "C:\AtomicRedTeam\invoke-atomicredteam\Invoke-AtomicRedTeam.psd1" -Force
$PSDefaultParameterValues = @{"Invoke-AtomicTest:PathToAtomicsFolder"="C:\AtomicRedTeam\atomics"}

# Load test data
$templateData = Import-Csv "C:\AtomicRedTeam\windows-index.csv"

# Create results directory
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$script:ResultsDir = Join-Path $OutputPath "CustomChain_${timestamp}"
New-Item -ItemType Directory -Path $script:ResultsDir -Force | Out-Null

# Parse techniques
$techniqueList = $Techniques -split ',' | ForEach-Object { $_.Trim() }
Write-Log "Starting custom attack chain with techniques: $($techniqueList -join ', ')"

$allResults = @()
$script:executionOrder = 0

foreach ($currentTechnique in $techniqueList) {
    $script:executionOrder++
    Write-Log "Executing technique $currentTechnique (Step $script:executionOrder of $($techniqueList.Count))"
    
    try {
        $results = Execute-Technique -TechniqueID $currentTechnique -TestData $templateData -ResultsDir $script:ResultsDir
        $allResults += $results
    }
    catch {
        Write-ErrorLog -TechniqueID $currentTechnique -TestNumber "N/A" `
                      -ErrorMessage "Critical error during technique execution" `
                      -ErrorDetails $_.Exception.Message
        # Continue with next technique
        continue
    }
}

# Export results
$resultsPath = Join-Path $script:ResultsDir "atomic_results.csv"
$allResults | Export-Csv -Path $resultsPath -NoTypeInformation

# Create execution summary
$summary = @{
    'TotalTechniques' = $techniqueList.Count
    'CompletedTechniques' = ($allResults | Select-Object -Property Technique -Unique).Count
    'TotalTests' = $allResults.Count
    'PassedTests' = ($allResults | Where-Object { $_.Status -eq 'Pass' }).Count
    'FailedTests' = ($allResults | Where-Object { $_.Status -eq 'Fail' }).Count
    'ExecutionOrder' = $techniqueList -join ' -> '
    'StartTime' = $timestamp
    'EndTime' = Get-Date -Format "yyyyMMdd_HHmmss"
}

$summaryPath = Join-Path $script:ResultsDir "execution_summary.json"
$summary | ConvertTo-Json | Out-File $summaryPath

Write-Log "Custom chain execution completed. Results saved to: $resultsPath"
Write-Log "Execution summary saved to: $summaryPath"

return $resultsPath