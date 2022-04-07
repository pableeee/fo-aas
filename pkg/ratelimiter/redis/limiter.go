package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	sessionKey = "user:%s"
	tokensKey  = "tokens:%s"
)

type redisClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Decr(ctx context.Context, key string) *redis.IntCmd
}

type RedisTokenRateLimiter struct {
	client redisClient
	every  time.Duration
	limit  int
}

type Option struct {
	Tokens   int
	Every    time.Duration
	Addr     string
	Password string
}

func New(opt *Option) (*RedisTokenRateLimiter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     opt.Addr,
		Password: opt.Password, // no password set
		DB:       0,            // use default DB
	})

	if _, err := rdb.Ping(context.TODO()).Result(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &RedisTokenRateLimiter{
		client: newTracingMiddleware(rdb),
		every:  opt.Every,
		limit:  opt.Tokens,
	}, nil
}

// Allow enforces a rate limit of N tokens per session.
func (t *RedisTokenRateLimiter) Allow(ctx context.Context, user string) bool {
	session := fmt.Sprintf(sessionKey, user)
	tokens := fmt.Sprintf(tokensKey, user)
	// try to create the session, but setting the NX option.
	ok, err := t.client.SetNX(ctx, session, time.Now().String(), t.every).Result()
	if err != nil {
		// redis client failed
		return false
	}

	if ok {
		// there was no previous session, set to the limit -1.
		_, err = t.client.Set(ctx, tokens, t.limit-1, time.Duration(0)).Result()
		if err != nil {
			// redis client failed
			return false
		}

		return true
	}

	res, err := t.client.Decr(ctx, tokens).Result()
	if err != nil {
		// redis client failed
		return false
	}

	if res <= 0 {
		return false
	}

	return true
}
