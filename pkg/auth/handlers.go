package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
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
// @Security BasicAuth
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
			Username:    user.Username,
			CreatedAt:   user.CreatedAt,
			LastLoginAt: user.LastLoginAt,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
