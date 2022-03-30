package ratelimiter

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const format = "2006-01-02T15:04:05.000Z"

func parseDate(date string) *time.Time {
	t1, err := time.Parse(
		format,
		date)

	if err != nil {
		panic(0)
	}

	return &t1
}

func mockClock(limiter *TokenRateLimiter, date string) {
	m := &clockMock{}
	limiter.clk = m
	m.On("Now").Return(parseDate(date))
}

func TestRateLimiter_FirstRequest(t *testing.T) {
	// 5 requests per second limit
	limiter := New(10)

	mockClock(limiter, "2000-01-01T00:00:00.000Z")

	// first requests from any user should always be allowed.
	assert.True(t, limiter.Allow("pable"))
	assert.True(t, limiter.Allow("fable"))
	assert.True(t, limiter.Allow("table"))
	assert.True(t, limiter.Allow("mable"))
}

func TestRateLimiter_LimitReachead(t *testing.T) {
	// 10 requests per second limit
	limiter := New(10)

	// a session for a user should allow up to 10 tokens.
	// "2000-01-01T00:00:00.100Z" being when the user was first seen.
	for i := 10; i < 20; i++ {
		mockClock(limiter, fmt.Sprintf("2000-01-01T00:00:00.%d0Z", i))
		assert.True(t, limiter.Allow("pable"))
	}

	// the 11th token withing the same second, musn't be allowed.
	mockClock(limiter, "2000-01-01T00:00:00.210Z")
	// user is throttled
	assert.False(t, limiter.Allow("pable"))
}

func TestRateLimiter_ReachLimit_SameTS(t *testing.T) {
	limiter := New(10)

	// a limit of 10 tokens should be allowed on a second
	for i := 0; i < 10; i++ {
		mockClock(limiter, "2000-01-01T00:00:00.100Z")
		assert.True(t, limiter.Allow("pable"))
	}

	// the 11th token withing the same second, musn't be allowed.
	mockClock(limiter, "2000-01-01T00:00:00.200Z")
	// user is throttled
	assert.False(t, limiter.Allow("pable"))
}

func TestRateLimiter_AllowedAfterBlocked(t *testing.T) {
	// 10 requests per second limit
	limiter := New(10)

	// a limit of 10 tokens should be allowed on a second
	for i := 0; i < 10; i++ {
		date := fmt.Sprintf("2000-01-01T00:00:00.%d00Z", i)
		mockClock(limiter, date)

		assert.True(t, limiter.Allow("pable"))
	}

	// the 11th token withing the same second, musn't be allowed.
	mockClock(limiter, "2000-01-01T00:00:00.900Z")
	// user is throttled
	assert.False(t, limiter.Allow("pable"))

	// one second after the first event, the session shoud have expired.
	// up to 10 tokens more have to be availble again
	for i := 0; i < 10; i++ {
		date := fmt.Sprintf("2000-01-01T00:00:01.%d00Z", i)
		mockClock(limiter, date)

		assert.True(t, limiter.Allow("pable"))
	}
}

func TestRateLimiter_ReachLimit_AllowedOtherUser(t *testing.T) {
	// 10 requests per second limit
	limiter := New(10)

	// a limit of 10 tokens should be allowed on a second
	for i := 0; i < 10; i++ {
		date := fmt.Sprintf("2000-01-01T00:00:00.%d00Z", i)
		mockClock(limiter, date)

		assert.True(t, limiter.Allow("pable"))
	}

	// the 11th token withing the same second, musn't be allowed.
	mockClock(limiter, "2000-01-01T00:00:00.900Z")
	// user is throttled
	assert.False(t, limiter.Allow("pable"))

	// since the rate limiting is user independent, 'pable' being throttled
	// musn't affect user 'jon'
	for i := 0; i < 10; i++ {
		assert.True(t, limiter.Allow("jon"))
	}

}
