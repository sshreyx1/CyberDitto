package handlers

import (
    "cyberditto-backend/internal/core/scan"
    "cyberditto-backend/internal/core/deploy"
    "encoding/json"
    "github.com/gorilla/mux"
    "net/http"
    "log"
)

type Handlers struct {
    scanService    *scan.Service
    deployService  *deploy.Service
}

func NewHandlers() *Handlers {
    return &Handlers{
        scanService:    scan.NewService(),
        deployService:  deploy.NewService(),
    }
}

func (h *Handlers) StartScan(w http.ResponseWriter, r *http.Request) {
    // Check for admin privileges
    elevated, err := h.scanService.CheckElevation()
    if err != nil || !elevated {
        log.Printf("Scan requires administrative privileges")
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusForbidden)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": "Administrative privileges required",
            "details": "Security scan requires administrative privileges to run.",
            "requires_elevation": true,
        })
        return
    }

    scanID, err := h.scanService.StartScan()
    if err != nil {
        log.Printf("Failed to start scan: %v", err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": err.Error(),
            "details": "Security scan failed. Please ensure you have administrative privileges.",
            "requires_elevation": true,
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "scan_id": scanID,
        "requires_elevation": true,
    })
}

func (h *Handlers) GetScanStatus(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    scanID := vars["id"]

    status, err := h.scanService.GetStatus(scanID)
    if err != nil {
        log.Printf("Error getting scan status: %v", err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "phase": "error",
            "progress": 0,
            "error": err.Error(),
            "details": "Failed to get scan status. The scan may have encountered an error.",
        })
        return
    }

    json.NewEncoder(w).Encode(status)
}

func (h *Handlers) GetScanResult(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    scanID := vars["id"]

    result, err := h.scanService.GetResult(scanID)
    if err != nil {
        log.Printf("Error getting scan result: %v", err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": err.Error(),
            "details": "Failed to retrieve scan results.",
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}