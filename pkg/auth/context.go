package auth

import (
	"context"
)

type contextKey string

const (
	userContextKey contextKey = "user"
)

// UserFromContext retrieves the user from the context
func UserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}

// ContextWithUser adds the user to the context
func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// SessionFromContext retrieves the session from the context by converting the user
func SessionFromContext(ctx context.Context) (*Session, bool) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, false
	}

	// Convert User to Session format
	session := &Session{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}

	return session, true
}
