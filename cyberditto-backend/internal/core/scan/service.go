package scan

import (
    "cyberditto-backend/pkg/powershell"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sync"
    "time"
)

type Service struct {
    scans       map[string]*ScanStatus
    results     map[string]*json.RawMessage
    mutex       sync.RWMutex
    outputDirs  map[string]string
    runner      *powershell.Runner
}

type ScanStatus struct {
    Phase       string  `json:"phase"`
    Progress    float64 `json:"progress"`
    Message     string  `json:"message,omitempty"`
    Error       string  `json:"error,omitempty"`
    Stage       string  `json:"stage,omitempty"`
}

func NewService() *Service {
    return &Service{
        scans:      make(map[string]*ScanStatus),
        results:    make(map[string]*json.RawMessage),
        outputDirs: make(map[string]string),
        runner:     powershell.NewRunner(),
    }
}

func (s *Service) StartScan() (string, error) {
    scanID := fmt.Sprintf("scan_%d", time.Now().UnixNano())
    
    log.Printf("Starting new scan with ID: %s", scanID)
    
    s.mutex.Lock()
    s.scans[scanID] = &ScanStatus{
        Phase:    "scanning",
        Progress: 0,
        Message:  "Starting scan...",
    }
    s.mutex.Unlock()

    go s.runScan(scanID)
    
    return scanID, nil
}

func (s *Service) runScan(scanID string) {
    // Initialize scan status
    s.updateStatus(scanID, "scanning", 0, "Starting scan...", "")

    outputPath, err := s.runner.RunConfigScript()
    if err != nil {
        log.Printf("Scan error: %v", err)
        s.updateStatus(scanID, "error", 0, "", fmt.Sprintf("Scan failed: %v", err))
        return
    }

    // Store the output directory
    s.mutex.Lock()
    s.outputDirs[scanID] = filepath.Dir(outputPath)
    s.mutex.Unlock()

    // Parse results
    result, err := s.parseScanResults(outputPath)
    if err != nil {
        log.Printf("Failed to parse scan results: %v", err)
        s.updateStatus(scanID, "error", 0, "", fmt.Sprintf("Failed to process results: %v", err))
        return
    }

    // Store results
    s.mutex.Lock()
    s.scans[scanID].Phase = "completed"
    s.scans[scanID].Progress = 100
    s.scans[scanID].Message = "Scan completed successfully"
    rawResult := json.RawMessage(result)
    s.results[scanID] = &rawResult
    s.mutex.Unlock()

    log.Printf("Scan %s completed successfully", scanID)
}

func (s *Service) GetStatus(scanID string) (*ScanStatus, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    status, exists := s.scans[scanID]
    if !exists {
        return nil, fmt.Errorf("scan not found: %s", scanID)
    }

    outputDir := s.outputDirs[scanID]
    if outputDir != "" {
        if progress, err := s.runner.GetProgressFile(outputDir); err == nil && progress != nil {
            return &ScanStatus{
                Phase:    progress.Phase,
                Progress: float64(progress.Progress),  // Convert int to float64
                Message:  progress.Status,
                Stage:    progress.Stage,
            }, nil
        }
    }

    return status, nil
}

func (s *Service) GetResult(scanID string) (map[string]interface{}, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    result, exists := s.results[scanID]
    if !exists {
        return nil, fmt.Errorf("scan result not found: %s", scanID)
    }

    var data map[string]interface{}
    if err := json.Unmarshal(*result, &data); err != nil {
        return nil, fmt.Errorf("failed to parse scan result: %v", err)
    }

    return data, nil
}

func (s *Service) parseScanResults(path string) ([]byte, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read scan results: %v", err)
    }
    return data, nil
}

func (s *Service) CheckElevation() (bool, error) {
    return s.runner.ValidateElevation(), nil
}

func (s *Service) updateStatus(scanID, phase string, progress float64, message string, err string) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    status := s.scans[scanID]
    if status == nil {
        log.Printf("Warning: Attempting to update status for non-existent scan: %s", scanID)
        return
    }
    
    status.Phase = phase
    status.Progress = progress
    status.Message = message
    if err != "" {
        status.Error = err
    }

    log.Printf("Updated scan %s status: phase=%s, progress=%.1f, message=%s, error=%s",
        scanID, phase, progress, message, err)
}