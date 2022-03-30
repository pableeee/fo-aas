package foaas

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type clientMock struct {
	mock.Mock
}

func (c *clientMock) Execute(ctx context.Context, method, url string, body io.Reader, headers map[string]string) (string, error) {
	args := c.Called()

	return args.String(0), args.Error(1)
}

func Test_FoaasOK(t *testing.T) {
	svc := New()
	m := &clientMock{}
	svc.client = m
	m.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(`{"message":"a message","subtitle":"a subtitle"}`, nil)

	res, err := svc.Get(context.Background(), "pable")
	assert.Nil(t, err)

	assert.Equal(t, "a message", res.Message)
	assert.Equal(t, "a subtitle", res.Subtitle)
}

func Test_FoaasFailToParse(t *testing.T) {
	svc := New()
	m := &clientMock{}
	svc.client = m
	m.On("Execute", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(`{"message":"a message"`, nil)

	res, err := svc.Get(context.Background(), "pable")
	assert.NotNil(t, err)
	assert.Nil(t, res)

}
