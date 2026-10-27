package middleware

import (
	"net/http"
	"slices"
	"strings"
)

// CORSConfig is the configuration consumed by CORS, independent of its source.
type CORSConfig interface {
	GetAppEnv() string
	GetAllowedOrigins() []string
}

var (
	corsMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions,
	}
	corsMethodsString = strings.Join(corsMethods, ", ")
)

// CORS permits any origin in explicitly local environments and otherwise uses
// exact origin and method allowlists. Denied origins receive no CORS permission
// headers; this is a browser access policy, not request authentication.
func CORS(config CORSConfig) func(http.Handler) http.Handler {
	isLocal := config.GetAppEnv() == "local"
	allowedOrigins := slices.Clone(config.GetAllowedOrigins())

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			method := r.Method
			if r.Method == http.MethodOptions {
				if requested := r.Header.Get("Access-Control-Request-Method"); requested != "" {
					method = requested
				}
			}

			w.Header().Add("Vary", "Origin")
			if origin != "" && (isLocal || (slices.Contains(allowedOrigins, origin) && slices.Contains(corsMethods, method))) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", corsMethodsString)
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
			}

			if r.Method == http.MethodOptions {
				w.Header().Add("Vary", "Access-Control-Request-Method")
				w.Header().Add("Vary", "Access-Control-Request-Headers")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
