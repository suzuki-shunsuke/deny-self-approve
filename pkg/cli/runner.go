package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/deny-self-approve/pkg/controller"
	"github.com/suzuki-shunsuke/go-ci-env/v3/cienv"
	"github.com/suzuki-shunsuke/urfave-cli-help-all/helpall"
	"github.com/urfave/cli/v3"
)

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
		Name:                 "deny-self-approve",
		Usage:                "Deny self-approvals on GitHub pull requests",
		Version:              r.LDFlags.Version + " (" + r.LDFlags.Commit + ")",
		Compiled:             compiledDate,
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			(&validateCommand{
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
