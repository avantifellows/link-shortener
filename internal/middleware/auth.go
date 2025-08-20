package middleware

import (
	"net/http"
	"os"
	"strings"
)

// AuthMiddleware validates bearer token for API access
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the expected token from environment
		expectedToken := os.Getenv("AUTH_TOKEN")
		if expectedToken == "" {
			http.Error(w, "AUTH_TOKEN environment variable not configured", http.StatusInternalServerError)
			return
		}

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != expectedToken {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Token is valid, continue to next handler
		next.ServeHTTP(w, r)
	})
}

// OptionalAuthMiddleware validates bearer token but allows requests without it for public endpoints
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the expected token from environment
		expectedToken := os.Getenv("AUTH_TOKEN")
		if expectedToken == "" {
			http.Error(w, "AUTH_TOKEN environment variable not configured", http.StatusInternalServerError)
			return
		}

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// If header is present, validate it
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Bearer token required", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token != expectedToken {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
		}

		// Either no auth header (public access) or valid token
		next.ServeHTTP(w, r)
	})
}

// ConditionalAuthMiddleware requires auth only for JSON API requests, not form submissions
func ConditionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a JSON API request
		isJSONRequest := strings.Contains(r.Header.Get("Accept"), "application/json") ||
			strings.Contains(r.Header.Get("Content-Type"), "application/json")

		// If it's not a JSON request (i.e., form submission), allow it through
		if !isJSONRequest {
			next.ServeHTTP(w, r)
			return
		}

		// For JSON requests, require authentication
		expectedToken := os.Getenv("AUTH_TOKEN")
		if expectedToken == "" {
			http.Error(w, "AUTH_TOKEN environment variable not configured", http.StatusInternalServerError)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required for API requests", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Bearer token required", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != expectedToken {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}