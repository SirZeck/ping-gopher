package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/SirZeck/ping-gopher/internal/auth"
)

type contextKey string

const (
	UserIDContextKey contextKey = "user_id"
	UserEmailContextKey contextKey = "user_email"
)

// CORSMiddleware enables Cross-Origin Resource Sharing headers for Web Dashboard requests.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Limit max incoming payload size to 1 MB (1048576 bytes) to prevent memory exhaustion
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates JWT Bearer tokens and attaches UserID to request context.
func AuthMiddleware(jwtSecret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			JSONError(w, http.StatusUnauthorized, "Missing Authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			JSONError(w, http.StatusUnauthorized, "Invalid Authorization header format. Expected 'Bearer <token>'")
			return
		}

		tokenString := parts[1]
		claims, err := auth.ValidateJWTToken(tokenString, jwtSecret)
		if err != nil {
			JSONError(w, http.StatusUnauthorized, fmt.Sprintf("Unauthorized: %v", err))
			return
		}

		// Attach user identity to request context
		ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailContextKey, claims.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserIDFromContext retrieves the authenticated UserID from request context.
func GetUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	val := r.Context().Value(UserIDContextKey)
	if val == nil {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}

	userIDStr, ok := val.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid user ID context type")
	}

	return uuid.Parse(userIDStr)
}
