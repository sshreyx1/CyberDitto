package models

import "time"

type SystemInfo struct {
    OSVersion   string `json:"osVersion"`
    CPUModel    string `json:"cpuModel"`
    Memory      int64  `json:"memory"`
    DiskSpace   int64  `json:"diskSpace"`
}

type NetworkInterface struct {
    Name       string `json:"name"`
    IPAddress  string `json:"ipAddress"`
    MACAddress string `json:"macAddress"`
}

type NetworkInfo struct {
    Interfaces []NetworkInterface `json:"interfaces"`
    OpenPorts  []int             `json:"openPorts"`
    DNSServers []string          `json:"dnsServers"`
    Services   []string          `json:"services"`
}

type FirewallInfo struct {
    Enabled bool `json:"enabled"`
    Rules   int  `json:"rules"`
}

type SecurityInfo struct {
    Firewall   FirewallInfo `json:"firewall"`
    Antivirus  []string     `json:"antivirus"`
    Updates    []string     `json:"updates"`
    UAC        struct {
        Enabled bool `json:"enabled"`
    } `json:"uac"`
}

type ScanResult struct {
    ID           string       `json:"id"`
    SystemInfo   SystemInfo   `json:"systemInfo"`
    NetworkInfo  NetworkInfo  `json:"networkInfo"`
    SecurityInfo SecurityInfo `json:"securityInfo"`
    CreatedAt    time.Time    `json:"createdAt"`
}

type ScanStatus struct {
    Phase    string  `json:"phase"`
    Progress float64 `json:"progress"`
    Message  string  `json:"message,omitempty"`
    Error    string  `json:"error,omitempty"`
}

type ResourceUsage struct {
    CPU    float64 `json:"cpu"`
    Memory float64 `json:"memory"`
    Disk   float64 `json:"disk"`
}

type DeploymentStatus struct {
    Status        string        `json:"status"`
    Progress      float64       `json:"progress"`
    Message       string        `json:"message,omitempty"`
    Error         string        `json:"error,omitempty"`
    VagrantID     string        `json:"vagrantId,omitempty"`
    ResourceUsage *ResourceUsage `json:"resourceUsage,omitempty"`
}