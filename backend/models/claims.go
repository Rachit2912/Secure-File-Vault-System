package models

import "github.com/golang-jwt/jwt/v5"

// claim data-structure for JWT info. :
type Claims struct {
	UserID   int `json:"userID"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
