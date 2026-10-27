package expression

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2"
	"github.com/shopspring/decimal"
)

const (
	MaxLength  = 512
	MaxNesting = 64
)

// Parser is immutable after construction and safe for concurrent use.
type Parser struct {
	grammar *participle.Parser[sumGrammar]
}

func NewParser() (*Parser, error) {
	grammar, err := participle.Build[sumGrammar](participle.Lexer(lexerDefinition{}))
	if err != nil {
		return nil, fmt.Errorf("build expression grammar: %w", err)
	}
	return &Parser{grammar: grammar}, nil
}

// Parse returns a typed AST without performing any calculations.
func (p *Parser) Parse(input string) (Node, error) {
	if !utf8.ValidString(input) {
		return nil, fmt.Errorf("%w: expression must be valid UTF-8", ErrInvalidInput)
	}
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("%w: expression must not be empty", ErrInvalidInput)
	}
	if utf8.RuneCountInString(input) > MaxLength {
		return nil, fmt.Errorf("%w: expression must not exceed %d characters", ErrInvalidInput, MaxLength)
	}
	parsed, err := p.grammar.ParseString("", input)
	if err != nil {
		var positional participle.Error
		if errors.As(err, &positional) {
			pos := positional.Position()
			return nil, &SyntaxError{Offset: pos.Offset, Line: pos.Line, Column: pos.Column, Message: positional.Message()}
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidExpression, err)
	}
	node := parsed.node()
	if !withinNesting(node, 0) {
		return nil, &SyntaxError{Line: 1, Column: 1, Message: fmt.Sprintf("AST nesting exceeds %d levels", MaxNesting)}
	}
	return node, nil
}

type decimalCapture struct {
	value decimal.Decimal
}

func (d *decimalCapture) Capture(values []string) error {
	value, err := decimal.NewFromString(values[0])
	if err != nil {
		return err
	}
	d.value = value
	return nil
}

type sumGrammar struct {
	First *productGrammar `parser:"@@"`
	Rest  []*sumTail      `parser:"@@*"`
}

type sumTail struct {
	Operator string          `parser:"@(\"+\" | \"-\")"`
	Right    *productGrammar `parser:"@@"`
}

type productGrammar struct {
	First *unaryGrammar  `parser:"@@"`
	Rest  []*productTail `parser:"@@*"`
}

type productTail struct {
	Operator string        `parser:"@(\"*\" | \"/\")"`
	Right    *unaryGrammar `parser:"@@"`
}

type unaryGrammar struct {
	Sign    string        `parser:"( @(\"+\" | \"-\")"`
	Operand *unaryGrammar `parser:"@@ )"`
	Power   *powerGrammar `parser:"| @@"`
}

type powerGrammar struct {
	Base     *postfixGrammar `parser:"@@"`
	Exponent *unaryGrammar   `parser:"( \"^\" @@ )?"`
}

type postfixGrammar struct {
	Primary *primaryGrammar `parser:"@@"`
	Percent bool            `parser:"@\"%\"?"`
}

type primaryGrammar struct {
	Number *decimalCapture `parser:"  @Number"`
	Group  *sumGrammar     `parser:"| \"(\" @@ \")\""`
	Sqrt   *sumGrammar     `parser:"| \"sqrt\" \"(\" @@ \")\""`
}

func (s *sumGrammar) node() Node {
	left := s.First.node()
	for _, tail := range s.Rest {
		left = Binary{Operator: BinaryOperator(tail.Operator), Left: left, Right: tail.Right.node()}
	}
	return left
}

func (p *productGrammar) node() Node {
	left := p.First.node()
	for _, tail := range p.Rest {
		left = Binary{Operator: BinaryOperator(tail.Operator), Left: left, Right: tail.Right.node()}
	}
	return left
}

func (u *unaryGrammar) node() Node {
	if u.Operand != nil {
		return Unary{Operator: UnaryOperator(u.Sign), Operand: u.Operand.node()}
	}
	return u.Power.node()
}

func (p *powerGrammar) node() Node {
	base := p.Base.node()
	if p.Exponent != nil {
		return Binary{Operator: Power, Left: base, Right: p.Exponent.node()}
	}
	return base
}

func (p *postfixGrammar) node() Node {
	node := p.Primary.node()
	if p.Percent {
		return Unary{Operator: Percentage, Operand: node}
	}
	return node
}

func (p *primaryGrammar) node() Node {
	switch {
	case p.Number != nil:
		return Number{Value: p.Number.value}
	case p.Group != nil:
		return p.Group.node()
	default:
		return Unary{Operator: SquareRoot, Operand: p.Sqrt.node()}
	}
}

func withinNesting(node Node, depth int) bool {
	if depth > MaxNesting {
		return false
	}
	switch n := node.(type) {
	case Unary:
		return withinNesting(n.Operand, depth+1)
	case Binary:
		return withinNesting(n.Left, depth+1) && withinNesting(n.Right, depth+1)
	default:
		return true
	}
}
