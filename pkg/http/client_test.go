package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type clientMock struct {
	mock.Mock
}

func (c *clientMock) Do(req *http.Request) (*http.Response, error) {
	args := c.Called()

	res, ok := args.Get(0).(*http.Response)
	if !ok {
		return nil, args.Error(1)
	}

	return res, args.Error(1)
}

func Test_HostUnavailable(t *testing.T) {
	m := &clientMock{}
	cl := New(&Options{}, logrus.New())
	cl.client = m

	m.On("Do", mock.Anything).
		Return("", fmt.Errorf("some error"))

	res, err := cl.Execute(context.Background(), http.MethodGet, "/ping", nil, nil)
	assert.Equal(t, "", res)
	assert.NotNil(t, err)
}

func Test_Available(t *testing.T) {
	m := &clientMock{}
	cl := New(&Options{}, logrus.New())
	cl.client = m

	pingResponse := `{"status":"ok","code":200}`

	body := io.NopCloser(strings.NewReader(pingResponse))

	m.On("Do", mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Status:     http.StatusText(http.StatusOK),
			Body:       body,
		}, nil)

	res, err := cl.Execute(context.Background(), http.MethodGet, "/ping", nil, nil)
	assert.Equal(t, pingResponse, res)
	assert.Nil(t, err)
}

func Test_RequestFail(t *testing.T) {
	m := &clientMock{}
	cl := New(&Options{}, logrus.New())
	cl.client = m

	pingResponse := `{"status":"StatusInternalServerError","code":500}`

	body := io.NopCloser(strings.NewReader(pingResponse))

	m.On("Do", mock.Anything).
		Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Status:     http.StatusText(http.StatusInternalServerError),
			Body:       body,
		}, nil)

	res, err := cl.Execute(context.Background(), http.MethodGet, "/ping", nil, nil)
	assert.Equal(t, "", res)
	assert.NotNil(t, err)

	var httpError *Error
	if errors.As(err, &httpError) {
		assert.Equal(t, http.StatusInternalServerError, httpError.Code)
		assert.Equal(t, http.StatusText(http.StatusInternalServerError), httpError.Status)
		assert.Equal(t, pingResponse, httpError.Body)
	}

}
