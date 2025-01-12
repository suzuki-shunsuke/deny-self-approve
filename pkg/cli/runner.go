package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/controller"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/github"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/log"
	"github.com/suzuki-shunsuke/go-ci-env/v3/cienv"
)

type CLI struct {
	LogLevel string `help:"The log level" enum:"debug,info,warn,error" default:"info"`
	LogColor string `help:"The log color" enum:"auto,always,never" default:"auto"`
	Repo     string `help:"The repository full name" short:"r"`
	PR       int    `help:"The pull request number"`
	Dismiss  bool   `help:"Dimiss the pull request" short:"d"`
	Version  bool   `help:"Show version" short:"v"`
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
	cli := &CLI{}
	kong.Parse(cli)
	if cli.Version {
		fmt.Fprintln(r.Stdout, r.LDFlags.Version)
		return nil
	}
	log.SetColor(cli.LogColor, r.LogE)
	log.SetLevel(cli.LogLevel, r.LogE)

	gh := &github.Client{}
	gh.Init(ctx, os.Getenv("GITHUB_TOKEN"))
	input := &controller.Input{
		Dismiss: cli.Dismiss,
		PR:      cli.PR,
	}
	// TODO Get a pull request number from a commit hash
	if err := r.getParamFromEnv(cli, input); err != nil {
		return err
	}
	ctrl := &controller.Controller{}
	ctrl.Init(afero.NewOsFs(), gh, r.Stdout, r.Stderr)
	return ctrl.Run(ctx, r.LogE, input)
}

// getParamFromEnv reads parameters from the environment variables and sets them to input.
// - input.RepoOwner
// - input.RepoName
// - input.PR
func (r *Runner) getParamFromEnv(cli *CLI, input *controller.Input) error {
	if cli.Repo != "" {
		// Read the repository full name from the command line argument --repo
		// Split the repository full name into the owner and the repository name
		o, n, ok := strings.Cut(cli.Repo, "/")
		if !ok {
			return fmt.Errorf("repo must be a repository full name like cli/cli: %s", cli.Repo)
		}
		input.RepoOwner = o
		input.RepoName = n
	}
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
			return err
		}
		input.PR = n
	}

	return nil
}
