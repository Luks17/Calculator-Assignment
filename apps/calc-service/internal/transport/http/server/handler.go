package server

import (
	"context"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/domain/expression"
	api "github.com/luks17/sezzle-assignment/apps/calc-service/internal/transport/http/server/gen"
)

type calculationHandler struct {
	calculator *expression.Calculator
}

var _ api.Handler = (*calculationHandler)(nil)

func (h *calculationHandler) Calculate(_ context.Context, request *api.CalculationRequest) (api.CalculateRes, error) {
	result, err := h.calculator.Calculate(request.Expression)
	if err != nil {
		return nil, err
	}
	return &api.CalculationResponse{Result: result.String()}, nil
}
