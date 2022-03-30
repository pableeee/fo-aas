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
)

var (
	ErrNotFound        = errors.New("not found")
	ErrTooManyRequests = errors.New("too many requests")
)

type Options struct {
	MaxRPM int
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	options Options
	client  httpClient
	cb      *breaker.Breaker
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

	return v
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// do the http request inside the CB scope
	err = c.cb.Run(func() error {
		resp, err = c.client.Do(req)

		return err
	})

	// check for cb errors, and replace error to avoid leaking impl details.
	if err != nil {
		if errors.Is(err, breaker.ErrBreakerOpen) {
			// TODO: track circuit open metrics
		}

		return nil, fmt.Errorf("request failed %w", err)
	}

	return resp, nil
}

func (c *Client) getBody(body io.ReadCloser) (string, error) {
	scanner := bufio.NewScanner(body)
	var response string
	for scanner.Scan() {
		response += scanner.Text()
	}

	return response, nil
}

func (c *Client) Execute(ctx context.Context, method, url string, body io.Reader, headers map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return "", err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	var res *http.Response
	res, err = c.do(ctx, req)

	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	responseBody, err := c.getBody(res.Body)
	if err != nil {
		return "", fmt.Errorf("error status: %d; %s", res.StatusCode, responseBody)
	}

	if res.StatusCode != http.StatusOK {
		err := &Error{Status: res.Status, Code: res.StatusCode, Body: responseBody}
		return "", fmt.Errorf("request failed %w", err)
	}

	return responseBody, nil
}
