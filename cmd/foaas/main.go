package foaas

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/oklog/run"
	"github.com/pableeee/fo-aas/pkg/api"
)

const (
	hostname = "HOSTNAME"
)

func Run() {
	var (
		logLevel = flag.String("loglevel", "debug", "logging level threshold")
		host     = flag.String("host", "0.0.0.0", "listen interface for http server")
		port     = flag.String("port", "8080", "http server port")
		every    = flag.Int("session-length", 1000, "lenght of the session (ms)")
		tokens   = flag.Int("tokens", 1000, "tokes availabe per session")
		timeout  = flag.Int("timeout", 3000, "connection timeout duration for request to foaas service (ms)")
		group    = run.Group{}
	)

	flag.Parse()

	logger, err := getLogger(*logLevel)
	if err != nil {
		log.Fatalf("unable to get logger")
	}

	srv := api.NewServer(&api.Config{
		Host:                      *host,
		Port:                      *port,
		Hostname:                  getOrDefault(hostname, "localhost"),
		HTTPServerShutdownTimeout: time.Second * 10,
		HTTPServerTimeout:         time.Second * 10,
		Every:                     time.Millisecond * time.Duration(*every),
		Tokens:                    *tokens,
		Timeout:                   time.Millisecond * time.Duration(*timeout),
	}, logger)

	// adds signal handler
	registerActor(&group, setupInterruptHandler)

	// adds http server to run group
	registerActor(&group, func(ctx context.Context) error {
		logger.Info("Staring http server")
		// ListenAndServe must return if/when the ctx is canceled
		return srv.ListenAndServe(ctx)
	})

	logger.Infof("exit: %s", group.Run())
}
