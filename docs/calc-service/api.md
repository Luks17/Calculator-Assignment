# Calculator API

Base URL: `http://localhost:8080`.
The [OpenAPI contract](../../contracts/calc/openapi/schema.yml) defines the REST API;
[Scalar](http://localhost:8081) provides interactive documentation when Docker is running.
Start it with `make dev-up`, or just `docker compose up -d api-reference` from the repository root.
The viewer serves the contract read-only; the backend must also be running to try requests.

## Examples

### Health check

```sh
curl -i http://localhost:8080/healthz
```

Returns HTTP `200` with no body. This operational endpoint is not part of the OpenAPI contract.

### Evaluate an expression

```sh
curl http://localhost:8080/api/v1/calculations \
  -H 'Content-Type: application/json' \
  -d '{"expression":"sqrt(81) + 2 ^ 3"}'
```

HTTP `200`:

```json
{"result":"17"}
```

### Decimal arithmetic

```sh
curl http://localhost:8080/api/v1/calculations \
  -H 'Content-Type: application/json' \
  -d '{"expression":"0.1 + 0.2"}'
```

HTTP `200`:

```json
{"result":"0.3"}
```

Expressions and results are JSON strings. Results have no unnecessary trailing zeroes.
The endpoint is stateless and does not store calculation history.

### Failed calculation

```sh
curl -i http://localhost:8080/api/v1/calculations \
  -H 'Content-Type: application/json' \
  -d '{"expression":"1 / 0"}'
```

HTTP `422`:

```json
{"code":"DIVISION_BY_ZERO","message":"division by zero is not allowed"}
```

## Supported expressions

Use decimal numbers, `+`, `-`, `*`, `/`, `^`, postfix `%`, parentheses, and `sqrt(...)`.
Whitespace is ignored. Decimal forms such as `12`, `0.5`, `.5`, and `5.` are accepted.
Scientific notation, implicit multiplication, other functions, and repeated postfix
percentages such as `10%%` are rejected.

| Expression | Result | Behavior |
| --- | --- | --- |
| `2 + 3 * 4` | `14` | Multiplication before addition |
| `(2 + 3) * 4` | `20` | Parentheses override precedence |
| `2^3^2` | `512` | Powers are right-associative |
| `-2^2` | `-4` | Powers bind before unary signs |
| `2^-3` | `0.125` | Negative integer exponents are allowed |
| `250 * 15%` | `37.5` | Percentage divides its operand by 100 |
| `200 + 10%` | `200.1` | Percentage is not a relative increase or modulo |

Precedence, highest first: grouping/`sqrt`, postfix `%`, `^`, unary signs,
`*`/`/`, then `+`/`-`.

## Precision and limits

- Decimal literals, addition, subtraction, multiplication, percentage, and positive
  integer powers use exact decimal arithmetic within the resource limits below.
- Division, negative powers (reciprocals), and square roots round at each operation
  to **16 fractional decimal places**, to nearest with ties away from zero.
  Square roots use scaled integer arithmetic, not binary floating point.
- Exponents must be integers in `[-1000, 1000]`. Negative square roots, invalid
  exponents, and `0^0` are domain errors. Zero raised to a negative power is a division-by-zero error.
- Expressions are limited to 512 Unicode characters. Parenthesis nesting and AST
  operator depth are each limited to 64 levels.
- Every normalized intermediate decimal, including intermediate powers, is limited
  to 4096 coefficient digits and a decimal exponent in `[-4096, 4096]`.
  Exceeding these arithmetic limits is a domain error.
- Request bodies are limited to 16 KiB, including JSON escaping.

## Errors

All errors use a JSON object with `code` and `message`, including request validation failures.

| HTTP | Code | Meaning |
| --- | --- | --- |
| `400` | `INVALID_REQUEST` | Invalid JSON, schema, content type, body size, or empty/oversized expression |
| `422` | `INVALID_EXPRESSION` | Syntax error (with line and column) or nesting limit |
| `422` | `DIVISION_BY_ZERO` | Division by zero, including zero to a negative power |
| `422` | `INVALID_DOMAIN` | Invalid mathematical domain or arithmetic resource limit |
| `500` | `INTERNAL_ERROR` | Unexpected failure; details stay in server logs |

Unknown routes and unsupported methods return HTTP `404`/`405` with `INVALID_REQUEST`.
Responses include `X-Request-ID` for correlating requests with backend logs.
