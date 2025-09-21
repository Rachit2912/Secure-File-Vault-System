package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"backend/internal/config"
	"backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// custom context keys for avoid collisioins :
type contextKey string

// exported keys for handlers :
const ContextUserIDKey = contextKey("userID")
const ContextUserRoleKey = contextKey("role")


// fn. for validating JWT & adding user info to context :
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// secret JWT key : 
		var jwtKey = []byte(config.AppConfig.JWTKey) 
		if len(jwtKey) == 0 {log.Fatal("JWT_KEY not found, plz set it in .env file")}

		// reading token from cookie:
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "Missing token cookie", http.StatusUnauthorized)
			return
		}
		tokenStr := cookie.Value

		// parsing & validating JWT : 
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


		//  storing userID  & role in context for handlers
		ctx := context.WithValue(r.Context(), ContextUserIDKey, claims.UserID)
		ctx = context.WithValue(ctx,ContextUserRoleKey,claims.Role)
		// calling next handler :
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
