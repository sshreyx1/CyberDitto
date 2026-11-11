package http

import (
	"context"

	appiam "github.com/sshreyx1/cyberditto/services/api/internal/app/iam"
)

type ctxKey int

const authContextKey ctxKey = iota

func withAuthContext(ctx context.Context, ac *appiam.AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, ac)
}

// authFromContext returns the authenticated user/membership attached by the
// session middleware. It is only ever nil if a handler that requires auth
// is reached without going through that middleware first, which is a
// programming error, not a request the client can trigger.
func authFromContext(ctx context.Context) *appiam.AuthContext {
	ac, _ := ctx.Value(authContextKey).(*appiam.AuthContext)
	return ac
}
