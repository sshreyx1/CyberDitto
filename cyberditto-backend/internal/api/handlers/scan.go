// internal/api/handlers/scan.go

package handlers

import (
    "cyberditto-backend/internal/core/scan"
    "cyberditto-backend/internal/core/deploy"
    "cyberditto-backend/internal/core/emulation" 
    "encoding/json"
    "github.com/gorilla/mux"
    "net/http"
)

type Handlers struct {
    scanService      *scan.Service
    deployService    *deploy.Service
    emulationService *emulation.Service  
}

func NewHandlers() *Handlers {
    projectRoot := "."  // You might need to adjust this path
    return &Handlers{
        scanService:      scan.NewService(),
        deployService:    deploy.NewService(),
        emulationService: emulation.NewService(projectRoot),  // Add this initialization
    }
}

func (h *Handlers) StartScan(w http.ResponseWriter, r *http.Request) {
    scanID, err := h.scanService.StartScan()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
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
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(status)
}

func (h *Handlers) GetScanResult(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    scanID := vars["id"]

    result, err := h.scanService.GetResult(scanID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(result)
}