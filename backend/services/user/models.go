package user

import "time"

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2"`
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateUserRequest represents user profile update request
type UpdateUserRequest struct {
	Name string `json:"name" binding:"required,min=2"`
}

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID int
	Email  string
	ExpiresAt int64
}
