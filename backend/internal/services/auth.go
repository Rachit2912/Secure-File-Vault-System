package services

import (
	"backend/internal/config"
	"backend/internal/models"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwt generator :
func GenerateJWT(userID int, username string, role string) (string, error) {
	jwtKeyStr := config.AppConfig.JWTKey
	if jwtKeyStr == "" {
		return "", fmt.Errorf("JWT_KEY environment variable is not configured")
	}
	jwtKey := []byte(jwtKeyStr)

	expiration := time.Now().Add(5 * time.Minute)
	claims := &models.Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
