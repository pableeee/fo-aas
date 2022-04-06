package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/pableeee/fo-aas/pkg/ratelimiter"
)

type RateLimiter interface {
	Allow(ctx context.Context, user string) bool
}

type RateLimiterMiddleware struct {
	limiter RateLimiter
}

// NewRateLimiterMiddleware creates a new http middleware that honors the provided limit.
func NewRateLimiterMiddleware(tokens int, every time.Duration) *RateLimiterMiddleware {
	lim := ratelimiter.New(tokens, every)
	return &RateLimiterMiddleware{
		limiter: ratelimiter.MewMetricMiddleware(lim),
	}
}

func (p *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := r.Header
		user, found := values["User"]
		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(500)*time.Millisecond)
		defer cancel()

		if !found {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		if !p.limiter.Allow(ctx, user[0]) {
			w.WriteHeader(http.StatusTooManyRequests)

			return
		}

		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}
