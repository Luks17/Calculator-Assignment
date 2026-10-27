# Calculator service

Go REST API at `http://localhost:8080`. `GET /healthz` checks availability;
`POST /api/v1/calculations` evaluates an expression.

**Docker:** follow the [root quick start](../../README.md#run-with-docker) to run
the stack, including [interactive OpenAPI docs](http://localhost:8081).

**Locally:** with Go 1.27.3+, run from this directory:

```sh
APP_ENV=local go run ./cmd/server
```

Tests: `make test` here uses local tools; `make test` at the repository root uses Docker.

See [setup and configuration](../../docs/coding/development.md),
[API examples](../../docs/calc-service/api.md), and [backend notes](../../docs/calc-service/backend.md).
