package foaas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	http_client "github.com/pableeee/fo-aas/pkg/http"
	log "github.com/sirupsen/logrus"
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
	logger *log.Logger
}

func New(logger *log.Logger) *Service {
	return &Service{
		client: http_client.New(&http_client.Options{
			// Transport config
			MaxIdleConns:        100,
			MaxConnsPerHost:     1000,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     5 * time.Second,
		}, logger),
		logger: logger,
	}
}

func (s *Service) Get(ctx context.Context, user string) (*Payload, error) {
	var msg Payload

	url := fmt.Sprintf(foasURL, user)
	res, err := s.client.Execute(ctx, http.MethodGet, url, nil, headers)

	if err != nil {
		s.logger.Infof("unable to complete http request %s", err.Error())

		return nil, fmt.Errorf("unable to complete http request %w", err)
	}

	if err = json.Unmarshal([]byte(res), &msg); err != nil {
		s.logger.Infof("unable to unmarshall reponse payload %s", err.Error())

		return nil, fmt.Errorf("unable to unmarshall reponse payload %w", err)
	}

	return &msg, nil
}
