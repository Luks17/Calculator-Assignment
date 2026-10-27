package expression

import (
	"fmt"
	"math/big"

	"github.com/shopspring/decimal"
)

const (
	// Precision is the number of fractional places for division and square roots.
	// Rounding is to nearest, with ties away from zero.
	Precision int32 = 16
	// MaxExponent bounds the absolute value of integer power operands.
	MaxExponent = 1000
	// MaxDigits bounds both coefficient digits and the absolute decimal scale of
	// every intermediate result, including intermediate powers.
	MaxDigits = 4096
)

// Evaluate calculates a parser-independent AST using only decimal and integer
// arithmetic. It does not modify decimal's process-wide precision settings.
func Evaluate(node Node) (decimal.Decimal, error) {
	return evaluate(node, 0)
}

func evaluate(node Node, depth int) (decimal.Decimal, error) {
	if depth > MaxNesting {
		return decimal.Decimal{}, fmt.Errorf("%w: evaluation nesting exceeds %d levels", ErrInvalidDomain, MaxNesting)
	}
	var result decimal.Decimal
	switch n := node.(type) {
	case Number:
		result = n.Value
	case Unary:
		operand, err := evaluate(n.Operand, depth+1)
		if err != nil {
			return decimal.Decimal{}, err
		}
		switch n.Operator {
		case Positive:
			result = operand
		case Negative:
			result = operand.Neg()
		case Percentage:
			result = operand.Shift(-2)
		case SquareRoot:
			if operand.IsNegative() {
				return decimal.Decimal{}, fmt.Errorf("%w: cannot calculate the square root of a negative number", ErrInvalidDomain)
			}
			result = squareRoot(operand)
		default:
			return decimal.Decimal{}, fmt.Errorf("unknown unary operator %q", n.Operator)
		}
	case Binary:
		left, err := evaluate(n.Left, depth+1)
		if err != nil {
			return decimal.Decimal{}, err
		}
		right, err := evaluate(n.Right, depth+1)
		if err != nil {
			return decimal.Decimal{}, err
		}
		switch n.Operator {
		case Add:
			result = left.Add(right)
		case Subtract:
			result = left.Sub(right)
		case Multiply:
			result = left.Mul(right)
		case Divide:
			if right.IsZero() {
				return decimal.Decimal{}, ErrDivisionByZero
			}
			result = left.DivRound(right, Precision)
		case Power:
			result, err = power(left, right)
			if err != nil {
				return decimal.Decimal{}, err
			}
		default:
			return decimal.Decimal{}, fmt.Errorf("unknown binary operator %q", n.Operator)
		}
	default:
		return decimal.Decimal{}, fmt.Errorf("unknown expression node %T", node)
	}
	return checked(result)
}

func checked(value decimal.Decimal) (decimal.Decimal, error) {
	if value.IsZero() {
		return decimal.Zero, nil
	}
	coefficient := value.Coefficient()
	digits := coefficient.String()
	end := len(digits)
	for digits[end-1] == '0' {
		end--
	}
	// Discard insignificant trailing zeroes before enforcing limits. Otherwise,
	// DivRound's padded 0.0100000000000000 would consume sixteen scale places
	// per multiplication instead of two.
	exponent := int64(value.Exponent()) + int64(len(digits)-end)
	count := end
	if digits[0] == '-' {
		count--
	}
	if exponent > MaxDigits || exponent < -MaxDigits || count > MaxDigits {
		return decimal.Decimal{}, fmt.Errorf("%w: calculation exceeds the %d-digit precision or scale limit", ErrInvalidDomain, MaxDigits)
	}
	if end != len(digits) {
		coefficient.SetString(digits[:end], 10)
	}
	return decimal.NewFromBigInt(coefficient, int32(exponent)), nil
}

func power(base, exponent decimal.Decimal) (decimal.Decimal, error) {
	if !exponent.IsInteger() || exponent.Abs().GreaterThan(decimal.NewFromInt(MaxExponent)) {
		return decimal.Decimal{}, fmt.Errorf("%w: exponents must be integers between -%d and %d", ErrInvalidDomain, MaxExponent, MaxExponent)
	}
	n := exponent.IntPart()
	if base.IsZero() {
		if n < 0 {
			return decimal.Decimal{}, ErrDivisionByZero
		}
		if n == 0 {
			return decimal.Decimal{}, fmt.Errorf("%w: zero to the power of zero is undefined", ErrInvalidDomain)
		}
	}
	negative := n < 0
	if negative {
		n = -n
	}
	result := decimal.NewFromInt(1)
	for n > 0 {
		var err error
		if n%2 == 1 {
			result, err = checked(result.Mul(base))
			if err != nil {
				return decimal.Decimal{}, err
			}
		}
		n /= 2
		if n > 0 {
			base, err = checked(base.Mul(base))
			if err != nil {
				return decimal.Decimal{}, err
			}
		}
	}
	if negative {
		result = decimal.NewFromInt(1).DivRound(result, Precision)
	}
	return result, nil
}

// squareRoot computes floor(sqrt(value) * 10^Precision) with integer arithmetic,
// then compares against the exact half-way point to round without floats or
// an iteration-dependent convergence tolerance.
func squareRoot(value decimal.Decimal) decimal.Decimal {
	coefficient := value.Coefficient()
	scale := int64(value.Exponent()) + 2*int64(Precision)
	numerator := new(big.Int).Set(coefficient)
	denominator := big.NewInt(1)
	if scale >= 0 {
		numerator.Mul(numerator, tenTo(scale))
	} else {
		denominator = tenTo(-scale)
	}
	radical := new(big.Int).Quo(numerator, denominator)
	root := new(big.Int).Sqrt(radical)

	// sqrt(numerator/denominator) >= root + 1/2 iff
	// 4*numerator >= denominator*(2*root+1)^2.
	midpoint := new(big.Int).Lsh(new(big.Int).Set(root), 1)
	midpoint.Add(midpoint, big.NewInt(1))
	midpoint.Mul(midpoint, midpoint)
	midpoint.Mul(midpoint, denominator)
	if new(big.Int).Lsh(numerator, 2).Cmp(midpoint) >= 0 {
		root.Add(root, big.NewInt(1))
	}
	return decimal.NewFromBigInt(root, -Precision)
}

func tenTo(exponent int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(exponent), nil)
}
