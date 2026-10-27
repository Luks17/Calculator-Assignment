# Backend notes

The Go service parses expressions with Participle and evaluates them with
`shopspring/decimal`. Parsing produces a typed AST without performing arithmetic.
The calculator runs the parse-then-evaluate workflow and is safe for concurrent use.

See [development](../coding/development.md) for setup, configuration, and generation commands,
and [the API guide](api.md) for arithmetic rules, limits, and error responses.
The [parser design](parser.md) covers the implemented grammar, AST, and validation.

## Structure

Paths below are relative to `apps/calc-service/`:

```text
cmd/server/                       # Load configuration, start the server, handle signals
internal/
  app/                            # Wire dependencies
  domain/expression/              # Lexer, parser, AST, evaluator, calculator
  infrastructure/config/          # Environment configuration
  transport/http/
    middleware/                   # Request logging, CORS, panic recovery
    server/                       # HTTP configuration, handlers, error mapping
      gen/                        # Generated ogen server; do not edit
```

Only the executable imports the concrete configuration adapter (`caarlos0/env`).
Consumers declare their own small configuration interfaces. `internal/app` wires
dependencies; the HTTP server package configures routing, limits, and middleware;
the executable owns startup and shutdown.

## Runtime behavior

Startup failures terminate the process. SIGINT/SIGTERM trigger graceful shutdown,
with up to five seconds to drain active requests before remaining connections are closed.

CORS settings are described in [application configuration](../coding/development.md#application-configuration).
The middleware permits `GET, POST, PUT, PATCH, DELETE, OPTIONS`, allows
`Content-Type` and `Authorization`, permits credentials, and exposes `X-Request-ID`.
Preflight OPTIONS requests return `204` without invoking the API handler.
Disallowed origins receive no CORS permission headers; ordinary requests still
reach the handler. CORS is a browser policy, not authentication.

## Logging and recovery

The executable creates a zerolog logger on the startup context. Logging middleware
clones it for each request, attaches a request ID through zerolog's `hlog` helpers,
and returns the ID in `X-Request-ID`. Handlers use `zerolog.Ctx(ctx)` and can add
request-local fields with `UpdateContext`.

Completed requests log method, path, remote address, user agent, origin, elapsed
time, response size, and status. Successes use info level, 4xx responses warn, and
5xx responses error. Logging wraps panic recovery so recovered failures are logged
as `500` responses. Panic details and stack traces remain in logs, not responses;
`http.ErrAbortHandler` is rethrown for the HTTP server to handle.
