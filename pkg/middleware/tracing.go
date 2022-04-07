package middleware

import (
	"net/http"

	"github.com/opentracing/opentracing-go"
)

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span, ctx := opentracing.StartSpanFromContext(
			r.Context(),
			r.Method,
		)
		defer span.Finish()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
