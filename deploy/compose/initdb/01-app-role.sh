#!/usr/bin/env bash
# Runs once, only on first initialization of an empty Postgres data volume
# (the postgres image's standard /docker-entrypoint-initdb.d behavior).
# Creates the non-superuser application role with LOGIN here, in dev-only
# scripted config, rather than baking a password into a goose migration
# (docs/MASTER_PLAN.md §6.2: the app connects as a non-superuser role).
# Migration 0001_init.sql creates this role NOLOGIN if it doesn't already
# exist yet; running this first means that CREATE is a no-op and the
# LOGIN/password set here is preserved.
set -euo pipefail

: "${CD_APP_DB_PASSWORD:?CD_APP_DB_PASSWORD must be set}"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-SQL
    DO \$\$
    BEGIN
      IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'cyberditto_app') THEN
        CREATE ROLE cyberditto_app LOGIN PASSWORD '${CD_APP_DB_PASSWORD}';
      ELSE
        ALTER ROLE cyberditto_app LOGIN PASSWORD '${CD_APP_DB_PASSWORD}';
      END IF;
    END
    \$\$;
SQL
