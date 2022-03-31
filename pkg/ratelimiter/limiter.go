package ratelimiter

import (
	"sync"
	"time"
)

type tokenCount struct {
	// when the user was first seen
	firstSeen *time.Time
	// amount of tokens left
	tokens int
}

type TokenRateLimiter struct {
	mutex *sync.Mutex
	users map[string]tokenCount
	every time.Duration
	limit int
	clk   clock
}

func New(tokens int, every time.Duration) *TokenRateLimiter {
	return &TokenRateLimiter{
		mutex: &sync.Mutex{},
		users: make(map[string]tokenCount),
		every: every,
		limit: tokens,
		clk:   new(defaultClock),
	}
}

// Allow enforces a rate limit of N tokens per session.
func (t *TokenRateLimiter) Allow(user string) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	now := t.clk.Now()
	count, found := t.users[user]

	if !found || now.Sub(*count.firstSeen) >= t.every {
		// since the session expired (or didn't exist), we refresh the tokens
		// minus this current request.
		t.users[user] = tokenCount{
			firstSeen: now,
			tokens:    t.limit - 1,
		}

		return true
	}

	if count.tokens == 0 {
		return false
	}

	t.users[user] = tokenCount{
		firstSeen: count.firstSeen,
		tokens:    count.tokens - 1,
	}

	return true
}
