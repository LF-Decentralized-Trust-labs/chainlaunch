package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

// @Summary Login user
// @Description Authenticates a user and returns a session cookie
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Login successful"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Invalid credentials"
// @Failure 405 {string} string "Method not allowed"
// @Router /auth/login [post]
// @BasePath /api/v1
func LoginHandler(authService *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		session, err := authService.Login(ctx, req.Username, req.Password)
		if err != nil {
			log.Printf("Error logging in: %v", err)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Sign the session ID
		signature := signSessionID(session.ID)
		signedSessionID := session.ID + "." + signature

		// Set session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     SessionCookieName,
			Value:    signedSessionID,
			Path:     "/",
			Expires:  session.ExpiresAt,
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteStrictMode,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LoginResponse{
			Message: "Login successful",
		})
	}
}

// @Summary Logout user
// @Description Invalidates the current session and clears the session cookie
// @Tags Authentication
// @Produce json
// @Security CookieAuth
// @Success 200 {object} LogoutResponse "Logout successful"
// @Failure 405 {string} string "Method not allowed"
// @Router /auth/logout [post]
// @BasePath /api/v1
func LogoutHandler(authService *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cookie, err := r.Cookie(SessionCookieName)
		if err == nil {
			authService.Logout(r.Context(), cookie.Value)
		}

		// Clear session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     SessionCookieName,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteStrictMode,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LogoutResponse{
			Message: "Logout successful",
		})
	}
}

// @Summary Get current user
// @Description Returns information about the currently authenticated user
// @Tags Authentication
// @Produce json
// @Security CookieAuth
// @Success 200 {object} UserResponse "User information"
// @Failure 401 {string} string "Unauthorized"
// @Router /auth/me [get]
// @BasePath /api/v1
func GetCurrentUserHandler(authService *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := SessionFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := authService.GetUserByUsername(session.Username)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		response := UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Role:        user.Role,
			CreatedAt:   user.CreatedAt,
			LastLoginAt: user.LastLoginAt,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// @Summary Create new user
// @Description Creates a new user with specified role (admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param user body CreateUserRequest true "User to create"
// @Success 201 {object} UserResponse "User created"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden - Requires admin role"
// @Router /auth/users [post]
// @BasePath /api/v1
func CreateUserHandler(authService *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		session, ok := SessionFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if session.Role != RoleAdmin {
			http.Error(w, "Forbidden - Requires admin role", http.StatusForbidden)
			return
		}

		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		user, err := authService.CreateUser(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Role:        user.Role,
			CreatedAt:   user.CreatedAt,
			LastLoginAt: user.LastLoginAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// @Summary List users
// @Description Returns a list of all users (admin only)
// @Tags Users
// @Produce json
// @Security CookieAuth
// @Success 200 {array} UserResponse "List of users"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden - Requires admin role"
// @Router /auth/users [get]
// @BasePath /api/v1
func ListUsersHandler(authService *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		session, ok := SessionFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if session.Role != RoleAdmin {
			http.Error(w, "Forbidden - Requires admin role", http.StatusForbidden)
			return
		}

		users, err := authService.ListUsers(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		responses := make([]UserResponse, len(users))
		for i, user := range users {
			responses[i] = UserResponse{
				ID:          user.ID,
				Username:    user.Username,
				Role:        user.Role,
				CreatedAt:   user.CreatedAt,
				LastLoginAt: user.LastLoginAt,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responses)
	}
}

// @Summary Update user
// @Description Updates an existing user (admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path int true "User ID"
// @Param user body UpdateUserRequest true "User updates"
// @Success 200 {object} UserResponse "User updated"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden - Requires admin role"
// @Failure 404 {string} string "User not found"
// @Router /auth/users/{id} [put]
// @BasePath /api/v1
func UpdateUserHandler(authService *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		session, ok := SessionFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if session.Role != RoleAdmin {
			http.Error(w, "Forbidden - Requires admin role", http.StatusForbidden)
			return
		}

		userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		var req UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		user, err := authService.UpdateUser(r.Context(), userID, &req)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			Role:        user.Role,
			CreatedAt:   user.CreatedAt,
			LastLoginAt: user.LastLoginAt,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// @Summary Delete user
// @Description Deletes a user (admin only)
// @Tags Users
// @Security CookieAuth
// @Param id path int true "User ID"
// @Success 204 "User deleted"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Forbidden - Requires admin role"
// @Failure 404 {string} string "User not found"
// @Router /auth/users/{id} [delete]
// @BasePath /api/v1
func DeleteUserHandler(authService *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		session, ok := SessionFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if session.Role != RoleAdmin {
			http.Error(w, "Forbidden - Requires admin role", http.StatusForbidden)
			return
		}

		userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Prevent self-deletion
		if userID == session.UserID {
			http.Error(w, "Cannot delete your own account", http.StatusForbidden)
			return
		}

		if err := authService.DeleteUser(r.Context(), userID); err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "User not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// ChangePasswordRequest represents the request to change a user's password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// @Summary Change own password
// @Description Allows a user to change their own password
// @Tags Authentication
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param request body ChangePasswordRequest true "Password change request"
// @Success 200 {object} map[string]string "Password changed successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {string} string "Invalid current password"
// @Router /auth/change-password [post]
// @BasePath /api/v1
func ChangePasswordHandler(authService *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get current user from context
		session, ok := SessionFromContext(r.Context())
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Get user from database to verify current password
		user, err := authService.GetUserByUsername(session.Username)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Verify current password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
			http.Error(w, "Invalid current password", http.StatusForbidden)
			return
		}

		// Update password
		if err := authService.UpdateUserPassword(r.Context(), session.Username, req.NewPassword); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Password changed successfully",
		})
	}
}
