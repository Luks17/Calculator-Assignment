# AI Prompts

I used GPT-6.1 Sol (xhigh) with OpenCode v2.1.4 for part of the work on this project.

## Prompts used for implementing Calc Web (frontend UI) - First iteration (no backend integration)

Initial prompt for it:

```
Create a beautiful and responsive calculator frontend based on Gnome Calculator at apps/calc-web, it should include support for a light and dark theme.

Use React as the framework and daisyui as your component library, both dependencies are already setup with vite.

Our supported operations will be:
- Addition
- Subtraction
- Multiplication
- Division
- Exponentiation
- Square root
- Percentage

This initial release will have no integration with the backend, display every expression result as '1' for now.

In the future, we will have a single endpoint on the backend: `POST /api/v1/calculations` which will accept an arithmetic expression. The backend will handle expression parsing and calculating the end result. No need to worry about this for now, but it's important to understand that numbers must be represented as strings to preserve precision though.

Follow a lightweight feature-slice design structure for this release, by 'lightweight' I mean that we have no need for an 'widgets' or 'entities' folder for this tiny project. We can start with 'app', 'pages', 'shared' and 'features' (if needed) folders. I'm thinking of something roughly like this:
src/
├── app/
│   ├── App.tsx
│   └── styles/
│       └── index.css
│
├── pages/
│   └── calculator/
│       ├── ui/
│       │   └── CalculatorPage.tsx
│       └── index.ts
│
├── features/
│   └── calculate/
│       ├── api/
│       │   └── calculate.ts
│       ├── model/
│       │   ├── calculator.reducer.ts
│       │   ├── calculator.reducer.test.ts
│       │   └── calculator.types.ts
│       ├── ui/
│       │   ├── Calculator.tsx
│       │   ├── CalculatorDisplay.tsx
│       │   └── CalculatorKeypad.tsx
│       └── index.ts
│
├── shared/
│   └── api/
│       ├── generated/
│       │   └── ...                 # OpenAPI-generated client
│       └── client.ts              # HTTP client configuration
│
└── main.tsx
```

I followed it up later with this:
```
Some things to improve:
- There is too much going on in the screen: 'All the essentials', 'basic', 'everyday math, without the distractions', 'keep things simple', etc. Keep the header and the actual calculator parts (including history and keyboard help). Remove everything else. Also cleanup unecessary text.
- Use lucide for icons instead of hardcoded svg.
- Install and use vitest for unit testing instead of node utilities
```

## Prompts using for implementing the Calc Service (backend API)

> It's important to mention that `docs/parser/design.md`, which is mentioned on the prompt, was rewritten and moved to `docs/calc-service/parser.md` during development. The original can be found by going back with Git.

Initial prompt for it:
```
Write a calculator microservice written in go at ./apps/calc-service.

The server was already generated using ogen based on contracts/calc/openapi/schema.yml. Follow the structure guidelines at docs/coding/structure/golang.md.

The main parts of this backend are:
- The expression parser, which is documented here: docs/parser/design.md.
- The expression evaluator: which must do the actual calculations

Use centralized error handling for the http server.
```

Following that prompt and some code iterations, I asked the following:
```
Okay, some structure worries I have:
- You are using 'config' directly in various parts instead of making the consumers use interfaces. Everthing from 'infrastructure' should not be imported directly from consumers.
- Instead of passing the logger around everywhere. I prefer using zerolog with contextual logging which can be injected on the request context as a middleware and retrieved or populated with zerolog helper functions. It also has hlog which contains useful http middleware helpers. I made this apps/calc-service/internal/transport/http/middleware/chain.go to help you chain middlewares and apps/calc-service/internal/transport/http/middleware/recover.go as an example of a middleware. Here is a logging middleware I used on a previous project for reference:
func Logging(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// wrap response writer to add status code
			ww := types.NewWrapResponseWriter(w)

			// attach logger to request context
			loggerCopy := logger.With().Logger()

			ctx := r.Context()

			reqID, err := common.GetReqIDContext(ctx)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				logger.Error().Err(fmt.Errorf("%w: %s", errs.ErrHTTPLoggingMiddleware, err.Error())).Send()
				return
			}

			ctx = loggerCopy.WithContext(ctx)
			r = r.WithContext(ctx)

			defer func() {
				l := zerolog.Ctx(r.Context())
				var ev *zerolog.Event
				if ww.StatusCode() < 400 {
					ev = l.Info()
				} else if ww.StatusCode() < 500 {
					ev = l.Warn()
				} else {
					ev = l.Error()
				}

				ev.Str("req_id", reqID.String()).
					Str("method", r.Method).
					Str("path", r.URL.Path).
					Str("ip", r.RemoteAddr).
					Str("user_agent", r.UserAgent()).
					Str("origin", r.Header.Get("Origin")).
					Dur("elapsed_ms", time.Since(start)).
					Int("status_code", ww.StatusCode()).
					Msg("incoming request")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
```

Finally, in another session, after noticing most tests written by the agent were super hard to read, I prompted:
```
Simplify golang tests by using stretchr/testify, make sure to use the latest version.

Also refactor any tests that could be table-driven, which is an important golang testing pattern.
```

## Prompts using for implementing the Calc Web UI (frontend) - Second iteration (with backend integration)

```
In cacl-web, I generated a http client for the api with hey-api. Now what's left is to stop returning a hardcoded 1 as every result and return an actual result. In the case of API errors, display it as a toast. Use tanstack query for handling server-side state.
```
