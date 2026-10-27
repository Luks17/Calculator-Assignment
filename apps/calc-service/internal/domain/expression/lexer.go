package expression

import (
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

const (
	numberToken lexer.TokenType = -2 - iota
	operatorToken
	sqrtToken
)

type lexerDefinition struct{}

func (lexerDefinition) Symbols() map[string]lexer.TokenType {
	return map[string]lexer.TokenType{
		"EOF": lexer.EOF, "Number": numberToken, "Operator": operatorToken, "Sqrt": sqrtToken,
	}
}

func (d lexerDefinition) Lex(filename string, r io.Reader) (lexer.Lexer, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return d.LexString(filename, string(input))
}

func (lexerDefinition) LexString(filename, input string) (lexer.Lexer, error) {
	return &expressionLexer{input: input, pos: lexer.Position{Filename: filename, Line: 1, Column: 1}}, nil
}

type expressionLexer struct {
	input string
	pos   lexer.Position
	depth int
}

func (l *expressionLexer) advance(n int) {
	l.pos.Advance(l.input[:n])
	l.input = l.input[n:]
}

func (l *expressionLexer) Next() (lexer.Token, error) {
	for len(l.input) > 0 {
		r, size := utf8.DecodeRuneInString(l.input)
		if !unicode.IsSpace(r) {
			break
		}
		l.advance(size)
	}
	if l.input == "" {
		return lexer.EOFToken(l.pos), nil
	}

	pos := l.pos
	kind, size := operatorToken, 1
	switch c := l.input[0]; {
	case isDigit(c) || c == '.':
		kind, size = numberToken, 0
		for size < len(l.input) && isDigit(l.input[size]) {
			size++
		}
		if size < len(l.input) && l.input[size] == '.' {
			size++
			for size < len(l.input) && isDigit(l.input[size]) {
				size++
			}
		}
		if l.input[:size] == "." {
			return lexer.Token{}, participle.Errorf(pos, "expected a decimal number")
		}
	case strings.HasPrefix(l.input, "sqrt"):
		kind, size = sqrtToken, len("sqrt")
	case strings.ContainsRune("+-*/^%()", rune(c)):
		if c == '(' {
			l.depth++
			if l.depth > MaxNesting {
				return lexer.Token{}, participle.Errorf(pos, "nesting exceeds %d levels", MaxNesting)
			}
		} else if c == ')' && l.depth > 0 {
			l.depth--
		}
	default:
		r, _ := utf8.DecodeRuneInString(l.input)
		return lexer.Token{}, participle.Errorf(pos, "unexpected character %q", r)
	}
	token := lexer.Token{Type: kind, Value: l.input[:size], Pos: pos}
	l.advance(size)
	return token, nil
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
