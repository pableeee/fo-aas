package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pableeee/fo-aas/pkg/foaas"
)

type MessageHandler struct {
	svc Service
}

type Service interface {
	Get(ctx context.Context, user string) (*foaas.Payload, error)
}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{
		svc: foaas.New(),
	}
}

// NewHandler returns a HandlerFunc that proxies the request to foaas
func (m *MessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user := r.Header["User"]

	res, err := m.svc.Get(r.Context(), user[0])
	if err != nil {
		// upstream request failed
		w.WriteHeader(http.StatusServiceUnavailable)

		return
	}

	body, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
