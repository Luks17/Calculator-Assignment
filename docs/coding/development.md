# Setup and development

Commands below run from the repository root unless stated otherwise.
You can run everything in Docker, use local tools, or mix the two for different tasks.

## Docker setup

Install Docker with Compose v2.24+ and Make. No local Go, Node.js, or pnpm installation is needed.

```sh
make env
make dev-up
```

`make env` copies each app's `.env.example` to `.env` and refuses to overwrite existing files.
Skip it if both files exist. If only one is missing, run that app's `env` target,
for example `make -C apps/calc-web env`.

`make dev-up` runs `docker compose up -d`. The first start builds the images and
downloads dependencies, which can take a few minutes. The containers use bind-mounted
source code: wgo/Delve restarts the backend on Go changes, and Vite reloads the frontend.

| Service | Default address |
| --- | --- |
| Frontend | <http://localhost:5173> |
| API | `http://localhost:8080` (`GET /healthz` for a health check) |
| OpenAPI reference (Scalar) | <http://localhost:8081> |
| Backend debugger (Delve) | `localhost:2345` |

Useful commands:

```sh
docker compose ps
docker compose logs -f calc-service calc-web
```

`make dev-down` stops and removes the containers but preserves dependency caches.
After Dockerfile changes, rebuild with `docker compose up -d --build`.
These are development containers, not a production deployment; do not expose the debugger publicly.

### Docker port overrides

Set these in your shell or a repository-root `.env` file (separate from the app `.env` files):

| Variable | Default | Host port for |
| --- | --- | --- |
| `CALC_WEB_PORT` | `5173` | Frontend |
| `CALC_SERVICE_PORT` | `8080` | API |
| `API_REFERENCE_PORT` | `8081` | Scalar |
| `CALC_SERVICE_DEBUG_PORT` | `2345` | Delve |

For example:

```sh
CALC_WEB_PORT=5174 make dev-up
```

These change host ports only. Keep `SERVER_PORT=8080` inside the backend container
so the port mapping and health check still work. If you change the API's host port,
also update `VITE_API_BASE_URL` in `apps/calc-web/.env` to the browser-accessible URL
and recreate the frontend with `docker compose up -d calc-web`.

## Run without Docker

Use separate terminals for the backend and frontend. Avoid running a local app and
its Docker container on the same port.

### Backend

Requires Go 1.27.3 or newer. From the repository root:

```sh
cd apps/calc-service
APP_ENV=local go run ./cmd/server
```

The API listens on `http://localhost:8080`. Dependencies are downloaded automatically.
Unlike Compose, `go run` does not load `.env` files. Export custom configuration
yourself, or load the copied file before running:

```sh
set -a
. ./.env
set +a
go run ./cmd/server
```

### Frontend

Requires Node.js 24+ and pnpm 12. In another terminal, from the repository root:

```sh
cd apps/calc-web
pnpm install --frozen-lockfile
pnpm dev
```

Open <http://localhost:5173>. Vite loads `apps/calc-web/.env` automatically;
without it, the generated client still defaults to `http://localhost:8080`.
To choose a local frontend port, use `pnpm dev --port 5174`. The `PORT` entry in
the example environment file is not read by the current Vite configuration.
Vite may choose the next free port if its default is busy; add `--strictPort` to fail instead.

The API reference is optional. To run just the viewer, use
`docker compose up -d api-reference`; without Docker, open the
[OpenAPI contract](../../contracts/calc/openapi/schema.yml) in your preferred OpenAPI viewer.

## Application configuration

Backend settings go in `apps/calc-service/.env` for Docker, or exported environment
variables for local execution:

| Variable | Default | Purpose |
| --- | --- | --- |
| `SERVER_HOST` | `0.0.0.0` | Listening interface |
| `SERVER_PORT` | `8080` | Listening port |
| `APP_ENV` | `production` | `local` allows any nonempty browser origin |
| `ALLOWED_ORIGINS` | Empty | Comma-separated exact origins outside local mode |

The supplied example uses `APP_ENV=local`. In other environments, cross-origin
access is denied unless the origin is listed, for example
`ALLOWED_ORIGINS=http://localhost:5173,http://localhost:8081`.
Entries are trimmed, but wildcards, subdomain patterns, and trailing slashes are not expanded.

Frontend settings go in `apps/calc-web/.env`:

| Variable | Default | Purpose |
| --- | --- | --- |
| `VITE_API_BASE_URL` | `http://localhost:8080` | API URL reached by the browser |

Use a browser-accessible hostname, not the Compose service name `calc-service`.
Restart Vite after changing frontend settings; production builds need to be rebuilt.
After changing app `.env` files in Docker, run `make dev-up` to recreate affected containers.

## Tests, builds, linting, and formatting

Root Make targets run in disposable Docker containers. They do not require the
development servers to be running. `APP` defaults to `calc-service`:

```sh
make test
make test APP=calc-web
```

To use installed local tools instead, call the app's Makefile directly:

```sh
make -C apps/calc-service test
make -C apps/calc-web test
```

The following targets work in either mode:

| Target | Backend | Frontend |
| --- | --- | --- |
| `test` | Go tests with the race detector | Vitest, single run |
| `build` | Compile Go packages | Type-check and build to `dist/` |
| `lint` | golangci-lint | ESLint |
| `format` | gofmt via golangci-lint | Prettier |
| `generate` | Generate the Go server from OpenAPI | Generate the TypeScript API client |

Local backend tests with `-race` require CGO and a C compiler. Local `lint` and
`format` require golangci-lint (the Docker image uses v2.15.0).
Frontend commands require `pnpm install` first. For watch mode, run
`pnpm test:watch` in `apps/calc-web`. You can also run `go vet ./...` in `apps/calc-service`.

### OpenAPI generation

[The shared contract](../../contracts/calc/openapi/schema.yml) is the source for the generated server and client.
After editing it, regenerate both sides:

```sh
make generate
make generate APP=calc-web
```

Or run `make generate` locally in each app directory. Generated files are checked
in, so generation is not required for the initial setup. Do not edit them by hand:

- Backend: `apps/calc-service/internal/transport/http/server/gen/`, generated by ogen; configured in `apps/calc-service/ogen.yml`.
- Frontend: `apps/calc-web/src/shared/api/generated/`, generated by Hey API; configured in `apps/calc-web/openapi-ts.config.ts`.

If frontend dependencies change while using Docker, update its dependency volume with
`docker compose run --rm --no-deps calc-web pnpm install --frozen-lockfile`.
