# Calculator frontend

React/TypeScript calculator at <http://localhost:5173>, backed by the API at
`http://localhost:8080`.

**Docker:** follow the [root quick start](../../README.md#run-with-docker) to run
both apps and the [interactive OpenAPI docs](http://localhost:8081).

**Locally:** with Node.js 24+ and pnpm 12, start the backend first, then run from this directory:

```sh
pnpm install --frozen-lockfile
pnpm dev
```

Tests: `make test` here uses local tools; `make test APP=calc-web` at the repository root uses Docker.

See [setup and configuration](../../docs/coding/development.md),
[API examples](../../docs/calc-service/api.md), and [frontend notes](../../docs/calc-web/frontend.md).
