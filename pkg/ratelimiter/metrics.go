package ratelimiter

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type allower interface {
	Allow(ctx context.Context, user string) bool
}

type MetricsMiddleware struct {
	histogram metrics.Histogram
	next      allower
}

func MewMetricMiddleware(next allower) *MetricsMiddleware {
	histogram := prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Subsystem: "rate_limiter",
		Name:      "allow_duration_seconds",
		Help:      "Seconds spent quering the rate limiter",
		Buckets:   stdprometheus.DefBuckets,
	}, []string{})

	return &MetricsMiddleware{histogram: histogram, next: next}
}

func (m *MetricsMiddleware) Allow(ctx context.Context, user string) bool {
	begin := time.Now()

	defer func() {
		took := time.Since(begin)
		m.histogram.Observe(took.Seconds())
	}()

	return m.next.Allow(ctx, user)
}
