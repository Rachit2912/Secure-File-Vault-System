package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/models"

	"github.com/golang-jwt/jwt/v5"
)

// private type for avoiding collisions
type contextKey string

// exported constant so handlers can use it
const ContextUserIDKey = contextKey("userID")

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// secret JWT key : 
		var jwtKey = []byte(os.Getenv("JWT_KEY")) 
		if len(jwtKey) == 0 {log.Fatal("JWT_KEY not found, plz set it in .env file")}

		// reading JWT from cookie
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Missing token cookie", http.StatusUnauthorized)
			return
		}
		tokenStr := cookie.Value

		// parsing JWT : 
		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}


		//  storing userID in context for handlers
		ctx := context.WithValue(r.Context(), ContextUserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
