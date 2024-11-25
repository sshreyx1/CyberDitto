package deploy


import (
    "cyberditto-backend/internal/models"
    "cyberditto-backend/pkg/vagrant"
    "fmt"
    "log"
    "sync"
    "time"
    "strings"
)


type DeploymentInfo struct {
    ID            string
    Name          string
    CreatedAt     time.Time
    LastActive    time.Time
    Status        string
    Progress      float64
    Message       string
    Error         string
    VagrantID     string
    ScanID        string
    ResourceUsage *models.ResourceUsage
    ScanResult    *models.ScanResult
}


type Service struct {
    deployments map[string]*DeploymentInfo
    vagrant     *vagrant.Manager
    mutex       sync.RWMutex
}


func NewService() *Service {
    return &Service{
        deployments: make(map[string]*DeploymentInfo),
        vagrant:     vagrant.NewManager(),
    }
}


func (s *Service) StartDeployment(scanID string) (string, error) {
    deployID := fmt.Sprintf("deploy_%d", time.Now().UnixNano())
    
    log.Printf("Starting new deployment with ID: %s for scan: %s", deployID, scanID)
    
    s.mutex.Lock()
    s.deployments[deployID] = &DeploymentInfo{
        ID:        deployID,
        CreatedAt: time.Now(),
        LastActive: time.Now(),
        Status:    "preparing",
        Progress:  0,
        Message:   "Preparing deployment...",
        ScanID:    scanID,
        ResourceUsage: &models.ResourceUsage{
            CPU:    0,
            Memory: 0,
            Disk:   0,
        },
    }
    s.mutex.Unlock()

    go s.runDeployment(deployID, scanID)
    
    return deployID, nil
}

func (s *Service) runDeployment(deployID, scanID string) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic in runDeployment: %v", r)
            s.updateStatus(deployID, "error", 0, "", fmt.Sprintf("Internal error: %v", r))
        }
    }()

    s.updateStatus(deployID, "initializing", 10, "Setting up Vagrant environment...", "")
    
    vagrantID, err := s.vagrant.Deploy(scanID)
    if err != nil {
        log.Printf("Deployment error: %v", err)
        s.updateStatus(deployID, "error", 0, "", fmt.Sprintf("Deployment failed: %v", err))
        return
    }

    s.mutex.Lock()
    if deployment := s.deployments[deployID]; deployment != nil {
        deployment.VagrantID = vagrantID
        deployment.Name = fmt.Sprintf("VM_%s", vagrantID)
        deployment.Status = "deploying"
        deployment.Progress = 50
        deployment.Message = "Creating virtual environment..."
    }
    s.mutex.Unlock()

    // Wait for VM to be ready
    retries := 0
    maxRetries := 60
    for retries < maxRetries {
        status, err := s.vagrant.GetStatus(vagrantID)
        if err != nil {
            log.Printf("Error checking VM status: %v", err)
            retries++
            time.Sleep(5 * time.Second)
            continue
        }

        if strings.ToLower(status) == "running" {
            log.Printf("VM deployed with ID: %s, Status: %s", vagrantID, status)
            
            s.mutex.Lock()
            if deployment := s.deployments[deployID]; deployment != nil {
                deployment.Status = "running"
                deployment.Progress = 100
                deployment.Message = "Virtual environment is ready"
                deployment.LastActive = time.Now()
                deployment.Name = fmt.Sprintf("VM_%s", vagrantID)
            }
            s.mutex.Unlock()
            return
        }

        retries++
        time.Sleep(5 * time.Second)
    }

    s.updateStatus(deployID, "error", 0, "", "Deployment timed out")
}


