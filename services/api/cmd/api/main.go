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
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/fake"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/postgres"
	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/token"
	appassets "github.com/sshreyx1/cyberditto/services/api/internal/app/assets"
	appiam "github.com/sshreyx1/cyberditto/services/api/internal/app/iam"
	apptwins "github.com/sshreyx1/cyberditto/services/api/internal/app/twins"
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

	auditRepo := postgres.NewAuditRepo(pool)
	iamSvc := appiam.NewService(postgres.NewIAMRepo(pool), auditRepo, argon2.New(), token.New())

	assetRepo := postgres.NewAssetRepo(pool)
	assetsSvc := appassets.NewService(assetRepo, auditRepo)

	// CD_LAB_PROVIDER defaults to "fake" (.env.example): the real Hyper-V
	// adapter is owner-blocked (docs/OWNER_TODO.md items 6-7). Twins and
	// their events persist in Postgres regardless of provider.
	labProvider := getenv("CD_LAB_PROVIDER", "fake")
	if labProvider != "fake" {
		slog.Error("unsupported CD_LAB_PROVIDER (only \"fake\" exists so far)", "provider", labProvider)
		os.Exit(1)
	}
	twinsSvc := apptwins.NewService(postgres.NewTwinRepo(pool), assetRepo, fake.LabProvider{}, labProvider, auditRepo)

	router := transporthttp.NewRouter(transporthttp.Deps{
		IAM:            iamSvc,
		Assets:         assetsSvc,
		Twins:          twinsSvc,
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
