package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/transport/http/middleware"
)

func TestRecoverDelegatesToErrorHandler(t *testing.T) {
	called := false
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	handler := middleware.Recover(func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		called = true
		require.Equal(t, req.Context(), ctx)
		require.Same(t, req, r)
		require.Error(t, err)
		require.ErrorContains(t, err, "private panic detail")
		require.ErrorContains(t, err, "runtime/debug.Stack")
		w.WriteHeader(http.StatusInternalServerError)
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("private panic detail")
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	require.True(t, called, "recovery must delegate response formatting")
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Empty(t, recorder.Body.String())
}

func TestRecoverPreservesAbortHandler(t *testing.T) {
	handler := middleware.Recover(func(context.Context, http.ResponseWriter, *http.Request, error) {
		require.FailNow(t, "http.ErrAbortHandler must not be converted to an error response")
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	require.PanicsWithValue(t, http.ErrAbortHandler, func() {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	})
}

func TestRecoverPassesThroughSuccess(t *testing.T) {
	handler := middleware.Recover(func(context.Context, http.ResponseWriter, *http.Request, error) {
		require.FailNow(t, "error handler must not run for successful responses")
	})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusNoContent, recorder.Code)
}
