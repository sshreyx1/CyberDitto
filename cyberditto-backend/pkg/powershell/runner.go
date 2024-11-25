package powershell

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "time"
)

type Runner struct {
    scriptPath string
    projectRoot string
}

func NewRunner() *Runner {
    // Get the current working directory
    cwd, err := os.Getwd()
    if err != nil {
        log.Printf("Error getting working directory: %v", err)
        return &Runner{scriptPath: "scripts/config_collector.ps1"}
    }
   
    // Navigate up to the project root from cmd/server
    projectRoot := filepath.Join(cwd, "..", "..")
    projectRoot, err = filepath.Abs(projectRoot)
    if err != nil {
        log.Printf("Error getting absolute path for project root: %v", err)
        return &Runner{scriptPath: "scripts/config_collector.ps1"}
    }
   
    // Construct the absolute path to the script
    scriptPath := filepath.Join(projectRoot, "scripts", "config_collector.ps1")
   
    // Log the paths for debugging
    log.Printf("Current working directory: %s", cwd)
    log.Printf("Project root directory: %s", projectRoot)
    log.Printf("Script path: %s", scriptPath)
   
    // Verify the script exists
    if _, err := os.Stat(scriptPath); err != nil {
        log.Printf("Error: Script not found at %s: %v", scriptPath, err)
        return &Runner{scriptPath: scriptPath} // Return with path anyway for error handling
    }

    log.Printf("Successfully found script at: %s", scriptPath)
    return &Runner{
        scriptPath: scriptPath,
        projectRoot: projectRoot,
    }
}

func (r *Runner) RunConfigScript() (string, error) {
    // Verify script exists before proceeding
    if _, err := os.Stat(r.scriptPath); err != nil {
        return "", fmt.Errorf("script not found at path %s: %v", r.scriptPath, err)
    }

    // Create output directory in project root
    outputDir := filepath.Join(r.projectRoot, "output", fmt.Sprintf("SecurityPosture_%s", time.Now().Format("20060102_150405")))
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create output directory: %v", err)
    }

    log.Printf("Using script path: %s", r.scriptPath)
    log.Printf("Output directory: %s", outputDir)

    // Copy script to output directory to ensure it has write permissions
    tempScriptPath := filepath.Join(outputDir, "config_collector.ps1")
    scriptContent, err := os.ReadFile(r.scriptPath)
    if err != nil {
        return "", fmt.Errorf("failed to read script: %v", err)
    }
    
    if err := os.WriteFile(tempScriptPath, scriptContent, 0644); err != nil {
        return "", fmt.Errorf("failed to copy script: %v", err)
    }

    // Run PowerShell script
    cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", tempScriptPath)
    cmd.Dir = outputDir

    // Capture and log output
    output, err := cmd.CombinedOutput()
    log.Printf("PowerShell Output: %s", string(output))
    
    if err != nil {
        log.Printf("PowerShell Error: %v", err)
        return "", fmt.Errorf("script execution failed: %v\nOutput: %s", err, output)
    }

    // Look for the JSON output file
    jsonPath := filepath.Join(outputDir, "security_posture.json")
    if _, err := os.Stat(jsonPath); err != nil {
        return "", fmt.Errorf("scan output file not found at %s: %v", jsonPath, err)
    }

    return jsonPath, nil
}

func (r *Runner) ValidateElevation() bool {
    cmd := exec.Command("powershell", "-Command",
        "[Security.Principal.WindowsIdentity]::GetCurrent().Groups -contains 'S-1-5-32-544'")
    return cmd.Run() == nil
}