func (s *Service) GetStatus(deployID string) (*models.DeploymentStatus, error) {
    s.mutex.RLock()
    deployment, exists := s.deployments[deployID]
    s.mutex.RUnlock()
   
    if !exists {
        log.Printf("Deployment not found: %s", deployID)
        return nil, fmt.Errorf("deployment not found: %s", deployID)
    }

    // Log initial deployment details
    log.Printf("Deployment %s initial state - Status: %s, Progress: %.1f%%, VagrantID: %s", 
        deployID, deployment.Status, deployment.Progress, deployment.VagrantID)

    status := &models.DeploymentStatus{
        Status:        deployment.Status,
        Progress:      deployment.Progress,
        Message:       deployment.Message,
        Error:         deployment.Error,
        VagrantID:     deployment.VagrantID,
        ResourceUsage: deployment.ResourceUsage,
    }

    log.Printf("Created status object for deployment %s: %+v", deployID, status)

    // Only check VM status if we have a VagrantID and are in a stable state
    if deployment.VagrantID != "" {
        log.Printf("Checking VM status for deployment %s with VagrantID %s", deployID, deployment.VagrantID)
        
        vmStatus, err := s.vagrant.GetStatus(deployment.VagrantID)
        if err != nil {
            log.Printf("Warning: Error getting VM status for deployment %s: %v", deployID, err)
            // Return current known status instead of error
            return status, nil
        }

        log.Printf("Retrieved VM status for deployment %s: %s", deployID, vmStatus)

        s.mutex.Lock()
        deployment.LastActive = time.Now()
        
        oldStatus := deployment.Status
        switch strings.ToLower(vmStatus) {
        case "running":
            deployment.Status = "running"
            deployment.Progress = 100
            deployment.Message = "Virtual environment is running"
            deployment.ResourceUsage = &models.ResourceUsage{
                CPU:    30,  // Example default values
                Memory: 40,
                Disk:   20,
            }
        case "poweroff", "aborted", "stopped":
            deployment.Status = "stopped"
            deployment.Progress = 0
            deployment.Message = "Virtual environment is stopped"
        case "saved":
            deployment.Status = "saved"
            deployment.Progress = 0
            deployment.Message = "Virtual environment is saved"
        case "not_created":
            deployment.Status = "error"
            deployment.Progress = 0
            deployment.Message = "Virtual environment not found"
        default:
            log.Printf("Unknown VM status received: %s", vmStatus)
        }
        
        log.Printf("Updated deployment %s status from %s to %s", deployID, oldStatus, deployment.Status)
        
        status.Status = deployment.Status
        status.Progress = deployment.Progress
        status.Message = deployment.Message
        status.ResourceUsage = deployment.ResourceUsage
        s.mutex.Unlock()

        log.Printf("Final status for deployment %s: Status=%s, Progress=%.1f%%, Message=%s", 
            deployID, status.Status, status.Progress, status.Message)
    } else {
        log.Printf("Deployment %s has no VagrantID, keeping current status", deployID)
    }

    return status, nil
}

func (s *Service) CancelDeployment(deployID string) error {
    s.mutex.Lock()
    deployment, exists := s.deployments[deployID]
    if !exists {
        s.mutex.Unlock()
        return fmt.Errorf("deployment not found")
    }


    // If deployment has a VM, destroy it
    vagrantID := deployment.VagrantID
    s.mutex.Unlock()


    if vagrantID != "" {
        log.Printf("Destroying VM for cancelled deployment: %s, vagrant ID: %s", deployID, vagrantID)
        if err := s.vagrant.Destroy(vagrantID); err != nil {
            log.Printf("Error destroying VM during cancellation: %v", err)
            // Continue with cancellation even if VM destroy fails
        }
    }


    s.mutex.Lock()
    deployment.Status = "cancelled"
    deployment.Progress = 0
    deployment.Message = "Deployment cancelled by user"
    deployment.VagrantID = ""
    s.mutex.Unlock()


    log.Printf("Successfully cancelled deployment: %s", deployID)
    return nil
}


func (s *Service) StartInstance(deployID string) error {
    s.mutex.Lock()
    deployment, exists := s.deployments[deployID]
    s.mutex.Unlock()


    if !exists {
        return fmt.Errorf("deployment not found")
    }


    if deployment.VagrantID == "" {
        return fmt.Errorf("no vagrant instance found")
    }


    log.Printf("Starting instance for deployment: %s, vagrant ID: %s", deployID, deployment.VagrantID)
   
    if err := s.vagrant.Start(deployment.VagrantID); err != nil {
        return err
    }


    s.updateStatus(deployID, "running", 100, "VM is running", "")
    return nil
}


