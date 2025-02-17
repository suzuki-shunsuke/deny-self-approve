package cli

import (
	"io"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type validateCommand struct {
	stdout io.Writer
	stderr io.Writer
	logE   *logrus.Entry
}

func (vc *validateCommand) command() *cli.Command {
	return &cli.Command{
		Name:   "validate",
		Usage:  "Validate if anyone who didn't push commits to the pull request approves it",
		Action: vc.action,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dismiss",
				Usage: "Dimiss self-approvals",
			},
		},
	}
}

func (vc *validateCommand) action(c *cli.Context) error {
	return commonAction(c, vc.logE, "validate", vc.stdout, vc.stderr, c.Bool("dismiss"))
}
