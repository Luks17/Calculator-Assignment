package expression

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrInvalidExpression = errors.New("invalid expression")
	ErrDivisionByZero    = errors.New("division by zero is not allowed")
	ErrInvalidDomain     = errors.New("invalid mathematical domain")
)

// SyntaxError locates a syntax problem using a zero-based byte offset and
// one-based line and column numbers.
type SyntaxError struct {
	Offset  int
	Line    int
	Column  int
	Message string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("invalid expression at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

func (e *SyntaxError) Unwrap() error { return ErrInvalidExpression }
