package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/cli"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/log"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	logE := log.New("deny-self-approve", version)
	if err := core(logE); err != nil {
		logerr.WithError(logE, err).Fatal("deny-self-approve failed")
	}
}

func core(logE *logrus.Entry) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	runner := cli.Runner{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		LDFlags: &cli.LDFlags{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
		LogE: logE,
	}
	return runner.Run(ctx) //nolint:wrapcheck
}
