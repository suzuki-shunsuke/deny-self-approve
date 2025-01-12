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

func (r *Runner) getParamFromEnv(cli *CLI, input *controller.Input) error {
	if cli.Repo != "" {
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
