package tracer

import (
	"net/http"
)

// TracingMiddleware extracts or generates trace details from a request.
// If the request is an external one, a new trace is generated.
func TraceMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			propagator := ExtractFromRequest(r)
			request, writer := propagator.Propagate(r, w)
			next.ServeHTTP(writer, request)
		})
	}
}
