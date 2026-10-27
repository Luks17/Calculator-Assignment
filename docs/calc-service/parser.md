# Expression parser design

The parser was implemented in
[`internal/domain/expression`](../../apps/calc-service/internal/domain/expression/)
to convert expression strings into a typed AST. Parsing and evaluation were kept
separate: parsing checked structure, while the evaluator handled mathematical validity.

## Implementation

- [Participle v2](https://github.com/alecthomas/participle) was used with a
  [custom lexer](../../apps/calc-service/internal/domain/expression/lexer.go)
  and a grammar defined by Go structs in
  [`parser.go`](../../apps/calc-service/internal/domain/expression/parser.go).
- The lexer recognized ASCII decimal numbers, `+ - * / ^ %`, parentheses, and
  the lowercase function `sqrt`. Unicode whitespace was ignored; unsupported
  characters and functions were rejected.
- Signs were kept separate from number tokens. Decimal forms such as `12`,
  `0.5`, `.5`, and `5.` were accepted; scientific notation and implicit
  multiplication were not supported.
- Number tokens were converted through a Participle `Capture` type using
  `decimal.NewFromString` from
  [shopspring/decimal](https://github.com/shopspring/decimal). No `float64`
  conversion was used.
- The parser was made immutable after construction and safe for concurrent use.
  The calculator created it once and reused it across calculations.

## Implemented grammar (conceptual EBNF)

```text
expression = sum ;
sum        = product, { ("+" | "-"), product } ;
product    = unary, { ("*" | "/"), unary } ;
unary      = ("+" | "-"), unary | power ;
power      = postfix, [ "^", unary ] ;
postfix    = primary, [ "%" ] ;
primary    = NUMBER | "(", expression, ")"
           | "sqrt", "(", expression, ")" ;
```

- Precedence was encoded from highest to lowest as grouping/`sqrt`, postfix `%`,
  `^`, unary signs, `*`/`/`, then `+`/`-`.
- Addition, subtraction, multiplication, and division were made left-associative.
  Powers were made right-associative: `2^3^2` was parsed as `2^(3^2)`.
- Unary signs were placed below powers, so `-2^2` was parsed as `-(2^2)`.
  A unary exponent also allowed `2^-3`.
- `%` was implemented as a single postfix percentage operator, **not modulo**:
  `200+10%` was parsed as `200 + (10%)`. Repeated postfix percentages such as
  `10%%` were rejected.

## AST and evaluation boundary

The Participle grammar structs were normalized into three parser-independent
node types defined in [`ast.go`](../../apps/calc-service/internal/domain/expression/ast.go):

- `Number` held a `decimal.Decimal` value.
- `Unary` held an operator and operand; signs, `sqrt`, and `%` used this node.
- `Binary` held an operator and left/right operands.

Grouping affected the tree structure without adding a separate parenthesis node.
The evaluator depended on these nodes, not Participle's grammar structs.

The calculator's workflow was implemented as `Parse` followed by `Evaluate`.
Arithmetic, rounding, division-by-zero checks, square-root domain checks, and
integer-only/bounded exponents were left to the evaluator. Consequently,
`1/0`, `sqrt(-1)`, `2^0.5`, `2^1001`, and `0^0` were accepted by the parser
but rejected during evaluation. See [API behavior](api.md#precision-and-limits).

## Validation and errors

- Empty or whitespace-only input, invalid UTF-8, and expressions over 512 Unicode
  characters were rejected with `ErrInvalidInput`.
- Parenthesis nesting was limited to 64 levels in the lexer. The normalized AST
  was also checked for a maximum operator depth of 64, covering deeply nested
  unary signs and powers even without parentheses.
- Incomplete expressions, unmatched parentheses, trailing tokens, and unsupported
  syntax were rejected with `SyntaxError`, which wrapped `ErrInvalidExpression`.
  Token and grammar errors included a zero-based byte offset and one-based line
  and column; the post-parse AST-depth check reported line 1, column 1.
- The HTTP layer mapped input errors to `400 INVALID_REQUEST` and syntax/nesting
  errors to `422 INVALID_EXPRESSION`, as described in [API errors](api.md#errors).

## Test coverage

The [parser tests](../../apps/calc-service/internal/domain/expression/parser_test.go)
checked AST shape, associativity, precedence, decimal literals, Unicode whitespace,
source positions, invalid syntax, and input/nesting boundaries. Separate cases
verified that parsing did not evaluate mathematically invalid expressions.

Accepted examples included `2+3*4`, `(2+3)*4`, `2^3^2`, `-2^2`, `2^-3`,
`sqrt(81)`, `250*15%`, and `0.1+0.2`.

Rejected examples included `2+`, `sqrt()`, `(2+3`, `2**3`, `foo(4)`, `10%%`,
`1e3`, `2(3)`, and `SQRT(4)`.
