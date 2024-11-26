// internal/api/handlers/emulation.go

package handlers

import (
    "cyberditto-backend/internal/core/emulation"
    "encoding/json"
    "github.com/gorilla/mux"
    "net/http"
    "log"
	"fmt"
)

type emulationRequest struct {
    DeploymentID string              `json:"deployment_id"`
    Mode         emulation.EmulationMode `json:"mode"`
    Target       string              `json:"target"`
    TestNumbers  string              `json:"test_numbers,omitempty"`
}

type emulationResponse struct {
    ExecutionID string `json:"execution_id"`
}

type errorResponse struct {
    Error string `json:"error"`
}

// StartEmulation handles the request to start a new emulation
func (h *Handlers) StartEmulation(w http.ResponseWriter, r *http.Request) {
    // Parse request body
    var req emulationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("Error decoding request body: %v", err)
        h.sendError(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate the mode
    if !isValidMode(req.Mode) {
        h.sendError(w, "Invalid emulation mode", http.StatusBadRequest)
        return
    }

    // Start the emulation
    executionID, err := h.emulationService.StartEmulation(
        req.DeploymentID,
        req.Mode,
        req.Target,
        req.TestNumbers,
    )
    if err != nil {
        log.Printf("Error starting emulation: %v", err)
        h.sendError(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Send success response
    h.sendJSON(w, emulationResponse{ExecutionID: executionID})
}

// GetEmulationStatus handles requests to check the status of an emulation
func (h *Handlers) GetEmulationStatus(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    executionID := vars["id"]

    if executionID == "" {
        h.sendError(w, "Execution ID is required", http.StatusBadRequest)
        return
    }

    status, err := h.emulationService.GetStatus(executionID)
    if err != nil {
        log.Printf("Error getting emulation status for %s: %v", executionID, err)
        h.sendError(w, err.Error(), http.StatusNotFound)
        return
    }

    h.sendJSON(w, status)
}

// GetEmulationResult handles requests to get the results of an emulation
func (h *Handlers) GetEmulationResult(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    executionID := vars["id"]

    if executionID == "" {
        h.sendError(w, "Execution ID is required", http.StatusBadRequest)
        return
    }

    result, err := h.emulationService.GetResult(executionID)
    if err != nil {
        log.Printf("Error getting emulation result for %s: %v", executionID, err)
        h.sendError(w, err.Error(), http.StatusNotFound)
        return
    }

    h.sendJSON(w, result)
}

// CancelEmulation handles requests to cancel an ongoing emulation
func (h *Handlers) CancelEmulation(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    executionID := vars["id"]

    if executionID == "" {
        h.sendError(w, "Execution ID is required", http.StatusBadRequest)
        return
    }

    if err := h.emulationService.CancelEmulation(executionID); err != nil {
        log.Printf("Error canceling emulation %s: %v", executionID, err)
        h.sendError(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

// GetAllEmulations handles requests to list all emulations
func (h *Handlers) GetAllEmulations(w http.ResponseWriter, r *http.Request) {
    executions, err := h.emulationService.GetAllEmulations()
    if err != nil {
        log.Printf("Error getting all emulations: %v", err)
        h.sendError(w, err.Error(), http.StatusInternalServerError)
        return
    }

    h.sendJSON(w, executions)
}

// Helper function to send JSON responses
func (h *Handlers) sendJSON(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(data); err != nil {
        log.Printf("Error encoding response: %v", err)
        h.sendError(w, "Internal server error", http.StatusInternalServerError)
    }
}

// Helper function to send error responses
func (h *Handlers) sendError(w http.ResponseWriter, message string, status int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(errorResponse{Error: message})
}

// Helper function to validate emulation mode
func isValidMode(mode emulation.EmulationMode) bool {
    switch mode {
    case emulation.ModeTactic, emulation.ModeTechnique, emulation.ModeCustom:
        return true
    default:
        return false
    }
}

// Helper function to validate execution ID
func (h *Handlers) validateExecutionID(executionID string) error {
    if executionID == "" {
        return fmt.Errorf("execution ID is required")
    }
    return nil
}

// UpdateHandlers adds the emulation service to an existing Handlers instance
func (h *Handlers) UpdateHandlers(emulationService *emulation.Service) {
    h.emulationService = emulationService
}