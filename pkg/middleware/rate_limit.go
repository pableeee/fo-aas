package middleware

import (
	"net/http"

	"github.com/pableeee/fo-aas/pkg/ratelimiter"
)

type RateLimiter interface {
	Allow(user string) bool
}

type RateLimiterMiddleware struct {
	limiter RateLimiter
}

// NewRateLimiterMiddleware creates a new http middleware that honors the provided limit.
func NewRateLimiterMiddleware(maxRMP int) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: ratelimiter.New(maxRMP),
	}
}

func (p *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		user, found := values["user"]
		if !found {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		if !p.limiter.Allow(user[0]) {
			w.WriteHeader(http.StatusTooManyRequests)

			return
		}

		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}
