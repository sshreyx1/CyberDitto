package vagrant

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "sync"
    "time"
)

type Manager struct {
    baseDir      string
    projectRoot  string
    activeVMDirs sync.Map
}

func NewManager() *Manager {
    cwd, err := os.Getwd()
    if err != nil {
        log.Printf("Error getting working directory: %v", err)
        return &Manager{baseDir: "vagrant"}
    }

    projectRoot := filepath.Join(cwd, "..", "..")
    projectRoot, err = filepath.Abs(projectRoot)
    if err != nil {
        log.Printf("Error getting absolute path for project root: %v", err)
        return &Manager{baseDir: "vagrant"}
    }

    vagrantDir := filepath.Join(projectRoot, "vagrant")
    if err := os.MkdirAll(vagrantDir, 0755); err != nil {
        log.Printf("Error creating vagrant directory: %v", err)
    }

    manager := &Manager{
        baseDir:      vagrantDir,
        projectRoot:  projectRoot,
        activeVMDirs: sync.Map{},
    }

    // Scan for existing VMs
    manager.scanExistingVMs()

    log.Printf("Initialized Vagrant Manager with baseDir: %s", vagrantDir)
    return manager
}

func (m *Manager) scanExistingVMs() {
    // Read all directories in vagrant folder
    entries, err := os.ReadDir(m.baseDir)
    if err != nil {
        log.Printf("Error scanning vagrant directory: %v", err)
        return
    }

    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        vmDir := filepath.Join(m.baseDir, entry.Name())
        vagrantfilePath := filepath.Join(vmDir, "Vagrantfile")

        // Check if directory contains a Vagrantfile
        if _, err := os.Stat(vagrantfilePath); err == nil {
            vmID := entry.Name()
            m.activeVMDirs.Store(vmID, vmDir)
            log.Printf("Found existing VM: %s in directory: %s", vmID, vmDir)
        }
    }
}

func (m *Manager) ListAllVMs() ([]VMInfo, error) {
    var vms []VMInfo

    m.activeVMDirs.Range(func(key, value interface{}) bool {
        vmID := key.(string)
        vmDir := value.(string)

        // Read VM metadata if exists
        metadataPath := filepath.Join(vmDir, "vm_metadata.json")
        var metadata VMMetadata
        if data, err := os.ReadFile(metadataPath); err == nil {
            if err := json.Unmarshal(data, &metadata); err == nil {
                status, _ := m.GetStatus(vmID)
                vms = append(vms, VMInfo{
                    ID:        vmID,
                    Name:      fmt.Sprintf("CyberDitto_%s", vmID),
                    Status:    status,
                    CreatedAt: metadata.CreatedAt,
                    ScanID:    metadata.ScanID,
                })
            }
        }
        return true
    })

    return vms, nil
}

