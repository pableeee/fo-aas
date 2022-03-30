package foaas

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/oklog/run"

	"github.com/sirupsen/logrus"
)

func getOrDefault(key, def string) string {
	if v, found := os.LookupEnv(key); found {
		return v
	}

	return def
}

func setupInterruptHandler(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		return fmt.Errorf("received signal %s", sig)
	case <-ctx.Done():
		return nil
	}
}

// registerActor adds the execute function to the run group.
// The execute function must accept a single context parameter,
// and should return when the context is done.
func registerActor(group *run.Group, execute func(context.Context) error) {
	ctx, cancel := context.WithCancel(context.Background())

	group.Add(
		func() error {
			return execute(ctx)
		}, func(e error) {
			// on interrupt, context is canceled to signal termination
			cancel()
		},
	)
}

func getLogger(logLevel string) (*logrus.Logger, error) {
	logger := logrus.New()

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}

	logger.SetLevel(level)
	logger.SetReportCaller(true)

	return logger, nil
}
