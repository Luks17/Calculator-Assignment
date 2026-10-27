# Design decisions and assumptions

I took the assignment's 2–4-hour recommendation seriously. I focused on the
calculator itself, reused familiar tools, and left out features and layers that
would not have added much value for this project.

## API contract and expression evaluation

I took a spec-first approach to the API, writing the OpenAPI contract and using it
to generate a type-safe frontend client and Go server bindings. This kept both
sides working from the same contract and saved me from writing that plumbing by
hand. I also exposed the contract through Scalar in Docker Compose, making it easy
to browse and try the API during local development.

I didn't like the idea of an endpoint for every arithmetic operation. I wanted an
actual expression calculator, not a UI limited to one operation at a time, and I
didn't want the frontend to parse expressions and coordinate requests for each
operation. An RPC-style API felt like a better fit for a calculator than a
resource-oriented REST API, so I used a single `POST /api/v1/calculations` endpoint
that accepted an expression and returned its result.

## Decimal arithmetic

One of my first concerns was avoiding binary floating-point errors. I used
`shopspring/decimal` for arithmetic and strings for expressions and results in the
API, so values would not lose precision through JavaScript numbers or JSON number
handling.

This was still a simple calculator, not a promise of exact results for every
possible calculation. Division, negative powers, and square roots rounded to 16
fractional decimal places per operation, which I considered generous for this
scope. The exact behavior and limits are documented in the [API guide](calc-service/api.md).

## UI and styling

The design was inspired by GNOME Calculator because I generally like how GNOME
applications are designed. I used Tailwind CSS and daisyUI to build the interface
quickly and get light/dark theme support without creating a custom theme system.

## Keeping the scope small

I considered a history endpoint with server-side persistence, but it would have
gone beyond what I wanted to spend the time on. Instead, I kept a simple
client-side history in memory. Persisting it in the browser would also have been
another feature, so I prioritized other work and accepted that history would be
lost on reload.

I also considered TanStack Start with a BFF layer, a pattern I use frequently and
like working with. For this application, though, it would not have brought enough
benefit to justify the extra layer within the time budget. A Vite frontend calling
the Go service directly was enough.

## Familiar structure, deliberately simplified

The initial commit contained boilerplate adapted from previous projects. I enjoy
working with this Dockerized monorepo setup: both applications live together, and
the development environment is straightforward to bring up.

On the frontend, I used a simplified feature-sliced structure. I left out folders
such as `entities` and `widgets` because this application did not need them.

For the Go service, I used a lightweight variation of hexagonal architecture and
DDD. I had used stricter versions before, including shared ports, separate use-case
layers, and CQRS, but deliberately left those out here. I wanted a familiar
structure that would be easy to maintain and grow while staying simple and
idiomatic to Go, rather than adding abstractions just to follow a pattern. The
guidance I followed is in [the Go structure notes](coding/structure/golang.md).

## Parser, logging, and errors

I used `alecthomas/participle` to build the expression parser. It handled much of
the tedious work of writing a parser from scratch and saved me time, while leaving
the arithmetic rules under my control. The implementation is described in the
[parser notes](calc-service/parser.md).

I chose `zerolog` because it is the Go logging library I most enjoy working with,
especially for contextual logging. I also centralized error handling in the HTTP
layer, a pattern I've grown to like: it kept response formatting and HTTP status
mapping in one place instead of spreading them across handlers.

## Unit tests

The assignment explicitly asked for unit tests. I used Vitest on the frontend and
Go's standard testing package with `testify` on the service. These were familiar
tools, and `testify`'s assertion helpers made the Go tests convenient to write and
read.
