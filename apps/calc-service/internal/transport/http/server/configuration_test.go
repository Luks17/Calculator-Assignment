package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/domain/expression"
)

type configStub string

func (c configStub) Address() string { return string(c) }

func (configStub) GetAppEnv() string { return "production" }

func (configStub) GetAllowedOrigins() []string {
	return []string{"http://localhost:5173", "http://localhost:8081"}
}

func testContext() context.Context {
	return zerolog.New(io.Discard).WithContext(context.Background())
}

func newHTTPServer(t *testing.T, ctx context.Context, cfg Config) *http.Server {
	t.Helper()
	calculator, err := expression.NewCalculator()
	require.NoError(t, err)
	srv, err := New(ctx, cfg, calculator)
	require.NoError(t, err)
	return srv
}

func TestNewServer(t *testing.T) {
	srv := newHTTPServer(t, testContext(), configStub("127.0.0.1:8080"))
	require.Equal(t, "127.0.0.1:8080", srv.Addr)
	for _, tt := range []struct {
		name    string
		timeout time.Duration
	}{
		{"read header timeout", srv.ReadHeaderTimeout},
		{"read timeout", srv.ReadTimeout},
		{"write timeout", srv.WriteTimeout},
		{"idle timeout", srv.IdleTimeout},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Positive(t, tt.timeout)
		})
	}
	recorder := httptest.NewRecorder()
	srv.Handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	require.Equal(t, http.StatusOK, recorder.Code)
}

func TestServerUsesContextualRequestLogging(t *testing.T) {
	var logs bytes.Buffer
	logger := zerolog.New(&logs).With().Str("service", "calc-service").Logger()
	srv := newHTTPServer(t, logger.WithContext(context.Background()), configStub("127.0.0.1:0"))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculations", strings.NewReader(`{"expression":"2+"}`))
	req.Header.Set("Content-Type", "application/json")
	srv.Handler.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusUnprocessableEntity, recorder.Code, "body: %s", recorder.Body)
	var entry struct {
		Service string `json:"service"`
		Level   string `json:"level"`
		ReqID   string `json:"req_id"`
		Status  int    `json:"status_code"`
	}
	require.NoError(t, json.Unmarshal(logs.Bytes(), &entry))
	require.Equal(t, "calc-service", entry.Service)
	require.Equal(t, "warn", entry.Level)
	require.Equal(t, http.StatusUnprocessableEntity, entry.Status)
	require.NotEmpty(t, entry.ReqID)
	require.Equal(t, recorder.Header().Get("X-Request-ID"), entry.ReqID)
}

func TestServerCORS(t *testing.T) {
	var logs bytes.Buffer
	logger := zerolog.New(&logs)
	srv := newHTTPServer(t, logger.WithContext(context.Background()), configStub("127.0.0.1:0"))
	for _, tt := range []struct {
		name, method, origin, preflightMethod, body string
		status                                      int
		allowed                                     bool
	}{
		{"Scalar preflight", http.MethodOptions, "http://localhost:8081", http.MethodPost, "", 204, true},
		{"web preflight", http.MethodOptions, "http://localhost:5173", http.MethodPost, "", 204, true},
		{"denied preflight", http.MethodOptions, "https://untrusted.example", http.MethodPost, "", 204, false},
		{"denied method", http.MethodOptions, "http://localhost:8081", http.MethodTrace, "", 204, false},
		{"success", http.MethodPost, "http://localhost:8081", "", `{"expression":"0.1+0.2"}`, 200, true},
		{"syntax error", http.MethodPost, "http://localhost:8081", "", `{"expression":"2+"}`, 422, true},
		{"invalid request", http.MethodPost, "http://localhost:8081", "", `{}`, 400, true},
		{"method not allowed", http.MethodGet, "http://localhost:8081", "", "", 405, true},
		{"untrusted actual request", http.MethodPost, "https://untrusted.example", "", `{"expression":"2+2"}`, 200, false},
		{"no origin", http.MethodPost, "", "", `{"expression":"2+2"}`, 200, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			logs.Reset()
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, "/api/v1/calculations", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", tt.origin)
			if tt.preflightMethod != "" {
				req.Header.Set("Access-Control-Request-Method", tt.preflightMethod)
				req.Header.Set("Access-Control-Request-Headers", "Content-Type")
			}
			srv.Handler.ServeHTTP(recorder, req)
			require.Equal(t, tt.status, recorder.Code, "body: %s", recorder.Body)
			wantOrigin := ""
			if tt.allowed {
				wantOrigin = tt.origin
			}
			require.Equal(t, wantOrigin, recorder.Header().Get("Access-Control-Allow-Origin"))
			if tt.method == http.MethodOptions {
				require.Empty(t, recorder.Body.String(), "preflight must not invoke the generated API handler or write a body")
			}
			var entry struct {
				Method string `json:"method"`
				ReqID  string `json:"req_id"`
				Status int    `json:"status_code"`
			}
			require.NoError(t, json.Unmarshal(logs.Bytes(), &entry))
			require.Equal(t, tt.method, entry.Method)
			require.Equal(t, tt.status, entry.Status)
			require.NotEmpty(t, entry.ReqID)
			require.Equal(t, recorder.Header().Get("X-Request-ID"), entry.ReqID)
		})
	}
}
