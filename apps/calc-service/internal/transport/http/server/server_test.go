package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/domain/expression"
	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/transport/http/middleware"
	api "github.com/luks17/sezzle-assignment/apps/calc-service/internal/transport/http/server/gen"
)

func newHandler(t *testing.T) http.Handler {
	t.Helper()
	return newHTTPServer(t, testContext(), configStub("127.0.0.1:0")).Handler
}

func calculationRequest(t *testing.T, input string) string {
	t.Helper()
	body, err := json.Marshal(api.CalculationRequest{Expression: input})
	require.NoError(t, err)
	return string(body)
}

func TestCalculationsHTTP(t *testing.T) {
	handler := newHandler(t)
	for _, tt := range []struct{ input, want string }{
		{"10 + 5 * 2", "20"}, {"sqrt(81) + 2 ^ 3", "17"},
		{"0.1 + 0.2", "0.3"}, {"200 + 10%", "200.1"},
		{"1 / 3", "0.3333333333333333"}, {"2^3^2", "512"},
		{strings.Repeat("1", expression.MaxLength), strings.Repeat("1", expression.MaxLength)},
	} {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculations", strings.NewReader(calculationRequest(t, tt.input)))
			req.Header.Set("Content-Type", "application/json; charset=utf-8")
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			require.Equal(t, http.StatusOK, recorder.Code, "body: %s", recorder.Body)
			var response api.CalculationResponse
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
			require.Equal(t, tt.want, response.Result)
		})
	}
}

func TestHTTPErrors(t *testing.T) {
	handler := newHandler(t)
	for _, tt := range []struct {
		name, body, contentType string
		status                  int
		code                    api.ErrorResponseCode
	}{
		{"missing body", "", "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"malformed JSON", `{"expression":`, "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"missing field", `{}`, "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"null body", `null`, "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"null expression", `{"expression":null}`, "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"wrong type", `{"expression":123}`, "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"unknown field", `{"expression":"2+2","extra":true}`, "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"trailing JSON", `{"expression":"2+2"}{}`, "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"wrong content type", `{"expression":"2+2"}`, "text/plain", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"missing content type", `{"expression":"2+2"}`, "", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"empty expression", calculationRequest(t, ""), "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"whitespace expression", calculationRequest(t, " \t\n"), "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"long expression", calculationRequest(t, strings.Repeat("1", 513)), "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"large body", strings.Repeat(" ", MaxBodyBytes) + `{"expression":"2+2"}`, "application/json", 400, api.ErrorResponseCodeINVALIDREQUEST},
		{"syntax", calculationRequest(t, "2+"), "application/json", 422, api.ErrorResponseCodeINVALIDEXPRESSION},
		{"unknown function", calculationRequest(t, "foo(2)"), "application/json", 422, api.ErrorResponseCodeINVALIDEXPRESSION},
		{"scientific notation", calculationRequest(t, "1e3"), "application/json", 422, api.ErrorResponseCodeINVALIDEXPRESSION},
		{"excessive nesting", calculationRequest(t, strings.Repeat("(", 65)+"1"+strings.Repeat(")", 65)), "application/json", 422, api.ErrorResponseCodeINVALIDEXPRESSION},
		{"division by zero", calculationRequest(t, "2/0"), "application/json", 422, api.ErrorResponseCodeDIVISIONBYZERO},
		{"negative square root", calculationRequest(t, "sqrt(-1)"), "application/json", 422, api.ErrorResponseCodeINVALIDDOMAIN},
		{"fractional power", calculationRequest(t, "2^0.5"), "application/json", 422, api.ErrorResponseCodeINVALIDDOMAIN},
		{"large exponent", calculationRequest(t, "2^1001"), "application/json", 422, api.ErrorResponseCodeINVALIDDOMAIN},
		{"large intermediate result", calculationRequest(t, "(2^1000)^1000"), "application/json", 422, api.ErrorResponseCodeINVALIDDOMAIN},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculations", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			assertErrorResponse(t, recorder, tt.status, tt.code)
		})
	}
}

func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, status int, code api.ErrorResponseCode) api.ErrorResponse {
	t.Helper()
	require.Equal(t, status, recorder.Code, "body: %s", recorder.Body)
	require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	var response api.ErrorResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Equal(t, code, response.Code)
	require.NotEmpty(t, response.Message)
	return response
}