type VMInfo struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"createdAt"`
    ScanID    string    `json:"scanId"`
}

type VMMetadata struct {
    CreatedAt time.Time `json:"createdAt"`
    ScanID    string    `json:"scanId"`
}

func (m *Manager) Deploy(scanID string) (string, error) {
    vmID := fmt.Sprintf("vm_%d", time.Now().UnixNano())
    deployDir := filepath.Join(m.baseDir, vmID)
    if err := os.MkdirAll(deployDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create deployment directory: %v", err)
    }

    m.activeVMDirs.Store(vmID, deployDir)

    // Save VM metadata
    metadata := VMMetadata{
        CreatedAt: time.Now(),
        ScanID:    scanID,
    }
    metadataBytes, err := json.Marshal(metadata)
    if err == nil {
        metadataPath := filepath.Join(deployDir, "vm_metadata.json")
        if err := os.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
            log.Printf("Warning: Failed to save VM metadata: %v", err)
        }
    }
   
    outputDir := filepath.Join(m.projectRoot, "output")
    scanResults, outputPath, err := m.findLatestScanResults(outputDir, scanID)
    if err != nil {
        return "", fmt.Errorf("failed to prepare scan results: %v", err)
    }
    log.Printf("Found scan results at: %s", outputPath)


    if err := m.initVagrantEnv(deployDir, vmID, scanResults); err != nil {
        return "", fmt.Errorf("failed to initialize vagrant environment: %v", err)
    }


    log.Printf("Starting Vagrant VM in directory: %s", deployDir)
    cmd := exec.Command("vagrant", "up")
    cmd.Dir = deployDir
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("Vagrant up failed with output: %s", string(output))
        return "", fmt.Errorf("vagrant up failed: %v\nOutput: %s", err, output)
    }


    // Verify VM is running
    status, err := m.GetStatus(vmID)
    if err != nil {
        log.Printf("Warning: Initial status check failed: %v", err)
    } else if status != "running" {
        log.Printf("Warning: VM status after creation is %s", status)
    }


    log.Printf("Successfully created VM with ID: %s in directory: %s", vmID, deployDir)
    return vmID, nil
}


func (m *Manager) findLatestScanResults(outputDir string, scanID string) (map[string]interface{}, string, error) {
    log.Printf("Searching for scan results in: %s", outputDir)
   
    entries, err := os.ReadDir(outputDir)
    if err != nil {
        return nil, "", fmt.Errorf("failed to read output directory: %v", err)
    }


    var latestTime int64
    var latestPath string


    for _, entry := range entries {
        if entry.IsDir() && strings.HasPrefix(entry.Name(), "SecurityPosture_") {
            jsonPath := filepath.Join(outputDir, entry.Name(), "security_posture.json")
            if info, err := os.Stat(jsonPath); err == nil {
                if info.ModTime().UnixNano() > latestTime {
                    latestTime = info.ModTime().UnixNano()
                    latestPath = jsonPath
                }
            }
        }
    }


    if latestPath == "" {
        return nil, "", fmt.Errorf("no scan results found for scan ID: %s", scanID)
    }


    log.Printf("Found latest scan results at: %s", latestPath)


    data, err := os.ReadFile(latestPath)
    if err != nil {
        return nil, "", fmt.Errorf("failed to read scan results: %v", err)
    }


    var results map[string]interface{}
    if err := json.Unmarshal(data, &results); err != nil {
        return nil, "", fmt.Errorf("failed to parse scan results: %v", err)
    }


    return results, latestPath, nil
}


func (m *Manager) initVagrantEnv(vmDir string, vmID string, scanResults map[string]interface{}) error {
    vagrantfileContent := fmt.Sprintf(`Vagrant.configure("2") do |config|
  config.vm.box = "StefanScherer/windows_11"
  config.vm.box_version = "2021.12.09"
  config.vm.define "%s" do |vm|
    vm.vm.provider "virtualbox" do |vb|
      vb.memory = "4096"
      vb.cpus = 2
      vb.gui = true
      vb.name = "CyberDitto_%s"
      vb.customize ["modifyvm", :id, "--vram", "128"]
      vb.customize ["modifyvm", :id, "--accelerate3d", "on"]
    end
  end


  config.vm.provision "file", source: "security_posture.json", destination: "C:/security_posture.json"
 
  config.vm.provision "shell", inline: <<-SHELL
    Write-Host "Setting up Windows 11 environment..."
  SHELL
