package handlers


import (
    "encoding/json"
    "github.com/gorilla/mux"
    "net/http"
    "log"
    "fmt"
    "strings"
)


type deployRequest struct {
    ScanID string `json:"scan_id"`
}


func (h *Handlers) StartDeploy(w http.ResponseWriter, r *http.Request) {
    var req deployRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }


    deployID, err := h.deployService.StartDeployment(req.ScanID)
    if err != nil {
        log.Printf("Error starting deployment: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }


    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "deployment_id": deployID,
    })
}


func (h *Handlers) GetDeployStatus(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    deployID := vars["id"]

    status, err := h.deployService.GetStatus(deployID)
    if err != nil {
        log.Printf("Error getting deployment status: %v", err)
        if strings.Contains(err.Error(), "not found") {
            http.Error(w, err.Error(), http.StatusNotFound)
            return
        }
        // Send a more detailed error message
        errResponse := map[string]string{
            "error": fmt.Sprintf("Failed to get deployment status: %v", err),
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(errResponse)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(status); err != nil {
        log.Printf("Error encoding deployment status: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
}


func (h *Handlers) GetAllDeployments(w http.ResponseWriter, r *http.Request) {
    deployments, err := h.deployService.GetAllDeployments()
    if err != nil {
        log.Printf("Error getting all deployments: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }


    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(deployments); err != nil {
        log.Printf("Error encoding deployments: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
}


func (h *Handlers) StartInstance(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    deployID := vars["id"]


    if err := h.deployService.StartInstance(deployID); err != nil {
        log.Printf("Error starting instance: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }


    w.WriteHeader(http.StatusOK)
}


func (h *Handlers) StopInstance(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    deployID := vars["id"]


    if err := h.deployService.StopInstance(deployID); err != nil {
        log.Printf("Error stopping instance: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }


    w.WriteHeader(http.StatusOK)
}


func (h *Handlers) DeleteInstance(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    deployID := vars["id"]


    if err := h.deployService.DeleteInstance(deployID); err != nil {
        log.Printf("Error deleting instance: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }


    w.WriteHeader(http.StatusOK)
}


func (h *Handlers) CancelDeploy(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    deployID := vars["id"]


    if err := h.deployService.CancelDeployment(deployID); err != nil {
        log.Printf("Error canceling deployment: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }


    w.WriteHeader(http.StatusOK)
}