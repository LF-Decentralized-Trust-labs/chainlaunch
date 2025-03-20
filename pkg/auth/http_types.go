package auth

import "time"

// LoginResponse represents the HTTP response for successful login
// @Description Login response
type LoginResponse struct {
	// Success message
	// @Example "Login successful"
	Message string `json:"message"`
}

// LogoutResponse represents the HTTP response for successful logout
// @Description Logout response
type LogoutResponse struct {
	// Success message
	// @Example "Logout successful"
	Message string `json:"message"`
}

// UserResponse represents the HTTP response for user information
// @Description User information response
type UserResponse struct {
	// Username of the user
	// @Example "admin"
	Username string `json:"username"`
	// Time when the user was created
	// @Example "2024-01-01T00:00:00Z"
	CreatedAt time.Time `json:"created_at"`
	// Last time the user logged in
	// @Example "2024-01-01T12:34:56Z"
	LastLoginAt time.Time `json:"last_login_at"`

}
