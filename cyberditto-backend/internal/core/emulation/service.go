// internal/core/emulation/service.go

package emulation

import (
    "cyberditto-backend/pkg/vagrant"
    "encoding/csv"
    "fmt"
    "io"
    "log"
    "os"
	"os/exec"
    "path/filepath"
    "sync"
    "time"
	"strings"
)

type Service struct {
    executions  map[string]*Execution
    executor    *Executor
    vagrant     *vagrant.Manager
    mutex       sync.RWMutex
    baseDir     string
    projectRoot string
}

func NewService(projectRoot string) *Service {
    vagrantMgr := vagrant.NewManager()
    return &Service{
        executions:  make(map[string]*Execution),
        executor:    NewExecutor(vagrantMgr, projectRoot),
        vagrant:     vagrantMgr,
        baseDir:     filepath.Join(projectRoot, "emulation_results"),
        projectRoot: projectRoot,
    }
}

func (s *Service) StartEmulation(deploymentID string, mode EmulationMode, target string, testNums string) (string, error) {
    // Validate inputs
    if err := s.validateRequest(mode, target, testNums); err != nil {
        return "", fmt.Errorf("invalid request: %v", err)
    }

    // Create execution ID and directory
    executionID := fmt.Sprintf("emulation_%d", time.Now().UnixNano())
    execDir, err := s.createExecutionDirectory(executionID)
    if err != nil {
        return "", err
    }

    // Create execution record
    execution := &Execution{
        ID:           executionID,
        DeploymentID: deploymentID,
        Mode:         mode,
        Target:       target,
        TestNumbers:  testNums,
        StartTime:    time.Now(),
        Status: &EmulationStatus{
            Phase:     "preparing",
            Progress:  0,
            Message:   "Preparing emulation environment...",
            StartTime: time.Now(),
        },
    }

    // Store execution
    s.mutex.Lock()
    s.executions[executionID] = execution
    s.mutex.Unlock()

    // Start execution in background
    go s.runEmulation(execution, execDir)

    return executionID, nil
}

func (s *Service) runEmulation(execution *Execution, execDir string) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic in emulation %s: %v", execution.ID, r)
            s.updateStatus(execution.ID, "error", 0, fmt.Sprintf("Internal error: %v", r))
        }
    }()

    // Prepare execution config
    config := ExecutionConfig{
        DeploymentID: execution.DeploymentID,
        Mode:         execution.Mode,
        Target:       execution.Target,
        TestNumbers:  execution.TestNumbers,
        OutputPath:   execDir,
    }

    // Execute tests
    s.updateStatus(execution.ID, "running", 25, "Executing attack simulation...")
    result, err := s.executor.Execute(config)
    if err != nil {
        s.updateStatus(execution.ID, "error", 0, fmt.Sprintf("Execution failed: %v", err))
        return
    }

    // Process results
    s.updateStatus(execution.ID, "processing", 75, "Processing results...")
    emulationResult, err := s.processResults(execution.ID, result.ResultsPath)
    if err != nil {
        s.updateStatus(execution.ID, "error", 0, fmt.Sprintf("Failed to process results: %v", err))
        return
    }

    // Update execution with results
    s.mutex.Lock()
    execution.Result = emulationResult
    execution.CompletedAt = time.Now()
    execution.Status.Phase = "completed"
    execution.Status.Progress = 100
    execution.Status.Message = "Emulation completed successfully"
    execution.Status.CompletedAt = time.Now()
    s.mutex.Unlock()
}

func (s *Service) GetStatus(executionID string) (*EmulationStatus, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    execution, exists := s.executions[executionID]
    if !exists {
        return nil, fmt.Errorf("execution not found: %s", executionID)
    }

    return execution.Status, nil
}

func (s *Service) GetResult(executionID string) (*EmulationResult, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    execution, exists := s.executions[executionID]
    if !exists {
        return nil, fmt.Errorf("execution not found: %s", executionID)
    }

    if execution.Result == nil {
        return nil, fmt.Errorf("results not yet available for: %s", executionID)
    }

    return execution.Result, nil
}

func (s *Service) processResults(executionID string, resultsPath string) (*EmulationResult, error) {
    file, err := os.Open(resultsPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open results file: %v", err)
    }
    defer file.Close()

    reader := csv.NewReader(file)
    var results []TestResult
    var totalTests, passedTests int

    // Skip header row
    _, err = reader.Read()
    if err != nil {
        return nil, fmt.Errorf("failed to read CSV header: %v", err)
    }

    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("failed to read CSV record: %v", err)
        }

        result := TestResult{
            Timestamp:     record[0],
            Status:        record[1],
            Tactic:        record[2],
            Technique:     record[3],
            TechniqueName: record[4],
            TestNumber:    record[5],
            TestName:      record[6],
            TestGUID:      record[7],
            ExecutorName:  record[8],
        }

        // Add error message if present
        if len(record) > 9 {
            result.ErrorMessage = record[9]
        }

        results = append(results, result)
        totalTests++
        if record[1] == "Pass" {
            passedTests++
        }
    }

    successRate := float64(0)
    if totalTests > 0 {
        successRate = (float64(passedTests) / float64(totalTests)) * 100
    }

    return &EmulationResult{
        ID:          executionID,
        Status:      "completed",
        Summary: TestSummary{
            TotalTests:  totalTests,
            PassedTests: passedTests,
            FailedTests: totalTests - passedTests,
            SuccessRate: successRate,
        },
        Results:     results,
        CreatedAt:   time.Now(),
        CompletedAt: time.Now(),
    }, nil
}

