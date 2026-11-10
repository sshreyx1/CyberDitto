// Command api runs the CyberDitto control-plane HTTP server.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	transporthttp "github.com/sshreyx1/cyberditto/services/api/internal/transport/http"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	addr := os.Getenv("CD_API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           transporthttp.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("starting cyberditto control plane", "addr", addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
}
