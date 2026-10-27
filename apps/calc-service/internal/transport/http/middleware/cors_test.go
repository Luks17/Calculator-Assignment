package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/transport/http/middleware"
)

type corsConfig struct {
	environment string
	origins     []string
}

func (c corsConfig) GetAppEnv() string           { return c.environment }
func (c corsConfig) GetAllowedOrigins() []string { return c.origins }

func TestCORS(t *testing.T) {
	for _, tt := range []struct {
		name, environment, method, origin, preflightMethod string
		allowed                                            bool
	}{
		{"allowed origin", "production", http.MethodPost, "https://trusted.example", "", true},
		{"Scalar origin", "production", http.MethodPost, "http://localhost:8081", "", true},
		{"allowed preflight", "production", http.MethodOptions, "https://trusted.example", http.MethodPost, true},
		{"bare OPTIONS", "production", http.MethodOptions, "https://trusted.example", "", true},
		{"untrusted origin", "production", http.MethodPost, "https://untrusted.example", "", false},
		{"untrusted preflight", "production", http.MethodOptions, "https://untrusted.example", http.MethodPost, false},
		{"origin prefix attack", "production", http.MethodPost, "https://trusted.example.evil", "", false},
		{"different scheme", "production", http.MethodPost, "http://trusted.example", "", false},
		{"different port", "production", http.MethodPost, "https://trusted.example:8443", "", false},
		{"different subdomain", "production", http.MethodPost, "https://other.trusted.example", "", false},
		{"unsupported method", "production", http.MethodTrace, "https://trusted.example", "", false},
		{"unsupported preflight method", "production", http.MethodOptions, "https://trusted.example", http.MethodTrace, false},
		{"no origin", "production", http.MethodPost, "", "", false},
		{"no origin preflight", "production", http.MethodOptions, "", http.MethodPost, false},
		{"opaque origin", "production", http.MethodPost, "null", "", false},
		{"local origin bypass", "local", http.MethodPost, "http://localhost:9999", "", true},
		{"local preflight bypass", "local", http.MethodOptions, "http://localhost:9999", http.MethodPost, true},
		{"local still needs an origin", "local", http.MethodPost, "", "", false},
		{"unknown environment is not local", "locla", http.MethodPost, "http://localhost:9999", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			cfg := corsConfig{environment: tt.environment, origins: []string{"https://trusted.example", "http://localhost:8081"}}
			handler := middleware.CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusAccepted)
			}))
			recorder := httptest.NewRecorder()
			recorder.Header().Add("Vary", "Accept-Encoding")
			req := httptest.NewRequest(tt.method, "/api/v1/calculations", nil)
			req.Header.Set("Origin", tt.origin)
			if tt.preflightMethod != "" {
				req.Header.Set("Access-Control-Request-Method", tt.preflightMethod)
				req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
			}
			handler.ServeHTTP(recorder, req)
			if tt.method == http.MethodOptions {
				require.False(t, nextCalled, "OPTIONS must short-circuit")
				require.Equal(t, http.StatusNoContent, recorder.Code)
				require.Empty(t, recorder.Body.String())
			} else {
				require.True(t, nextCalled, "ordinary requests must reach the handler, including denied origins")
				require.Equal(t, http.StatusAccepted, recorder.Code)
			}

			permissions := map[string]string{
				"Access-Control-Allow-Origin":      tt.origin,
				"Access-Control-Allow-Methods":     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Expose-Headers":    "X-Request-ID",
			}
			for name, want := range permissions {
				if !tt.allowed {
					want = ""
				}
				require.Equal(t, want, recorder.Header().Get(name), name)
			}
			vary := strings.Join(recorder.Header().Values("Vary"), ", ")
			for _, value := range []string{"Accept-Encoding", "Origin"} {
				require.Contains(t, vary, value, "Vary")
			}
			if tt.method == http.MethodOptions {
				for _, value := range []string{"Access-Control-Request-Method", "Access-Control-Request-Headers"} {
					require.Contains(t, vary, value, "preflight Vary")
				}
			}
		})
	}
}

func TestCORSDefaultsDenyAndDoesNotExpandWildcards(t *testing.T) {
	for _, tt := range []struct {
		name    string
		origins []string
	}{
		{"no origins", nil},
		{"wildcard", []string{"*"}},
		{"subdomain wildcard", []string{"https://*.example.com"}},
		{"trailing slash", []string{"https://trusted.example/"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.CORS(corsConfig{environment: "production", origins: tt.origins})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set("Origin", "https://trusted.example")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			require.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestCORSSnapshotsConfiguration(t *testing.T) {
	origins := []string{"https://trusted.example"}
	handler := middleware.CORS(corsConfig{environment: "production", origins: origins})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	origins[0] = "https://untrusted.example"
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Origin", "https://trusted.example")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	require.Equal(t, "https://trusted.example", recorder.Header().Get("Access-Control-Allow-Origin"), "middleware configuration changed after construction")
}
