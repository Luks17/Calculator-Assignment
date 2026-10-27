# Frontend notes

The responsive, GNOME-inspired calculator uses React, TypeScript, Vite, TanStack
Query, Tailwind CSS, daisyUI, and Lucide icons. Styling uses daisyUI's built-in
light/dark themes and component classes with Tailwind utilities, without custom
CSS rules or theme overrides.

See [development](../coding/development.md) for setup, configuration, tests, and client generation.

## Calculator behavior

- Type an expression, use the on-screen keypad, or use keyboard shortcuts.
- `Enter` or `=` submits; `Escape` clears; `Backspace` deletes. `*` and `/` enter
  multiplication and division, `^` enters a power, and `R` inserts a square root.
  Standard text-editing shortcuts work in the expression field.
- A successful result replaces the expression. History keeps the original expression
  and result for up to 20 recent calculations; click an entry to edit and reuse it.
- History is kept in memory and is lost on reload. Only the theme preference is persisted locally.

Supported arithmetic and percentage semantics are documented in [the API guide](../calc-service/api.md).

## API integration

Calculations call `POST /api/v1/calculations` through the generated Hey API/ofetch client.
Set `VITE_API_BASE_URL` to change the default API URL (`http://localhost:8080`);
the backend must allow the frontend's origin through local mode or `ALLOWED_ORIGINS`.

The frontend translates `−`, `×`, `÷`, and `√(...)` to `-`, `*`, `/`, and `sqrt(...)`.
Operands, expressions, and results remain strings. Only the backend parses and
evaluates arithmetic, so the frontend preserves the API's decimal precision.

TanStack Query mutations track request status and errors. Keyboard and keypad
submissions wait for the API result; repeated submissions are ignored while a
request is pending. Editing or clearing an expression prevents a late response
from replacing it. Failed calculations leave the expression and history unchanged.
API messages (or a fallback for network failures) appear in a dismissible toast
that also closes automatically after six seconds.

## Structure

Paths below are relative to `apps/calc-web/`:

```text
src/
  app/                  # Composition, Query provider, global themed styles
  pages/calculator/     # Calculator page layout
  features/calculate/   # API integration, string-based model, keypad, history
  shared/api/           # Public API boundary and generated Hey API client
  shared/theme/         # Persistent theme switch
  main.tsx
```

Pages consume features through public `index.ts` entry points. Inside
`features/calculate/`, `api/` adapts the SDK and defines mutation options,
`model/` owns the reducer and mutation workflow, and `ui/` renders the calculator
and error toast. The Query provider lives in `app/providers/`. Tests use Vitest.

Generation replaces `src/shared/api/generated/` and formats the output with Prettier.
Generated files are checked in and should not be edited manually.
