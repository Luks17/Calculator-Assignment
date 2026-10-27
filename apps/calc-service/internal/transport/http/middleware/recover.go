package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
)

// Recover delegates panics to the supplied error handler, preserving centralized
// response formatting. Stack traces are included in the error for server-side
// logging, not written to the client. http.ErrAbortHandler is left to net/http.
func Recover(handleError func(context.Context, http.ResponseWriter, *http.Request, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					if rec == http.ErrAbortHandler {
						panic(rec)
					}

					handleError(r.Context(), w, r, fmt.Errorf("request panic: %v\n%s", rec, debug.Stack()))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
