package cli

import (
	"io"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type dismissCommand struct {
	stdout io.Writer
	stderr io.Writer
	logE   *logrus.Entry
}

func (dc *dismissCommand) command() *cli.Command {
	return &cli.Command{
		Name:   "dismiss",
		Usage:  "Dismiss self-approvals",
		Action: dc.action,
	}
}

func (dc *dismissCommand) action(c *cli.Context) error {
	return commonAction(c, dc.logE, "dismiss", dc.stdout, dc.stderr, true)
}
