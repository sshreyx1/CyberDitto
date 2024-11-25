package scan

import (
    "cyberditto-backend/internal/models"
    "cyberditto-backend/pkg/powershell"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "sync"
    "time"
    "bytes"
)

type Service struct {
    scans   map[string]*models.ScanStatus
    results map[string]*models.ScanResult
    mutex   sync.RWMutex
}

func NewService() *Service {
    return &Service{
        scans:   make(map[string]*models.ScanStatus),
        results: make(map[string]*models.ScanResult),
    }
}

func (s *Service) StartScan() (string, error) {
    scanID := fmt.Sprintf("scan_%d", time.Now().UnixNano())
    
    log.Printf("Starting new scan with ID: %s", scanID)
    
    s.mutex.Lock()
    s.scans[scanID] = &models.ScanStatus{
        Phase:    "starting",
        Progress: 0,
        Message:  "Initializing scan...",
    }
    s.mutex.Unlock()

    go s.runScan(scanID)
    
    return scanID, nil
}

func (s *Service) runScan(scanID string) {
    runner := powershell.NewRunner()
    
    // Log current working directory
    if cwd, err := os.Getwd(); err == nil {
        log.Printf("Scan Service working directory: %s", cwd)
    }
    
    // Update status for script execution
    s.updateStatus(scanID, "scanning", 10, "Running system scan...")
    
    outputPath, err := runner.RunConfigScript()
    if err != nil {
        log.Printf("Scan error: %v", err)
        s.updateStatus(scanID, "error", 0, "", fmt.Sprintf("Scan failed: %v", err))
        return
    }

    log.Printf("Scan completed successfully. Output at: %s", outputPath)

    // Parse results
    s.updateStatus(scanID, "processing", 50, "Processing scan results...")
    
    result, err := s.parseScanResults(outputPath)
    if err != nil {
        log.Printf("Failed to parse scan results: %v", err)
        s.updateStatus(scanID, "error", 0, "", fmt.Sprintf("Failed to process results: %v", err))
        return
    }

    // Store results
    s.mutex.Lock()
    result.ID = scanID
    result.CreatedAt = time.Now()
    s.results[scanID] = result
    s.mutex.Unlock()

    log.Printf("Scan %s completed and results stored successfully", scanID)
    s.updateStatus(scanID, "completed", 100, "Scan completed successfully")
}

func (s *Service) GetStatus(scanID string) (*models.ScanStatus, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    status, exists := s.scans[scanID]
    if !exists {
        return nil, fmt.Errorf("scan not found: %s", scanID)
    }
    return status, nil
}

func (s *Service) GetResult(scanID string) (*models.ScanResult, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    result, exists := s.results[scanID]
    if !exists {
        return nil, fmt.Errorf("scan result not found: %s", scanID)
    }
    return result, nil
}

func (s *Service) updateStatus(scanID, phase string, progress float64, message string, err ...string) {
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
    if len(err) > 0 {
        status.Error = err[0]
    }

    log.Printf("Updated scan %s status: phase=%s, progress=%.1f, message=%s, error=%v",
        scanID, phase, progress, message, err)
}

func (s *Service) parseScanResults(path string) (*models.ScanResult, error) {
    log.Printf("Parsing scan results from: %s", path)
    
    // Read file content
    data, err := os.ReadFile(path)
    if err != nil {
        log.Printf("Error reading scan results file: %v", err)
        return nil, fmt.Errorf("failed to read scan results: %v", err)
    }

    // Remove UTF-8 BOM if present
    data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

    // Try to read the file content for debugging
    log.Printf("File content preview (first 100 bytes): %q", string(data[:min(len(data), 100)]))

    var result models.ScanResult
    decoder := json.NewDecoder(bytes.NewReader(data))
    decoder.DisallowUnknownFields()
    
    if err := decoder.Decode(&result); err != nil {
        // Try to identify the specific JSON parsing error
        log.Printf("JSON parsing error: %v", err)
        // Log the full content for debugging
        log.Printf("Full file content: %s", string(data))
        return nil, fmt.Errorf("failed to parse scan results: %v", err)
    }

    // Verify required fields are present
    if err := validateScanResult(&result); err != nil {
        return nil, fmt.Errorf("invalid scan results: %v", err)
    }

    log.Printf("Successfully parsed scan results: OS=%s, Network Interfaces=%d",
        result.SystemInfo.OSVersion,
        len(result.NetworkInfo.Interfaces))

    return &result, nil
}

func validateScanResult(result *models.ScanResult) error {
    if result.SystemInfo.OSVersion == "" {
        return fmt.Errorf("missing OS version")
    }
    if result.SystemInfo.CPUModel == "" {
        return fmt.Errorf("missing CPU model")
    }
    if result.SystemInfo.Memory == 0 {
        return fmt.Errorf("invalid memory value")
    }
    if result.SystemInfo.DiskSpace == 0 {
        return fmt.Errorf("invalid disk space value")
    }
    return nil
}

func (s *Service) CancelScan(scanID string) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    status, exists := s.scans[scanID]
    if !exists {
        return fmt.Errorf("scan not found: %s", scanID)
    }

    if status.Phase == "completed" || status.Phase == "error" {
        return fmt.Errorf("cannot cancel scan in phase: %s", status.Phase)
    }

    status.Phase = "cancelled"
    status.Progress = 0
    status.Message = "Scan cancelled by user"

    log.Printf("Scan %s cancelled by user", scanID)
    return nil
}

func (s *Service) CleanupOldScans(maxAge time.Duration) {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    now := time.Now()
    for scanID, result := range s.results {
        if now.Sub(result.CreatedAt) > maxAge {
            delete(s.results, scanID)
            delete(s.scans, scanID)
            log.Printf("Cleaned up old scan: %s", scanID)
        }
    }
}

// Helper function for min
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}