package ratelimiter

import (
	"fmt"
	"sync"
	"time"
)

type tokenCount struct {
	// when the user was first seen
	firstSeen *time.Time
	// when the user was first seen
	expires time.Time
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

func New(rps int) *TokenRateLimiter {
	return &TokenRateLimiter{
		mutex: &sync.Mutex{},
		users: make(map[string]tokenCount),
		//every: time.Duration(time.Second / time.Duration(rps)),
		every: time.Duration(time.Second),
		limit: rps,
		clk:   new(defaultClock),
	}
}

// Allow enforces the the rate limit, by consuming tokens
func (t *TokenRateLimiter) Allow(user string) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	now := t.clk.Now()
	count, found := t.users[user]

	if !found || now.Sub(*count.firstSeen) >= t.every {
		// since the session expired, we refresh the tokens
		// minus this current request.
		t.users[user] = tokenCount{
			firstSeen: now,
			tokens:    t.limit - 1,
		}

		return true
	}

	fmt.Printf("first seen %s", count.firstSeen.String())
	since := now.Sub(*count.firstSeen)
	fmt.Printf("%d > %d :%+v \n", since, t.every, now.Sub(*count.firstSeen) > t.every)

	if count.tokens == 0 {
		return false
	}

	t.users[user] = tokenCount{
		firstSeen: count.firstSeen,
		tokens:    count.tokens - 1,
	}

	return true
}
