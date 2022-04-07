package foaas

import (
	"context"
	"io"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type tracingMiddleware struct {
	delegate httpClient
}

func newTracingMiddleware(c httpClient) *tracingMiddleware {
	return &tracingMiddleware{
		delegate: c,
	}
}

func (t *tracingMiddleware) Execute(ctx context.Context, method, url string, body io.Reader, headers map[string]string) (string, error) {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		method,
	)
	defer span.Finish()

	span.LogFields(
		log.String("url", url),
	)

	for k, v := range headers {
		span.LogFields(
			log.String(k, v),
		)
	}

	res, err := t.delegate.Execute(ctx, method, url, body, headers)
	if err != nil {
		span.SetTag("error", true)
		span.LogFields(
			log.String("error-message", err.Error()),
		)
	}

	return res, err
}
