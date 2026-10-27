package app_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/app"
)

type configStub string

func (c configStub) Address() string           { return string(c) }
func (configStub) GetAppEnv() string           { return "production" }
func (configStub) GetAllowedOrigins() []string { return []string{"http://localhost:8081"} }

func TestNewWiresCalculator(t *testing.T) {
	ctx := zerolog.New(io.Discard).WithContext(context.Background())
	srv, err := app.New(ctx, configStub("127.0.0.1:8080"))
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:8080", srv.Addr)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculations", strings.NewReader(`{"expression":"0.1+0.2"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost:8081")
	recorder := httptest.NewRecorder()
	srv.Handler.ServeHTTP(recorder, req)
	var response struct {
		Result string `json:"result"`
	}
	require.Equal(t, http.StatusOK, recorder.Code, "body: %s", recorder.Body)
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, "0.3", response.Result)
	require.Equal(t, "http://localhost:8081", recorder.Header().Get("Access-Control-Allow-Origin"))
	require.NotEmpty(t, recorder.Header().Get("X-Request-ID"))
}
