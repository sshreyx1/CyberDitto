package api

import (
	"cyberditto-backend/internal/api/handlers"
	"cyberditto-backend/internal/api/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"net/http"
)

func SetupRouter() http.Handler {
	r := mux.NewRouter()
	h := handlers.NewHandlers()

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

	// Route handlers
	api.HandleFunc("/scan", h.StartScan).Methods("POST")
	api.HandleFunc("/scan/{id}/status", h.GetScanStatus).Methods("GET")
	api.HandleFunc("/scan/{id}/result", h.GetScanResult).Methods("GET")
	api.HandleFunc("/deploy", h.StartDeploy).Methods("POST")
	api.HandleFunc("/deploy/{id}/status", h.GetDeployStatus).Methods("GET")
	api.HandleFunc("/deploy/{id}/start", h.StartInstance).Methods("POST")
	api.HandleFunc("/deploy/{id}/stop", h.StopInstance).Methods("POST")
	api.HandleFunc("/deploy/{id}", h.DeleteInstance).Methods("DELETE")
	api.HandleFunc("/deploy/{id}/cancel", h.CancelDeploy).Methods("POST")
	api.HandleFunc("/deployments", h.GetAllDeployments).Methods("GET")

	return corsMiddleware.Handler(r)
}