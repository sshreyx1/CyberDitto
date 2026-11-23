-- name: AdvisoryXactLock :exec
SELECT pg_advisory_xact_lock(hashtext(sqlc.arg(lock_key)::text));
