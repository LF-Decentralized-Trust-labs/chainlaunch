package auth

import "time"

// User represents an authenticated user in the service layer
type User struct {
	ID          int64
	Username    string
	Password    string
	CreatedAt   time.Time
	LastLoginAt time.Time
}

// LoginRequest represents the login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Session represents an authenticated session in the service layer
type Session struct {
	ID        string
	Token     string
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
}
