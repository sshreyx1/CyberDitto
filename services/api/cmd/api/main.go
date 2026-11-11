// Command api runs the CyberDitto control-plane HTTP server.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/argon2"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/postgres"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/token"
	appiam "github.com/sshreyx1/cyberditto/services/api/internal/app/iam"
	transporthttp "github.com/sshreyx1/cyberditto/services/api/internal/transport/http"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	addr := getenv("CD_API_ADDR", ":8080")
	databaseURL := os.Getenv("CD_APP_DATABASE_URL")
	if databaseURL == "" {
		slog.Error("CD_APP_DATABASE_URL is required (the cyberditto_app role; see .env.example)")
		os.Exit(1)
	}
	secure := getenv("CD_ENV", "dev") != "dev"
	allowedOrigins := strings.Split(getenv("CD_ALLOWED_ORIGINS", "http://localhost:3000"), ",")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		slog.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	iamSvc := appiam.NewService(
		postgres.NewIAMRepo(pool),
		postgres.NewAuditRepo(pool),
		argon2.New(),
		token.New(),
	)

	router := transporthttp.NewRouter(transporthttp.Deps{
		IAM:            iamSvc,
		Secure:         secure,
		AllowedOrigins: allowedOrigins,
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("starting cyberditto control plane", "addr", addr, "secure_cookies", secure)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
