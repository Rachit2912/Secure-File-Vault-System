package models

import "github.com/golang-jwt/jwt/v5"

// Claims defines the structure of JWT payload used
// Contains user identity and role information.
type Claims struct {
	UserID   int    `json:"userID"`   // unique user ID
	Username string `json:"username"` // unique username of the user
	Role     string `json:"role"`     // user role (admin/user)
	jwt.RegisteredClaims              // standard JWT fields
}
