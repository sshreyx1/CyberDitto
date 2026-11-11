// Command migrate applies goose migrations against CD_DATABASE_URL.
// Usage: migrate [up|down|status] (default: up).
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver, required by goose
	"github.com/pressly/goose/v3"

	"github.com/sshreyx1/cyberditto/services/api/internal/adapters/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dsn := os.Getenv("CD_DATABASE_URL")
	if dsn == "" {
		slog.Error("CD_DATABASE_URL is required")
		os.Exit(1)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	goose.SetBaseFS(postgres.MigrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		slog.Error("set dialect", "error", err)
		os.Exit(1)
	}

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "up":
		err = goose.Up(db, "migrations")
	case "down":
		err = goose.Down(db, "migrations")
	case "status":
		err = goose.Status(db, "migrations")
	case "grant-app-login":
		err = grantAppLogin(db)
	default:
		slog.Error("unknown migrate command", "cmd", cmd, "want", "up|down|status|grant-app-login")
		os.Exit(1)
	}
	if err != nil {
		slog.Error("migrate", "cmd", cmd, "error", err)
		os.Exit(1)
	}
	slog.Info("migrate ok", "cmd", cmd)
}

// grantAppLogin gives the cyberditto_app role (created NOLOGIN by
// 0001_init.sql) a password and LOGIN rights, sourced from CD_APP_DB_PASSWORD
// — never from a migration file (docs/MASTER_PLAN.md §6.2: no secrets in
// committed SQL). The dev compose stack does the equivalent via
// deploy/compose/initdb/01-app-role.sh (which runs before any migration, on
// first container init); this subcommand is the same operation for CI and
// for any environment that provisions the role after running migrations.
func grantAppLogin(db *sql.DB) error {
	password := os.Getenv("CD_APP_DB_PASSWORD")
	if password == "" {
		return fmt.Errorf("CD_APP_DB_PASSWORD is required for grant-app-login")
	}
	// ALTER ROLE does not accept a bind parameter for PASSWORD; the value
	// comes only from our own env var, never user input, so building this
	// one statement with fmt.Sprintf (quoted via a literal, not
	// interpolated into identifiers) is the standard pattern — pgx/lib/pq
	// have no parameterized ALTER ROLE support.
	_, err := db.ExecContext(context.Background(),
		fmt.Sprintf(`ALTER ROLE cyberditto_app LOGIN PASSWORD %s`, quoteLiteral(password)))
	return err
}

// quoteLiteral escapes a string for use as a single-quoted SQL literal by
// doubling embedded single quotes (the standard SQL escaping, valid with
// Postgres's default standard_conforming_strings=on) — this is the one
// place in the codebase that must build SQL text rather than use a bind
// parameter, since ALTER ROLE ... PASSWORD has no parameterized form.
func quoteLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
