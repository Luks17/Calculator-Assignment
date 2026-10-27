package expression_test

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/domain/expression"
)

func newParser(t *testing.T) *expression.Parser {
	t.Helper()
	parser, err := expression.NewParser()
	require.NoError(t, err)
	return parser
}

func number(value string) expression.Number {
	return expression.Number{Value: decimal.RequireFromString(value)}
}

func TestParserAST(t *testing.T) {
	parser := newParser(t)
	tests := []struct {
		input string
		want  expression.Node
	}{
		{"2+3*4", expression.Binary{Operator: expression.Add, Left: number("2"), Right: expression.Binary{Operator: expression.Multiply, Left: number("3"), Right: number("4")}}},
		{"(2+3)*4", expression.Binary{Operator: expression.Multiply, Left: expression.Binary{Operator: expression.Add, Left: number("2"), Right: number("3")}, Right: number("4")}},
		{"2^3^2", expression.Binary{Operator: expression.Power, Left: number("2"), Right: expression.Binary{Operator: expression.Power, Left: number("3"), Right: number("2")}}},
		{"-2^2", expression.Unary{Operator: expression.Negative, Operand: expression.Binary{Operator: expression.Power, Left: number("2"), Right: number("2")}}},
		{"2^-3", expression.Binary{Operator: expression.Power, Left: number("2"), Right: expression.Unary{Operator: expression.Negative, Operand: number("3")}}},
		{"sqrt(81)", expression.Unary{Operator: expression.SquareRoot, Operand: number("81")}},
		{"250*15%", expression.Binary{Operator: expression.Multiply, Left: number("250"), Right: expression.Unary{Operator: expression.Percentage, Operand: number("15")}}},
		{"-2%^2", expression.Unary{Operator: expression.Negative, Operand: expression.Binary{Operator: expression.Power, Left: expression.Unary{Operator: expression.Percentage, Operand: number("2")}, Right: number("2")}}},
		{"0.1+0.2", expression.Binary{Operator: expression.Add, Left: number("0.1"), Right: number("0.2")}},
		{"3-2-1", expression.Binary{Operator: expression.Subtract, Left: expression.Binary{Operator: expression.Subtract, Left: number("3"), Right: number("2")}, Right: number("1")}},
		{"8/4/2", expression.Binary{Operator: expression.Divide, Left: expression.Binary{Operator: expression.Divide, Left: number("8"), Right: number("4")}, Right: number("2")}},
		{"+-.5", expression.Unary{Operator: expression.Positive, Operand: expression.Unary{Operator: expression.Negative, Operand: number(".5")}}},
		{"5.", number("5.")},
		{"\t\n .5 \u2003", number(".5")},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parser.Parse(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParserRejectsInvalidSyntax(t *testing.T) {
	parser := newParser(t)
	for _, input := range []string{
		"2+", "sqrt()", "(2+3", "2**3", "foo(4)", "10%%", "1e3",
		".", "1..2", "2 3", "2(3)", "sqrt 4", "SQRT(4)", "sqrtx(4)",
		"2)", "()", "*2", "2//3", "2^", "%2", "2,3", "2 & 3", "NaN", "∞", "１２", "1_000",
	} {
		t.Run(input, func(t *testing.T) {
			_, err := parser.Parse(input)
			var syntax *expression.SyntaxError
			require.ErrorIs(t, err, expression.ErrInvalidExpression)
			require.ErrorAs(t, err, &syntax)
			require.Positive(t, syntax.Line)
			require.Positive(t, syntax.Column)
		})
	}
}

func TestParserSourcePositions(t *testing.T) {
	parser := newParser(t)
	for _, tt := range []struct {
		input                string
		offset, line, column int
	}{
		{"1 +\n  @", 6, 2, 3},
		{"\u2003@", 3, 1, 2},
		{"2+", 2, 1, 3},
	} {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parser.Parse(tt.input)
			var syntax *expression.SyntaxError
			require.ErrorAs(t, err, &syntax)
			require.Equal(t, tt.offset, syntax.Offset, "offset")
			require.Equal(t, tt.line, syntax.Line, "line")
			require.Equal(t, tt.column, syntax.Column, "column")
		})
	}
}

func TestParserInputLimits(t *testing.T) {
	parser := newParser(t)
	for _, tt := range []struct {
		name, input string
		wantErr     error
	}{
		{"empty", "", expression.ErrInvalidInput},
		{"whitespace only", " \t\n", expression.ErrInvalidInput},
		{"too long", strings.Repeat("1", expression.MaxLength+1), expression.ErrInvalidInput},
		{"invalid UTF-8", string([]byte{0xff}), expression.ErrInvalidInput},
		{"maximum length", strings.Repeat("1", expression.MaxLength), nil},
		{"maximum Unicode length", strings.Repeat("\u2003", expression.MaxLength-1) + "1", nil},
		{"maximum parentheses", strings.Repeat("(", expression.MaxNesting) + "1" + strings.Repeat(")", expression.MaxNesting), nil},
		{"maximum unary nesting", strings.Repeat("-", expression.MaxNesting) + "1", nil},
		{"excessive parentheses", strings.Repeat("(", expression.MaxNesting+1) + "1" + strings.Repeat(")", expression.MaxNesting+1), expression.ErrInvalidExpression},
		{"excessive unary nesting", strings.Repeat("-", expression.MaxNesting+1) + "1", expression.ErrInvalidExpression},
		{"excessive power nesting", strings.Repeat("1^", expression.MaxNesting+1) + "1", expression.ErrInvalidExpression},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse(tt.input)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestParserDoesNotEvaluate(t *testing.T) {
	parser := newParser(t)
	for _, input := range []string{"1/0", "sqrt(-1)", "2^0.5", "2^1001", "0^0"} {
		t.Run(input, func(t *testing.T) {
			_, err := parser.Parse(input)
			require.NoError(t, err, "mathematical validity is the evaluator's responsibility")
		})
	}
}
