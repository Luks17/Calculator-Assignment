# Calculator

A responsive, GNOME-inspired calculator built with React, TypeScript, Vite,
TanStack Query, Tailwind CSS, daisyUI and Lucide icons. Uses daisyUI's built-in `light` and `dark` themes and
component classes with Tailwind utilities, with no custom CSS rules or theme
overrides. Includes persistent theme selection, keyboard input,
editable expressions and up to 20 recent calculations for the current session.

## Development

Requires Node.js 24+ and pnpm 12.

```sh
pnpm install
pnpm dev
pnpm test
pnpm test:watch
pnpm lint
pnpm build
```

The repository's Docker workflow also supports `make test APP=calc-web`,
`make build APP=calc-web` and `make lint APP=calc-web`.

## HTTP client generation

The shared contract at `../../contracts/calc/openapi/schema.yml` is the source
for the Hey API ofetch client, calculation SDK and TypeScript types. Regenerate
them with `pnpm generate` from this directory, or `make generate APP=calc-web`
from the repository root using Docker.

Generation is configured in `openapi-ts.config.ts`. Output is checked in under
`src/shared/api/generated/` and formatted with Prettier. Do not edit generated
files directly; regeneration replaces the directory.

## API integration

Calculations use the generated client to call `POST /api/v1/calculations`.
The API defaults to `http://localhost:8080`; set `VITE_API_BASE_URL` in `.env`
to use a different browser-accessible URL, then restart Vite (or rebuild for
production). Run `calc-service` with `APP_ENV=local`, or allow the frontend's
origin through `ALLOWED_ORIGINS`, so the browser can access the API.

The frontend translates `−`, `×`, `÷` and `√(...)` to the API's `-`, `*`, `/`
and `sqrt(...)` notation. All operands, expressions and results stay strings;
only the backend parses and evaluates arithmetic. Results preserve the API's
decimal precision.

TanStack Query mutations own request status and errors; both keyboard and keypad
submissions wait for the API result. Repeated submits
are ignored while a request is pending, and editing or clearing the expression
prevents a late response from replacing it. Failed calculations leave the
expression and history unchanged. API error messages (or a fallback for network
errors) appear in a dismissible toast that closes automatically after six seconds.

## Calculator behavior

Supported input: addition, subtraction, multiplication, division, exponentiation,
square root, percentage, parentheses, decimal numbers and sign changes.

Type directly into the expression field, use the on-screen keys, or use your
keyboard. `Enter` or `=` submits, `Escape` clears, and `Backspace` deletes.
The display shows a single expression. After calculation, the result replaces
that expression; history retains the original expression and its result.
`*` and `/` map to multiplication and division; `^` is exponentiation;
`R` inserts a square root. Standard text editing shortcuts work in the input.
History entries can be clicked to edit and reuse their expression. History is
not persisted; the theme preference is the only value stored locally.

## Structure

```text
src/
  app/                  # Composition, Query provider and global themed styles
  pages/calculator/     # Calculator page layout
  features/calculate/   # API integration, string-based model, keypad and history
  shared/api/           # Public API boundary and generated Hey API client
  shared/theme/         # Persistent theme switch
  main.tsx
```

Pages consume features through their public `index.ts` entry points.
Within `features/calculate`, `api/` adapts the generated SDK and defines mutation
options, `model/` owns the reducer and mutation workflow, and `ui/` renders the
calculator and error toast. The Query provider lives in `app/providers/`.
Tests use Vitest. `pnpm test` runs the suite once; `pnpm test:watch` watches for changes.
