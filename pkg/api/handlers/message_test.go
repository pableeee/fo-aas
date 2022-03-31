package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pableeee/fo-aas/pkg/foaas"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type serviceMock struct {
	mock.Mock
}

func (s *serviceMock) Get(ctx context.Context, user string) (*foaas.Payload, error) {
	args := s.Called()
	t, _ := args.Get(0).(*foaas.Payload)

	return t, args.Error(1)
}

func setup() (*httptest.ResponseRecorder, *http.Request) {
	rr := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(context.TODO(), "GET", "/message", nil)
	if err != nil {
		panic(0)
	}

	req.Header.Add("user", "pable")

	return rr, req
}

func Test_FoaasUnavailable(t *testing.T) {
	handler := NewMessageHandler(&Options{}, logrus.New())
	m := &serviceMock{}
	handler.svc = m

	m.On("Get", mock.Anything, mock.Anything).Return(
		nil, fmt.Errorf("some internal error"),
	)

	rec, req := setup()

	handler.ServeHTTP(rec, req)

	// Check the status code is what we expect.
	if status := rec.Code; status != http.StatusServiceUnavailable {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusServiceUnavailable)
	}

}

func Test_FoaasOK(t *testing.T) {
	handler := NewMessageHandler(&Options{}, logrus.New())
	m := &serviceMock{}
	handler.svc = m

	m.On("Get", mock.Anything, mock.Anything).Return(
		&foaas.Payload{Message: "a message", Subtitle: "a subtitle"},
		nil,
	)

	rec, req := setup()

	handler.ServeHTTP(rec, req)

	// Check the status code is what we expect.
	if status := rec.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusServiceUnavailable)
	}

	var msg foaas.Payload
	err := json.Unmarshal(rec.Body.Bytes(), &msg)
	assert.Nil(t, err)

	assert.Equal(t, msg.Message, "a message")
	assert.Equal(t, msg.Subtitle, "a subtitle")
}
