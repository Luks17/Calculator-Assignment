package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/rs/zerolog"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/domain/expression"
	api "github.com/luks17/sezzle-assignment/apps/calc-service/internal/transport/http/server/gen"
)

// handleError is the single place translating errors into API responses,
// including ogen decoding/validation failures and recovered panics.
func handleError(ctx context.Context, w http.ResponseWriter, _ *http.Request, err error) {
	status := http.StatusInternalServerError
	response := api.ErrorResponse{
		Code:    api.ErrorResponseCodeINTERNALERROR,
		Message: "An unexpected error occurred",
	}
	var decodeError *ogenerrors.DecodeRequestError
	switch {
	case errors.As(err, &decodeError):
		status, response.Code = http.StatusBadRequest, api.ErrorResponseCodeINVALIDREQUEST
		response.Message = "Invalid request body; expression must be a string of 1 to 512 characters"
	case errors.Is(err, expression.ErrInvalidInput):
		status, response.Code = http.StatusBadRequest, api.ErrorResponseCodeINVALIDREQUEST
		response.Message = err.Error()
	case errors.Is(err, expression.ErrInvalidExpression):
		status, response.Code = http.StatusUnprocessableEntity, api.ErrorResponseCodeINVALIDEXPRESSION
		response.Message = err.Error()
	case errors.Is(err, expression.ErrDivisionByZero):
		status, response.Code = http.StatusUnprocessableEntity, api.ErrorResponseCodeDIVISIONBYZERO
		response.Message = err.Error()
	case errors.Is(err, expression.ErrInvalidDomain):
		status, response.Code = http.StatusUnprocessableEntity, api.ErrorResponseCodeINVALIDDOMAIN
		response.Message = err.Error()
	default:
		zerolog.Ctx(ctx).Error().Err(err).Msg("request failed")
	}
	writeError(ctx, w, status, response)
}

func writeError(ctx context.Context, w http.ResponseWriter, status int, response api.ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("write error response")
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	writeError(r.Context(), w, http.StatusNotFound, api.ErrorResponse{
		Code: api.ErrorResponseCodeINVALIDREQUEST, Message: "Route not found",
	})
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request, allowed string) {
	w.Header().Set("Allow", allowed)
	writeError(r.Context(), w, http.StatusMethodNotAllowed, api.ErrorResponse{
		Code: api.ErrorResponseCodeINVALIDREQUEST, Message: "Method not allowed",
	})
}
