# Calculator

A responsive, GNOME-inspired calculator built with React, TypeScript, Vite,
Tailwind CSS, daisyUI and Lucide icons. Uses daisyUI's built-in `light` and `dark` themes and
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
files directly; regeneration replaces the directory. The client is not yet
connected to the calculator UI; the mock behavior below is unchanged.

## Preview behavior

**Every submitted expression returns the string `"1"`.** There is no network
integration or client-side arithmetic evaluator. All operands, expressions and
results remain strings, including long decimals and trailing zeros. The
`features/calculate/api/calculate.ts` function is the mock boundary to replace
when `POST /api/v1/calculations` is integrated. Backend expression syntax and
transport types should be confirmed against the API contract at that point.

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
  app/                  # Application composition and global themed styles
  pages/calculator/     # Calculator page layout
  features/calculate/   # Mock boundary, string-based model, keypad and history
  shared/api/generated/ # Hey API client, SDK and types from the shared contract
  shared/theme/         # Persistent theme switch
  main.tsx
```

Pages consume features through their public `index.ts` entry points.
Tests use Vitest. `pnpm test` runs the suite once; `pnpm test:watch` watches for changes.
