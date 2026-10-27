// Package server adapts the generated ogen server to the calculator domain.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/domain/expression"
	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/transport/http/middleware"
	api "github.com/luks17/sezzle-assignment/apps/calc-service/internal/transport/http/server/gen"
)

// MaxBodyBytes allows escaped JSON expressions while bounding decoder memory use.
const MaxBodyBytes = 16 << 10

type Config interface {
	Address() string
	GetAppEnv() string
	GetAllowedOrigins() []string
}

// New configures the HTTP server without starting it. The executable owns its
// lifecycle; this module owns routing, middleware, and HTTP resource limits.
func New(ctx context.Context, cfg Config, calculator *expression.Calculator) (*http.Server, error) {
	generated, err := api.NewServer(&calculationHandler{calculator: calculator},
		api.WithErrorHandler(handleError),
		api.WithNotFound(notFound),
		api.WithMethodNotAllowed(methodNotAllowed),
	)
	if err != nil {
		return nil, fmt.Errorf("create API server: %w", err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				methodNotAllowed(w, r, "GET, HEAD")
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		}
		generated.ServeHTTP(w, r)
	})
	return &http.Server{
		Addr: cfg.Address(),
		Handler: middleware.NewChain(
			middleware.Logging(zerolog.Ctx(ctx)),
			middleware.CORS(cfg),
			middleware.Recover(handleError),
		).Apply(handler),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    8 << 10,
	}, nil
}
