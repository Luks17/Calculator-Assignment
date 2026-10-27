package expression

import "github.com/shopspring/decimal"

// Calculator owns the parse-then-evaluate workflow. It is safe for concurrent use.
type Calculator struct {
	parser *Parser
}

func NewCalculator() (*Calculator, error) {
	parser, err := NewParser()
	if err != nil {
		return nil, err
	}
	return &Calculator{parser: parser}, nil
}

func (c *Calculator) Calculate(input string) (decimal.Decimal, error) {
	node, err := c.parser.Parse(input)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return Evaluate(node)
}
