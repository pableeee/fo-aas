package http

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrTooManyRequests = errors.New("too many requests")
)

type Options struct {
	MaxIdleConns        int
	MaxConnsPerHost     int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	client httpClient
}

func New(opt *Options, logger *log.Logger) *Client {
	return &Client{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        opt.MaxIdleConns,
				MaxConnsPerHost:     opt.MaxConnsPerHost,
				MaxIdleConnsPerHost: opt.MaxIdleConnsPerHost,
				IdleConnTimeout:     opt.IdleConnTimeout,
			},
		},
	}
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	resp, err = c.client.Do(req)

	// check for cb errors, and replace error to avoid leaking impl details.
	if err != nil {
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
	// client trace to log whether the request's underlying tcp connection was re-used
	// clientTrace := &httptrace.ClientTrace{
	// 	GotConn: func(info httptrace.GotConnInfo) {
	// 		log.Debugf("conn was reused: %t", info.Reused)
	// 	},
	// }
	// ctx = httptrace.WithClientTrace(ctx, clientTrace)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		log.Infof("request failed: %s", err.Error())

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
		log.Infof("error reading response body: %s", err.Error())

		return "", fmt.Errorf("error status: %d; %s", res.StatusCode, responseBody)
	}

	if res.StatusCode != http.StatusOK {
		err := &Error{Status: res.Status, Code: res.StatusCode, Body: responseBody}
		log.Infof("request returned an error: %+v", err)

		return "", fmt.Errorf("request failed %w", err)
	}

	return responseBody, nil
}
