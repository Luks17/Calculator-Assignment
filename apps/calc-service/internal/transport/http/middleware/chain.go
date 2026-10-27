package middleware

import "net/http"

type Chain [](func(http.Handler) http.Handler)

func NewChain(middlewares ...func(http.Handler) http.Handler) Chain {
	return Chain(middlewares)
}

func (m Chain) Apply(handler http.Handler) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		handler = m[i](handler)
	}
	return handler
}
