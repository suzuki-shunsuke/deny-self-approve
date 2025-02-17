package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/controller"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/github"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/log"
	"github.com/suzuki-shunsuke/go-ci-env/v3/cienv"
	"github.com/suzuki-shunsuke/urfave-cli-help-all/helpall"
	"github.com/urfave/cli/v2"
)

type CLI struct {
	LogLevel string `help:"The log level" enum:"debug,info,warn,error" default:"info"`
	LogColor string `help:"The log color" enum:"auto,always,never" default:"auto"`
	Repo     string `help:"The repository full name" short:"r"`
	PR       int    `help:"The pull request number"`
	Version  bool   `help:"Show version" short:"v"`
	Validate struct {
		Dismiss bool `help:"Dimiss the pull request" short:"d"`
	} `cmd:"" help:"Validate if anyone who didn't push commits to the pull request approves it"`
	Dismiss struct{} `cmd:"" help:"Dismiss self-approvals"`
}

type Runner struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	LDFlags *LDFlags
	LogE    *logrus.Entry
}

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

// Run is the entrypoint of the CLI.
// It reads parameters from command line arguments and environment variables.
// It also reads parameters from the CI environment if it's running in a CI environment.
// Environment variables:
// - GITHUB_TOKEN: GitHub Access token
// https://github.com/suzuki-shunsuke/go-ci-env/tree/main/cienv
func (r *Runner) Run(ctx context.Context) error {
	compiledDate, err := time.Parse(time.RFC3339, r.LDFlags.Date)
	if err != nil {
		compiledDate = time.Now()
	}
	app := cli.App{
		Name:     "deny-self-approve",
		Usage:    "Deny self-approvals on GitHub pull requests",
		Version:  r.LDFlags.Version + " (" + r.LDFlags.Commit + ")",
		Compiled: compiledDate,
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
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			(&validateCommand{
				stdout: r.Stdout,
				stderr: r.Stderr,
				logE:   r.LogE,
			}).command(),
			(&dismissCommand{
				stdout: r.Stdout,
				stderr: r.Stderr,
				logE:   r.LogE,
			}).command(),
			(&versionCommand{
				stdout:  r.Stdout,
				version: r.LDFlags.Version,
				commit:  r.LDFlags.Commit,
			}).command(),
			(&completionCommand{
				logE:   r.LogE,
				stdout: r.Stdout,
			}).command(),
			helpall.New(nil),
		},
	}
	return app.RunContext(ctx, os.Args) //nolint:wrapcheck
}

func setRepo(repo string, input *controller.Input) error {
	if repo == "" {
		return nil
	}
	// Read the repository full name from the command line argument --repo
	// Split the repository full name into the owner and the repository name
	o, n, ok := strings.Cut(repo, "/")
	if !ok {
		return fmt.Errorf("repo must be a repository full name like cli/cli: %s", repo)
	}
	input.RepoOwner = o
	input.RepoName = n
	return nil
}

// getParamFromEnv reads parameters from the environment variables and sets them to input.
// - input.RepoOwner
// - input.RepoName
// - input.PR
func getParamFromEnv(input *controller.Input) error {
	if input.RepoOwner != "" && input.PR != 0 {
		return nil
	}
	// Read parameters from the CI environment
	pt := cienv.Get(nil)
	if pt == nil {
		return nil
	}
	if input.RepoOwner == "" {
		input.RepoOwner = pt.RepoOwner()
	}
	if input.RepoName == "" {
		input.RepoName = pt.RepoName()
	}
	if input.PR <= 0 {
		n, err := pt.PRNumber()
		if err != nil {
			return fmt.Errorf("get a pull request number: %w", err)
		}
		input.PR = n
	}

	return nil
}

func commonAction(c *cli.Context, logE *logrus.Entry, cmd string, stdout, stderr io.Writer, dismiss bool) error {
	log.SetLevel(c.String("log-level"), logE)
	log.SetColor(c.String("log-color"), logE)
	gh := &github.Client{}
	gh.Init(c.Context, os.Getenv("GITHUB_TOKEN"))

	input := &controller.Input{
		PR:      c.Int("pr"),
		Command: cmd,
		Dismiss: dismiss,
	}

	if err := setRepo(c.String("repo"), input); err != nil {
		return err
	}

	// TODO Get a pull request number from a commit hash
	if err := getParamFromEnv(input); err != nil {
		return err
	}

	ctrl := &controller.Controller{}
	ctrl.Init(afero.NewOsFs(), gh, stdout, stderr)
	return ctrl.Run(c.Context, logE, input) //nolint:wrapcheck
}
