package httpadapter

import (
	"net/http"
	"strings"
)

// CORSConfig holds CORS middleware configuration.
type CORSConfig struct {
	// AllowedOrigins is the list of origins allowed to make requests.
	// Use ["*"] to allow all origins (development only).
	AllowedOrigins []string
}

// CORSMiddleware returns a middleware that handles CORS headers and preflight requests.
func CORSMiddleware(cfg CORSConfig) func(http.Handler) http.Handler {
	allowedSet := make(map[string]struct{}, len(cfg.AllowedOrigins))
	allowAll := false
	for _, origin := range cfg.AllowedOrigins {
		if origin == "*" {
			allowAll = true
			break
		}
		allowedSet[strings.ToLower(origin)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Determine if origin is allowed.
			var allowedOrigin string
			if allowAll {
				allowedOrigin = "*"
			} else if origin != "" {
				if _, ok := allowedSet[strings.ToLower(origin)]; ok {
					allowedOrigin = origin
				}
			}

			// Set CORS headers if origin is allowed.
			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
				w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
				w.Header().Set("Access-Control-Max-Age", "86400")

				// Allow credentials only for specific origins, not wildcard.
				if allowedOrigin != "*" {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			// Handle preflight requests.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
