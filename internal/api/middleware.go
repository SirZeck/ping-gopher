package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/SirZeck/ping-gopher/internal/auth"
)

type contextKey string

const (
	UserIDContextKey    contextKey = "user_id"
	UserEmailContextKey contextKey = "user_email"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}

var globalAuthRateLimiter = &rateLimiter{
	requests: make(map[string][]time.Time),
}

func init() {
	// Background ticker to evict stale IP entries from memory map every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			globalAuthRateLimiter.mu.Lock()
			now := time.Now()
			cutoff := now.Add(-1 * time.Minute)
			for ip, times := range globalAuthRateLimiter.requests {
				hasRecent := false
				for _, t := range times {
					if t.After(cutoff) {
						hasRecent = true
						break
					}
				}
				if !hasRecent {
					delete(globalAuthRateLimiter.requests, ip)
				}
			}
			globalAuthRateLimiter.mu.Unlock()
		}
	}()
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	clientIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(clientIP); err == nil {
		clientIP = host
	}
	return clientIP
}

// RateLimitMiddleware enforces a sliding-window rate limit (requestsPerSec) per client IP address.
func RateLimitMiddleware(requestsPerSec int, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		now := time.Now()
		cutoff := now.Add(-1 * time.Second)

		globalAuthRateLimiter.mu.Lock()
		times := globalAuthRateLimiter.requests[clientIP]

		validTimes := make([]time.Time, 0, len(times))
		for _, t := range times {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}

		if len(validTimes) >= requestsPerSec {
			globalAuthRateLimiter.requests[clientIP] = validTimes
			globalAuthRateLimiter.mu.Unlock()
			JSONError(w, http.StatusTooManyRequests, "Rate limit exceeded. Please slow down your requests.")
			return
		}

		validTimes = append(validTimes, now)
		globalAuthRateLimiter.requests[clientIP] = validTimes
		globalAuthRateLimiter.mu.Unlock()

		next.ServeHTTP(w, r)
	}
}

// CORSMiddleware enables Cross-Origin Resource Sharing headers for Web Dashboard requests.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOrigin == "" {
			allowedOrigin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
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
