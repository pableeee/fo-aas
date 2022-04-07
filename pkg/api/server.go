package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pableeee/fo-aas/pkg/api/handlers"
	"github.com/pableeee/fo-aas/pkg/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	router  *mux.Router
	logger  *log.Logger
	config  *Config
	handler http.Handler
}

type Config struct {
	HTTPServerTimeout         time.Duration
	HTTPServerShutdownTimeout time.Duration
	Host                      string
	Port                      string
	Hostname                  string
	Tokens                    int
	Every                     time.Duration
	Timeout                   time.Duration
}

func NewServer(c *Config, l *log.Logger) *Server {
	return &Server{
		router: mux.NewRouter(),
		logger: l,
		config: c,
	}
}

func (s *Server) registerHandlers(ctx context.Context) {
	// observability endpoints.
	s.router.Handle("/metrics", promhttp.Handler())
	s.router.HandleFunc("/ping", s.pingHandler).Methods("GET")

	// domain router, to aviod tracking the operative endpoints above.
	domainSubrouter := s.router.PathPrefix("").Subrouter()

	s.registerMiddlewares(domainSubrouter)

	// register domain endpoints
	domainSubrouter.Handle("/message",
		handlers.NewMessageHandler(
			&handlers.Options{Timeout: s.config.Timeout}, s.logger,
		),
	).Methods("GET")
}

func (s *Server) JSONResponse(w http.ResponseWriter, r *http.Request, result interface{}, responseCode int) {
	body, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(responseCode)
	_, _ = w.Write(body)
}

func (s *Server) registerMiddlewares(router *mux.Router) {
	// logging middleware
	router.Use(
		func(h http.Handler) http.Handler {
			return gorilla_handlers.LoggingHandler(os.Stdout, h)
		},
	)

	// metrics middleware
	prom := middleware.PrometheusMetricsMiddleware(s.config.Hostname)
	router.Use(prom.Handler)

	// tracing middleware
	router.Use(
		func(h http.Handler) http.Handler {
			return middleware.TracingMiddleware(h)
		},
	)

	// rate limiting middleware
	limiter := middleware.NewRateLimiterMiddleware(s.config.Tokens, s.config.Every)
	router.Use(limiter.Handler)
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	// go s.startMetricsServer()

	s.registerHandlers(ctx)

	s.handler = s.router

	// create the http server
	srv := s.startServer()

	// wait for context to be done
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), s.config.HTTPServerShutdownTimeout)
	defer cancel()

	s.logger.Info("Shutting down HTTP/HTTPS server")

	// determine if the http server was started
	if srv != nil {
		if err := srv.Shutdown(ctx); err != nil {
			s.logger.Warn(fmt.Sprintf("HTTP server graceful shutdown failed %s", err.Error()))

			return err
		}
	}

	return nil
}

func (s *Server) startServer() *http.Server {
	// determine if the port is specified
	if s.config.Port == "0" {
		// move on immediately
		return nil
	}

	srv := &http.Server{
		Addr:         s.config.Host + ":" + s.config.Port,
		WriteTimeout: s.config.HTTPServerTimeout,
		ReadTimeout:  s.config.HTTPServerTimeout,
		IdleTimeout:  2 * s.config.HTTPServerTimeout,
		Handler:      s.handler,
	}

	// start the server in the background
	go func() {
		s.logger.Info("Starting HTTP Server ", fmt.Sprintf("addr:%s", srv.Addr))
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal(fmt.Sprintf("HTTP server crashed %s", err.Error()))
		}
	}()

	// return the server and routine
	return srv
}

func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	s.JSONResponse(w, r, map[string]string{"status": "OK"}, http.StatusOK)
}