func (s *Service) StopInstance(deployID string) error {
    s.mutex.Lock()
    deployment, exists := s.deployments[deployID]
    s.mutex.Unlock()


    if !exists {
        return fmt.Errorf("deployment not found")
    }


    if deployment.VagrantID == "" {
        return fmt.Errorf("no vagrant instance found")
    }


    log.Printf("Stopping instance for deployment: %s, vagrant ID: %s", deployID, deployment.VagrantID)
   
    if err := s.vagrant.Stop(deployment.VagrantID); err != nil {
        return err
    }


    s.updateStatus(deployID, "stopped", 0, "VM is stopped", "")
    return nil
}


func (s *Service) DeleteInstance(deployID string) error {
    s.mutex.Lock()
    deployment, exists := s.deployments[deployID]
    if !exists {
        s.mutex.Unlock()
        return fmt.Errorf("deployment not found")
    }


    if deployment.VagrantID == "" {
        s.mutex.Unlock()
        return fmt.Errorf("no vagrant instance found")
    }


    vagrantID := deployment.VagrantID
    s.mutex.Unlock()


    log.Printf("Deleting instance for deployment: %s, vagrant ID: %s", deployID, vagrantID)
   
    if err := s.vagrant.Destroy(vagrantID); err != nil {
        return fmt.Errorf("failed to destroy VM: %v", err)
    }


    s.mutex.Lock()
    delete(s.deployments, deployID)
    s.mutex.Unlock()


    return nil
}


func (s *Service) CleanupOldDeployments(maxAge time.Duration) {
    s.mutex.Lock()
    defer s.mutex.Unlock()


    for deployID, deployment := range s.deployments {
        if time.Since(deployment.CreatedAt) > maxAge {
            if deployment.VagrantID != "" {
                if err := s.vagrant.Destroy(deployment.VagrantID); err != nil {
                    log.Printf("Error cleaning up old deployment %s: %v", deployID, err)
                    continue
                }
            }
            delete(s.deployments, deployID)
            log.Printf("Cleaned up old deployment: %s", deployID)
        }
    }
}


func (s *Service) updateStatus(deployID, status string, progress float64, message string, err string) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
   
    deployment, exists := s.deployments[deployID]
    if !exists {
        log.Printf("Warning: Attempting to update status for non-existent deployment: %s", deployID)
        return
    }


    deployment.Status = status
    deployment.Progress = progress
    deployment.Message = message
    deployment.Error = err


    log.Printf("Updated deployment %s status: status=%s, progress=%.1f, message=%s, error=%s",
        deployID, status, progress, message, err)
}


func (s *Service) GetAllDeployments() ([]DeploymentInfo, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    log.Printf("Getting all deployments. Current count: %d", len(s.deployments))

    deployments := make([]DeploymentInfo, 0, len(s.deployments))
    for deployID, d := range s.deployments {
        log.Printf("Processing deployment %s - VagrantID: %s, Status: %s",
            deployID, d.VagrantID, d.Status)
        
        deployment := *d // Create a copy of deployment info
        
        // Only check VM status if we have a VagrantID
        if d.VagrantID != "" {
            if status, err := s.vagrant.GetStatus(d.VagrantID); err == nil {
                oldStatus := d.Status
                deployment.Status = strings.ToLower(status)
                deployment.LastActive = time.Now()
                
                // Update progress based on status
                switch deployment.Status {
                case "running":
                    deployment.Progress = 100
                    deployment.Message = "Virtual environment is running"
                case "stopped":
                    deployment.Progress = 0
                    deployment.Message = "Virtual environment is stopped"
                case "error":
                    deployment.Progress = 0
                    deployment.Message = "Error in virtual environment"
                }
                
                log.Printf("Updated status for deployment %s: %s -> %s",
                    deployID, oldStatus, deployment.Status)
            }
        }

        // Ensure resource usage exists
        if deployment.ResourceUsage == nil {
            deployment.ResourceUsage = &models.ResourceUsage{
                CPU:    30,
                Memory: 40,
                Disk:   20,
            }
        }

        // Update the name to match VM name if it exists
        if deployment.VagrantID != "" {
            deployment.Name = fmt.Sprintf("VM_%s", deployment.VagrantID)
        }

        deployments = append(deployments, deployment)
        log.Printf("Added deployment to list: %+v", deployment)
    }

    log.Printf("Returning %d deployments", len(deployments))
    return deployments, nil
}