# Expression Parser Design

Converts an expression string into a typed AST. Parsing must not evaluate expressions.

## Implementation

- Use [participle/v2](https://github.com/alecthomas/participle) with a small custom lexer and grammar defined by Go structs.
- Recognize decimal numbers, `+ - * / ^ %`, parentheses, and only the function `sqrt`. Ignore whitespace; reject all other tokens.
- Keep signs separate from number tokens. Accept decimal literals such as `12`, `0.5`, `.5`, and `5.`; do not accept scientific notation unless explicitly added later.
- Convert number tokens with `decimal.NewFromString` (from [github.com/shopspring/decimal](https://github.com/shopspring/decimal)) through a Participle `Capture` type. Never use `float64`.

## Grammar (conceptual EBNF)

```
expression = sum ;
sum        = product, { ("+" | "-"), product } ;
product    = unary, { ("*" | "/"), unary } ;
unary      = ("+" | "-"), unary | power ;
power      = postfix, [ "^", unary ] ;
postfix    = primary, [ "%" ] ;
primary    = NUMBER | "(", expression, ")"
           | "sqrt", "(", expression, ")" ;
```

- Precedence (high to low): grouping/`sqrt`, postfix `%`, `^`, unary signs, `*`/`/`, `+`/`-`.
- `^` is right-associative (`2^3^2` means `2^(3^2)`); `-2^2` means `-(2^2)`; `2^-3` is valid syntax.
- `%` is postfix percentage, **not modulo**: `200+10%` becomes `200 + (10%)`.

## AST and boundaries

- Return an AST with `Number(decimal.Decimal)`, `Unary(operator, operand)`, and `Binary(operator, left, right)` nodes; represent `sqrt` and `%` as unary operators.
- Participle grammar structs may be normalized into these nodes so the evaluator does not depend on parsing internals.
- Report syntax errors with source positions. Reject empty/incomplete input, unmatched parentheses, and trailing tokens; enforce the API's 512-character input limit and a reasonable nesting limit.
- Evaluator responsibilities (out of scope): arithmetic, rounding, division by zero, square-root domain checks, and enforcing integer-only/bounded exponents.

## Essential parser tests

Valid: `2+3*4`, `(2+3)*4`, `2^3^2`, `-2^2`, `2^-3`, `sqrt(81)`, `250*15%`, `0.1+0.2`.

Invalid: `2+`, `sqrt()`, `(2+3`, `2**3`, `foo(4)`, `10%%`, `1e3`.
