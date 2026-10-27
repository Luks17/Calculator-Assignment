package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/transport/http/middleware"
)

func TestChain(t *testing.T) {
	for _, tt := range []struct {
		name        string
		middlewares []string
		want        []string
	}{
		{"empty", nil, []string{"handler"}},
		{"single", []string{"first"}, []string{"first before", "handler", "first after"}},
		{"multiple", []string{"first", "second"}, []string{"first before", "second before", "handler", "second after", "first after"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			var wrappers []func(http.Handler) http.Handler
			for _, name := range tt.middlewares {
				wrappers = append(wrappers, func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						calls = append(calls, name+" before")
						next.ServeHTTP(w, r)
						calls = append(calls, name+" after")
					})
				})
			}
			handler := middleware.NewChain(wrappers...).Apply(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls = append(calls, "handler")
				w.WriteHeader(http.StatusNoContent)
			}))
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
			require.Equal(t, tt.want, calls)
			require.Equal(t, http.StatusNoContent, recorder.Code)
		})
	}
}
