package middleware_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/stretchr/testify/require"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/transport/http/middleware"
)

func readLogs(t *testing.T, logs *bytes.Buffer) []map[string]any {
	t.Helper()
	decoder := json.NewDecoder(logs)
	var entries []map[string]any
	for {
		var entry map[string]any
		err := decoder.Decode(&entry)
		if err == io.EOF {
			return entries
		}
		require.NoError(t, err)
		entries = append(entries, entry)
	}
}

func TestLogging(t *testing.T) {
	for _, tt := range []struct {
		name, body, level string
		status            int
	}{
		{"implicit empty response", "", "info", 0},
		{"implicit body response", "result", "info", 0},
		{"success", "", "info", 204},
		{"client error", "", "warn", 422},
		{"server error", "", "error", 500},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := zerolog.New(&logs).With().Str("service", "calc-service").Logger()
			handler := middleware.Logging(&logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id, ok := hlog.IDFromRequest(r)
				require.True(t, ok, "request ID was not attached to context")
				require.Equal(t, id.String(), w.Header().Get("X-Request-ID"))
				zerolog.Ctx(r.Context()).UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Str("calculation", "test")
				})
				if tt.status != 0 {
					w.WriteHeader(tt.status)
				}
				if tt.body != "" {
					_, err := io.WriteString(w, tt.body)
					require.NoError(t, err)
				}
			}))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculations", nil)
			req.RemoteAddr = "192.0.2.1:12345"
			req.Header.Set("User-Agent", "calculator-test")
			req.Header.Set("Origin", "http://localhost:5173")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			logger.Info().Msg("root log")

			entries := readLogs(t, &logs)
			require.Len(t, entries, 2, "want request log and root log")
			status := tt.status
			if status == 0 {
				status = http.StatusOK
			}
			want := map[string]any{
				"level": tt.level, "message": "incoming request", "service": "calc-service",
				"method": http.MethodPost, "path": "/api/v1/calculations", "ip": req.RemoteAddr,
				"user_agent": "calculator-test", "origin": "http://localhost:5173",
				"status_code": float64(status), "bytes_written": float64(len(tt.body)),
				"req_id": recorder.Header().Get("X-Request-ID"), "calculation": "test",
			}
			for key, value := range want {
				require.Equal(t, value, entries[0][key], "request log %s", key)
			}
			require.IsType(t, float64(0), entries[0]["elapsed_ms"], "request log is missing elapsed time")
			require.Equal(t, status, recorder.Code)
			require.Equal(t, tt.body, recorder.Body.String())
			for _, key := range []string{"req_id", "method", "calculation"} {
				require.NotContains(t, entries[1], key, "request field leaked into root logger")
			}
		})
	}
}

func TestLoggingIsolatesConcurrentRequestContexts(t *testing.T) {
	var logs bytes.Buffer
	logger := zerolog.New(zerolog.SyncWriter(&logs))
	handler := middleware.Logging(&logger)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		zerolog.Ctx(r.Context()).UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("marker", r.URL.Path)
		})
		zerolog.Ctx(r.Context()).Info().Msg("handler log")
	}))
	const requests = 32
	var wg sync.WaitGroup
	for i := range requests {
		wg.Go(func() {
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, fmt.Sprintf("/requests/%d", i), nil))
		})
	}
	wg.Wait()
	entries := readLogs(t, &logs)
	require.Len(t, entries, 2*requests)
	counts := make(map[string]int)
	for _, entry := range entries {
		id, ok := entry["req_id"].(string)
		require.True(t, ok, "request ID must be a string")
		require.NotEmpty(t, id)
		require.Equal(t, entry["path"], entry["marker"], "request fields are cross-contaminated")
		counts[id]++
	}
	require.Len(t, counts, requests, "unique request IDs")
	for id, count := range counts {
		require.Equal(t, 2, count, "request %s must have handler log and access log", id)
	}
}

func TestLoggingPreservesResponseController(t *testing.T) {
	var logs bytes.Buffer
	logger := zerolog.New(&logs)
	handler := middleware.Logging(&logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		require.NoError(t, http.NewResponseController(w).Flush())
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	require.True(t, recorder.Flushed, "logging wrapper did not forward Flush")
	entries := readLogs(t, &logs)
	require.Len(t, entries, 1)
	require.Equal(t, float64(http.StatusOK), entries[0]["status_code"], "flushed response must be logged as 200")
}
