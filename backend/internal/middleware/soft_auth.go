package middleware

import (
	"context"
	"fmt"
	"net/http"

	"backend/internal/config"
	"backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// ParseJWTFromRequest attempts to read the "token" cookie and parse JWT.
// Returns userID, role, nil on success. Returns non-nil error when no token or invalid.
func ParseJWTFromRequest(r *http.Request) (int, string, error) {
	// JWT secret
	jwtKey := []byte(config.AppConfig.JWTKey)
	if len(jwtKey) == 0 {
		return 0, "", fmt.Errorf("JWT_KEY not configured")
	}

	// reading cookie  for token: 
	cookie, err := r.Cookie("token")
	if err != nil {
		return 0, "", err 
	}
	tokenStr := cookie.Value

	// parsing into claims : 
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return 0, "", fmt.Errorf("invalid token: %v", err)
	}

	return claims.UserID, claims.Role, nil
}

// SoftAuthMiddleware will parse JWT if present and set user id & role in context.
// It will NOT reject requests without a valid token — it continues as guest.
func SoftAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uid, role, err := ParseJWTFromRequest(r); err == nil {
			// set values in context only if token parsed OK
			ctx := context.WithValue(r.Context(), ContextUserIDKey, uid)
			ctx = context.WithValue(ctx, ContextUserRoleKey, role)
			r = r.WithContext(ctx)
		}
		// continue in all cases (guest or authenticated)
		next.ServeHTTP(w, r)
	})
}
