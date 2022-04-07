package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/pableeee/fo-aas/pkg/ratelimiter"
	"github.com/pableeee/fo-aas/pkg/ratelimiter/redis"
)

type RateLimiter interface {
	Allow(ctx context.Context, user string) bool
}

type RateLimiterMiddleware struct {
	limiter RateLimiter
}

// NewRateLimiterMiddleware creates a new http middleware that honors the provided limit.
func NewRateLimiterMiddleware(tokens int, every time.Duration) *RateLimiterMiddleware {
	lim, _ := redis.New(&redis.Option{
		Tokens:   tokens,
		Every:    every,
		Addr:     "localhost:6379",
		Password: "foobar",
	})
	return &RateLimiterMiddleware{
		limiter: ratelimiter.MewMetricMiddleware(lim),
	}
}

func (p *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := r.Header
		user, found := values["User"]
		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(500)*time.Second)
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