end`, vmID, vmID)


    vagrantfilePath := filepath.Join(vmDir, "Vagrantfile")
    if err := os.WriteFile(vagrantfilePath, []byte(vagrantfileContent), 0644); err != nil {
        return fmt.Errorf("failed to write Vagrantfile: %v", err)
    }


    jsonData, err := json.Marshal(scanResults)
    if err != nil {
        return fmt.Errorf("failed to marshal scan results: %v", err)
    }


    if err := os.WriteFile(filepath.Join(vmDir, "security_posture.json"), jsonData, 0644); err != nil {
        return fmt.Errorf("failed to copy scan results: %v", err)
    }


    log.Printf("Created Vagrantfile and copied scan results at: %s", vmDir)
    return nil
}


func (m *Manager) GetStatus(vmID string) (string, error) {
    dirInterface, exists := m.activeVMDirs.Load(vmID)
    if !exists {
        return "", fmt.Errorf("VM directory not found for ID: %s", vmID)
    }
    vmDir := dirInterface.(string)


    // Try VBoxManage showvminfo first
    vboxCmd := exec.Command("VBoxManage", "showvminfo", fmt.Sprintf("CyberDitto_%s", vmID), "--machinereadable")
    vboxOutput, err := vboxCmd.Output()
    if err == nil {
        status := parseVBoxStatus(string(vboxOutput))
        if status != "unknown" {
            log.Printf("VM %s status (VBox): %s", vmID, status)
            return status, nil
        }
    }


    // Fallback to vagrant status
    cmd := exec.Command("vagrant", "status")
    cmd.Dir = vmDir
    output, err := cmd.CombinedOutput()
    if err != nil {
        globalCmd := exec.Command("vagrant", "global-status", "--prune")
        globalOutput, err := globalCmd.CombinedOutput()
        if err != nil {
            return "unknown", fmt.Errorf("failed to get status: %v", err)
        }
        return parseVagrantStatus(string(globalOutput), vmID)
    }


    return parseVagrantStatus(string(output), vmID)
}


func parseVBoxStatus(output string) string {
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if strings.Contains(line, "VMState=") {
            state := strings.Split(line, "=")[1]
            state = strings.Trim(state, "\"")
            switch strings.ToLower(state) {
            case "running":
                return "running"
            case "poweroff", "aborted":
                return "stopped"
            case "saved":
                return "saved"
            case "notcreated":
                return "not_created"
            }
        }
    }
    return "unknown"
}


func parseVagrantStatus(output string, vmID string) (string, error) {
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        if strings.Contains(line, vmID) || strings.Contains(line, "default") {
            line = strings.ToLower(line)
            if strings.Contains(line, "running") {
                return "running", nil
            } else if strings.Contains(line, "poweroff") || strings.Contains(line, "aborted") || strings.Contains(line, "stopped") {
                return "stopped", nil
            } else if strings.Contains(line, "saved") {
                return "saved", nil
            } else if strings.Contains(line, "not created") {
                return "not_created", nil
            }
        }
    }
    return "not_created", nil
}


func (m *Manager) Start(vmID string) error {
    dirInterface, exists := m.activeVMDirs.Load(vmID)
    if !exists {
        return fmt.Errorf("VM directory not found for ID: %s", vmID)
    }
    vmDir := dirInterface.(string)


    // Try VBoxManage first
    vboxCmd := exec.Command("VBoxManage", "startvm", fmt.Sprintf("CyberDitto_%s", vmID))
    if err := vboxCmd.Run(); err == nil {
        return nil
    }


    // Fallback to vagrant
    cmd := exec.Command("vagrant", "up")
    cmd.Dir = vmDir
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("failed to start VM: %v\nOutput: %s", err, output)
    }
    return nil
}


func (m *Manager) Stop(vmID string) error {
    dirInterface, exists := m.activeVMDirs.Load(vmID)
    if !exists {
        return fmt.Errorf("VM directory not found for ID: %s", vmID)
    }
    vmDir := dirInterface.(string)


    // Try VBoxManage first
    vboxCmd := exec.Command("VBoxManage", "controlvm", fmt.Sprintf("CyberDitto_%s", vmID), "poweroff")
    if err := vboxCmd.Run(); err == nil {
        return nil
    }


    // Fallback to vagrant
    cmd := exec.Command("vagrant", "halt", "-f")
    cmd.Dir = vmDir
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("failed to stop VM: %v\nOutput: %s", err, output)
    }
    return nil
}


func (m *Manager) Destroy(vmID string) error {
    dirInterface, exists := m.activeVMDirs.Load(vmID)
    if !exists {
        return fmt.Errorf("VM directory not found for ID: %s", vmID)
    }
    vmDir := dirInterface.(string)


    // Try VBoxManage first
    vboxCmd := exec.Command("VBoxManage", "unregistervm", fmt.Sprintf("CyberDitto_%s", vmID), "--delete")
    if err := vboxCmd.Run(); err == nil {
        if err := os.RemoveAll(vmDir); err != nil {
            log.Printf("Warning: failed to remove VM directory: %v", err)
        }
        m.activeVMDirs.Delete(vmID)
        return nil
    }


    // Fallback to vagrant
    cmd := exec.Command("vagrant", "destroy", "-f")
    cmd.Dir = vmDir
    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("failed to destroy VM: %v\nOutput: %s", err, output)
    }


    if err := os.RemoveAll(vmDir); err != nil {
        log.Printf("Warning: failed to remove VM directory: %v", err)
    }
    m.activeVMDirs.Delete(vmID)
    return nil
}