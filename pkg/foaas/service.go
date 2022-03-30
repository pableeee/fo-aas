package foaas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	http_client "github.com/pableeee/fo-aas/pkg/http"
)

const (
	foasURL = "https://foaas.com/fascinating/%s"
)

var headers = map[string]string{"Accept": "application/json"}

type Payload struct {
	Message  string `json:"message"`
	Subtitle string `json:"subtitle"`
}

type httpClient interface {
	Execute(ctx context.Context, method, url string, body io.Reader, headers map[string]string) (string, error)
}

type Service struct {
	client httpClient
}

func New() *Service {
	return &Service{
		client: http_client.New(&http_client.Options{
			// no outgoing client rate limit to enforce
			MaxRPM: 0,
		}),
	}
}

func (s *Service) Get(ctx context.Context, user string) (*Payload, error) {
	var msg Payload

	url := fmt.Sprintf(foasURL, user)
	res, err := s.client.Execute(ctx, http.MethodGet, url, nil, headers)

	if err != nil {
		return nil, fmt.Errorf("unable to complete http request %w", err)
	}

	if err = json.Unmarshal([]byte(res), &msg); err != nil {
		return nil, fmt.Errorf("unable to unmarshall reponse payload %w", err)
	}

	return &msg, nil
}
