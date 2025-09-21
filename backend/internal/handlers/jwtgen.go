package handlers

import (
	"backend/models"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwt generator :
func GenerateJWT(userID int, username string) (string, error) {

	// JWT secret key : 
	var jwtKey = []byte(os.Getenv("JWT_KEY")) 
	if len(jwtKey) == 0 {log.Fatal("JWT_KEY not found, plz set it in .env file")}

	expiration := time.Now().Add(5 * time.Minute)
	claims := &models.Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
