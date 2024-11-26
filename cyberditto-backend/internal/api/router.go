// internal/api/router.go

package api

import (
    "cyberditto-backend/internal/api/handlers"
    "cyberditto-backend/internal/api/middleware"
    "github.com/gorilla/mux"
    "github.com/rs/cors"
    "net/http"
    "path/filepath"
    "os"
)

func SetupRouter() http.Handler {
    r := mux.NewRouter()
    
    // Get project root directory
    cwd, err := os.Getwd()
    if err != nil {
        panic("Failed to get working directory: " + err.Error())
    }
    
    // Navigate up to project root from cmd/server
    projectRoot := filepath.Join(cwd, "..", "..")
    projectRoot, err = filepath.Abs(projectRoot)
    if err != nil {
        panic("Failed to get absolute path: " + err.Error())
    }
    
    // Initialize handlers with project root
    h := handlers.NewHandlers(projectRoot)

    // API subrouter
    api := r.PathPrefix("/api").Subrouter()

    // CORS middleware
    corsMiddleware := cors.New(cors.Options{
        AllowedOrigins: []string{"http://localhost:5173"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{"Content-Type", "Authorization"},
    })

    // Apply middleware
    api.Use(middleware.Logging)
    api.Use(middleware.Auth)

    // Scan routes
    api.HandleFunc("/scan", h.StartScan).Methods("POST")
    api.HandleFunc("/scan/{id}/status", h.GetScanStatus).Methods("GET")
    api.HandleFunc("/scan/{id}/result", h.GetScanResult).Methods("GET")

    // Deploy routes
    api.HandleFunc("/deploy", h.StartDeploy).Methods("POST")
    api.HandleFunc("/deploy/{id}/status", h.GetDeployStatus).Methods("GET")
    api.HandleFunc("/deploy/{id}/start", h.StartInstance).Methods("POST")
    api.HandleFunc("/deploy/{id}/stop", h.StopInstance).Methods("POST")
    api.HandleFunc("/deploy/{id}", h.DeleteInstance).Methods("DELETE")
    api.HandleFunc("/deploy/{id}/cancel", h.CancelDeploy).Methods("POST")
    api.HandleFunc("/deployments", h.GetAllDeployments).Methods("GET")

    // Emulation routes
    api.HandleFunc("/emulation", h.StartEmulation).Methods("POST")
    api.HandleFunc("/emulation/{id}/status", h.GetEmulationStatus).Methods("GET")
    api.HandleFunc("/emulation/{id}/result", h.GetEmulationResult).Methods("GET")
    api.HandleFunc("/emulation/{id}/cancel", h.CancelEmulation).Methods("POST")
    api.HandleFunc("/emulations", h.GetAllEmulations).Methods("GET")

    return corsMiddleware.Handler(r)
}