func TestRoutingAndHealth(t *testing.T) {
	handler := newHandler(t)
	for _, tt := range []struct {
		method, path, allow string
		status              int
	}{
		{http.MethodGet, "/healthz", "", 200},
		{http.MethodHead, "/healthz", "", 200},
		{http.MethodPost, "/healthz", "GET, HEAD", 405},
		{http.MethodGet, "/missing", "", 404},
		{http.MethodGet, "/api/v1/calculations", "POST", 405},
	} {
		t.Run(tt.method+tt.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(tt.method, tt.path, nil))
			require.Equal(t, tt.status, recorder.Code)
			if tt.status >= 400 {
				assertErrorResponse(t, recorder, tt.status, api.ErrorResponseCodeINVALIDREQUEST)
			}
			require.Equal(t, tt.allow, recorder.Header().Get("Allow"))
		})
	}
}

func TestUnexpectedErrorIsLoggedButNotExposed(t *testing.T) {
	var logs bytes.Buffer
	logger := zerolog.New(&logs).With().Str("req_id", "test-request").Logger()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/calculations", nil).WithContext(logger.WithContext(context.Background()))
	handleError(req.Context(), recorder, req, errors.New("private internal detail"))
	response := assertErrorResponse(t, recorder, 500, api.ErrorResponseCodeINTERNALERROR)
	require.NotContains(t, response.Message, "private")
	require.Contains(t, logs.String(), "private internal detail")
	require.Contains(t, logs.String(), `"req_id":"test-request"`, "error log must use the contextual logger")
}

func TestRecoveryUsesCentralErrorHandler(t *testing.T) {
	var logs bytes.Buffer
	logger := zerolog.New(&logs)
	wrapped := middleware.NewChain(middleware.Logging(&logger), middleware.Recover(handleError)).Apply(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("private panic detail")
	}))
	recorder := httptest.NewRecorder()
	wrapped.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/calculations", nil))
	response := assertErrorResponse(t, recorder, 500, api.ErrorResponseCodeINTERNALERROR)
	require.NotContains(t, response.Message, "private", "panic detail must not be exposed")
	decoder := json.NewDecoder(&logs)
	var failure, access struct {
		RequestID string `json:"req_id"`
		Level     string `json:"level"`
		Error     string `json:"error"`
		Status    int    `json:"status_code"`
	}
	require.NoError(t, decoder.Decode(&failure))
	require.NoError(t, decoder.Decode(&access))
	require.NotEmpty(t, failure.RequestID)
	require.Equal(t, failure.RequestID, access.RequestID)
	require.Equal(t, http.StatusInternalServerError, access.Status)
	require.Equal(t, "error", access.Level)
	require.Contains(t, failure.Error, "private panic detail")
	require.Contains(t, failure.Error, "runtime/debug.Stack")
}

func TestWrappedDomainErrors(t *testing.T) {
	for _, tt := range []struct {
		name   string
		err    error
		status int
		code   api.ErrorResponseCode
	}{
		{"invalid input", expression.ErrInvalidInput, http.StatusBadRequest, api.ErrorResponseCodeINVALIDREQUEST},
		{"invalid expression", expression.ErrInvalidExpression, http.StatusUnprocessableEntity, api.ErrorResponseCodeINVALIDEXPRESSION},
		{"division by zero", expression.ErrDivisionByZero, http.StatusUnprocessableEntity, api.ErrorResponseCodeDIVISIONBYZERO},
		{"invalid domain", expression.ErrInvalidDomain, http.StatusUnprocessableEntity, api.ErrorResponseCodeINVALIDDOMAIN},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/calculations", nil)
			handleError(req.Context(), recorder, req, fmt.Errorf("calculate: %w", tt.err))
			assertErrorResponse(t, recorder, tt.status, tt.code)
		})
	}
}

func TestCalculationOverNetwork(t *testing.T) {
	srv := httptest.NewServer(newHandler(t))
	defer srv.Close()
	response, err := srv.Client().Post(srv.URL+"/api/v1/calculations", "application/json", strings.NewReader(`{"expression":"0.1+0.2"}`))
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, response.Body.Close())
	}()
	var result api.CalculationResponse
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.NoError(t, json.NewDecoder(response.Body).Decode(&result))
	require.Equal(t, "0.3", result.Result)
}
