package http

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eapache/go-resiliency/breaker"
	"golang.org/x/time/rate"
)

var (
	ErrNotFound        = errors.New("user not found")
	ErrTooManyRequests = errors.New("too many requests")
	ErrUnknownError    = errors.New("unknown error")
	ErrCircuitOpen     = errors.New("circuit open")
)

type RateLimiter interface {
	Wait(ctx context.Context) (err error)
}

// dummy rate limiter, used when no limits are specified.
type dummyLimiter int

func (n *dummyLimiter) Wait(ctx context.Context) (err error) {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

type Options struct {
	URL     string
	Method  string
	Body    io.Reader
	Headers map[string]string
	MaxRPM  int
}

type Client struct {
	options     Options
	ratelimiter RateLimiter
	client      *http.Client
	cb          *breaker.Breaker
}

func New(opt *Options) *Client {
	v := &Client{
		options: *opt,
		cb:      breaker.New(3, 1, 5*time.Second),
	}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}

	v.client = &http.Client{Transport: tr}

	if opt.MaxRPM > 0 {
		v.ratelimiter = rate.NewLimiter(rate.Every(time.Minute/time.Duration(opt.MaxRPM)), 1)
	} else {
		v.ratelimiter = new(dummyLimiter)
	}

	return v
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// This is a blocking call only if limit is defined, to honor the rate limit
	err := c.ratelimiter.Wait(ctx)
	if err != nil {
		return nil, err
	}

	var resp *http.Response

	// do the http request inside the CB scope
	err = c.cb.Run(func() error {
		resp, err = c.client.Do(req)

		return err
	})

	// check for cb errors, and replace error
	if err != nil {
		if errors.Is(err, breaker.ErrBreakerOpen) {
			err = ErrCircuitOpen
		}

		return nil, err
	}

	return resp, nil
}

func (c *Client) getBody(resp *http.Response) (string, error) {
	switch resp.StatusCode {

	case http.StatusNotFound:
		return "", ErrNotFound

	case http.StatusTooManyRequests:
		return "", ErrTooManyRequests

	case http.StatusOK:
		scanner := bufio.NewScanner(resp.Body)
		var response string
		for scanner.Scan() {
			response += scanner.Text()
		}

		return response, nil

	default:
		scanner := bufio.NewScanner(resp.Body)
		var er string
		for scanner.Scan() {
			er += scanner.Text()
		}

		return er, ErrUnknownError
	}
}

func (c *Client) Execute(ctx context.Context, method, url string, body io.Reader, headers map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, c.options.Method, c.options.URL, c.options.Body)
	if err != nil {
		return "", err
	}

	for key, value := range c.options.Headers {
		req.Header.Add(key, value)
	}

	var res *http.Response
	res, err = c.do(ctx, req)

	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	responseBody, err := c.getBody(res)
	if err != nil {
		return "", fmt.Errorf("error status: %d; %s", res.StatusCode, responseBody)
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error status: %d %s; error: %s", res.StatusCode, c.options.Body, err.Error())
	}

	return responseBody, nil
}
