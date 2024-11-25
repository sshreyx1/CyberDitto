package handlers

import (
    "cyberditto-backend/internal/core/scan"
    "cyberditto-backend/internal/core/deploy"
    "encoding/json"
    "github.com/gorilla/mux"
    "net/http"
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