package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// SetupRoutes sets up the authentication routes
func SetupRoutes(r chi.Router, authService *AuthService) {
	// Public routes
	r.Post("/login", LoginHandler(authService))

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(authService))

		r.Post("/logout", LogoutHandler(authService))
		r.Get("/me", GetCurrentUserHandler(authService))
		r.Post("/change-password", ChangePasswordHandler(authService))

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(AdminOnlyMiddleware)

			r.Post("/users", CreateUserHandler(authService))
			r.Get("/users", ListUsersHandler(authService))
			r.Put("/users/{id}", UpdateUserHandler(authService))
			r.Delete("/users/{id}", DeleteUserHandler(authService))
		})
	})
}

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
