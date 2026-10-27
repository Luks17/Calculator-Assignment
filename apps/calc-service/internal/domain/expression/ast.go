// Package expression parses and evaluates decimal calculator expressions.
package expression

import "github.com/shopspring/decimal"

// Node is a calculator AST node, independent of the parser's grammar.
type Node interface {
	expressionNode()
}

type Number struct {
	Value decimal.Decimal
}

func (Number) expressionNode() {}

type UnaryOperator string

const (
	Positive   UnaryOperator = "+"
	Negative   UnaryOperator = "-"
	SquareRoot UnaryOperator = "sqrt"
	Percentage UnaryOperator = "%"
)

type Unary struct {
	Operator UnaryOperator
	Operand  Node
}

func (Unary) expressionNode() {}

type BinaryOperator string

const (
	Add      BinaryOperator = "+"
	Subtract BinaryOperator = "-"
	Multiply BinaryOperator = "*"
	Divide   BinaryOperator = "/"
	Power    BinaryOperator = "^"
)

type Binary struct {
	Operator BinaryOperator
	Left     Node
	Right    Node
}

func (Binary) expressionNode() {}
