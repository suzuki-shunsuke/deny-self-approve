package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/controller"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/github"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/log"
	"github.com/urfave/cli/v3"
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
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "log level. One of 'debug', 'info', 'warn', 'error'",
				Value: "info",
			},
			&cli.StringFlag{
				Name:  "log-color",
				Usage: "Log color. One of 'auto' (default), 'always', 'never'",
				Value: "auto",
			},
			&cli.StringFlag{
				Name:    "repo",
				Usage:   "The repository full name",
				Aliases: []string{"r"},
			},
			&cli.StringFlag{
				Name:  "pr",
				Usage: "The pull request number",
			},
			&cli.StringFlag{
				Name:  "reliable-apps",
				Usage: "Reliable apps. Comma-separated list of GitHub login names",
			},
			&cli.StringFlag{
				Name:  "unreliable-machine-users",
				Usage: "Unreliable machine users. Comma-separated list of GitHub login names",
			},
		},
	}
}

func (vc *validateCommand) toMap(s string) map[string]struct{} {
	a := strings.Split(s, ",")
	m := make(map[string]struct{}, len(a))
	for _, v := range a {
		if v == "" {
			continue
		}
		m[v] = struct{}{}
	}
	return m
}

func (vc *validateCommand) action(ctx context.Context, c *cli.Command) error {
	if err := log.Set(vc.logE, c.String("log-level"), c.String("log-color")); err != nil {
		return fmt.Errorf("set up a logger: %w", err)
	}
	gh := &github.Client{}
	gh.Init(ctx, os.Getenv("GITHUB_TOKEN"))

	input := &controller.Input{
		PR:                     c.Int("pr"),
		ReliableApps:           vc.toMap(c.String("reliable-apps")),
		UnreliableMachineUsers: vc.toMap(c.String("unreliable-machine-users")),
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
	return ctrl.Run(ctx, vc.logE, input) //nolint:wrapcheck
}
