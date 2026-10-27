# Calculator service

Run from this directory:

```sh
go run ./cmd/server
```

`SERVER_HOST` defaults to `0.0.0.0`; `SERVER_PORT` defaults to `8080`.
Environment variables must be exported (the service does not load `.env` files;
Docker Compose loads them). SIGINT/SIGTERM trigger a five-second graceful shutdown.

```sh
curl http://localhost:8080/api/v1/calculations \
  -H 'Content-Type: application/json' \
  -d '{"expression":"sqrt(81) + 2 ^ 3"}'
# {"result":"17"}
```

`GET /healthz` returns 200. Calculations are stateless and results are decimal
strings without unnecessary trailing zeroes.

For interactive OpenAPI documentation at <http://localhost:8081>, run
`docker compose up -d api-reference` from the repository root (also included in
`make dev-up`). Scalar serves the contract in `contracts/calc/openapi`, mounted
read-only into the documentation container.

## Configuration and logging

Only the executable imports the concrete environment configuration adapter.
Configuration is loaded with `caarlos0/env`. Each consumer declares its own
configuration interface explicitly; Go's implicit interface satisfaction connects
them without shared interface embedding or concrete infrastructure dependencies.

Initialization has one owner per concern: `internal/app` wires dependencies,
`internal/transport/http/server` configures HTTP routing, limits, and a single
middleware chain, and `cmd/server/main.go` starts the server and handles signals.
Startup failures terminate the process. Shutdown drains active requests for up
to five seconds, then forcibly closes remaining connections if necessary.

`APP_ENV` defaults to `production`. Cross-origin access is denied by default;
set `ALLOWED_ORIGINS` to a comma-separated list of exact origins, for example
`http://localhost:5173,http://localhost:8081` for the web app and Scalar. Entries
are trimmed; wildcards, subdomain patterns, and trailing slashes are not expanded.
Only `APP_ENV=local` permits any nonempty origin, as shown in `.env.example`.
Update an existing `.env` with these settings before restarting Docker Compose;
for `go run`, export the variables yourself.

CORS supports `GET, POST, PUT, PATCH, DELETE, OPTIONS`, allows `Content-Type` and
`Authorization`, permits credentials, and exposes `X-Request-ID` to browser code.
OPTIONS requests return 204 without invoking the API handler. Disallowed origins
receive no CORS permission headers, so the browser blocks cross-origin access;
ordinary requests still reach the handler. CORS is not authentication.

The executable attaches a zerolog logger to the startup context. HTTP logging
middleware clones it per request, attaches a request ID using zerolog's `hlog`
helpers, and returns that ID in `X-Request-ID`. Handlers retrieve their logger
with `zerolog.Ctx(ctx)` and can add request-local fields with `UpdateContext`.
Completed requests log method, path, remote address, user agent, origin, elapsed
time, response bytes, and status: info for success, warn for 4xx, error for 5xx.
Logging wraps recovery so recovered panics are recorded as 500 responses.
Panic details and stack traces stay in contextual logs, never client responses;
`http.ErrAbortHandler` is rethrown for the HTTP server to handle.

## Arithmetic and limits

- Decimal literals and `+`, `-`, `*`, and postfix `%` use exact decimal arithmetic.
  `%` divides its operand by 100, not modulo: `200+10%` is `200.1`.
- Division, negative integer powers (reciprocals), and square roots round to
  **16 fractional decimal places**, to nearest with ties away from zero.
  Rounding occurs at each such operation, not just when serializing the result.
  Square roots are calculated with scaled integer arithmetic, never binary floats.
- Positive integer powers are exact. Exponents must be integers in `[-1000, 1000]`.
  `^` is right-associative, `-2^2` is `-4`, and `2^-3` is valid.
- Division by zero and zero raised to a negative exponent are division-by-zero
  errors. Negative square roots, noninteger/out-of-range exponents, and `0^0`
  are domain errors.
- Expressions contain at most 512 Unicode characters. Parenthesis nesting and
  AST operator depth are each limited to 64 levels. Scientific notation, implicit
  multiplication, other functions, and repeated postfix percentages are rejected.
- Every normalized intermediate decimal (including intermediate integer powers) is limited
  to 4096 coefficient digits and a decimal exponent in `[-4096, 4096]`.
  Exceeding these bounds is a domain error; bounds are checked during powers to
  prevent nested exponentiation from exhausting memory.
- Request bodies are limited to 16 KiB, allowing JSON-escaped expressions while
  bounding memory used by the generated decoder.

## Errors

The transport's centralized error handler emits the contract's `{code, message}`
JSON envelope for both generated request-validation failures and domain errors:

| HTTP | Code | Meaning |
| --- | --- | --- |
| 400 | `INVALID_REQUEST` | Invalid JSON/schema/content type, body limit, empty or oversized input |
| 422 | `INVALID_EXPRESSION` | Syntax error (with line and column), nesting limit |
| 422 | `DIVISION_BY_ZERO` | Division by zero |
| 422 | `INVALID_DOMAIN` | Invalid mathematical domain or arithmetic resource limit |
| 500 | `INTERNAL_ERROR` | Unexpected error; details are logged, not returned |

Routing errors use the same envelope with `INVALID_REQUEST` and HTTP 404/405.
Generated files in `internal/transport/http/server/gen` are not edited by hand.
Use `make generate` after changing the OpenAPI contract.

## Verification

```sh
go test -race ./...
go vet ./...
go build ./...
```