func (s *Service) CancelEmulation(executionID string) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    execution, exists := s.executions[executionID]
    if !exists {
        return fmt.Errorf("execution not found: %s", executionID)
    }

    if execution.Status.Phase == "completed" || execution.Status.Phase == "error" {
        return fmt.Errorf("cannot cancel completed execution: %s", executionID)
    }

    execution.Status.Phase = "cancelled"
    execution.Status.Progress = 0
    execution.Status.Message = "Emulation cancelled by user"
    execution.Status.CompletedAt = time.Now()

    // Attempt to clean up any running processes on the VM
    if err := s.cleanupExecution(execution); err != nil {
        log.Printf("Warning: cleanup after cancellation failed for %s: %v", executionID, err)
    }

    return nil
}

func (s *Service) cleanupExecution(execution *Execution) error {
    vagrantCmd := fmt.Sprintf(`
        vagrant powershell -c "
            Get-Process | Where-Object {$_.Name -like '*atomic*'} | Stop-Process -Force;
            Remove-Item -Path 'C:\AtomicScripts\*' -Force -ErrorAction SilentlyContinue
        " -- %s`, execution.DeploymentID)

    cmd := exec.Command("powershell", "-Command", vagrantCmd)
    cmd.Dir = filepath.Join(s.projectRoot, "vagrant")
    
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("cleanup failed: %v, output: %s", err, string(output))
    }

    return nil
}


func (s *Service) updateStatus(executionID string, phase string, progress float64, message string) {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    if execution, exists := s.executions[executionID]; exists {
        execution.Status.Phase = phase
        execution.Status.Progress = progress
        execution.Status.Message = message

        if phase == "completed" || phase == "error" {
            execution.Status.CompletedAt = time.Now()
        }

        log.Printf("Updated emulation status - ID: %s, Phase: %s, Progress: %.2f%%, Message: %s",
            executionID, phase, progress, message)
    }
}

func (s *Service) validateRequest(mode EmulationMode, target string, testNums string) error {
    switch mode {
    case ModeTactic:
        if target == "" {
            return fmt.Errorf("tactic name is required")
        }
    case ModeTechnique:
        if target == "" {
            return fmt.Errorf("technique ID is required")
        }
        if !s.isValidTechnique(target) {
            return fmt.Errorf("invalid technique ID: %s", target)
        }
    case ModeCustom:
        techniques := strings.Split(target, ",")
        for _, technique := range techniques {
            if !s.isValidTechnique(strings.TrimSpace(technique)) {
                return fmt.Errorf("invalid technique ID in chain: %s", technique)
            }
        }
    default:
        return fmt.Errorf("invalid mode: %s", mode)
    }
    return nil
}

func (s *Service) isValidTechnique(techniqueID string) bool {
    // TODO: Implement technique validation against ATT&CK framework
    // For now, just check basic format (T####)
    return len(techniqueID) == 5 && techniqueID[0] == 'T' && 
           s.isNumeric(techniqueID[1:])
}

func (s *Service) isNumeric(str string) bool {
    for _, char := range str {
        if char < '0' || char > '9' {
            return false
        }
    }
    return true
}

func (s *Service) createExecutionDirectory(executionID string) (string, error) {
    execDir := filepath.Join(s.baseDir, executionID)
    if err := os.MkdirAll(execDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create execution directory: %v", err)
    }
    return execDir, nil
}

func (s *Service) CleanupOldEmulations(maxAge time.Duration) {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    now := time.Now()
    for executionID, execution := range s.executions {
        if now.Sub(execution.StartTime) > maxAge {
            // Clean up result files
            execDir := filepath.Join(s.baseDir, executionID)
            if err := os.RemoveAll(execDir); err != nil {
                log.Printf("Warning: failed to cleanup directory for %s: %v", executionID, err)
            }
            delete(s.executions, executionID)
            log.Printf("Cleaned up old emulation: %s", executionID)
        }
    }
}

func (s *Service) GetAllEmulations() ([]*Execution, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    executions := make([]*Execution, 0, len(s.executions))
    for _, execution := range s.executions {
        executions = append(executions, execution)
    }

    return executions, nil
}