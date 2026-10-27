// Package app wires the calculator and its HTTP server.
package app

import (
	"context"
	"net/http"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/domain/expression"
	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/transport/http/server"
)

type Config interface {
	Address() string
	GetAppEnv() string
	GetAllowedOrigins() []string
}

func New(ctx context.Context, cfg Config) (*http.Server, error) {
	calculator, err := expression.NewCalculator()
	if err != nil {
		return nil, err
	}
	return server.New(ctx, cfg, calculator)
}
