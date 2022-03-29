package middleware

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter interface {
	Allow() bool
}

type RateLimiterMiddleware struct {
	limiter RateLimiter
}

// NewRateLimiterMiddleware creates a new http middleware that honors the provided limit.
func NewRateLimiterMiddleware(maxRMP int) *RateLimiterMiddleware {
	return &RateLimiterMiddleware{
		limiter: rate.NewLimiter(rate.Every(time.Minute/time.Duration(maxRMP)), 1),
	}
}

func (p *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !p.limiter.Allow() {
			w.WriteHeader(http.StatusTooManyRequests)

			return
		}

		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}
