package cli

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/controller"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/github"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/log"
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
	logE := dc.logE
	log.SetLevel(c.String("log-level"), logE)
	log.SetColor(c.String("log-color"), logE)
	gh := &github.Client{}
	gh.Init(c.Context, os.Getenv("GITHUB_TOKEN"))

	input := &controller.Input{
		PR:      c.Int("pr"),
		Command: "dismiss",
		Dismiss: true,
	}

	if err := setRepo(c.String("repo"), input); err != nil {
		return err
	}

	// TODO Get a pull request number from a commit hash
	if err := getParamFromEnv(input); err != nil {
		return err
	}

	ctrl := &controller.Controller{}
	ctrl.Init(afero.NewOsFs(), gh, dc.stdout, dc.stderr)
	return ctrl.Run(c.Context, logE, input) //nolint:wrapcheck
}
