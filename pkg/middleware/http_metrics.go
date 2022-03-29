package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/felixge/httpsnoop"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/gorilla/mux"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

var invalidChars = regexp.MustCompile(`[^a-zA-Z0-9]+`)

type MetricsMiddleware struct {
	Histogram metrics.Histogram
	Counter   metrics.Counter
	host      string
}

func PrometheusMetricsMiddleware(host string) *MetricsMiddleware {
	// used for monitoring and alerting (RED method)
	histogram := prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "Seconds spent serving HTTP requests.",
		Buckets:   stdprometheus.DefBuckets,
	}, []string{"method", "path", "status", "hostname"})
	// used for horizontal pod auto-scaling (Kubernetes HPA v2)
	counter := prometheus.NewCounterFrom(
		stdprometheus.CounterOpts{
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "The total number of HTTP requests.",
		},
		[]string{"status", "hostname"},
	)

	return &MetricsMiddleware{
		host:      host,
		Histogram: histogram,
		Counter:   counter,
	}
}

func (p *MetricsMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			path    = p.getRouteName(r)
			metrics = httpsnoop.CaptureMetrics(next, w, r)
			status  = strconv.Itoa(metrics.Code)
		)
		p.Histogram.With(r.Method, path, status, p.host).Observe(metrics.Duration.Seconds())
		p.Counter.With(status, p.host).Add(1)
	})
}

func (m *MetricsMiddleware) getRouteName(r *http.Request) string {
	if mux.CurrentRoute(r) != nil {
		if name := mux.CurrentRoute(r).GetName(); len(name) > 0 {
			return urlToLabel(name)
		}
		if path, err := mux.CurrentRoute(r).GetPathTemplate(); err == nil {
			if len(path) > 0 {
				return urlToLabel(path)
			}
		}
	}

	return urlToLabel(r.RequestURI)
}

func urlToLabel(path string) string {
	result := invalidChars.ReplaceAllString(path, "_")
	result = strings.ToLower(strings.Trim(result, "_"))
	if result == "" {
		result = "root"
	}

	return result
}
