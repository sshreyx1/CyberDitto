package postgres

import "embed"

// MigrationsFS embeds the goose migration files into the binary so
// deployments only need the compiled binary, not a separate SQL directory.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
