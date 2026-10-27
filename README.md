# Calculator

A React/TypeScript frontend and Go REST API for evaluating mathematical expressions.

## Run with Docker

Requires Docker with Compose v2.24+ and Make. Go, Node.js, and pnpm run inside the containers.

From the repository root:

```sh
make env
make dev-up
```

`make env` is a first-time setup step; skip it if both apps already have `.env` files.
The first start builds the development images. Both apps reload on source changes.

| Service | Default URL |
| --- | --- |
| Frontend | <http://localhost:5173> |
| Backend health check | <http://localhost:8080/healthz> |
| Interactive OpenAPI docs (Scalar) | <http://localhost:8081> |

Stop the stack with `make dev-down`.

## Development commands

Root Make commands run in Docker and default to the backend:

```sh
make test
make test APP=calc-web
```

`build`, `lint`, `format`, and `generate` work the same way.
To use local tools instead, run Make in the app directory, for example
`make -C apps/calc-service test` or `make -C apps/calc-web test`.

## Documentation

- [Design decisions and assumptions](docs/design-decisions.md)
- [Setup, local execution, configuration, and commands](docs/coding/development.md)
- [API calls, arithmetic behavior, and errors](docs/calc-service/api.md)
- [Backend notes](docs/calc-service/backend.md) · [Frontend notes](docs/calc-web/frontend.md)
