package auth

import (
	"net/http"
)

// AdminOnlyMiddleware ensures the user has admin role
func AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if user.Role != RoleAdmin {
			http.Error(w, "Forbidden - Requires admin role", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
