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
	logE := vc.logE
	log.SetLevel(c.String("log-level"), logE)
	log.SetColor(c.String("log-color"), logE)
	gh := &github.Client{}
	gh.Init(c.Context, os.Getenv("GITHUB_TOKEN"))

	input := &controller.Input{
		PR:                  c.Int("pr"),
		Command:             "validate",
		Dismiss:             c.Bool("dismiss"),
		IgnoreUnknownCommit: c.Bool("ignore-unknown-commit"),
	}

	if err := setRepo(c.String("repo"), input); err != nil {
		return err
	}

	// TODO Get a pull request number from a commit hash
	if err := getParamFromEnv(input); err != nil {
		return err
	}

	ctrl := &controller.Controller{}
	ctrl.Init(afero.NewOsFs(), gh, vc.stdout, vc.stderr)
	return ctrl.Run(c.Context, logE, input) //nolint:wrapcheck
}
