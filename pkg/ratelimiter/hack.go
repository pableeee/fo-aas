package ratelimiter

import (
	"time"

	"github.com/stretchr/testify/mock"
)

// Clock abstracts away the os Now function, to enable mocking for testing purposes.
type clock interface {
	Now() *time.Time
}

type defaultClock int

func (d *defaultClock) Now() *time.Time {
	t := time.Now()

	return &t
}

type clockMock struct {
	mock.Mock
}

func (c *clockMock) Now() *time.Time {
	args := c.Called()

	t, _ := args.Get(0).(*time.Time)

	return t
}
