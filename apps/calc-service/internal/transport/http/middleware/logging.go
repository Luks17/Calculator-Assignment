package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Logging attaches an independent contextual logger and request ID to each
// request, then logs the completed response. Place it outside recovery so that
// access logs include recovered failures' final status codes.
func Logging(logger *zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return NewChain(
			hlog.NewHandler(*logger),
			hlog.RequestIDHandler("req_id", "X-Request-ID"),
			requestLogFields,
			hlog.AccessHandler(logRequest),
		).Apply(next)
	}
}

func requestLogFields(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zerolog.Ctx(r.Context()).UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("ip", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Str("origin", r.Header.Get("Origin"))
		})
		next.ServeHTTP(w, r)
	})
}

func logRequest(r *http.Request, status, size int, elapsed time.Duration) {
	// An empty handler implicitly sends 200, even without Write or WriteHeader.
	if status == 0 {
		status = http.StatusOK
	}
	logger := zerolog.Ctx(r.Context())
	var event *zerolog.Event
	switch {
	case status < http.StatusBadRequest:
		event = logger.Info()
	case status < http.StatusInternalServerError:
		event = logger.Warn()
	default:
		event = logger.Error()
	}
	event.Dur("elapsed_ms", elapsed).
		Int("status_code", status).
		Int("bytes_written", size).
		Msg("incoming request")
}
