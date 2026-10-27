package expression_test

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/domain/expression"
)

func newCalculator(t *testing.T) *expression.Calculator {
	t.Helper()
	calculator, err := expression.NewCalculator()
	require.NoError(t, err)
	return calculator
}

func TestCalculator(t *testing.T) {
	calculator := newCalculator(t)
	for _, tt := range []struct{ input, want string }{
		{"2+3*4", "14"}, {"(2+3)*4", "20"}, {"2^3^2", "512"},
		{"-2^2", "-4"}, {"(-2)^2", "4"}, {"2^-3", "0.125"},
		{"sqrt(81)", "9"}, {"sqrt(81) + 2 ^ 3", "17"},
		{"250*15%", "37.5"}, {"200+10%", "200.1"}, {"0.1+0.2", "0.3"},
		{".5+5.", "5.5"}, {"+--2", "2"}, {"3-2-1", "0"}, {"8/4/2", "1"},
		{"sqrt(sqrt(81))", "3"}, {"sqrt(0)", "0"}, {"sqrt(.25)", "0.5"},
		{"sqrt(2)", "1.414213562373095"}, {"sqrt(3)", "1.7320508075688773"},
		{"sqrt(81129638414606699710187514626049)", "9007199254740993"},
		{"sqrt(1.0000000000000001000000000000000025)", "1.0000000000000001"},
		{"sqrt(1.0000000000000001000000000000000024)", "1"},
		{"sqrt(0.00000000000000000000000000000001)", "0.0000000000000001"},
		{"sqrt(0.0000000000000000000000000000000025)", "0.0000000000000001"},
		{"sqrt(0.0000000000000000000000000000000024)", "0"},
		{"1/3", "0.3333333333333333"}, {"2/3", "0.6666666666666667"},
		{"10/3", "3.3333333333333333"}, {"-2/3", "-0.6666666666666667"},
		{"1/20000000000000000", "0.0000000000000001"},
		{"-1/20000000000000000", "-0.0000000000000001"},
		{"1/3*3", "0.9999999999999999"},
		{"0.000000000000000001*100", "0.0000000000000001"},
		{"1.123456789012345678+0.000000000000000001", "1.123456789012345679"},
		{"3^-1", "0.3333333333333333"}, {"(-2)^-3", "-0.125"},
		{"2^0", "1"}, {"0^3", "0"}, {"(-2)^3", "-8"}, {"2^3.0", "8"},
		{"1^1000", "1"}, {"1^-1000", "1"}, {"(1/100)^1000", "0." + strings.Repeat("0", 1999) + "1"},
		{"(2^1000)^4", new(big.Int).Lsh(big.NewInt(1), 4000).String()},
		{"1.000000000000000000^1000", "1"},
		{"9007199254740993+1", "9007199254740994"},
		{"(10+5)%", "0.15"}, {"2*3^2", "18"},
	} {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel() // Exercise a single shared, immutable Participle parser.
			got, err := calculator.Calculate(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.want, got.String())
		})
	}
}

func TestCalculatorEvaluationErrors(t *testing.T) {
	calculator := newCalculator(t)
	for _, tt := range []struct {
		input string
		want  error
	}{
		{"1/0", expression.ErrDivisionByZero}, {"1/(2-2)", expression.ErrDivisionByZero},
		{"0^-1", expression.ErrDivisionByZero}, {"1/0%", expression.ErrDivisionByZero},
		{"sqrt(-1)", expression.ErrInvalidDomain}, {"0^0", expression.ErrInvalidDomain},
		{"2^0.5", expression.ErrInvalidDomain}, {"2^3%", expression.ErrInvalidDomain},
		{"2^1001", expression.ErrInvalidDomain}, {"2^-1001", expression.ErrInvalidDomain},
		{"2^99999999999999999999999999999999", expression.ErrInvalidDomain},
		{"(2^1000)^1000", expression.ErrInvalidDomain},
		{"(0.01^1000)^1000", expression.ErrInvalidDomain},
		{"(10^1000)^5", expression.ErrInvalidDomain},
	} {
		t.Run(tt.input, func(t *testing.T) {
			_, err := calculator.Calculate(tt.input)
			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestEvaluateWithoutParser(t *testing.T) {
	node := expression.Binary{Operator: expression.Add, Left: number("0.1"), Right: number("0.2")}
	got, err := expression.Evaluate(node)
	require.NoError(t, err)
	require.Equal(t, "0.3", got.String())
}

func TestEvaluateRejectsInvalidNodes(t *testing.T) {
	for _, tt := range []struct {
		name    string
		node    expression.Node
		wantErr error
	}{
		{"exponent too large", expression.Number{Value: decimal.New(1, expression.MaxDigits+1)}, expression.ErrInvalidDomain},
		{"exponent too small", expression.Number{Value: decimal.New(1, -expression.MaxDigits-1)}, expression.ErrInvalidDomain},
		{"too many digits", expression.Number{Value: decimal.RequireFromString(strings.Repeat("9", expression.MaxDigits+1))}, expression.ErrInvalidDomain},
		{"nil node", nil, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := expression.Evaluate(tt.node)
			require.Error(t, err)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func FuzzCalculate(f *testing.F) {
	calculator, err := expression.NewCalculator()
	require.NoError(f, err)
	for _, input := range []string{
		"2+3*4", "2^3^2", "-2^2", "2^-3", "sqrt(81)", "250*15%", "0.1+0.2",
		"sqrt(-1)", "(2^1000)^1000", "1/0", "2+", "1e3", "", strings.Repeat("(", 100),
	} {
		f.Add(input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		_, err := calculator.Calculate(input)
		require.True(t, err == nil || errors.Is(err, expression.ErrInvalidInput) ||
			errors.Is(err, expression.ErrInvalidExpression) ||
			errors.Is(err, expression.ErrDivisionByZero) || errors.Is(err, expression.ErrInvalidDomain),
			"unexpected error for %q: %v", input, err)
	})
}
