package handlers

import (
	"context"
	"encoding/json"

	"net/http"

	"github.com/pableeee/fo-aas/pkg/foaas"
	log "github.com/sirupsen/logrus"
)

type MessageHandler struct {
	svc    Service
	logger *log.Logger
}

type Service interface {
	Get(ctx context.Context, user string) (*foaas.Payload, error)
}

func NewMessageHandler(logger *log.Logger) *MessageHandler {
	return &MessageHandler{
		svc:    foaas.New(logger),
		logger: logger,
	}
}

// NewHandler returns a HandlerFunc that proxies the request to foaas
func (m *MessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := r.Header["User"]

	res, err := m.svc.Get(r.Context(), user[0])
	if err != nil {
		// upstream request failed
		m.logger.Infof("foaas service failed %s", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)

		return
	}

	body, err := json.Marshal(res)
	if err != nil {
		m.logger.Errorf("error masharling response %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
