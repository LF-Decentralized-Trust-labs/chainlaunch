package auth

import (
	"context"
)

type contextKey string

const sessionContextKey contextKey = "session"

// WithSession adds a session to the context
func WithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}

// SessionFromContext retrieves a session from the context
func SessionFromContext(ctx context.Context) (*Session, bool) {
	session, ok := ctx.Value(sessionContextKey).(*Session)
	return session, ok
}
