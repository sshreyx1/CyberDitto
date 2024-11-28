package powershell

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "time"
)

type Runner struct {
    scriptPath   string
    projectRoot  string
}

type ScanProgress struct {
    Stage    string  `json:"stage"`
    Progress int     `json:"progress"`
    Status   string  `json:"status"`
    Phase    string  `json:"phase"`
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
        return &Runner{scriptPath: scriptPath}
    }

    log.Printf("Successfully found script at: %s", scriptPath)
    return &Runner{
        scriptPath:  scriptPath,
        projectRoot: projectRoot,
    }
}

func (r *Runner) RunConfigScript() (string, error) {
    // Verify script exists before proceeding
    if _, err := os.Stat(r.scriptPath); err != nil {
        return "", fmt.Errorf("script not found at path %s: %v", r.scriptPath, err)
    }

    // Create output directory
    outputDir := filepath.Join(r.projectRoot, "output", fmt.Sprintf("SecurityPosture_%s", time.Now().Format("20060102_150405")))
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create output directory: %v", err)
    }

    // Copy script to output directory
    tempScriptPath := filepath.Join(outputDir, "config_collector.ps1")
    scriptContent, err := os.ReadFile(r.scriptPath)
    if err != nil {
        return "", fmt.Errorf("failed to read script: %v", err)
    }

    if err := os.WriteFile(tempScriptPath, scriptContent, 0644); err != nil {
        return "", fmt.Errorf("failed to copy script: %v", err)
    }

    // Run PowerShell script
    cmd := exec.Command("powershell.exe",
        "-WindowStyle", "Hidden",
        "-NoProfile",
        "-ExecutionPolicy", "Bypass",
        "-NonInteractive",
        "-NoLogo",
        "-File", tempScriptPath)
    cmd.Dir = outputDir

    // Create channels for progress tracking
    progressChan := make(chan ScanProgress)
    errorChan := make(chan error)
    doneChan := make(chan bool)
    logChan := make(chan string)

    // Start the script
    if err := cmd.Start(); err != nil {
        return "", fmt.Errorf("failed to start script: %v", err)
    }

    // Monitor progress file in separate goroutine
    go func() {
        defer close(progressChan)
        defer close(logChan)
        
        var lastProgress int
        var lastStage string
        progressFile := filepath.Join(outputDir, "progress.json")
        logFile := filepath.Join(outputDir, "scan.log")
        
        for {
            select {
            case <-doneChan:
                return
            default:
                // Check progress file
                if content, err := os.ReadFile(progressFile); err == nil {
                    var progress ScanProgress
                    if err := json.Unmarshal(content, &progress); err == nil {
                        if progress.Progress != lastProgress || progress.Stage != lastStage {
                            progressChan <- progress
                            lastProgress = progress.Progress
                            lastStage = progress.Stage
                        }
                    }
                }

                // Check log file
                if content, err := os.ReadFile(logFile); err == nil {
                    logChan <- string(content)
                }

                time.Sleep(500 * time.Millisecond)
            }
        }
    }()

    // Wait for completion in separate goroutine
    go func() {
        if err := cmd.Wait(); err != nil {
            errorChan <- err
        }
        close(doneChan)
    }()

    // Monitor progress and completion
    timeout := time.After(5 * time.Minute)
    var lastProgress int
    var lastLog string

    for {
        select {
        case progress, ok := <-progressChan:
            if ok {
                lastProgress = progress.Progress
                log.Printf("Scan progress: %s - %d%% - %s", 
                    progress.Stage, 
                    progress.Progress, 
                    progress.Status)
            }
        case logContent, ok := <-logChan:
            if ok && logContent != lastLog {
                lastLog = logContent
                log.Printf("Scan log: %s", logContent)
            }
        case err := <-errorChan:
            return "", fmt.Errorf("script execution failed: %v", err)
        case <-timeout:
            cmd.Process.Kill()
            return "", fmt.Errorf("scan timed out after 5 minutes")
        case <-doneChan:
            // Final check for JSON file
            jsonPath := filepath.Join(outputDir, "security_posture.json")
            if _, err := os.Stat(jsonPath); err != nil {
                return "", fmt.Errorf("scan output file not found: %v", err)
            }

            // Validate JSON content
            content, err := os.ReadFile(jsonPath)
            if err != nil {
                return "", fmt.Errorf("failed to read scan results: %v", err)
            }

            var result map[string]interface{}
            if err := json.Unmarshal(content, &result); err != nil {
                return "", fmt.Errorf("invalid scan results format: %v", err)
            }

            if lastProgress < 100 {
                log.Printf("Scan completed successfully: 100%%")
            }
            return jsonPath, nil
        }
    }
}

func (r *Runner) ValidateElevation() bool {
    cmd := exec.Command("powershell.exe",
        "-WindowStyle", "Hidden",
        "-Command",
        "[Security.Principal.WindowsIdentity]::GetCurrent().Groups -contains 'S-1-5-32-544'")
    
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    
    err := cmd.Run()
    if err != nil {
        log.Printf("Elevation check failed: %v\nError output: %s", err, stderr.String())
        return false
    }
    
    return true
}

func (r *Runner) GetLogFile(outputDir string) string {
    logFile := filepath.Join(outputDir, "scan.log")
    content, err := os.ReadFile(logFile)
    if err != nil {
        return ""
    }
    return string(content)
}

func (r *Runner) GetProgressFile(outputDir string) (*ScanProgress, error) {
    progressFile := filepath.Join(outputDir, "progress.json")
    content, err := os.ReadFile(progressFile)
    if err != nil {
        return nil, err
    }

    var progress ScanProgress
    if err := json.Unmarshal(content, &progress); err != nil {
        return nil, err
    }

    return &progress, nil
}