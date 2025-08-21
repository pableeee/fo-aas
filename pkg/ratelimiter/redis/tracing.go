package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type tracingMiddleware struct {
	delegate redisClient
}

func newTracingMiddleware(c redisClient) *tracingMiddleware {
	return &tracingMiddleware{
		delegate: c,
	}
}

func spanFromContext(ctx context.Context, operation, key string, expiration time.Duration) (opentracing.Span, context.Context) {
	span, ctx := opentracing.StartSpanFromContext(
		ctx,
		operation,
	)

	span.LogFields(
		log.String("key", key),
		//log.String("value",value),
		log.Int("expiration", int(expiration)),
	)

	return span, ctx
}

func (t *tracingMiddleware) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	span, ctx := spanFromContext(ctx, "SETNX", key, expiration)
	defer span.Finish()

	res := t.delegate.SetNX(ctx, key, value, expiration)
	failed := false        
	if _, err := res.Result(); err != nil {
		failed = true
	}
	span.SetTag("error", failed)

	return res
}

func (t *tracingMiddleware) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	span, ctx := spanFromContext(ctx, "SET", key, expiration)
	defer span.Finish()

	res := t.delegate.Set(ctx, key, value, expiration)
	failed := false        
	if _, err := res.Result(); err != nil {
		failed = true
	}
	span.SetTag("error", failed)

	return res
}

func (t *tracingMiddleware) Decr(ctx context.Context, key string) *redis.IntCmd {
	span, ctx := spanFromContext(ctx, "DECR", key, time.Duration(0))
	defer span.Finish()

	res := t.delegate.Decr(ctx, key)
	failed := false        
	if _, err := res.Result(); err != nil {
		failed = true
	}
	span.SetTag("error", failed)

	return res
}
