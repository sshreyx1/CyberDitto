// Package http wires HTTP transport adapters to application use cases.
// It must not contain business logic (see docs/MASTER_PLAN.md §5.3).
package http

import (
	"encoding/json"
	"net/http"

	"github.com/sshreyx1/cyberditto/services/api/internal/domain"
)

// NewRouter builds the control-plane's HTTP routes.
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	return mux
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(domain.NewHealth())
}
