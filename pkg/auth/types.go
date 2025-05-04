package auth

import "time"

// Role represents user roles in the system
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleViewer  Role = "viewer"
)

// User represents an authenticated user in the service layer
type User struct {
	ID          int64
	Username    string
	Password    string
	Role        Role
	CreatedAt   time.Time
	LastLoginAt time.Time
}

// LoginRequest represents the login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the HTTP response for successful login
type LoginResponse struct {
	Message string `json:"message"`
}

// LogoutResponse represents the HTTP response for successful logout
type LogoutResponse struct {
	Message string `json:"message"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	Role     Role   `json:"role" validate:"required,oneof=admin manager viewer"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Username string `json:"username,omitempty"`
	Role     Role   `json:"role,omitempty" validate:"omitempty,oneof=admin manager viewer"`
}

// UserResponse represents the HTTP response for user information
type UserResponse struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Role        Role      `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	LastLoginAt time.Time `json:"last_login_at"`
}

// Session represents an authenticated session in the service layer
type Session struct {
	ID        string
	Token     string
	Username  string
	UserID    int64
	Role      Role
	CreatedAt time.Time
	ExpiresAt time.Time
}
