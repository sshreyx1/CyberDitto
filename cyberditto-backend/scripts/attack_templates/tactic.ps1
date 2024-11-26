# scripts/attack_templates/tactic.ps1
param(
    [string]$Tactic,
    [string]$OutputPath = "C:\AtomicResults"
)

$ErrorActionPreference = "Continue"
$ProgressPreference = "SilentlyContinue"

# Import Atomic Red Team module
Import-Module "C:\AtomicRedTeam\invoke-atomicredteam\Invoke-AtomicRedTeam.psd1" -Force
$PSDefaultParameterValues = @{"Invoke-AtomicTest:PathToAtomicsFolder"="C:\AtomicRedTeam\atomics"}

# Load test data
$templateData = Import-Csv "C:\AtomicRedTeam\windows-index.csv"

# Filter techniques for the given tactic
$techniques = $templateData | Where-Object { $_.Tactic -eq $Tactic } | Select-Object Technique -Unique

# Create results directory
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$resultsDir = Join-Path $OutputPath "Results_${timestamp}"
New-Item -ItemType Directory -Path $resultsDir -Force | Out-Null

# Initialize results array
$results = @()

foreach ($techItem in $techniques) {
    $technique = $techItem.Technique
    Write-Host "Processing Technique: $technique"
    
    $techniqueTests = $templateData | Where-Object { $_.Technique -eq $technique }
    foreach ($test in $techniqueTests) {
        Write-Host "Running Test $($test.TestNumber): $($test.TestName)"
        
        try {
            # Run test and capture output
            $testOutput = Invoke-AtomicTest $test.Technique -TestNumbers $test.TestNumber -GetPrereqs
            $testResult = Invoke-AtomicTest $test.Technique -TestNumbers $test.TestNumber
            
            $results += [PSCustomObject]@{
                'Timestamp' = (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
                'Status' = if ($?) { "Pass" } else { "Fail" }
                'Tactic' = $test.Tactic
                'Technique' = $test.Technique
                'TechniqueName' = $test.'Technique Name'
                'TestNumber' = $test.TestNumber
                'TestName' = $test.TestName
                'TestGUID' = $test.'Test GUID'
                'ExecutorName' = $test.'Executor Name'
            }
        }
        catch {
            # Log error and continue with next test
            $results += [PSCustomObject]@{
                'Timestamp' = (Get-Date -Format "yyyy-MM-dd HH:mm:ss")
                'Status' = "Fail"
                'Tactic' = $test.Tactic
                'Technique' = $test.Technique
                'TechniqueName' = $test.'Technique Name'
                'TestNumber' = $test.TestNumber
                'TestName' = $test.TestName
                'TestGUID' = $test.'Test GUID'
                'ExecutorName' = $test.'Executor Name'
            }
        }
    }
}

# Export results
$resultsPath = Join-Path $resultsDir "atomic_results.csv"
$results | Export-Csv -Path $resultsPath -NoTypeInformation

Write-Host "Results saved to: $resultsPath"
return $resultsPath